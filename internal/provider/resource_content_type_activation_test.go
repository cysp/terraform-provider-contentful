package provider_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
)

type contentTypeOperation uint8

const (
	contentTypeOperationGet contentTypeOperation = iota
	contentTypeOperationPut
	contentTypeOperationActivate
)

type contentTypeEvent struct {
	Operation      contentTypeOperation
	Method         string
	Path           string
	Version        int64
	VersionPresent bool
}

type contentTypeRawRequest struct {
	contentTypeEvent

	VersionValues []string
	Body          []byte
	ContentLength int64
}

func TestAccContentTypeResourceFailedCreateActivationRecoversExactDraft(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	handler := &contentTypeActivationTestHandler{delegate: server}
	handler.failActivation.Store(true)

	configVariables := contentTypeActivationConfigVariables("create-activation-failure")

	var providerFactoryCalls atomic.Int64

	ContentfulProviderMockedResourceTestWithFactoryCounter(t, handler, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
		Steps: []resource.TestStep{
			{
				ConfigDirectory:    config.StaticDirectory("testdata/TestAccContentTypeResourceCreate/1"),
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
					require.JSONEq(t, `{
					"name":"Test",
					"description":"Test content type (create-activation-failure)",
					"displayField":"name",
					"fields":[
						{"id":"name","name":"Name","type":"Symbol","localized":false,"required":true,"disabled":false,"omitted":false,"validations":[]},
						{"id":"flags","name":"Flags","type":"Array","localized":false,"required":false,"disabled":false,"omitted":false,"validations":[],"items":{"type":"Symbol","validations":[]}}
					]
				}`, string(handler.lastPutBody()))
					handler.failActivation.Store(false)
					handler.resetRequestHistory()
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceCreate/1"),
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
				Check: resource.ComposeTestCheckFunc(
					contentTypeActivationRequestAndVersionsCheck(handler, 0, 1, []int64{1}),
					func(*terraform.State) error {
						require.GreaterOrEqual(t, providerFactoryCalls.Load(), int64(2), "recovery must survive provider restart and private-state serialization")

						return nil
					},
				),
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
	}, &providerFactoryCalls)
}

func TestAccContentTypeResourceCreateUsesExactPositiveReturnedVersion(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")
	offset := &contentTypeVersionOffsetAdapter{delegate: server, offset: 3}
	handler := &contentTypeActivationTestHandler{delegate: offset}
	variables := contentTypeActivationConfigVariables("create-positive-version")

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{Steps: []resource.TestStep{
		{
			ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceCreate/1"), ConfigVariables: variables,
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
				require.JSONEq(t, `{
					"name":"Test",
					"description":"Test content type (create-positive-version)",
					"displayField":"name",
					"fields":[
						{"id":"name","name":"Name","type":"Symbol","localized":false,"required":true,"disabled":false,"omitted":false,"validations":[]},
						{"id":"flags","name":"Flags","type":"Array","localized":false,"required":false,"disabled":false,"omitted":false,"validations":[],"items":{"type":"Symbol","validations":[]}}
					]
				}`, string(raw[0].Body))
				require.Empty(t, raw[1].Body)
				require.Zero(t, raw[1].ContentLength)
				require.Equal(t, []string{"4"}, raw[1].VersionValues)

				return nil
			},
		},
		{
			PreConfig:       handler.resetRequestHistory,
			ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceCreate/1"), ConfigVariables: variables,
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

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{Steps: []resource.TestStep{
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

func TestAccContentTypeResourceDraftRateLimitDoesNotCreateActivationAuthority(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")
	handler := &contentTypeActivationTestHandler{delegate: server}
	variables := contentTypeActivationConfigVariables("draft-rate-limit")

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
		Steps: []resource.TestStep{
			{ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/1"), ConfigVariables: variables},
			{
				PreConfig: func() {
					handler.resetRequestHistory()
					handler.rateLimitPut.Store(true)
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"), ConfigVariables: variables,
				ExpectError: regexp.MustCompile(`Failed to update content type`),
			},
			{
				PreConfig: func() {
					require.Equal(t, 1, handler.eventCount(contentTypeOperationPut), "the draft 429 must not be retried")
					require.Zero(t, handler.eventCount(contentTypeOperationActivate), "a rejected draft must not authorize activation")
					handler.resetRequestHistory()
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"), ConfigVariables: variables,
				Check: contentTypeActivationRequestAndVersionsCheck(handler, 1, 1, []int64{3}),
			},
		},
	})
}

func TestAccContentTypeResourceActivationRateLimitRetainsExactAuthority(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")
	handler := &contentTypeActivationTestHandler{delegate: server}
	variables := contentTypeActivationConfigVariables("activation-rate-limit")

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
		Steps: []resource.TestStep{
			{ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/1"), ConfigVariables: variables},
			{
				PreConfig: func() {
					handler.resetRequestHistory()
					handler.rateLimitActivation.Store(true)
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"), ConfigVariables: variables,
				ExpectError: regexp.MustCompile(`Failed to activate content type`),
			},
			{
				PreConfig: func() {
					require.Equal(t, 1, handler.eventCount(contentTypeOperationPut))
					require.Equal(t, []int64{3}, handler.activationVersionHistory(), "the activation 429 must not be retried")
					handler.resetRequestHistory()
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"), ConfigVariables: variables,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionUpdate),
				}},
				Check: contentTypeActivationRequestAndVersionsCheck(handler, 0, 1, []int64{3}),
			},
		},
	})
}

func TestAccContentTypeResourceHigherPostActivationVersionIsAccepted(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")
	higher := &contentTypeHigherPostActivationVersionAdapter{delegate: server, server: server}
	handler := &contentTypeActivationTestHandler{delegate: higher}
	variables := contentTypeActivationConfigVariables("higher-post-activation-version")

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{
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

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{
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

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceCreate/1"), ConfigVariables: variables,
			},
			{
				PreConfig: func() {
					require.Equal(t, []contentTypeEvent{
						{Operation: contentTypeOperationPut, Method: http.MethodPut, Path: "/spaces/space/environments/environment/content_types/contradictory-create-activation-publication", Version: -1},
						{Operation: contentTypeOperationActivate, Method: http.MethodPut, Path: "/spaces/space/environments/environment/content_types/contradictory-create-activation-publication/published", Version: 1, VersionPresent: true},
					}, handler.eventHistory())
					handler.resetRequestHistory()
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceCreate/1"), ConfigVariables: variables,
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
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceCreate/1"), ConfigVariables: variables,
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

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{
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

func removeContentTypePublishedVersion(response map[string]any) {
	delete(mustContentTypeResponseObject(response["sys"]), "publishedVersion")
}

func TestAccContentTypeResourceAmbiguousCreateActivationReconcilesByRead(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	handler := &contentTypeActivationTestHandler{delegate: server}
	handler.activationFailureAfterSuccessStatus.Store(http.StatusUnprocessableEntity)

	configVariables := contentTypeActivationConfigVariables("ambiguous-create-activation")

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
		Steps: []resource.TestStep{
			{
				ConfigDirectory:    config.StaticDirectory("testdata/TestAccContentTypeResourceCreate/1"),
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

func TestAccContentTypeResourceFailedUpdateActivationRecoversExactDraftWithoutRefresh(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	handler := &contentTypeActivationTestHandler{delegate: server}
	configVariables := contentTypeActivationConfigVariables("update-activation-failure")

	var providerFactoryCalls atomic.Int64

	ContentfulProviderMockedResourceTestWithFactoryCounter(t, handler, resource.TestCase{
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
					require.JSONEq(t, `{
						"name":"Test",
						"description":"Test content type (update-activation-failure)",
						"displayField":"name",
						"fields":[
							{"id":"name","name":"Name","type":"Symbol","localized":false,"required":true,"disabled":false,"omitted":false,"validations":[]},
							{"id":"slug","name":"Slug","type":"Symbol","localized":false,"required":true,"disabled":false,"omitted":false,"validations":[]}
						]
					}`, string(handler.lastPutBody()))
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
						require.GreaterOrEqual(t, providerFactoryCalls.Load(), int64(3), "recovery must survive provider restart and private-state serialization")

						return nil
					},
				),
			},
		},
	}, &providerFactoryCalls)
}

func TestAccContentTypeResourceFailedUpdateActivationRecoversExactDraftAfterRefresh(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	handler := &contentTypeActivationTestHandler{delegate: server}
	variables := contentTypeActivationConfigVariables("update-activation-refresh-recovery")

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{Steps: []resource.TestStep{
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

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{Steps: []resource.TestStep{
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

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{
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

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{
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

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{
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

func TestAccContentTypeResourceRefreshDisabledRecoveryAfterCommittedActivationDoesNotActivateNewerVersion(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	handler := &contentTypeActivationTestHandler{delegate: server}
	variables := contentTypeActivationConfigVariables("ambiguous-activation-no-refresh")

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{
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

func TestAccContentTypeResourceTimeoutChangeLeavesExternalDeactivationObservational(t *testing.T) {
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
						XContentfulVersion: cm.NewOptInt(2),
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

func TestAccContentTypeResourceLeavesExternalPendingDraftObservational(t *testing.T) {
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
			ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceCreate/1"), ConfigVariables: variables,
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
			ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceCreate/1"),
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

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{AdditionalCLIOptions: additionalCLIOptions, Steps: []resource.TestStep{
		{
			ConfigDirectory:    config.StaticDirectory("testdata/TestAccContentTypeResourceCreate/1"),
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
			ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceCreate/1"),
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
			ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceCreate/1"),
			ConfigVariables: variables,
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionNoop),
			}},
			Check: contentTypeActivationRequestCheck(handler, 0, 0),
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
					tracedRace <- putExternalContentTypeDraft(request, server, "draft-put-race", "External racing draft", "Must not be overwritten or activated")
				}
			},
			ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"), ConfigVariables: variables,
			ExpectError: regexp.MustCompile(`Failed to update content type`),
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

func TestAccContentTypeResourceChangedDraftVersionMismatchRevokesPendingAuthority(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	handler := &contentTypeActivationTestHandler{delegate: server}
	variables := contentTypeActivationConfigVariables("changed-draft-version-mismatch")
	tracedRace := make(chan error, 1)

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{
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
					handler.resetRequestHistory()
					handler.beforePut = func(request *http.Request) {
						tracedRace <- putExternalContentTypeDraft(
							request, server, "changed-draft-version-mismatch", "External racing draft", "Must revoke the old marker",
						)
					}
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/1"), ConfigVariables: variables,
				ExpectError: regexp.MustCompile(`VersionMismatch`),
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
					require.Zero(t, handler.eventCount(contentTypeOperationActivate))
					handler.beforePut = nil
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
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/1"),
				ConfigVariables: configVariables,
			},
			{
				PreConfig: func() {
					handler.beforeActivation = func(request *http.Request) {
						raceSetupResult <- putExternalContentTypeDraft(request, server, "activation-race", "Racing external draft", "Must not be activated")
					}
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"),
				ConfigVariables: configVariables,
				ExpectError:     regexp.MustCompile(`Failed to activate content type`),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionUpdate),
						plancheck.ExpectUnknownValue("contentful_content_type.test", tfjsonpath.New("published_version")),
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

func putExternalContentTypeDraft(request *http.Request, server *cmt.Server, contentTypeID, name, description string) error {
	currentResponse, err := server.Handler().GetContentType(request.Context(), cm.GetContentTypeParams{
		SpaceID: "space", EnvironmentID: "environment", ContentTypeID: contentTypeID,
	})
	if err != nil {
		return fmt.Errorf("get Content Type before racing request: %w", err)
	}

	current, ok := currentResponse.(*cm.ContentType)
	if !ok {
		return fmt.Errorf("%w: got %T while reading draft before race", errUnexpectedContentTypeResponseType, currentResponse)
	}

	wireCurrent, err := json.Marshal(current)
	if err != nil {
		return fmt.Errorf("encode draft before racing request: %w", err)
	}

	var racingDraft cm.ContentTypeRequestData

	err = json.Unmarshal(wireCurrent, &racingDraft)
	if err != nil {
		return fmt.Errorf("decode draft before racing request: %w", err)
	}

	racingDraft.Name = name
	racingDraft.Description = cm.NewOptNilString(description)

	response, err := server.Handler().PutContentType(request.Context(), &racingDraft, cm.PutContentTypeParams{
		SpaceID: "space", EnvironmentID: "environment", ContentTypeID: contentTypeID, XContentfulVersion: cm.NewOptInt(current.Sys.Version),
	})
	if err != nil {
		return fmt.Errorf("create racing external draft: %w", err)
	}

	if _, ok := response.(*cm.ContentTypeStatusCode); !ok {
		return fmt.Errorf("%w: got %T while creating racing external draft", errUnexpectedContentTypeResponseType, response)
	}

	return nil
}

func putIdenticalExternalContentTypeDraft(t *testing.T, server *cmt.Server, contentTypeID string, version int) {
	t.Helper()

	currentResponse, err := server.Handler().GetContentType(t.Context(), cm.GetContentTypeParams{
		SpaceID: "space", EnvironmentID: "environment", ContentTypeID: contentTypeID,
	})
	require.NoError(t, err)

	current, ok := currentResponse.(*cm.ContentType)
	require.True(t, ok)
	require.Equal(t, version, current.Sys.Version)

	wireCurrent, err := json.Marshal(current)
	require.NoError(t, err)

	var identicalDraft cm.ContentTypeRequestData
	require.NoError(t, json.Unmarshal(wireCurrent, &identicalDraft))

	response, err := server.Handler().PutContentType(t.Context(), &identicalDraft, cm.PutContentTypeParams{
		SpaceID: "space", EnvironmentID: "environment", ContentTypeID: contentTypeID, XContentfulVersion: cm.NewOptInt(version),
	})
	require.NoError(t, err)
	require.IsType(t, &cm.ContentTypeStatusCode{}, response)
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
	activationVersionMismatch           atomic.Bool
	rateLimitPut                        atomic.Bool
	rateLimitActivation                 atomic.Bool
	activationFailureAfterSuccessStatus atomic.Int64
	activationFailureReturned           atomic.Bool

	mu               sync.Mutex
	events           []contentTypeEvent
	rawRequests      []contentTypeRawRequest
	putBodies        [][]byte
	beforeActivation func(*http.Request)
	beforePut        func(*http.Request)
	// mutatePutResponse is deliberately an HTTP-test-only fault injector. It
	// models a CMA success response that contradicts the accepted request.
	mutatePutResponse        func(string) string
	mutateActivationResponse func(string) string
	mutateReadResponse       func(string) string
}

type contentTypeVersionOffsetAdapter struct {
	delegate          http.Handler
	offset            int
	jumpOnDraftUpdate int

	mu     sync.Mutex
	jumped bool
}

type contentTypeHigherPostActivationVersionAdapter struct {
	delegate http.Handler
	server   *cmt.Server
	shot     entryOneShot
}

func (h *contentTypeVersionOffsetAdapter) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	isContentType := strings.Contains(request.URL.Path, "/content_types/") &&
		!strings.HasSuffix(request.URL.Path, "/editor_interface")
	if !isContentType || (request.Method != http.MethodGet && request.Method != http.MethodPut) {
		h.delegate.ServeHTTP(responseWriter, request)

		return
	}

	h.mu.Lock()
	offset := h.offset
	h.mu.Unlock()

	versionValues := request.Header.Values("X-Contentful-Version")
	isDraftUpdate := request.Method == http.MethodPut && !strings.HasSuffix(request.URL.Path, "/published") && len(versionValues) == 1

	if len(versionValues) == 1 {
		version, err := strconv.Atoi(versionValues[0])
		if err != nil {
			panic(fmt.Sprintf("parse offset Content Type version: %v", err))
		}

		request.Header.Set("X-Contentful-Version", strconv.Itoa(version-offset))
	}

	recorder := httptest.NewRecorder()
	h.delegate.ServeHTTP(recorder, request)

	if recorder.Code < http.StatusOK || recorder.Code >= http.StatusMultipleChoices {
		for name, values := range recorder.Header() {
			responseWriter.Header()[name] = append([]string(nil), values...)
		}

		responseWriter.WriteHeader(recorder.Code)
		_, _ = responseWriter.Write(recorder.Body.Bytes())

		return
	}

	h.mu.Lock()
	if isDraftUpdate && h.jumpOnDraftUpdate > 0 && !h.jumped {
		h.offset += h.jumpOnDraftUpdate
		h.jumped = true
	}

	offset = h.offset
	h.mu.Unlock()

	var payload map[string]any

	err := json.Unmarshal(recorder.Body.Bytes(), &payload)
	if err != nil {
		panic(fmt.Sprintf("decode offset Content Type response: %v", err))
	}

	sys := mustContentTypeResponseObject(payload["sys"])
	for _, name := range []string{"version", "publishedVersion"} {
		if value, present := sys[name].(float64); present {
			sys[name] = int(value) + offset
		}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		panic(fmt.Sprintf("encode offset Content Type response: %v", err))
	}

	for name, values := range recorder.Header() {
		responseWriter.Header()[name] = append([]string(nil), values...)
	}

	responseWriter.WriteHeader(recorder.Code)
	_, _ = responseWriter.Write(body)
}

func (h *contentTypeHigherPostActivationVersionAdapter) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPut || !strings.HasSuffix(request.URL.Path, "/published") || !h.shot.take() {
		h.delegate.ServeHTTP(responseWriter, request)

		return
	}

	recorder := httptest.NewRecorder()
	h.delegate.ServeHTTP(recorder, request)

	if recorder.Code < http.StatusOK || recorder.Code >= http.StatusMultipleChoices {
		for name, values := range recorder.Header() {
			responseWriter.Header()[name] = append([]string(nil), values...)
		}

		responseWriter.WriteHeader(recorder.Code)
		_, _ = responseWriter.Write(recorder.Body.Bytes())

		return
	}

	contentTypeID := strings.TrimSuffix(strings.TrimPrefix(request.URL.Path, "/spaces/space/environments/environment/content_types/"), "/published")

	currentResponse, err := h.server.Handler().GetContentType(request.Context(), cm.GetContentTypeParams{
		SpaceID: "space", EnvironmentID: "environment", ContentTypeID: contentTypeID,
	})
	if err != nil {
		panic(fmt.Sprintf("read Content Type after activation: %v", err))
	}

	current, ok := currentResponse.(*cm.ContentType)
	if !ok {
		panic(fmt.Sprintf("unexpected Content Type response after activation: %T", currentResponse))
	}

	for range 2 {
		wireCurrent, marshalErr := json.Marshal(current)
		if marshalErr != nil {
			panic(fmt.Sprintf("encode Content Type draft: %v", marshalErr))
		}

		var draft cm.ContentTypeRequestData

		unmarshalErr := json.Unmarshal(wireCurrent, &draft)
		if unmarshalErr != nil {
			panic(fmt.Sprintf("decode Content Type draft: %v", unmarshalErr))
		}

		putResponse, putErr := h.server.Handler().PutContentType(request.Context(), &draft, cm.PutContentTypeParams{
			SpaceID: "space", EnvironmentID: "environment", ContentTypeID: contentTypeID,
			XContentfulVersion: cm.NewOptInt(current.Sys.Version),
		})
		if putErr != nil {
			panic(fmt.Sprintf("write Content Type draft: %v", putErr))
		}

		status, statusOK := putResponse.(*cm.ContentTypeStatusCode)
		if !statusOK {
			panic(fmt.Sprintf("unexpected Content Type draft response: %T", putResponse))
		}

		current = &status.Response
	}

	body, err := json.Marshal(current)
	if err != nil {
		panic(fmt.Sprintf("encode high Content Type response: %v", err))
	}

	for name, values := range recorder.Header() {
		responseWriter.Header()[name] = append([]string(nil), values...)
	}

	responseWriter.WriteHeader(recorder.Code)
	_, _ = responseWriter.Write(body)
}

func (h *contentTypeActivationTestHandler) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodGet && strings.Contains(request.URL.Path, "/content_types/") && h.activationFailureReturned.Load() {
		h.recordEvent(contentTypeOperationGet, request)
	}

	if request.Method == http.MethodGet && strings.Contains(request.URL.Path, "/content_types/") &&
		!strings.HasSuffix(request.URL.Path, "/editor_interface") {
		if mutateReadResponse := h.takeReadResponseMutator(); mutateReadResponse != nil {
			recorder := httptest.NewRecorder()
			h.delegate.ServeHTTP(recorder, request)

			for name, values := range recorder.Header() {
				responseWriter.Header()[name] = append([]string(nil), values...)
			}

			responseWriter.WriteHeader(recorder.Code)
			_, _ = responseWriter.Write([]byte(mutateReadResponse(recorder.Body.String())))

			return
		}
	}

	if request.Method == http.MethodPut && strings.HasSuffix(request.URL.Path, "/published") {
		h.serveActivation(responseWriter, request)

		return
	}

	if request.Method == http.MethodPut &&
		strings.Contains(request.URL.Path, "/content_types/") &&
		!strings.HasSuffix(request.URL.Path, "/editor_interface") {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			panic(fmt.Sprintf("read content type PUT body: %v", err))
		}

		request.Body = io.NopCloser(bytes.NewReader(body))
		h.recordPutBody(body)
		h.recordEvent(contentTypeOperationPut, request)

		if h.rateLimitPut.Swap(false) {
			message := "Injected content type draft rate limit"
			_ = cmt.WriteContentfulManagementErrorResponse(responseWriter, http.StatusTooManyRequests, "RateLimitExceeded", &message, nil)

			return
		}

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

	if h.rateLimitActivation.Swap(false) {
		h.activationFailureReturned.Store(true)

		message := "Injected content type activation rate limit"
		_ = cmt.WriteContentfulManagementErrorResponse(responseWriter, http.StatusTooManyRequests, "RateLimitExceeded", &message, nil)

		return
	}

	if beforeActivation := h.takeBeforeActivation(); beforeActivation != nil {
		beforeActivation(request)
	}

	if h.activationVersionMismatch.Swap(false) {
		message := "Injected initial content type activation VersionMismatch"
		_ = cmt.WriteContentfulManagementErrorResponse(responseWriter, http.StatusConflict, "VersionMismatch", &message, nil)

		return
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
	body, err := io.ReadAll(request.Body)
	if err != nil {
		panic(fmt.Sprintf("read content type event body: %v", err))
	}

	request.Body = io.NopCloser(bytes.NewReader(body))

	version := int64(0)

	versionValues := append([]string(nil), request.Header.Values("X-Contentful-Version")...)
	versionPresent := len(versionValues) > 0

	if operation != contentTypeOperationGet {
		version = -1

		if versionPresent {
			parsedVersion, err := strconv.ParseInt(request.Header.Get("X-Contentful-Version"), 10, 64)
			if err == nil {
				version = parsedVersion
			}
		}
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	event := contentTypeEvent{
		Operation: operation, Method: request.Method, Path: request.URL.Path, Version: version, VersionPresent: versionPresent,
	}
	h.events = append(h.events, event)
	h.rawRequests = append(h.rawRequests, contentTypeRawRequest{
		contentTypeEvent: event,
		VersionValues:    versionValues,
		Body:             append([]byte(nil), body...),
		ContentLength:    request.ContentLength,
	})
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
	h.rawRequests = nil
	h.putBodies = nil
	h.mu.Unlock()
	h.activationFailureReturned.Store(false)
}

func (h *contentTypeActivationTestHandler) recordPutBody(body []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.putBodies = append(h.putBodies, append([]byte(nil), body...))
}

func (h *contentTypeActivationTestHandler) lastPutBody() []byte {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.putBodies) == 0 {
		return nil
	}

	return append([]byte(nil), h.putBodies[len(h.putBodies)-1]...)
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

func (h *contentTypeActivationTestHandler) rawRequestHistory() []contentTypeRawRequest {
	h.mu.Lock()
	defer h.mu.Unlock()

	return append([]contentTypeRawRequest(nil), h.rawRequests...)
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

func (h *contentTypeActivationTestHandler) takeReadResponseMutator() func(string) string {
	h.mu.Lock()
	defer h.mu.Unlock()

	mutator := h.mutateReadResponse
	h.mutateReadResponse = nil

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

		for _, request := range handler.rawRequestHistory() {
			switch request.Operation {
			case contentTypeOperationGet:
			case contentTypeOperationPut:
				if request.Method != http.MethodPut || len(request.Body) == 0 || request.ContentLength <= 0 {
					return fmt.Errorf("%w: malformed content-type draft request: method=%s body=%d content-length=%d", errUnexpectedContentTypeRequestCount, request.Method, len(request.Body), request.ContentLength)
				}

				if len(request.VersionValues) > 1 {
					return fmt.Errorf("%w: content-type draft request has %d version headers", errUnexpectedContentTypeRequestCount, len(request.VersionValues))
				}
			case contentTypeOperationActivate:
				if request.Method != http.MethodPut || len(request.Body) != 0 || request.ContentLength != 0 {
					return fmt.Errorf("%w: malformed content-type activation request: method=%s body=%d content-length=%d", errUnexpectedContentTypeRequestCount, request.Method, len(request.Body), request.ContentLength)
				}

				if len(request.VersionValues) != 1 || request.VersionValues[0] != strconv.FormatInt(request.Version, 10) {
					return fmt.Errorf("%w: content-type activation version headers=%v parsed-version=%d", errUnexpectedContentTypePublicationVersions, request.VersionValues, request.Version)
				}
			}
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
