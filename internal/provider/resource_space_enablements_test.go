package provider_test

import (
	"bytes"
	"context"
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
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"
)

var (
	errSpaceEnablementsRequestMismatch = errors.New("space enablements request mismatch")
	errSpaceEnablementsStoredMismatch  = errors.New("stored space enablements mismatch")
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

func TestAccSpaceEnablementsResourceRejectsInvalidCreateRequests(t *testing.T) {
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
		"unequal paired members configured": {
			values: "cross_space_links = true\n  space_templates = false",
			expected: map[string]bool{
				"crossSpaceLinks": true,
				"spaceTemplates":  false,
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(100))
			require.NoError(t, err)
			server.RegisterSpaceEnvironment("space", "master")

			recorder := &spaceEnablementsPutRecorder{next: server}
			ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{
				Steps: []resource.TestStep{{
					PreConfig:   func() { recorder.expectNextRequestFields(test.expected) },
					Config:      spaceEnablementsTestConfig(test.values),
					ExpectError: regexp.MustCompile(`Failed to create space enablements(?s:.*)Error: ValidationFailed`),
				}},
			})
		})
	}
}

func TestAccSpaceEnablementsResourceRejectsOneSidedUpdateWithoutMutationAndConverges(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(100))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")

	recorder := &spaceEnablementsPutRecorder{next: server}
	ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					recorder.expectNextRequestFields(map[string]bool{
						"crossSpaceLinks": false,
						"spaceTemplates":  false,
					})
				},
				Config: spaceEnablementsTestConfig("cross_space_links = false\n  space_templates = false"),
				Check: requireStoredSpaceEnablements(server.Handler(), 2, map[string]bool{
					"crossSpaceLinks": false,
					"spaceTemplates":  false,
				}),
			},
			{
				PreConfig: func() {
					recorder.expectNextRequestFields(map[string]bool{
						"crossSpaceLinks": true,
						"spaceTemplates":  false,
					})
				},
				Config:      spaceEnablementsTestConfig("cross_space_links = true"),
				ExpectError: regexp.MustCompile(`Failed to update space enablements(?s:.*)Error: ValidationFailed`),
			},
			{
				PreConfig: func() {
					require.NoError(t, requireStoredSpaceEnablementsFields(server.Handler(), 2, map[string]bool{
						"crossSpaceLinks": false,
						"spaceTemplates":  false,
					}))
					recorder.expectNextRequestFields(map[string]bool{
						"crossSpaceLinks": true,
						"spaceTemplates":  true,
					})
				},
				Config: spaceEnablementsTestConfig("cross_space_links = true\n  space_templates = true"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: requireStoredSpaceEnablements(server.Handler(), 3, map[string]bool{
					"crossSpaceLinks": true,
					"spaceTemplates":  true,
				}),
			},
		},
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
				PreConfig: func() {
					recorder.expectNextRequestFields(map[string]bool{
						"crossSpaceLinks": true,
						"spaceTemplates":  true,
						"suggestConcepts": true,
					})
				},
				Config: spaceEnablementsTestConfig("suggest_concepts = true"),
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
	next http.Handler
	mu   sync.Mutex

	expectedFields  map[string]bool
	requestObserved bool
	validationError error
}

func (r *spaceEnablementsPutRecorder) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodPut && request.URL.Path == "/spaces/space/enablements" {
		body, err := io.ReadAll(request.Body)
		if err == nil {
			request.Body = io.NopCloser(bytes.NewReader(body))

			var decoded map[string]json.RawMessage
			if json.Unmarshal(body, &decoded) == nil {
				r.mu.Lock()
				r.validateRequestLocked(decoded)
				r.mu.Unlock()
			}
		}
	}

	r.next.ServeHTTP(responseWriter, request)
}

func (r *spaceEnablementsPutRecorder) expectNextRequestFields(expected map[string]bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.validateExpectedRequestLocked()

	if r.validationError != nil {
		return
	}

	r.expectedFields = expected
	r.requestObserved = false
}

func (r *spaceEnablementsPutRecorder) handlerError() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.validateExpectedRequestLocked()

	return r.validationError
}

func (r *spaceEnablementsPutRecorder) validateRequestLocked(actual map[string]json.RawMessage) {
	if r.validationError != nil {
		return
	}

	if r.expectedFields == nil || r.requestObserved {
		r.validationError = fmt.Errorf("%w: recorded more PUTs than expected", errSpaceEnablementsRequestMismatch)

		return
	}

	r.requestObserved = true
	r.validationError = compareSpaceEnablementsRequestFields(actual, r.expectedFields)
}

func (r *spaceEnablementsPutRecorder) validateExpectedRequestLocked() {
	if r.expectedFields == nil || r.requestObserved || r.validationError != nil {
		return
	}

	r.validationError = fmt.Errorf("%w: expected PUT was not recorded", errSpaceEnablementsRequestMismatch)
}

func compareSpaceEnablementsRequestFields(actual map[string]json.RawMessage, expected map[string]bool) error {
	if len(actual) != len(expected) {
		return fmt.Errorf("%w: PUT sent %d fields, want %d", errSpaceEnablementsRequestMismatch, len(actual), len(expected))
	}

	for key, expectedEnabled := range expected {
		body, ok := actual[key]
		if !ok {
			return fmt.Errorf("%w: PUT omitted %s", errSpaceEnablementsRequestMismatch, key)
		}

		var field map[string]json.RawMessage

		err := json.Unmarshal(body, &field)
		if err != nil {
			return fmt.Errorf("decode %s: %w", key, err)
		}

		if len(field) != 1 {
			return fmt.Errorf("%w: PUT sent %s with %d members, want exactly enabled", errSpaceEnablementsRequestMismatch, key, len(field))
		}

		enabledBody, ok := field["enabled"]
		if !ok {
			return fmt.Errorf("%w: PUT omitted %s.enabled", errSpaceEnablementsRequestMismatch, key)
		}

		var actualEnabled bool

		err = json.Unmarshal(enabledBody, &actualEnabled)
		if err != nil {
			return fmt.Errorf("decode %s.enabled: %w", key, err)
		}

		if actualEnabled != expectedEnabled {
			return fmt.Errorf("%w: PUT sent %s.enabled=%t, want %t", errSpaceEnablementsRequestMismatch, key, actualEnabled, expectedEnabled)
		}
	}

	return nil
}

func requireStoredSpaceEnablements(
	handler *cmt.Handler,
	expectedVersion int,
	expectedFields map[string]bool,
) resource.TestCheckFunc {
	return func(*terraform.State) error {
		return requireStoredSpaceEnablementsFields(handler, expectedVersion, expectedFields)
	}
}

func requireStoredSpaceEnablementsFields(
	handler *cmt.Handler,
	expectedVersion int,
	expectedFields map[string]bool,
) error {
	response, err := handler.GetSpaceEnablements(context.Background(), cm.GetSpaceEnablementsParams{SpaceID: "space"})
	if err != nil {
		return fmt.Errorf("get stored Space Enablements: %w", err)
	}

	enablements, ok := response.(*cm.SpaceEnablement)
	if !ok {
		return fmt.Errorf(
			"%w: response has type %T, want *contentfulmanagement.SpaceEnablement",
			errSpaceEnablementsStoredMismatch,
			response,
		)
	}

	if enablements.Sys.Version != expectedVersion {
		return fmt.Errorf(
			"%w: version is %d, want %d",
			errSpaceEnablementsStoredMismatch,
			enablements.Sys.Version,
			expectedVersion,
		)
	}

	actualFields := map[string]cm.OptSpaceEnablementField{
		"crossSpaceLinks":   enablements.CrossSpaceLinks,
		"spaceTemplates":    enablements.SpaceTemplates,
		"studioExperiences": enablements.StudioExperiences,
		"suggestConcepts":   enablements.SuggestConcepts,
	}

	for key, actual := range actualFields {
		expectedEnabled, expected := expectedFields[key]

		field, present := actual.Get()
		if present != expected {
			return fmt.Errorf(
				"%w: field %s presence is %t, want %t",
				errSpaceEnablementsStoredMismatch,
				key,
				present,
				expected,
			)
		}

		if present && field.Enabled != expectedEnabled {
			return fmt.Errorf(
				"%w: field %s is %t, want %t",
				errSpaceEnablementsStoredMismatch,
				key,
				field.Enabled,
				expectedEnabled,
			)
		}
	}

	return nil
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
