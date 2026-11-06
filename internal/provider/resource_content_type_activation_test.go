package provider_test

import (
	"net/http"
	"regexp"
	"strings"
	"testing"

	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

func TestAccContentTypeResourceCreateUsesExactPositiveReturnedVersion(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")
	offset := &contentTypeVersionOffsetAdapter{delegate: server, offset: 3}
	handler := &contentTypeActivationTestHandler{delegate: offset}
	variables := contentTypeActivationConfigVariables("create-positive-version")

	testAccMockedResource(t, handler, resource.TestCase{Steps: []resource.TestStep{
		{
			ConfigDirectory: config.StaticDirectory("testdata/content_type_draft"), ConfigVariables: variables,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("contentful_content_type.test", tfjsonpath.New("published_version"), knownvalue.Int64Exact(4)),
			},
			Check: func(*terraform.State) error {
				require.Equal(t, []contentTypeEvent{
					{Operation: contentTypeOperationPut, Method: http.MethodPut, Path: "/spaces/space/environments/environment/content_types/create-positive-version", Version: -1},
					{Operation: contentTypeOperationActivate, Method: http.MethodPut, Path: "/spaces/space/environments/environment/content_types/create-positive-version/published", Version: 4, VersionPresent: true},
				}, handler.eventHistory())
				raw := handler.rawRequestHistory()
				require.Len(t, raw, 2)
				require.JSONEq(t, testJSON(map[string]any{
					"name":         "Test",
					"description":  "Test content type (create-positive-version)",
					"displayField": "name",
					"fields": []any{
						map[string]any{
							"id":          "name",
							"name":        "Name",
							"type":        "Symbol",
							"localized":   false,
							"required":    true,
							"disabled":    false,
							"omitted":     false,
							"validations": []any{},
						},
						map[string]any{
							"id":          "flags",
							"name":        "Flags",
							"type":        "Array",
							"localized":   false,
							"required":    false,
							"disabled":    false,
							"omitted":     false,
							"validations": []any{},
							"items":       map[string]any{"type": "Symbol", "validations": []any{}},
						},
					},
				}), string(raw[0].Body))
				require.Empty(t, raw[1].Body)
				require.Zero(t, raw[1].ContentLength)
				require.Equal(t, []string{"4"}, raw[1].VersionValues)

				return nil
			},
		},
		{
			PreConfig:       handler.resetRequestHistory,
			ConfigDirectory: config.StaticDirectory("testdata/content_type_draft"), ConfigVariables: variables,
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionNoop),
			}},
			Check: contentTypeActivationRequestCheck(handler, 0, 0),
		},
	}})
}

func TestAccContentTypeResourceUpdateUsesExactArbitraryPositiveReturnedVersion(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")
	offset := &contentTypeVersionOffsetAdapter{delegate: server, jumpOnDraftUpdate: 40}
	handler := &contentTypeActivationTestHandler{delegate: offset}
	variables := contentTypeActivationConfigVariables("update-arbitrary-positive-version")

	testAccMockedResource(t, handler, resource.TestCase{Steps: []resource.TestStep{
		{ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/1"), ConfigVariables: variables},
		{
			PreConfig:       handler.resetRequestHistory,
			ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"), ConfigVariables: variables,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("contentful_content_type.test", tfjsonpath.New("published_version"), knownvalue.Int64Exact(43)),
			},
			Check: func(*terraform.State) error {
				require.Equal(t, []contentTypeEvent{
					{Operation: contentTypeOperationPut, Method: http.MethodPut, Path: "/spaces/space/environments/environment/content_types/update-arbitrary-positive-version", Version: 2, VersionPresent: true},
					{Operation: contentTypeOperationActivate, Method: http.MethodPut, Path: "/spaces/space/environments/environment/content_types/update-arbitrary-positive-version/published", Version: 43, VersionPresent: true},
				}, handler.eventHistory())

				return nil
			},
		},
		{
			PreConfig:       handler.resetRequestHistory,
			ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"), ConfigVariables: variables,
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionNoop),
			}},
			Check: contentTypeActivationRequestCheck(handler, 0, 0),
		},
	}})
}

func TestAccContentTypeResourceHigherPostActivationVersionIsAccepted(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")
	higher := &contentTypeHigherPostActivationVersionAdapter{delegate: server, server: server}
	handler := &contentTypeActivationTestHandler{delegate: higher}
	variables := contentTypeActivationConfigVariables("higher-post-activation-version")

	testAccMockedResource(t, handler, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
		Steps: []resource.TestStep{
			{ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/1"), ConfigVariables: variables},
			{
				PreConfig: func() {
					handler.resetRequestHistory()
					higher.shot.arm()
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"), ConfigVariables: variables,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("contentful_content_type.test", tfjsonpath.New("published_version"), knownvalue.Int64Exact(3)),
				},
				Check: contentTypeActivationRequestAndVersionsCheck(handler, 1, 1, []int64{3}),
			},
			{
				PreConfig:       handler.resetRequestHistory,
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/4"), ConfigVariables: variables,
				Check: func(*terraform.State) error {
					require.Equal(t, []contentTypeEvent{
						{Operation: contentTypeOperationPut, Method: http.MethodPut, Path: "/spaces/space/environments/environment/content_types/higher-post-activation-version", Version: 6, VersionPresent: true},
						{Operation: contentTypeOperationActivate, Method: http.MethodPut, Path: "/spaces/space/environments/environment/content_types/higher-post-activation-version/published", Version: 7, VersionPresent: true},
					}, handler.eventHistory())

					return nil
				},
			},
			{
				PreConfig:       handler.resetRequestHistory,
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/4"), ConfigVariables: variables,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionNoop),
				}},
				Check: contentTypeActivationRequestCheck(handler, 0, 0),
			},
		},
	})
}

func TestAccContentTypeResourceInitialActivationVersionMismatchRevokesAuthority(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")
	handler := &contentTypeActivationTestHandler{delegate: server}
	variables := contentTypeActivationConfigVariables("initial-activation-version-mismatch")

	testAccMockedResource(t, handler, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
		Steps: []resource.TestStep{
			{ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/1"), ConfigVariables: variables},
			{
				PreConfig: func() {
					handler.resetRequestHistory()
					handler.activationVersionMismatch.Store(true)
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"), ConfigVariables: variables,
				ExpectError: regexp.MustCompile(`(?s)VersionMismatch.*revoked activation authority`),
			},
			{
				PreConfig: func() {
					require.Equal(t, []contentTypeEvent{
						{Operation: contentTypeOperationPut, Method: http.MethodPut, Path: "/spaces/space/environments/environment/content_types/initial-activation-version-mismatch", Version: 2, VersionPresent: true},
						{Operation: contentTypeOperationActivate, Method: http.MethodPut, Path: "/spaces/space/environments/environment/content_types/initial-activation-version-mismatch/published", Version: 3, VersionPresent: true},
					}, handler.eventHistory())
					handler.resetRequestHistory()
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"), ConfigVariables: variables,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionNoop),
				}},
				Check: contentTypeActivationRequestCheck(handler, 0, 0),
			},
		},
	})
}

// A successful draft PUT that returns a contradictory body is not safe to
// publish. The response is checkpointed truthfully, but later observation
// cannot authorize its activation merely because the transport succeeded.
func TestAccContentTypeResourceDoesNotActivateContradictoryDraftPutResponse(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		initialDirectory string
		updateDirectory  string
		mutate           func(map[string]any)
		expectError      string
	}{
		"name": {
			"testdata/TestAccContentTypeResourceUpdate/1", "testdata/TestAccContentTypeResourceUpdate/2",
			func(response map[string]any) { response["name"] = "Contradictory" },
			`Unexpected Contentful content type response`,
		},
		"metadata taxonomy": {
			"testdata/TestAccContentTypeResourceUnknownMetadata/1", "testdata/TestAccContentTypeResourceUnknownMetadata/2",
			func(response map[string]any) {
				mustContentTypeResponseObject(response["metadata"])["taxonomy"] = []any{map[string]any{
					"sys":      map[string]any{"type": "Link", "id": "contradictory", "linkType": "TaxonomyConcept"},
					"required": false,
				}}
			},
			`Unexpected Contentful content type response`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
			require.NoError(t, err)
			server.RegisterSpaceEnvironment("space", "environment")
			handler := &contentTypeActivationTestHandler{delegate: server}
			variables := contentTypeActivationConfigVariables("contradictory-put-response-" + strings.ReplaceAll(name, " ", "-"))
			testAccMockedResource(t, handler, resource.TestCase{Steps: []resource.TestStep{
				{ConfigDirectory: config.StaticDirectory(test.initialDirectory), ConfigVariables: variables},
				{
					PreConfig: func() {
						handler.resetRequestHistory()
						handler.mutatePutResponse = contentTypeResponseMutator(t, test.mutate)
					},
					ConfigDirectory: config.StaticDirectory(test.updateDirectory), ConfigVariables: variables,
					ExpectError: regexp.MustCompile(test.expectError),
				},
				{
					PreConfig: func() {
						require.Equal(t, 1, handler.eventCount(contentTypeOperationPut))
						require.Equal(t, 0, handler.eventCount(contentTypeOperationActivate))
						handler.mutatePutResponse = nil
						handler.resetRequestHistory()
					},
					ConfigDirectory: config.StaticDirectory(test.updateDirectory), ConfigVariables: variables,
					Check: contentTypeActivationRequestCheck(handler, 0, 0),
				},
			}})
		})
	}
}

func TestAccContentTypeResourceContradictoryCreateActivationRevokesRecoveryAuthority(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")
	handler := &contentTypeActivationTestHandler{delegate: server}
	handler.mutateActivationResponse = contentTypeResponseMutator(t, removeContentTypePublishedVersion)
	variables := contentTypeActivationConfigVariables("contradictory-create-activation-publication")

	testAccMockedResource(t, handler, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/content_type_draft"), ConfigVariables: variables,
			},
			{
				PreConfig: func() {
					require.Equal(t, []contentTypeEvent{
						{Operation: contentTypeOperationPut, Method: http.MethodPut, Path: "/spaces/space/environments/environment/content_types/contradictory-create-activation-publication", Version: -1},
						{Operation: contentTypeOperationActivate, Method: http.MethodPut, Path: "/spaces/space/environments/environment/content_types/contradictory-create-activation-publication/published", Version: 1, VersionPresent: true},
					}, handler.eventHistory())
					handler.resetRequestHistory()
				},
				ConfigDirectory: config.StaticDirectory("testdata/content_type_draft"), ConfigVariables: variables,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionNoop),
				}},
				Check: contentTypeActivationRequestCheck(handler, 0, 0),
			},
			{
				PreConfig: func() {
					require.Empty(t, handler.eventHistory(), "a response version beyond the marked draft must revoke recovery authority")
					handler.resetRequestHistory()
				},
				ConfigDirectory: config.StaticDirectory("testdata/content_type_draft"), ConfigVariables: variables,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionNoop),
				}},
				Check: contentTypeActivationRequestCheck(handler, 0, 0),
			},
		},
	})
}

func TestAccContentTypeResourceRejectsContradictoryPutUpdateActivationPublication(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")
	handler := &contentTypeActivationTestHandler{delegate: server}
	variables := contentTypeActivationConfigVariables("contradictory-put-update-activation-publication")

	testAccMockedResource(t, handler, resource.TestCase{Steps: []resource.TestStep{
		{ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/1"), ConfigVariables: variables},
		{
			PreConfig: func() {
				handler.resetRequestHistory()
				handler.mutateActivationResponse = contentTypeResponseMutator(t, removeContentTypePublishedVersion)
			},
			ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"), ConfigVariables: variables,
			ExpectError: regexp.MustCompile(`Unexpected Contentful content type activation response`),
		},
		{
			PreConfig: func() {
				require.Equal(t, []contentTypeEvent{
					{Operation: contentTypeOperationPut, Method: http.MethodPut, Path: "/spaces/space/environments/environment/content_types/contradictory-put-update-activation-publication", Version: 2, VersionPresent: true},
					{Operation: contentTypeOperationActivate, Method: http.MethodPut, Path: "/spaces/space/environments/environment/content_types/contradictory-put-update-activation-publication/published", Version: 3, VersionPresent: true},
				}, handler.eventHistory())
				handler.resetRequestHistory()
			},
			ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"), ConfigVariables: variables,
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionNoop),
			}},
			Check: contentTypeActivationRequestCheck(handler, 0, 0),
		},
	}})
}

func TestAccContentTypeResourceNonAdvancingActivationResponseRevokesAuthority(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")
	handler := &contentTypeActivationTestHandler{delegate: server}
	variables := contentTypeActivationConfigVariables("nonadvancing-activation-response")

	testAccMockedResource(t, handler, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
		Steps: []resource.TestStep{
			{ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/1"), ConfigVariables: variables},
			{
				PreConfig: func() {
					handler.resetRequestHistory()
					handler.mutateActivationResponse = contentTypeResponseMutator(t, func(response map[string]any) {
						sys := mustContentTypeResponseObject(response["sys"])
						sys["version"] = 3
						sys["publishedVersion"] = 3
					})
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"), ConfigVariables: variables,
				ExpectError: regexp.MustCompile(`current version must be greater than the published version`),
			},
			{
				PreConfig:       handler.resetRequestHistory,
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"), ConfigVariables: variables,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionNoop),
				}},
				Check: contentTypeActivationRequestCheck(handler, 0, 0),
			},
		},
	})
}
