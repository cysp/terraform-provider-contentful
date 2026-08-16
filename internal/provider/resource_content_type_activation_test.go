package provider_test

import (
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
)

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
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceCreate/1"),
				ConfigVariables: configVariables,
				ExpectError:     regexp.MustCompile(`Failed to activate content type`),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionReplace),
					},
				},
			},
		},
	})
}

func TestAccContentTypeResourceAmbiguousCreateActivationRemainsReplaceable(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	handler := &contentTypeActivationTestHandler{delegate: server}
	handler.failActivationAfterSuccess.Store(true)

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
					require.Equal(t, int64(1), handler.puts.Load())
					require.Equal(t, int64(2), handler.activations.Load())
					require.Equal(t, []int64{1, 1}, handler.activationVersionHistory())

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
					handler.puts.Store(0)
					handler.activations.Store(0)
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
					handler.failActivationAfterSuccess.Store(true)
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdate/2"),
				ConfigVariables: configVariables,
				ExpectError:     regexp.MustCompile(`Failed to activate content type`),
			},
			{
				PreConfig: func() {
					require.Equal(t, int64(1), handler.puts.Load())
					require.Equal(t, int64(2), handler.activations.Load())
					require.Equal(t, []int64{3, 3}, handler.activationVersionHistory())

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

func TestAccContentTypeResourceReactivatesDeactivatedContentType(t *testing.T) {
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

					handler.puts.Store(0)
					handler.activations.Store(0)
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceCreate/1"),
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
						DisplayField: "external",
						Fields:       []cm.ContentTypeRequestDataFieldsItem{},
					}, cm.PutContentTypeParams{
						SpaceID:            "space",
						EnvironmentID:      "environment",
						ContentTypeID:      "external-draft",
						XContentfulVersion: 2,
					})
					require.NoError(t, putErr)
					require.IsType(t, &cm.ContentTypeStatusCode{}, response)

					handler.puts.Store(0)
					handler.activations.Store(0)
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

					require.Equal(t, int64(3), handler.lastActivationVersion.Load())

					handler.beforeActivation = nil
					handler.puts.Store(0)
					handler.activations.Store(0)
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
						if got := handler.lastActivationVersion.Load(); got != 5 {
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

type contentTypeActivationTestHandler struct {
	delegate                   http.Handler
	failActivation             atomic.Bool
	failActivationAfterSuccess atomic.Bool
	puts                       atomic.Int64
	activations                atomic.Int64
	lastActivationVersion      atomic.Int64
	activationVersionsMu       sync.Mutex
	activationVersions         []int64
	beforeActivation           func(*http.Request)
	beforeActivationOnce       sync.Once
}

func (h *contentTypeActivationTestHandler) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodPut && strings.HasSuffix(request.URL.Path, "/published") {
		h.serveActivation(responseWriter, request)

		return
	}

	if request.Method == http.MethodPut &&
		strings.Contains(request.URL.Path, "/content_types/") &&
		!strings.HasSuffix(request.URL.Path, "/editor_interface") {
		h.puts.Add(1)
	}

	h.delegate.ServeHTTP(responseWriter, request)
}

func (h *contentTypeActivationTestHandler) serveActivation(responseWriter http.ResponseWriter, request *http.Request) {
	h.activations.Add(1)
	h.recordActivationVersion(request)

	if h.beforeActivation != nil {
		h.beforeActivationOnce.Do(func() {
			h.beforeActivation(request)
		})
	}

	if h.failActivation.Load() {
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

	if !h.failActivationAfterSuccess.Load() {
		h.delegate.ServeHTTP(responseWriter, request)

		return
	}

	recorder := httptest.NewRecorder()
	h.delegate.ServeHTTP(recorder, request)

	if recorder.Code >= http.StatusOK && recorder.Code < http.StatusMultipleChoices &&
		h.failActivationAfterSuccess.CompareAndSwap(true, false) {
		message := "Injected content type activation response failure"
		_ = cmt.WriteContentfulManagementErrorResponse(
			responseWriter,
			http.StatusBadGateway,
			"ServerError",
			&message,
			nil,
		)

		return
	}

	for name, values := range recorder.Header() {
		responseWriter.Header()[name] = append([]string(nil), values...)
	}

	responseWriter.WriteHeader(recorder.Code)
	_, _ = responseWriter.Write(recorder.Body.Bytes())
}

func (h *contentTypeActivationTestHandler) recordActivationVersion(request *http.Request) {
	version, err := strconv.ParseInt(request.Header.Get("X-Contentful-Version"), 10, 64)
	if err != nil {
		version = -1
	}

	h.lastActivationVersion.Store(version)
	h.activationVersionsMu.Lock()
	h.activationVersions = append(h.activationVersions, version)
	h.activationVersionsMu.Unlock()
}

func (h *contentTypeActivationTestHandler) activationVersionHistory() []int64 {
	h.activationVersionsMu.Lock()
	defer h.activationVersionsMu.Unlock()

	return append([]int64(nil), h.activationVersions...)
}

func (h *contentTypeActivationTestHandler) resetRequestHistory() {
	h.puts.Store(0)
	h.activations.Store(0)
	h.activationVersionsMu.Lock()
	h.activationVersions = nil
	h.activationVersionsMu.Unlock()
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
		if got := handler.puts.Load(); got != wantPuts {
			return fmt.Errorf("%w: got %d content-type PUTs, want %d", errUnexpectedContentTypeRequestCount, got, wantPuts)
		}

		if got := handler.activations.Load(); got != wantActivations {
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

			return nil
		},
	)
}
