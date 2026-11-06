package provider_test

import (
	"fmt"
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

func TestAccContentTypeResourceDraftRateLimitRetriesSameRequestBeforeActivation(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")
	handler := &contentTypeActivationTestHandler{delegate: server}
	variables := contentTypeActivationConfigVariables("draft-rate-limit")
	testAccMockedResource(t, handler, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{
			Plan: resource.PlanOptions{
				NoRefresh: true,
			},
		},
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/1"),
				ConfigVariables: variables,
			},
			{
				PreConfig: func() {
					handler.resetRequestHistory()
					handler.rateLimitPut.Store(true)
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"),
				ConfigVariables: variables,
				Check: func(state *terraform.State) error {
					raw := handler.rawRequestHistory()
					require.Len(t, raw, 3)
					require.Equal(t, contentTypeOperationPut, raw[0].Operation)
					require.JSONEq(t, `{"name":"Test","description":"Test content type (draft-rate-limit)","displayField":"name","fields":[{"id":"name","name":"Name","type":"Symbol","required":true,"localized":false,"disabled":false,"omitted":false,"validations":[]},{"id":"slug","name":"Slug","type":"Symbol","required":true,"localized":false,"disabled":false,"omitted":false,"validations":[]}]}`, string(raw[0].Body))
					require.Equal(t, []string{"2"}, raw[0].VersionValues)
					require.Equal(t, raw[0], raw[1], "429 retries the exact draft body and version")
					require.Equal(t, contentTypeOperationActivate, raw[2].Operation)
					require.Equal(t, []string{"3"}, raw[2].VersionValues)

					return resource.TestCheckResourceAttr("contentful_content_type.test", "published_version", "3")(state)
				},
			},
			{
				PreConfig:       handler.resetRequestHistory,
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"),
				ConfigVariables: variables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionNoop),
					},
				},
				Check: contentTypeActivationRequestCheck(handler, 0, 0),
			},
		},
	})
}

func TestAccContentTypeResourceActivationRateLimitRetriesSameVersion(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")
	handler := &contentTypeActivationTestHandler{delegate: server}
	variables := contentTypeActivationConfigVariables("activation-rate-limit")
	testAccMockedResource(t, handler, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{
			Plan: resource.PlanOptions{
				NoRefresh: true,
			},
		},
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/1"),
				ConfigVariables: variables,
			},
			{
				PreConfig: func() {
					handler.resetRequestHistory()
					handler.rateLimitActivation.Store(true)
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"),
				ConfigVariables: variables,
				Check: func(state *terraform.State) error {
					raw := handler.rawRequestHistory()
					require.Len(t, raw, 3)
					require.Equal(t, contentTypeOperationPut, raw[0].Operation)
					require.Equal(t, []string{"2"}, raw[0].VersionValues)
					require.Equal(t, contentTypeOperationActivate, raw[1].Operation)
					require.Equal(t, []string{"3"}, raw[1].VersionValues)
					require.Empty(t, raw[1].Body)
					require.Equal(t, raw[1], raw[2], "activation retries only the acknowledged draft version")

					return resource.TestCheckResourceAttr("contentful_content_type.test", "published_version", "3")(state)
				},
			},
			{
				PreConfig:       handler.resetRequestHistory,
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"),
				ConfigVariables: variables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionNoop),
					},
				},
				Check: contentTypeActivationRequestCheck(handler, 0, 0),
			},
		},
	})
}

func TestAccContentTypeResourceDraftPutRaceDoesNotRetryAgainstNewerVersion(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	handler := &contentTypeActivationTestHandler{delegate: server}
	variables := contentTypeActivationConfigVariables("draft-put-race")
	tracedRace := make(chan error, 1)

	testAccMockedResource(t, handler, resource.TestCase{Steps: []resource.TestStep{
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

	testAccMockedResource(t, handler, resource.TestCase{
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

func TestAccContentTypeResourceCreateRetriesRateLimit(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")
	handler := &contentTypeActivationTestHandler{delegate: server}
	handler.rateLimitPut.Store(true)
	testAccMockedResource(t, handler, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/1"),
				ConfigVariables: contentTypeActivationConfigVariables("create-rate-limit"),
				Check: func(state *terraform.State) error {
					raw := handler.rawRequestHistory()
					require.Len(t, raw, 3)
					require.Equal(t, contentTypeOperationPut, raw[0].Operation)
					require.JSONEq(t, `{"name":"Test","description":"Test content type (create-rate-limit)","displayField":"name","fields":[{"id":"name","name":"Name","type":"Symbol","required":true,"localized":false,"disabled":false,"omitted":false,"validations":[]}]}`, string(raw[0].Body))
					require.Empty(t, raw[0].VersionValues)
					require.Equal(t, raw[0], raw[1])
					require.Equal(t, contentTypeOperationActivate, raw[2].Operation)
					require.Equal(t, []string{"1"}, raw[2].VersionValues)

					return resource.TestCheckResourceAttr("contentful_content_type.test", "published_version", "1")(state)
				},
			},
		},
	})
}

func TestAccContentTypeResourceActivationRateLimitThenVersionMismatchRevokesAuthority(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")
	handler := &contentTypeActivationTestHandler{delegate: server}
	variables := contentTypeActivationConfigVariables("rate-limit-then-conflict")
	testAccMockedResource(t, handler, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{
			Plan: resource.PlanOptions{
				NoRefresh: true,
			},
		},
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/1"),
				ConfigVariables: variables,
			},
			{
				PreConfig: func() {
					handler.resetRequestHistory()
					handler.rateLimitActivation.Store(true)
					handler.activationVersionMismatch.Store(true)
					handler.activationFailureReturned.Store(true)
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"),
				ConfigVariables: variables,
				ExpectError:     regexp.MustCompile(`(?s)VersionMismatch.*revoked activation authority`),
			},
			{
				PreConfig: func() {
					raw := handler.rawRequestHistory()
					require.Len(t, raw, 3, "429 and 409 must not cause a recovery GET")
					require.Equal(t, contentTypeOperationPut, raw[0].Operation)
					require.Equal(t, []string{"2"}, raw[0].VersionValues)
					require.Equal(t, contentTypeOperationActivate, raw[1].Operation)
					require.Equal(t, []string{"3"}, raw[1].VersionValues)
					require.Equal(t, raw[1], raw[2])
					handler.resetRequestHistory()
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"),
				ConfigVariables: variables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionNoop),
					},
				},
				Check: func(state *terraform.State) error {
					require.Empty(t, handler.eventHistory(), "409 revokes activation authority")

					return resource.TestCheckResourceAttr("contentful_content_type.test", "published_version", "1")(state)
				},
			},
		},
	})
}
