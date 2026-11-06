package provider_test

import (
	"encoding/json"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

func TestAccContentTypeResourceTimeoutOnlyUpdateDoesNotMutate(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	handler := &contentTypeActivationTestHandler{delegate: server}
	configVariables := contentTypeActivationConfigVariables("timeout-only-update")

	testAccMockedResource(t, handler, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/content_type_draft"),
				ConfigVariables: configVariables,
			},
			{
				PreConfig: func() {
					handler.resetRequestHistory()
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceTimeout/1"),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"contentful_content_type.test",
						tfjsonpath.New("published_version"),
						knownvalue.Int64Exact(1),
					),
				},
				Check: contentTypeActivationRequestCheck(handler, 0, 0),
			},
			{
				PreConfig: func() {
					handler.resetRequestHistory()
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceTimeout/2"),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionUpdate),
						plancheck.ExpectUnknownValue("contentful_content_type.test", tfjsonpath.New("published_version")),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"contentful_content_type.test",
						tfjsonpath.New("published_version"),
						knownvalue.Int64Exact(3),
					),
				},
				Check: contentTypeActivationRequestCheck(handler, 1, 1),
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceTimeout/2"),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccContentTypeResourceUnknownMetadataFinalPlanHandlesIndependentActivationDrift(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	handler := &contentTypeActivationTestHandler{delegate: server}
	configVariables := contentTypeActivationConfigVariables("unknown-metadata")

	testAccMockedResource(t, handler, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUnknownMetadata/1"),
				ConfigVariables: configVariables,
			},
			{
				PreConfig: func() {
					response, deactivateErr := server.Handler().DeactivateContentType(t.Context(), cm.DeactivateContentTypeParams{
						SpaceID: "space", EnvironmentID: "environment", ContentTypeID: "unknown-metadata",
					})
					require.NoError(t, deactivateErr)
					require.IsType(t, &cm.ContentType{}, response)
					handler.resetRequestHistory()
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUnknownMetadata/2"),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("terraform_data.metadata", plancheck.ResourceActionCreate),
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionUpdate),
						plancheck.ExpectUnknownValue("contentful_content_type.test", tfjsonpath.New("published_version")),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"contentful_content_type.test",
						tfjsonpath.New("published_version"),
						knownvalue.Int64Exact(4),
					),
				},
				Check: contentTypeActivationRequestAndVersionsCheck(handler, 1, 1, []int64{4}),
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUnknownMetadata/2"),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccContentTypeResourceManagedDraftTimeoutUpdateMutatesOnce(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	handler := &contentTypeActivationTestHandler{delegate: server}
	configVariables := contentTypeActivationConfigVariables("managed-draft-timeout")

	testAccMockedResource(t, handler, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/content_type_draft"),
				ConfigVariables: configVariables,
			},
			{
				PreConfig: func() {
					handler.resetRequestHistory()
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceTimeout/2"),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionUpdate),
						plancheck.ExpectUnknownValue("contentful_content_type.test", tfjsonpath.New("published_version")),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"contentful_content_type.test",
						tfjsonpath.New("published_version"),
						knownvalue.Int64Exact(3),
					),
				},
				Check: contentTypeActivationRequestCheck(handler, 1, 1),
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceTimeout/2"),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccContentTypeResourceTimeoutChangeLeavesExternalDeactivationObservational(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	handler := &contentTypeActivationTestHandler{delegate: server}
	configVariables := contentTypeActivationConfigVariables("deactivated-content-type")

	testAccMockedResource(t, handler, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/content_type_draft"),
				ConfigVariables: configVariables,
			},
			{
				PreConfig: func() {
					response, deactivateErr := server.Handler().DeactivateContentType(t.Context(), cm.DeactivateContentTypeParams{
						SpaceID:       "space",
						EnvironmentID: "environment",
						ContentTypeID: "deactivated-content-type",
					})
					require.NoError(t, deactivateErr)
					require.IsType(t, &cm.ContentType{}, response)

					handler.resetRequestHistory()
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceTimeout/1"),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue("contentful_content_type.test", tfjsonpath.New("published_version"), knownvalue.Null()),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("contentful_content_type.test", tfjsonpath.New("published_version"), knownvalue.Null()),
				},
				Check: contentTypeActivationRequestCheck(handler, 0, 0),
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceTimeout/1"),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccContentTypeResourceReconcilesExternalDraftBeforeActivation(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	handler := &contentTypeActivationTestHandler{delegate: server}
	configVariables := contentTypeActivationConfigVariables("external-draft")

	testAccMockedResource(t, handler, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/content_type_draft"),
				ConfigVariables: configVariables,
			},
			{
				PreConfig: func() {
					response, putErr := server.Handler().PutContentType(t.Context(), &cm.ContentTypeRequestData{
						Name:         "Externally changed",
						Description:  cm.NewOptNilString("External draft"),
						DisplayField: "name",
						Fields: []cm.ContentTypeRequestDataFieldsItem{
							{
								ID:        "name",
								Name:      "Name",
								Type:      "Symbol",
								Required:  cm.NewOptBool(true),
								Localized: cm.NewOptBool(false),
							},
							{
								ID:   "flags",
								Name: "Flags",
								Type: "Array",
								Items: cm.NewOptContentTypeRequestDataFieldsItemItems(
									cm.ContentTypeRequestDataFieldsItemItems{Type: cm.NewOptString("Symbol")},
								),
								Required:  cm.NewOptBool(false),
								Localized: cm.NewOptBool(false),
							},
						},
					}, cm.PutContentTypeParams{
						SpaceID:            "space",
						EnvironmentID:      "environment",
						ContentTypeID:      "external-draft",
						XContentfulVersion: cm.NewOptInt(2),
					})
					require.NoError(t, putErr)
					require.IsType(t, &cm.ContentTypeStatusCode{}, response)

					handler.resetRequestHistory()
				},
				ConfigDirectory: config.StaticDirectory("testdata/content_type_draft"),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionUpdate),
						plancheck.ExpectUnknownValue("contentful_content_type.test", tfjsonpath.New("published_version")),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"contentful_content_type.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("Test"),
					),
					statecheck.ExpectKnownValue(
						"contentful_content_type.test",
						tfjsonpath.New("published_version"),
						knownvalue.Int64Exact(4),
					),
				},
				Check: contentTypeActivationRequestAndVersionsCheck(handler, 1, 1, []int64{4}),
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/content_type_draft"),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccContentTypeResourceLeavesExternalPendingDraftObservational(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")
	handler := &contentTypeActivationTestHandler{delegate: server}
	variables := contentTypeActivationConfigVariables("external-pending-no-drift")
	testAccMockedResource(t, handler, resource.TestCase{Steps: []resource.TestStep{
		{ConfigDirectory: config.StaticDirectory("testdata/content_type_draft"), ConfigVariables: variables},
		{
			PreConfig: func() {
				getResponse, getErr := server.Handler().GetContentType(t.Context(), cm.GetContentTypeParams{
					SpaceID:       "space",
					EnvironmentID: "environment",
					ContentTypeID: "external-pending-no-drift",
				})
				require.NoError(t, getErr)

				contentType, ok := getResponse.(*cm.ContentType)
				require.True(t, ok)

				// Re-encode the exact current modeled representation as an external
				// PUT. The only meaningful change is the new draft version.
				wireResponse, marshalErr := json.Marshal(contentType)
				require.NoError(t, marshalErr)

				var identicalDraft cm.ContentTypeRequestData
				require.NoError(t, json.Unmarshal(wireResponse, &identicalDraft))

				response, putErr := server.Handler().PutContentType(t.Context(), &identicalDraft, cm.PutContentTypeParams{
					SpaceID:            "space",
					EnvironmentID:      "environment",
					ContentTypeID:      "external-pending-no-drift",
					XContentfulVersion: cm.NewOptInt(2),
				})
				require.NoError(t, putErr)
				require.IsType(t, &cm.ContentTypeStatusCode{}, response)
				handler.resetRequestHistory()
			},
			ConfigDirectory: config.StaticDirectory("testdata/content_type_draft"), ConfigVariables: variables,
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionNoop),
				plancheck.ExpectKnownValue("contentful_content_type.test", tfjsonpath.New("published_version"), knownvalue.Int64Exact(1)),
			}},
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("contentful_content_type.test", tfjsonpath.New("fields"), knownvalue.ListSizeExact(2)),
				statecheck.ExpectKnownValue("contentful_content_type.test", tfjsonpath.New("published_version"), knownvalue.Int64Exact(1)),
			},
			Check: contentTypeActivationRequestCheck(handler, 0, 0),
		},
		{
			ConfigDirectory: config.StaticDirectory("testdata/content_type_draft"),
			ConfigVariables: variables,
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionNoop),
			}},
		},
	}})
}

func TestAccContentTypeResourceImportUnpublishedDoesNotActivate(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	contentTypeID := "import-unpublished"
	draft := seedContentTypeDraft(t, server, contentTypeID)
	require.Equal(t, 1, draft.Sys.Version)
	require.False(t, draft.Sys.PublishedVersion.IsSet())

	handler := &contentTypeActivationTestHandler{delegate: server}
	variables := contentTypeActivationConfigVariables(contentTypeID)
	additionalCLIOptions := &resource.AdditionalCLIOptions{}

	testAccMockedResource(t, handler, resource.TestCase{AdditionalCLIOptions: additionalCLIOptions, Steps: []resource.TestStep{
		{
			ConfigDirectory:    config.StaticDirectory("testdata/content_type_draft"),
			ConfigVariables:    variables,
			ResourceName:       "contentful_content_type.test",
			ImportState:        true,
			ImportStateId:      "space/environment/" + contentTypeID,
			ImportStatePersist: true,
		},
		{
			PreConfig: func() {
				require.Equal(t, 0, handler.eventCount(contentTypeOperationPut))
				require.Equal(t, 0, handler.eventCount(contentTypeOperationActivate))
				handler.resetRequestHistory()
			},
			ConfigDirectory: config.StaticDirectory("testdata/content_type_draft"),
			ConfigVariables: variables,
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionNoop),
			}},
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("contentful_content_type.test", tfjsonpath.New("published_version"), knownvalue.Null()),
			},
			Check: contentTypeActivationRequestCheck(handler, 0, 0),
		},
		{
			PreConfig: func() {
				additionalCLIOptions.Plan.NoRefresh = true

				handler.resetRequestHistory()
			},
			ConfigDirectory: config.StaticDirectory("testdata/content_type_draft"),
			ConfigVariables: variables,
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionNoop),
			}},
			Check: contentTypeActivationRequestCheck(handler, 0, 0),
		},
	}})
}
