package provider_test

import (
	"encoding/json"
	"net/http"
	"regexp"
	"sync"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type environmentStatusReadyTestHandler struct {
	mu sync.Mutex

	spaceID       string
	environmentID string

	statusesToServe []string
	statusesServed  []string
}

func (h *environmentStatusReadyTestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	environmentPath := "/spaces/" + h.spaceID + "/environments/" + h.environmentID

	if r.Method == http.MethodGet && r.URL.Path == environmentPath {
		h.mu.Lock()
		defer h.mu.Unlock()

		status := h.statusesToServe[len(h.statusesToServe)-1]
		if len(h.statusesServed) < len(h.statusesToServe) {
			status = h.statusesToServe[len(h.statusesServed)]
		}

		h.statusesServed = append(h.statusesServed, status)

		environment := cm.Environment{
			Sys:  cm.NewEnvironmentSys(h.spaceID, h.environmentID, status),
			Name: h.environmentID,
		}

		w.Header().Set("Content-Type", "application/vnd.contentful.management.v1+json")
		w.WriteHeader(http.StatusOK)

		err := json.NewEncoder(w).Encode(environment)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	http.NotFound(w, r)
}

func (h *environmentStatusReadyTestHandler) StatusesServed() []string {
	h.mu.Lock()
	defer h.mu.Unlock()

	return append([]string(nil), h.statusesServed...)
}

func TestAccEnvironmentStatusReadyDataSourceRead(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)

	server.SetEnvironment("0p38pssr0fi3", "master", "ready", cm.EnvironmentData{
		Name: "master",
	})

	configVariables := config.Variables{
		"space_id":       config.StringVariable("0p38pssr0fi3"),
		"environment_id": config.StringVariable("master"),
	}

	testAccMockedResource(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.contentful_environment_status_ready.test", tfjsonpath.New("id"), knownvalue.StringExact("0p38pssr0fi3/master")),
					statecheck.ExpectKnownValue("data.contentful_environment_status_ready.test", tfjsonpath.New("space_id"), knownvalue.StringExact("0p38pssr0fi3")),
					statecheck.ExpectKnownValue("data.contentful_environment_status_ready.test", tfjsonpath.New("environment_id"), knownvalue.StringExact("master")),
					statecheck.ExpectKnownValue("data.contentful_environment_status_ready.test", tfjsonpath.New("status"), knownvalue.StringExact("ready")),
				},
			},
		},
	})
}

func TestAccEnvironmentStatusReadyDataSourcePolling(t *testing.T) {
	t.Parallel()

	server := &environmentStatusReadyTestHandler{
		spaceID:         "space-id",
		environmentID:   "environment-id",
		statusesToServe: []string{"queued", "ready"},
	}

	configVariables := config.Variables{
		"space_id":       config.StringVariable("space-id"),
		"environment_id": config.StringVariable("environment-id"),
	}

	testAccMockedResource(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.contentful_environment_status_ready.test", tfjsonpath.New("id"), knownvalue.StringExact("space-id/environment-id")),
					statecheck.ExpectKnownValue("data.contentful_environment_status_ready.test", tfjsonpath.New("space_id"), knownvalue.StringExact("space-id")),
					statecheck.ExpectKnownValue("data.contentful_environment_status_ready.test", tfjsonpath.New("environment_id"), knownvalue.StringExact("environment-id")),
					statecheck.ExpectKnownValue("data.contentful_environment_status_ready.test", tfjsonpath.New("status"), knownvalue.StringExact("ready")),
				},
			},
		},
	})

	statuses := server.StatusesServed()
	require.GreaterOrEqual(t, len(statuses), 2)
	assert.Equal(t, []string{"queued", "ready"}, statuses[:2])

	for _, status := range statuses[2:] {
		assert.Equal(t, "ready", status)
	}
}

func TestAccEnvironmentStatusReadyDataSourceFailedStopsImmediately(t *testing.T) {
	t.Parallel()

	server := &environmentStatusReadyTestHandler{
		spaceID:         "space-id",
		environmentID:   "environment-id",
		statusesToServe: []string{"failed"},
	}

	testAccMockedResource(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: `
data "contentful_environment_status_ready" "test" {
  space_id       = "space-id"
  environment_id = "environment-id"

  timeouts = {
    read = "100ms"
  }
}
`,
				ExpectError: regexp.MustCompile(`Contentful environment failed to become ready`),
			},
		},
	})

	assert.Equal(t, []string{"failed"}, server.StatusesServed())
}

func TestAccEnvironmentStatusReadyDataSourceInProgressTimesOut(t *testing.T) {
	t.Parallel()

	server := &environmentStatusReadyTestHandler{
		spaceID:         "space-id",
		environmentID:   "environment-id",
		statusesToServe: []string{"inProgress"},
	}

	testAccMockedResource(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: `
data "contentful_environment_status_ready" "test" {
  space_id       = "space-id"
  environment_id = "environment-id"

  timeouts = {
    read = "100ms"
  }
}
`,
				ExpectError: regexp.MustCompile(`Timed out waiting for environment to become ready`),
			},
		},
	})

	assert.Equal(t, []string{"inProgress"}, server.StatusesServed())
}

func TestAccEnvironmentStatusReadyDataSourceNotFound(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)

	configVariables := config.Variables{
		"space_id":       config.StringVariable("0p38pssr0fi3"),
		"environment_id": config.StringVariable("nonexistent"),
	}

	testAccMockableResource(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ExpectError:     regexp.MustCompile(`Failed to read environment`),
			},
		},
	})
}
