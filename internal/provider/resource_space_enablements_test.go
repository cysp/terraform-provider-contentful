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
	"sync/atomic"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"
)

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

func TestAccSpaceEnablementsResourceRequiresCoupledValuesBeforeContentful(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		values    string
		wantError string
	}{
		"neither configured on create": {
			wantError: "Missing coupled space enablements",
		},
		"only cross space links configured": {
			values:    "cross_space_links = true",
			wantError: "Missing coupled space enablement",
		},
		"unequal values configured": {
			values:    "cross_space_links = true\n  space_templates = false",
			wantError: "Unequal coupled space enablements",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer()
			require.NoError(t, err)
			var requestCount atomic.Int64
			handler := http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
				if request.Method == http.MethodPut && request.URL.Path == "/spaces/space/enablements" {
					requestCount.Add(1)
				}
				server.ServeHTTP(responseWriter, request)
			})

			ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{
				Steps: []resource.TestStep{{
					Config:      spaceEnablementsTestConfig(test.values),
					ExpectError: regexp.MustCompile(regexp.QuoteMeta(test.wantError)),
				}},
			})

			require.Zero(t, requestCount.Load())
		})
	}
}

func TestAccSpaceEnablementsResourceRejectsInvalidUpdateBeforeContentful(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		values    string
		wantError string
	}{
		"configuration cannot combine with state": {
			values:    "cross_space_links = true",
			wantError: "Missing coupled space enablement",
		},
		"configured values must be equal": {
			values:    "cross_space_links = true\n  space_templates = false",
			wantError: "Unequal coupled space enablements",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer()
			require.NoError(t, err)
			server.RegisterSpaceEnvironment("space", "master")

			var requestCount atomic.Int64
			handler := http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
				if request.Method == http.MethodPut && request.URL.Path == "/spaces/space/enablements" {
					requestCount.Add(1)
				}
				server.ServeHTTP(responseWriter, request)
			})

			ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{
				Steps: []resource.TestStep{
					{Config: spaceEnablementsTestConfig("cross_space_links = false\n  space_templates = false")},
					{
						PreConfig:   func() { requestCount.Store(0) },
						Config:      spaceEnablementsTestConfig(test.values),
						ExpectError: regexp.MustCompile(regexp.QuoteMeta(test.wantError)),
					},
				},
			})

			require.Zero(t, requestCount.Load())
		})
	}
}

func TestAccSpaceEnablementsResourceRetransmitsImportedCoupledValues(t *testing.T) {
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
				Check:  recorder.checkImportedPairRequest(),
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

func (r *spaceEnablementsPutRecorder) checkImportedPairRequest() resource.TestCheckFunc {
	return func(*terraform.State) error {
		r.mu.Lock()
		defer r.mu.Unlock()

		if len(r.bodies) != 1 {
			return fmt.Errorf("recorded %d space enablement PUTs, want exactly 1", len(r.bodies))
		}

		for _, key := range []string{"crossSpaceLinks", "spaceTemplates"} {
			var field struct {
				Enabled bool `json:"enabled"`
			}
			body, ok := r.bodies[0][key]
			if !ok {
				return fmt.Errorf("space enablement PUT omitted %s", key)
			}
			if err := json.Unmarshal(body, &field); err != nil {
				return fmt.Errorf("decode %s: %w", key, err)
			}
			if !field.Enabled {
				return fmt.Errorf("space enablement PUT sent %s.enabled=false, want true", key)
			}
		}

		if body, ok := r.bodies[0]["suggestConcepts"]; !ok || !bytes.Contains(body, []byte(`"enabled":true`)) {
			return errors.New("space enablement PUT did not send configured suggestConcepts=true")
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
