package provider_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
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

var (
	errUnexpectedContentTypeRequestCount        = errors.New("unexpected content type request count")
	errUnexpectedContentTypeResponseType        = errors.New("unexpected content type response type")
	errUnexpectedContentTypePublicationVersions = errors.New("unexpected content type publication versions")
	errExternalDraftSentinelLost                = errors.New("external unmodeled draft sentinel was not preserved through activation")
)

type contentTypeOperation uint8

const (
	contentTypeOperationGet contentTypeOperation = iota
	contentTypeOperationPut
	contentTypeOperationActivate
)

type contentTypeEvent struct {
	Operation contentTypeOperation
	Version   int64
}

func TestAccContentTypeResourceFailedCreateActivationRemainsReplaceable(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	handler := &contentTypeActivationTestHandler{delegate: server}
	handler.failActivation.Store(true)

	configVariables := contentTypeActivationConfigVariables("create-activation-failure")

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceCreate/1"),
				ConfigVariables: configVariables,
				ExpectError:     regexp.MustCompile(`Failed to activate content type`),
			},
			{
				PreConfig: func() {
					requireNoConfirmingGetAfterActivationFailure(t, handler)
					handler.failActivation.Store(false)
					handler.resetRequestHistory()
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceCreate/1"),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionReplace),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"contentful_content_type.test",
						tfjsonpath.New("published_version"),
						knownvalue.Int64Exact(1),
					),
				},
				Check: contentTypeActivationRequestCheck(handler, 1, 1),
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceCreate/1"),
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

// A successful draft PUT that returns a contradictory body is not safe to
// publish. The response is checkpointed for recovery, but must not be
// activated merely because the transport status was successful.
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
		"wrong current version": {
			"testdata/TestAccContentTypeResourceUpdate/1", "testdata/TestAccContentTypeResourceUpdate/2",
			func(response map[string]any) { mustContentTypeResponseObject(response["sys"])["version"] = 99 },
			`Unexpected Contentful content type draft response`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
			require.NoError(t, err)
			server.RegisterSpaceEnvironment("space", "environment")
			handler := &contentTypeActivationTestHandler{delegate: server}
			variables := contentTypeActivationConfigVariables("contradictory-put-response-" + strings.ReplaceAll(name, " ", "-"))
			ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{Steps: []resource.TestStep{
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
					Check: contentTypeActivationRequestAndVersionsCheck(handler, 0, 1, []int64{3}),
				},
			}})
		})
	}
}

func TestAccContentTypeResourceRejectsContradictoryCreateActivationPublication(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")
	handler := &contentTypeActivationTestHandler{delegate: server}
	handler.mutateActivationResponse = contentTypeResponseMutator(t, removeContentTypePublishedVersion)
	variables := contentTypeActivationConfigVariables("contradictory-create-activation-publication")

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{Steps: []resource.TestStep{
		{
			ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceCreate/1"), ConfigVariables: variables,
			ExpectError: regexp.MustCompile(`Unexpected Contentful content type activation response`),
		},
		{
			PreConfig: func() {
				require.Equal(t, []contentTypeEvent{
					{Operation: contentTypeOperationPut, Version: 1},
					{Operation: contentTypeOperationActivate, Version: 1},
				}, handler.eventHistory())
				handler.resetRequestHistory()
			},
			ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceCreate/1"), ConfigVariables: variables,
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionReplace),
			}},
			Check: contentTypeActivationRequestAndVersionsCheck(handler, 1, 1, []int64{1}),
		},
	}})
}

func TestAccContentTypeResourceRejectsContradictoryPutUpdateActivationPublication(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")
	handler := &contentTypeActivationTestHandler{delegate: server}
	variables := contentTypeActivationConfigVariables("contradictory-put-update-activation-publication")

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{Steps: []resource.TestStep{
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
					{Operation: contentTypeOperationPut, Version: 2},
					{Operation: contentTypeOperationActivate, Version: 3},
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

func TestAccContentTypeResourceRejectsContradictoryActivationOnlyPublication(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")
	handler := &contentTypeActivationTestHandler{delegate: server}
	variables := contentTypeActivationConfigVariables("contradictory-activation-only-publication")
	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{Steps: []resource.TestStep{
		{ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceCreate/1"), ConfigVariables: variables},
		{
			PreConfig: func() {
				_, deactivateErr := server.Handler().DeactivateContentType(t.Context(), cm.DeactivateContentTypeParams{SpaceID: "space", EnvironmentID: "environment", ContentTypeID: "contradictory-activation-only-publication"})
				require.NoError(t, deactivateErr)
				handler.resetRequestHistory()
				handler.mutateActivationResponse = contentTypeResponseMutator(t, removeContentTypePublishedVersion)
			},
			ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceCreate/1"), ConfigVariables: variables,
			ExpectError: regexp.MustCompile(`Unexpected Contentful content type activation response`),
		},
		{
			PreConfig: func() {
				require.Equal(t, 0, handler.eventCount(contentTypeOperationPut))
				require.Equal(t, 1, handler.eventCount(contentTypeOperationActivate))
				response, getErr := server.Handler().GetContentType(t.Context(), cm.GetContentTypeParams{SpaceID: "space", EnvironmentID: "environment", ContentTypeID: "contradictory-activation-only-publication"})
				require.NoError(t, getErr)

				contentType, ok := response.(*cm.ContentType)
				require.True(t, ok)

				publishedVersion, published := contentType.Sys.PublishedVersion.Get()
				require.True(t, published)
				require.Equal(t, 3, publishedVersion)
				require.Equal(t, 4, contentType.Sys.Version)

				handler.mutateActivationResponse = nil
				handler.resetRequestHistory()
			},
			ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceCreate/1"), ConfigVariables: variables,
			Check: contentTypeActivationRequestCheck(handler, 0, 0),
		},
	}})
}

func removeContentTypePublishedVersion(response map[string]any) {
	delete(mustContentTypeResponseObject(response["sys"]), "publishedVersion")
}

func TestAccContentTypeResourceAmbiguousCreateActivationRemainsReplaceable(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	handler := &contentTypeActivationTestHandler{delegate: server}
	handler.activationFailureAfterSuccessStatus.Store(http.StatusUnprocessableEntity)

	configVariables := contentTypeActivationConfigVariables("ambiguous-create-activation")

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceCreate/1"),
				ConfigVariables: configVariables,
				ExpectError:     regexp.MustCompile(`Failed to activate content type`),
			},
			{
				PreConfig: func() {
					require.Equal(t, 1, handler.eventCount(contentTypeOperationPut))
					require.Equal(t, 1, handler.eventCount(contentTypeOperationActivate))
					require.Equal(t, []int64{1}, handler.activationVersionHistory())

					handler.resetRequestHistory()
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceCreate/1"),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionReplace),
					},
				},
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
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceCreate/1"),
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

func TestAccContentTypeResourceRecoversFailedUpdateActivation(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	handler := &contentTypeActivationTestHandler{delegate: server}
	configVariables := contentTypeActivationConfigVariables("update-activation-failure")

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{
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
					handler.failActivation.Store(true)
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"),
				ConfigVariables: configVariables,
				ExpectError:     regexp.MustCompile(`Failed to activate content type`),
			},
			{
				PreConfig: func() {
					requireNoConfirmingGetAfterActivationFailure(t, handler)
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"),
				ConfigVariables: configVariables,
				ExpectError:     regexp.MustCompile(`Failed to activate content type`),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(
							"contentful_content_type.test",
							tfjsonpath.New("published_version"),
							knownvalue.Int64Exact(3),
						),
					},
				},
			},
			{
				PreConfig: func() {
					handler.failActivation.Store(false)
					handler.resetRequestHistory()
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionUpdate),
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
				Check: contentTypeActivationRequestCheck(handler, 0, 1),
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"),
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

// The failed Update's response state and private version must be sufficient for
// recovery even before a normal refresh. This isolates the mutation checkpoint
// from the refresh-based recovery path exercised above.
func TestAccContentTypeResourceCheckpointsFailedUpdateWithoutRefresh(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	handler := &contentTypeActivationTestHandler{delegate: server}
	configVariables := contentTypeActivationConfigVariables("update-checkpoint-no-refresh")

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{
			Plan: resource.PlanOptions{NoRefresh: true},
		},
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/1"),
				ConfigVariables: configVariables,
			},
			{
				PreConfig: func() {
					handler.failActivation.Store(true)
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"),
				ConfigVariables: configVariables,
				ExpectError:     regexp.MustCompile(`Failed to activate content type`),
			},
			{
				PreConfig: func() {
					handler.failActivation.Store(false)
					handler.resetRequestHistory()
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionUpdate),
					plancheck.ExpectKnownValue(
						"contentful_content_type.test",
						tfjsonpath.New("published_version"),
						knownvalue.Int64Exact(3),
					),
				}},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"contentful_content_type.test",
						tfjsonpath.New("published_version"),
						knownvalue.Int64Exact(3),
					),
				},
				Check: contentTypeActivationRequestAndVersionsCheck(handler, 0, 1, []int64{3}),
			},
		},
	})
}

func TestAccContentTypeResourceRecoversAmbiguousUpdateActivation(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	handler := &contentTypeActivationTestHandler{delegate: server}
	configVariables := contentTypeActivationConfigVariables("ambiguous-update-activation")

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{
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

func TestAccContentTypeResourceTimeoutOnlyUpdateDoesNotMutate(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	handler := &contentTypeActivationTestHandler{delegate: server}
	configVariables := contentTypeActivationConfigVariables("timeout-only-update")

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceCreate/1"),
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

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{
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

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceCreate/1"),
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

func TestAccContentTypeResourceTimeoutChangeReactivatesExternalDeactivationWithoutDraftPut(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	handler := &contentTypeActivationTestHandler{delegate: server}
	configVariables := contentTypeActivationConfigVariables("deactivated-content-type")

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceCreate/1"),
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
				Check: contentTypeActivationRequestAndVersionsCheck(handler, 0, 1, []int64{3}),
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

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceCreate/1"),
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
						XContentfulVersion: 2,
					})
					require.NoError(t, putErr)
					require.IsType(t, &cm.ContentTypeStatusCode{}, response)

					handler.resetRequestHistory()
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceCreate/1"),
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
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceCreate/1"),
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

// This is intentional ownership policy: a newer externally-created draft with
// no provider-modeled drift is activated in place, rather than overwritten by
// a complete PUT. Consequently any values outside this schema survive.
func TestAccContentTypeResourceActivatesExternalPendingDraftWithoutModeledDrift(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")
	handler := &contentTypeActivationTestHandler{delegate: server}
	variables := contentTypeActivationConfigVariables("external-pending-no-drift")
	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{Steps: []resource.TestStep{
		{ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceCreate/1"), ConfigVariables: variables},
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
				// PUT. The only meaningful change is the new draft version; the
				// sidecar below represents data outside the provider schema.
				wireResponse, marshalErr := json.Marshal(contentType)
				require.NoError(t, marshalErr)

				var identicalDraft cm.ContentTypeRequestData
				require.NoError(t, json.Unmarshal(wireResponse, &identicalDraft))

				response, putErr := server.Handler().PutContentType(t.Context(), &identicalDraft, cm.PutContentTypeParams{
					SpaceID:            "space",
					EnvironmentID:      "environment",
					ContentTypeID:      "external-pending-no-drift",
					XContentfulVersion: 2,
				})
				require.NoError(t, putErr)
				require.IsType(t, &cm.ContentTypeStatusCode{}, response)
				handler.unmodeledDraftSentinel.Store(true)
				handler.resetRequestHistory()
			},
			ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceCreate/1"), ConfigVariables: variables,
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionUpdate),
				plancheck.ExpectKnownValue("contentful_content_type.test", tfjsonpath.New("published_version"), knownvalue.Int64Exact(3)),
			}},
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("contentful_content_type.test", tfjsonpath.New("fields"), knownvalue.ListSizeExact(2)),
				statecheck.ExpectKnownValue("contentful_content_type.test", tfjsonpath.New("published_version"), knownvalue.Int64Exact(3)),
			},
			Check: func(state *terraform.State) error {
				err := contentTypeActivationRequestAndVersionsCheck(handler, 0, 1, []int64{3})(state)
				if err != nil {
					return err
				}

				if !handler.sentinelObserved.Load() {
					return errExternalDraftSentinelLost
				}

				return nil
			},
		},
		{
			ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceCreate/1"),
			ConfigVariables: variables,
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionNoop),
			}},
		},
	}})
}

func TestAccContentTypeResourceDraftPutRaceDoesNotRetryAgainstNewerVersion(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	handler := &contentTypeActivationTestHandler{delegate: server}
	variables := contentTypeActivationConfigVariables("draft-put-race")
	tracedRace := make(chan error, 1)

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{Steps: []resource.TestStep{
		{ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/1"), ConfigVariables: variables},
		{
			PreConfig: func() {
				handler.resetRequestHistory()
				handler.beforePut = func(request *http.Request) {
					currentResponse, getErr := server.Handler().GetContentType(request.Context(), cm.GetContentTypeParams{
						SpaceID: "space", EnvironmentID: "environment", ContentTypeID: "draft-put-race",
					})
					if getErr != nil {
						tracedRace <- fmt.Errorf("get draft before racing PUT: %w", getErr)

						return
					}

					current, ok := currentResponse.(*cm.ContentType)
					if !ok {
						tracedRace <- fmt.Errorf("%w: got %T while reading draft before race", errUnexpectedContentTypeResponseType, currentResponse)

						return
					}

					wireCurrent, marshalErr := json.Marshal(current)
					if marshalErr != nil {
						tracedRace <- fmt.Errorf("encode draft before racing PUT: %w", marshalErr)

						return
					}

					var racingDraft cm.ContentTypeRequestData

					unmarshalErr := json.Unmarshal(wireCurrent, &racingDraft)
					if unmarshalErr != nil {
						tracedRace <- fmt.Errorf("decode draft before racing PUT: %w", unmarshalErr)

						return
					}

					racingDraft.Name = "External racing draft"
					racingDraft.Description = cm.NewOptNilString("Must not be overwritten or activated")

					response, putErr := server.Handler().PutContentType(request.Context(), &racingDraft, cm.PutContentTypeParams{
						SpaceID: "space", EnvironmentID: "environment", ContentTypeID: "draft-put-race", XContentfulVersion: current.Sys.Version,
					})
					if putErr != nil {
						tracedRace <- fmt.Errorf("create racing draft before Terraform PUT: %w", putErr)

						return
					}

					if _, ok := response.(*cm.ContentTypeStatusCode); !ok {
						tracedRace <- fmt.Errorf("%w: got %T while creating racing draft", errUnexpectedContentTypeResponseType, response)

						return
					}

					tracedRace <- nil
				}
			},
			ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"), ConfigVariables: variables,
			ExpectError: regexp.MustCompile(`Failed to save content type draft`),
		},
		{
			PreConfig: func() {
				select {
				case raceErr := <-tracedRace:
					require.NoError(t, raceErr)
				default:
					require.Fail(t, "draft PUT race callback did not report a result")
				}

				require.Equal(t, 1, handler.eventCount(contentTypeOperationPut))
				require.Equal(t, 0, handler.eventCount(contentTypeOperationActivate))

				response, getErr := server.Handler().GetContentType(t.Context(), cm.GetContentTypeParams{
					SpaceID: "space", EnvironmentID: "environment", ContentTypeID: "draft-put-race",
				})
				require.NoError(t, getErr)

				contentType, ok := response.(*cm.ContentType)
				require.True(t, ok)
				require.Equal(t, "External racing draft", contentType.Name)
				require.Equal(t, 3, contentType.Sys.Version)
				publishedVersion, published := contentType.Sys.PublishedVersion.Get()
				require.True(t, published)
				require.Equal(t, 1, publishedVersion)

				handler.beforePut = nil
				handler.resetRequestHistory()
			},
			ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"), ConfigVariables: variables,
			Check: contentTypeActivationRequestAndVersionsCheck(handler, 1, 1, []int64{4}),
		},
	}})
}

func TestAccContentTypeResourceActivationRaceDoesNotPublishInterveningDraft(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	handler := &contentTypeActivationTestHandler{delegate: server}
	configVariables := contentTypeActivationConfigVariables("activation-race")
	raceSetupResult := make(chan error, 1)

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceCreate/1"),
				ConfigVariables: configVariables,
			},
			{
				PreConfig: func() {
					response, deactivateErr := server.Handler().DeactivateContentType(t.Context(), cm.DeactivateContentTypeParams{
						SpaceID:       "space",
						EnvironmentID: "environment",
						ContentTypeID: "activation-race",
					})
					require.NoError(t, deactivateErr)
					require.IsType(t, &cm.ContentType{}, response)

					handler.beforeActivation = func(request *http.Request) {
						raceResponse, putErr := server.Handler().PutContentType(request.Context(), &cm.ContentTypeRequestData{
							Name:         "Racing external draft",
							Description:  cm.NewOptNilString("Must not be activated"),
							DisplayField: "external",
							Fields:       []cm.ContentTypeRequestDataFieldsItem{},
						}, cm.PutContentTypeParams{
							SpaceID:            "space",
							EnvironmentID:      "environment",
							ContentTypeID:      "activation-race",
							XContentfulVersion: 3,
						})
						if putErr != nil {
							raceSetupResult <- fmt.Errorf("create racing external draft: %w", putErr)

							return
						}

						if _, ok := raceResponse.(*cm.ContentTypeStatusCode); !ok {
							raceSetupResult <- fmt.Errorf(
								"%w: got %T while creating racing external draft",
								errUnexpectedContentTypeResponseType,
								raceResponse,
							)

							return
						}

						raceSetupResult <- nil
					}
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceCreate/1"),
				ConfigVariables: configVariables,
				ExpectError:     regexp.MustCompile(`Failed to activate content type`),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(
							"contentful_content_type.test",
							tfjsonpath.New("published_version"),
							knownvalue.Int64Exact(3),
						),
					},
				},
			},
			{
				PreConfig: func() {
					select {
					case setupErr := <-raceSetupResult:
						require.NoError(t, setupErr)
					default:
						require.Fail(t, "intervening draft callback did not report a result")
					}

					require.Equal(t, int64(3), handler.lastVersion(contentTypeOperationActivate))

					handler.beforeActivation = nil
					handler.resetRequestHistory()
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceCreate/1"),
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
						knownvalue.Int64Exact(5),
					),
				},
				Check: resource.ComposeTestCheckFunc(
					contentTypeActivationRequestCheck(handler, 1, 1),
					func(_ *terraform.State) error {
						if got := handler.lastVersion(contentTypeOperationActivate); got != 5 {
							return fmt.Errorf("%w: got activation version %d, want 5", errUnexpectedContentTypeRequestCount, got)
						}

						response, getErr := server.Handler().GetContentType(t.Context(), cm.GetContentTypeParams{
							SpaceID:       "space",
							EnvironmentID: "environment",
							ContentTypeID: "activation-race",
						})
						if getErr != nil {
							return fmt.Errorf("get activation-race content type: %w", getErr)
						}

						contentType, ok := response.(*cm.ContentType)
						if !ok {
							return fmt.Errorf("%w: got %T", errUnexpectedContentTypeResponseType, response)
						}

						publishedVersion, ok := contentType.Sys.PublishedVersion.Get()
						if !ok || publishedVersion != 5 || contentType.Sys.Version != 6 {
							return fmt.Errorf("%w: published=%d present=%t current=%d, want published=5 present=true current=6", errUnexpectedContentTypePublicationVersions, publishedVersion, ok, contentType.Sys.Version)
						}

						return nil
					},
				),
			},
		},
	})
}

func mustContentTypeResponseObject(value any) map[string]any {
	converted, ok := value.(map[string]any)
	if !ok {
		panic(fmt.Sprintf("unexpected injected content type response object %T", value))
	}

	return converted
}

func contentTypeResponseMutator(t *testing.T, mutate func(map[string]any)) func(string) string {
	t.Helper()

	return func(body string) string {
		var response map[string]any

		decodeErr := json.Unmarshal([]byte(body), &response)
		if decodeErr != nil {
			t.Errorf("decode injected content type response: %v", decodeErr)

			return body
		}

		mutate(response)

		mutated, err := json.Marshal(response)
		if err != nil {
			t.Errorf("encode injected content type response: %v", err)

			return body
		}

		return string(mutated)
	}
}

type contentTypeActivationTestHandler struct {
	delegate                            http.Handler
	failActivation                      atomic.Bool
	activationFailureAfterSuccessStatus atomic.Int64
	activationFailureReturned           atomic.Bool

	mu               sync.Mutex
	events           []contentTypeEvent
	beforeActivation func(*http.Request)
	beforePut        func(*http.Request)
	// mutatePutResponse is deliberately an HTTP-test-only fault injector. It
	// models a CMA success response that contradicts the accepted request.
	mutatePutResponse        func(string) string
	mutateActivationResponse func(string) string
	// A mock-only sidecar: a provider PUT clears it; activation observes it.
	unmodeledDraftSentinel atomic.Bool
	sentinelObserved       atomic.Bool
}

func (h *contentTypeActivationTestHandler) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodGet && strings.Contains(request.URL.Path, "/content_types/") && h.activationFailureReturned.Load() {
		h.recordEvent(contentTypeOperationGet, request)
	}

	if request.Method == http.MethodPut && strings.HasSuffix(request.URL.Path, "/published") {
		h.serveActivation(responseWriter, request)

		return
	}

	if request.Method == http.MethodPut &&
		strings.Contains(request.URL.Path, "/content_types/") &&
		!strings.HasSuffix(request.URL.Path, "/editor_interface") {
		h.recordEvent(contentTypeOperationPut, request)
		h.unmodeledDraftSentinel.Store(false)

		if beforePut := h.takeBeforePut(); beforePut != nil {
			beforePut(request)
		}

		if mutatePutResponse := h.takePutResponseMutator(); mutatePutResponse != nil {
			recorder := httptest.NewRecorder()
			h.delegate.ServeHTTP(recorder, request)

			for name, values := range recorder.Header() {
				responseWriter.Header()[name] = append([]string(nil), values...)
			}

			responseWriter.WriteHeader(recorder.Code)
			_, _ = responseWriter.Write([]byte(mutatePutResponse(recorder.Body.String())))

			return
		}
	}

	h.delegate.ServeHTTP(responseWriter, request)
}

func (h *contentTypeActivationTestHandler) serveActivation(responseWriter http.ResponseWriter, request *http.Request) {
	h.recordEvent(contentTypeOperationActivate, request)
	h.sentinelObserved.Store(h.unmodeledDraftSentinel.Load())

	if beforeActivation := h.takeBeforeActivation(); beforeActivation != nil {
		beforeActivation(request)
	}

	if h.failActivation.Load() {
		h.activationFailureReturned.Store(true)

		message := "Injected content type activation failure"
		_ = cmt.WriteContentfulManagementErrorResponse(
			responseWriter,
			http.StatusUnprocessableEntity,
			"ValidationFailed",
			&message,
			nil,
		)

		return
	}

	mutateActivationResponse := h.takeActivationResponseMutator()
	if h.activationFailureAfterSuccessStatus.Load() == 0 && mutateActivationResponse == nil {
		h.delegate.ServeHTTP(responseWriter, request)

		return
	}

	recorder := httptest.NewRecorder()
	h.delegate.ServeHTTP(recorder, request)

	if mutateActivationResponse != nil {
		for name, values := range recorder.Header() {
			responseWriter.Header()[name] = append([]string(nil), values...)
		}

		responseWriter.WriteHeader(recorder.Code)
		_, _ = responseWriter.Write([]byte(mutateActivationResponse(recorder.Body.String())))

		return
	}

	if recorder.Code >= http.StatusOK && recorder.Code < http.StatusMultipleChoices {
		failureStatus := int(h.activationFailureAfterSuccessStatus.Swap(0))
		if failureStatus != 0 {
			h.activationFailureReturned.Store(true)

			message := "Injected content type activation response failure"
			_ = cmt.WriteContentfulManagementErrorResponse(
				responseWriter,
				failureStatus,
				"ServerError",
				&message,
				nil,
			)

			return
		}
	}

	for name, values := range recorder.Header() {
		responseWriter.Header()[name] = append([]string(nil), values...)
	}

	responseWriter.WriteHeader(recorder.Code)
	_, _ = responseWriter.Write(recorder.Body.Bytes())
}

func (h *contentTypeActivationTestHandler) recordEvent(operation contentTypeOperation, request *http.Request) {
	version := int64(0)

	if operation != contentTypeOperationGet {
		var err error

		version, err = strconv.ParseInt(request.Header.Get("X-Contentful-Version"), 10, 64)
		if err != nil {
			version = -1
		}
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	h.events = append(h.events, contentTypeEvent{Operation: operation, Version: version})
}

func (h *contentTypeActivationTestHandler) activationVersionHistory() []int64 {
	events := h.eventHistory()
	versions := make([]int64, 0, len(events))

	for _, event := range events {
		if event.Operation == contentTypeOperationActivate {
			versions = append(versions, event.Version)
		}
	}

	return versions
}

func (h *contentTypeActivationTestHandler) resetRequestHistory() {
	h.mu.Lock()
	h.events = nil
	h.mu.Unlock()
	h.activationFailureReturned.Store(false)
}

func requireNoConfirmingGetAfterActivationFailure(t *testing.T, handler *contentTypeActivationTestHandler) {
	t.Helper()
	require.Equal(t, 0, handler.eventCount(contentTypeOperationGet), "provider issued a confirming GET after an activation failure")
	handler.activationFailureReturned.Store(false)
}

func (h *contentTypeActivationTestHandler) eventHistory() []contentTypeEvent {
	h.mu.Lock()
	defer h.mu.Unlock()

	return append([]contentTypeEvent(nil), h.events...)
}

func (h *contentTypeActivationTestHandler) eventCount(operation contentTypeOperation) int {
	count := 0

	for _, event := range h.eventHistory() {
		if event.Operation == operation {
			count++
		}
	}

	return count
}

func (h *contentTypeActivationTestHandler) lastVersion(operation contentTypeOperation) int64 {
	events := h.eventHistory()

	for _, event := range slices.Backward(events) {
		if event.Operation == operation {
			return event.Version
		}
	}

	return -1
}

func (h *contentTypeActivationTestHandler) takeBeforePut() func(*http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()

	hook := h.beforePut
	h.beforePut = nil

	return hook
}

func (h *contentTypeActivationTestHandler) takeBeforeActivation() func(*http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()

	hook := h.beforeActivation
	h.beforeActivation = nil

	return hook
}

func (h *contentTypeActivationTestHandler) takePutResponseMutator() func(string) string {
	h.mu.Lock()
	defer h.mu.Unlock()

	mutator := h.mutatePutResponse
	h.mutatePutResponse = nil

	return mutator
}

func (h *contentTypeActivationTestHandler) takeActivationResponseMutator() func(string) string {
	h.mu.Lock()
	defer h.mu.Unlock()

	mutator := h.mutateActivationResponse
	h.mutateActivationResponse = nil

	return mutator
}

func contentTypeActivationConfigVariables(contentTypeID string) config.Variables {
	return config.Variables{
		"space_id":             config.StringVariable("space"),
		"environment_id":       config.StringVariable("environment"),
		"test_content_type_id": config.StringVariable(contentTypeID),
	}
}

func contentTypeActivationRequestCheck(handler *contentTypeActivationTestHandler, wantPuts, wantActivations int64) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if got := int64(handler.eventCount(contentTypeOperationPut)); got != wantPuts {
			return fmt.Errorf("%w: got %d content-type PUTs, want %d", errUnexpectedContentTypeRequestCount, got, wantPuts)
		}

		if got := int64(handler.eventCount(contentTypeOperationActivate)); got != wantActivations {
			return fmt.Errorf("%w: got %d activations, want %d", errUnexpectedContentTypeRequestCount, got, wantActivations)
		}

		return nil
	}
}

func contentTypeActivationRequestAndVersionsCheck(
	handler *contentTypeActivationTestHandler,
	wantPuts int64,
	wantActivations int64,
	wantActivationVersions []int64,
) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		contentTypeActivationRequestCheck(handler, wantPuts, wantActivations),
		func(_ *terraform.State) error {
			if got := handler.activationVersionHistory(); !slices.Equal(got, wantActivationVersions) {
				return fmt.Errorf(
					"%w: got activation versions %v, want %v",
					errUnexpectedContentTypePublicationVersions,
					got,
					wantActivationVersions,
				)
			}

			wantOperations := make([]contentTypeOperation, 0, wantPuts+wantActivations)
			for range wantPuts {
				wantOperations = append(wantOperations, contentTypeOperationPut)
			}

			for range wantActivations {
				wantOperations = append(wantOperations, contentTypeOperationActivate)
			}

			gotOperations := make([]contentTypeOperation, 0, len(wantOperations))

			for _, event := range handler.eventHistory() {
				if event.Operation == contentTypeOperationPut || event.Operation == contentTypeOperationActivate {
					gotOperations = append(gotOperations, event.Operation)
				}
			}

			if !slices.Equal(gotOperations, wantOperations) {
				return fmt.Errorf(
					"%w: got operation order %v, want %v",
					errUnexpectedContentTypeRequestCount,
					gotOperations,
					wantOperations,
				)
			}

			return nil
		},
	)
}
