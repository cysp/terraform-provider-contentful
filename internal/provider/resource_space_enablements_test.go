package provider_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sync"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"
)

var errSpaceEnablementsRequestMismatch = errors.New("space enablements request mismatch")

//nolint:paralleltest
func TestAccSpaceEnablementsResourceImport(t *testing.T) {
	parallelWhenMocked(t)

	server, _ := cmt.NewContentfulManagementServer()

	configVariables := config.Variables{
		"space_id": config.StringVariable("0p38pssr0fi3"),
	}

	server.SetSpaceEnablements("0p38pssr0fi3", cm.SpaceEnablementData{
		CrossSpaceLinks: cm.NewOptSpaceEnablementField(cm.SpaceEnablementField{Enabled: true}),
		SpaceTemplates:  cm.NewOptSpaceEnablementField(cm.SpaceEnablementField{Enabled: true}),
	})

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ResourceName:    "contentful_space_enablements.test",
				ImportState:     true,
				ImportStateId:   "0p38pssr0fi3",
			},
		},
	})
}

func TestAccSpaceEnablementsResourceAllowsOmittedAndOneSidedCreate(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		values   string
		expected map[string]bool
	}{
		"all enablements omitted": {
			expected: map[string]bool{},
		},
		"only cross space links configured": {
			values: "cross_space_links = true",
			expected: map[string]bool{
				"crossSpaceLinks": true,
			},
		},
		"only space templates configured": {
			values: "space_templates = false",
			expected: map[string]bool{
				"spaceTemplates": false,
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer()
			require.NoError(t, err)
			server.RegisterSpaceEnvironment("space", "master")

			recorder := &spaceEnablementsPutRecorder{next: server}
			ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{
				Steps: []resource.TestStep{{
					Config: spaceEnablementsTestConfig(test.values),
					Check:  recorder.checkRequestFields(test.expected),
				}},
			})
		})
	}
}

func TestAccSpaceEnablementsResourceSendsStatePreservedSiblingOnOneSidedUpdate(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(100))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")

	recorder := &spaceEnablementsPutRecorder{next: server}
	ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{
		Steps: []resource.TestStep{
			{Config: spaceEnablementsTestConfig("cross_space_links = false\n  space_templates = false")},
			{
				PreConfig: func() { recorder.reset() },
				Config:    spaceEnablementsTestConfig("cross_space_links = true"),
				Check: recorder.checkRequestFields(map[string]bool{
					"crossSpaceLinks": true,
					"spaceTemplates":  false,
				}),
			},
		},
	})
}

func TestAccSpaceEnablementsResourceSendsUnequalValues(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(100))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")

	recorder := &spaceEnablementsPutRecorder{next: server}
	ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{
		Steps: []resource.TestStep{{
			Config: spaceEnablementsTestConfig("cross_space_links = true\n  space_templates = false"),
			Check: recorder.checkRequestFields(map[string]bool{
				"crossSpaceLinks": true,
				"spaceTemplates":  false,
			}),
		}},
	})
}

func TestAccSpaceEnablementsResourceRetransmitsImportedKnownValues(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(100))
	require.NoError(t, err)
	server.SetSpaceEnablements("space", cm.SpaceEnablementData{
		CrossSpaceLinks: cm.NewOptSpaceEnablementField(cm.SpaceEnablementField{Enabled: true}),
		SpaceTemplates:  cm.NewOptSpaceEnablementField(cm.SpaceEnablementField{Enabled: true}),
		SuggestConcepts: cm.NewOptSpaceEnablementField(cm.SpaceEnablementField{Enabled: false}),
	})

	recorder := &spaceEnablementsPutRecorder{next: server}
	ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:             spaceEnablementsTestConfig(""),
				ResourceName:       "contentful_space_enablements.test",
				ImportState:        true,
				ImportStateId:      "space",
				ImportStatePersist: true,
			},
			{
				Config: spaceEnablementsTestConfig("suggest_concepts = true"),
				Check: recorder.checkRequestFields(map[string]bool{
					"crossSpaceLinks": true,
					"spaceTemplates":  true,
					"suggestConcepts": true,
				}),
			},
		},
	})
}

func spaceEnablementsTestConfig(values string) string {
	return fmt.Sprintf(`
resource "contentful_space_enablements" "test" {
  space_id = "space"
  %s
}
`, values)
}

type spaceEnablementsPutRecorder struct {
	next   http.Handler
	mu     sync.Mutex
	bodies []map[string]json.RawMessage
}

func (r *spaceEnablementsPutRecorder) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodPut && request.URL.Path == "/spaces/space/enablements" {
		body, err := io.ReadAll(request.Body)
		if err == nil {
			request.Body = io.NopCloser(bytes.NewReader(body))

			var decoded map[string]json.RawMessage
			if json.Unmarshal(body, &decoded) == nil {
				r.mu.Lock()
				r.bodies = append(r.bodies, decoded)
				r.mu.Unlock()
			}
		}
	}

	r.next.ServeHTTP(responseWriter, request)
}

func (r *spaceEnablementsPutRecorder) reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.bodies = nil
}

func (r *spaceEnablementsPutRecorder) checkRequestFields(expected map[string]bool) resource.TestCheckFunc {
	return func(*terraform.State) error {
		r.mu.Lock()
		defer r.mu.Unlock()

		if len(r.bodies) != 1 {
			return fmt.Errorf("%w: recorded %d PUTs, want exactly 1", errSpaceEnablementsRequestMismatch, len(r.bodies))
		}

		if len(r.bodies[0]) != len(expected) {
			return fmt.Errorf("%w: PUT sent %d fields, want %d", errSpaceEnablementsRequestMismatch, len(r.bodies[0]), len(expected))
		}

		for key, expectedEnabled := range expected {
			var field struct {
				Enabled bool `json:"enabled"`
			}

			body, ok := r.bodies[0][key]
			if !ok {
				return fmt.Errorf("%w: PUT omitted %s", errSpaceEnablementsRequestMismatch, key)
			}

			err := json.Unmarshal(body, &field)
			if err != nil {
				return fmt.Errorf("decode %s: %w", key, err)
			}

			if field.Enabled != expectedEnabled {
				return fmt.Errorf("%w: PUT sent %s.enabled=%t, want %t", errSpaceEnablementsRequestMismatch, key, field.Enabled, expectedEnabled)
			}
		}

		return nil
	}
}

//nolint:paralleltest
func TestAccSpaceEnablementsResourceImportNotFound(t *testing.T) {
	parallelWhenMocked(t)

	server, _ := cmt.NewContentfulManagementServer()

	configVariables := config.Variables{
		"space_id": config.StringVariable("0p38pssr0fi3"),
	}

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ResourceName:    "contentful_space_enablements.test",
				ImportState:     true,
				ImportStateId:   "nonexistent",
				ExpectError:     regexp.MustCompile(`Cannot import non-existent remote object`),
			},
		},
	})
}

func TestAccSpaceEnablementsResourceCreateUpdateDelete(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	server.RegisterSpaceEnvironment("0p38pssr0fi3", "master")

	configVariables := config.Variables{
		"space_id": config.StringVariable("0p38pssr0fi3"),
	}

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
			},
		},
	})
}
