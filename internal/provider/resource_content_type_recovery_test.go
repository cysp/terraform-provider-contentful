package provider_test

import (
	"net/http"
	"regexp"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
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

func TestAccContentTypeResourceFailedCreateActivationRecoversExactDraft(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	handler := &contentTypeActivationTestHandler{delegate: server}
	handler.failActivation.Store(true)

	configVariables := contentTypeActivationConfigVariables("create-activation-failure")

	testAccMockedResource(t, handler, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
		Steps: []resource.TestStep{
			{
				ConfigDirectory:    config.StaticDirectory("testdata/content_type_draft"),
				ConfigVariables:    configVariables,
				ExpectNonEmptyPlan: true,
			},
			{
				PreConfig: func() {
					requireNoConfirmingGetAfterActivationFailure(t, handler)
					require.Equal(t, []contentTypeEvent{
						{
							Operation: contentTypeOperationPut,
							Method:    http.MethodPut,
							Path:      "/spaces/space/environments/environment/content_types/create-activation-failure",
							Version:   -1,
						},
						{
							Operation:      contentTypeOperationActivate,
							Method:         http.MethodPut,
							Path:           "/spaces/space/environments/environment/content_types/create-activation-failure/published",
							Version:        1,
							VersionPresent: true,
						},
					}, handler.eventHistory())
					require.JSONEq(t, testJSON(map[string]any{
						"name":         "Test",
						"description":  "Test content type (create-activation-failure)",
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
					}), string(handler.lastPutBody()))
					handler.failActivation.Store(false)
					handler.resetRequestHistory()
				},
				ConfigDirectory: config.StaticDirectory("testdata/content_type_draft"),
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
				Check: contentTypeActivationRequestAndVersionsCheck(handler, 0, 1, []int64{1}),
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

func TestAccContentTypeResourceAmbiguousCreateActivationReconcilesByRead(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	handler := &contentTypeActivationTestHandler{delegate: server}
	handler.activationFailureAfterSuccessStatus.Store(http.StatusUnprocessableEntity)

	configVariables := contentTypeActivationConfigVariables("ambiguous-create-activation")

	testAccMockedResource(t, handler, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
		Steps: []resource.TestStep{
			{
				ConfigDirectory:    config.StaticDirectory("testdata/content_type_draft"),
				ConfigVariables:    configVariables,
				ExpectNonEmptyPlan: true,
			},
			{
				PreConfig: func() {
					require.Equal(t, 1, handler.eventCount(contentTypeOperationPut))
					require.Equal(t, 1, handler.eventCount(contentTypeOperationActivate))
					require.Equal(t, []int64{1}, handler.activationVersionHistory())

					handler.resetRequestHistory()
				},
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("contentful_content_type.test", "published_version", "1"),
					contentTypeActivationRequestAndVersionsCheck(handler, 0, 0, nil),
				),
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

func TestAccContentTypeResourceFailedUpdateActivationRecoversExactDraftWithoutRefresh(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	handler := &contentTypeActivationTestHandler{delegate: server}
	configVariables := contentTypeActivationConfigVariables("update-activation-failure")

	testAccMockedResource(t, handler, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{
			Plan: resource.PlanOptions{NoRefresh: true},
		},
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/1"),
				ConfigVariables: configVariables,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"contentful_content_type.test",
						tfjsonpath.New("published_version"),
						knownvalue.Int64Exact(1),
					),
				},
			},
			{
				PreConfig: func() {
					handler.resetRequestHistory()
					handler.failActivation.Store(true)
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"),
				ConfigVariables: configVariables,
				ExpectError:     regexp.MustCompile(`Failed to activate content type`),
			},
			{
				PreConfig: func() {
					require.Equal(t, []contentTypeEvent{
						{
							Operation:      contentTypeOperationPut,
							Method:         http.MethodPut,
							Path:           "/spaces/space/environments/environment/content_types/update-activation-failure",
							Version:        2,
							VersionPresent: true,
						},
						{
							Operation:      contentTypeOperationActivate,
							Method:         http.MethodPut,
							Path:           "/spaces/space/environments/environment/content_types/update-activation-failure/published",
							Version:        3,
							VersionPresent: true,
						},
					}, handler.eventHistory())
					require.JSONEq(t, testJSON(map[string]any{
						"name":         "Test",
						"description":  "Test content type (update-activation-failure)",
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
								"id":          "slug",
								"name":        "Slug",
								"type":        "Symbol",
								"localized":   false,
								"required":    true,
								"disabled":    false,
								"omitted":     false,
								"validations": []any{},
							},
						},
					}), string(handler.lastPutBody()))
					handler.failActivation.Store(false)
					handler.resetRequestHistory()
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"),
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
				Check: resource.ComposeTestCheckFunc(
					contentTypeActivationRequestAndVersionsCheck(handler, 0, 1, []int64{3}),
					func(*terraform.State) error {
						require.Equal(t, []contentTypeEvent{{
							Operation:      contentTypeOperationActivate,
							Method:         http.MethodPut,
							Path:           "/spaces/space/environments/environment/content_types/update-activation-failure/published",
							Version:        3,
							VersionPresent: true,
						}}, handler.eventHistory())

						return nil
					},
				),
			},
		},
	})
}

func TestAccContentTypeResourceFailedUpdateActivationRecoversExactDraftAfterRefresh(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	handler := &contentTypeActivationTestHandler{delegate: server}
	variables := contentTypeActivationConfigVariables("update-activation-refresh-recovery")

	testAccMockedResource(t, handler, resource.TestCase{Steps: []resource.TestStep{
		{ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/1"), ConfigVariables: variables},
		{
			PreConfig: func() {
				handler.resetRequestHistory()
				handler.failActivation.Store(true)
			},
			ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"), ConfigVariables: variables,
			ExpectError: regexp.MustCompile(`Failed to activate content type`),
		},
		{
			PreConfig: func() {
				require.Equal(t, 1, handler.eventCount(contentTypeOperationPut))
				require.Equal(t, []int64{3}, handler.activationVersionHistory())
				handler.failActivation.Store(false)
				handler.resetRequestHistory()
			},
			ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"), ConfigVariables: variables,
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionUpdate),
				plancheck.ExpectUnknownValue("contentful_content_type.test", tfjsonpath.New("published_version")),
			}},
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("contentful_content_type.test", tfjsonpath.New("published_version"), knownvalue.Int64Exact(3)),
			},
			Check: contentTypeActivationRequestAndVersionsCheck(handler, 0, 1, []int64{3}),
		},
		{
			ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"), ConfigVariables: variables,
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionNoop),
			}},
		},
	}})
}

func TestAccContentTypeResourceExternalAdvanceRevokesPendingActivationAfterRefresh(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	handler := &contentTypeActivationTestHandler{delegate: server}
	variables := contentTypeActivationConfigVariables("activation-external-advance-refresh")

	testAccMockedResource(t, handler, resource.TestCase{Steps: []resource.TestStep{
		{ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/1"), ConfigVariables: variables},
		{
			PreConfig: func() {
				handler.resetRequestHistory()
				handler.failActivation.Store(true)
			},
			ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"), ConfigVariables: variables,
			ExpectError: regexp.MustCompile(`Failed to activate content type`),
		},
		{
			PreConfig: func() {
				handler.failActivation.Store(false)
				putIdenticalExternalContentTypeDraft(t, server, "activation-external-advance-refresh", 3)
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

func TestAccContentTypeResourcePublicationTupleChangeAtMarkedVersionRevokesAuthority(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")
	handler := &contentTypeActivationTestHandler{delegate: server}
	variables := contentTypeActivationConfigVariables("activation-tuple-change")
	cliOptions := &resource.AdditionalCLIOptions{}

	testAccMockedResource(t, handler, resource.TestCase{
		AdditionalCLIOptions: cliOptions,
		Steps: []resource.TestStep{
			{ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/1"), ConfigVariables: variables},
			{
				PreConfig: func() {
					handler.resetRequestHistory()
					handler.failActivation.Store(true)
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"), ConfigVariables: variables,
				ExpectError: regexp.MustCompile(`Failed to activate content type`),
			},
			{
				PreConfig: func() {
					handler.failActivation.Store(false)
					handler.mutateReadResponse = contentTypeResponseMutator(t, removeContentTypePublishedVersion)
					handler.resetRequestHistory()
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"), ConfigVariables: variables,
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
					cliOptions.Plan.NoRefresh = true

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

func TestAccContentTypeResourceExternalAdvanceRevokesPendingActivationWithoutRefresh(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	handler := &contentTypeActivationTestHandler{delegate: server}
	variables := contentTypeActivationConfigVariables("activation-external-advance-no-refresh")

	testAccMockedResource(t, handler, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
		Steps: []resource.TestStep{
			{ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/1"), ConfigVariables: variables},
			{
				PreConfig: func() {
					handler.resetRequestHistory()
					handler.failActivation.Store(true)
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"), ConfigVariables: variables,
				ExpectError: regexp.MustCompile(`Failed to activate content type`),
			},
			{
				PreConfig: func() {
					handler.failActivation.Store(false)
					putIdenticalExternalContentTypeDraft(t, server, "activation-external-advance-no-refresh", 3)
					handler.resetRequestHistory()
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"), ConfigVariables: variables,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionUpdate),
				}},
				ExpectError: regexp.MustCompile(`VersionMismatch`),
			},
			{
				PreConfig: func() {
					require.Equal(t, 0, handler.eventCount(contentTypeOperationPut))
					require.Equal(t, []int64{3}, handler.activationVersionHistory())
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

func TestAccContentTypeResourceExternalActivationOfMarkedDraftClearsRecovery(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	handler := &contentTypeActivationTestHandler{delegate: server}
	variables := contentTypeActivationConfigVariables("external-marked-activation")
	additionalCLIOptions := &resource.AdditionalCLIOptions{}

	testAccMockedResource(t, handler, resource.TestCase{
		AdditionalCLIOptions: additionalCLIOptions,
		Steps: []resource.TestStep{
			{ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/1"), ConfigVariables: variables},
			{
				PreConfig: func() {
					handler.resetRequestHistory()
					handler.failActivation.Store(true)
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"), ConfigVariables: variables,
				ExpectError: regexp.MustCompile(`Failed to activate content type`),
			},
			{
				PreConfig: func() {
					handler.failActivation.Store(false)

					response, activateErr := server.Handler().ActivateContentType(t.Context(), cm.ActivateContentTypeParams{
						SpaceID: "space", EnvironmentID: "environment", ContentTypeID: "external-marked-activation", XContentfulVersion: 3,
					})
					require.NoError(t, activateErr)
					require.IsType(t, &cm.ContentTypeStatusCode{}, response)

					handler.resetRequestHistory()
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"), ConfigVariables: variables,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionNoop),
				}},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("contentful_content_type.test", tfjsonpath.New("published_version"), knownvalue.Int64Exact(3)),
				},
				Check: contentTypeActivationRequestCheck(handler, 0, 0),
			},
			{
				PreConfig: func() {
					additionalCLIOptions.Plan.NoRefresh = true

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

func TestAccContentTypeResourceAmbiguousUpdateActivationReconcilesByRead(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	handler := &contentTypeActivationTestHandler{delegate: server}
	configVariables := contentTypeActivationConfigVariables("ambiguous-update-activation")

	testAccMockedResource(t, handler, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/1"),
				ConfigVariables: configVariables,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"contentful_content_type.test",
						tfjsonpath.New("published_version"),
						knownvalue.Int64Exact(1),
					),
				},
				Check: contentTypeActivationRequestAndVersionsCheck(handler, 1, 1, []int64{1}),
			},
			{
				PreConfig: func() {
					handler.resetRequestHistory()
					handler.activationFailureAfterSuccessStatus.Store(http.StatusBadGateway)
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"),
				ConfigVariables: configVariables,
				ExpectError:     regexp.MustCompile(`Failed to activate content type`),
			},
			{
				PreConfig: func() {
					require.Equal(t, 1, handler.eventCount(contentTypeOperationPut))
					require.Equal(t, 1, handler.eventCount(contentTypeOperationActivate))
					require.Equal(t, []int64{3}, handler.activationVersionHistory())

					handler.resetRequestHistory()
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionNoop),
						plancheck.ExpectKnownValue(
							"contentful_content_type.test",
							tfjsonpath.New("published_version"),
							knownvalue.Int64Exact(3),
						),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"contentful_content_type.test",
						tfjsonpath.New("published_version"),
						knownvalue.Int64Exact(3),
					),
				},
				Check: contentTypeActivationRequestAndVersionsCheck(handler, 0, 0, nil),
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionNoop),
					},
				},
				Check: contentTypeActivationRequestAndVersionsCheck(handler, 0, 0, nil),
			},
		},
	})
}

func TestAccContentTypeResourceRefreshDisabledRecoveryAfterCommittedActivationDoesNotActivateNewerVersion(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	handler := &contentTypeActivationTestHandler{delegate: server}
	variables := contentTypeActivationConfigVariables("ambiguous-activation-no-refresh")

	testAccMockedResource(t, handler, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
		Steps: []resource.TestStep{
			{ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/1"), ConfigVariables: variables},
			{
				PreConfig: func() {
					handler.resetRequestHistory()
					handler.activationFailureAfterSuccessStatus.Store(http.StatusBadGateway)
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"), ConfigVariables: variables,
				ExpectError: regexp.MustCompile(`Failed to activate content type`),
			},
			{
				PreConfig: func() {
					require.Equal(t, 1, handler.eventCount(contentTypeOperationPut))
					require.Equal(t, []int64{3}, handler.activationVersionHistory())
					handler.resetRequestHistory()
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"), ConfigVariables: variables,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionUpdate),
				}},
				ExpectError: regexp.MustCompile(`VersionMismatch`),
			},
			{
				PreConfig: func() {
					require.Equal(t, []contentTypeEvent{{
						Operation:      contentTypeOperationActivate,
						Method:         http.MethodPut,
						Path:           "/spaces/space/environments/environment/content_types/ambiguous-activation-no-refresh/published",
						Version:        3,
						VersionPresent: true,
					}}, handler.eventHistory(), "recovery must use the marked version exactly once and never repeat the draft PUT")
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
