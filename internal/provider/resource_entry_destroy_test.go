package provider_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sync"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

var errEntryDeleteAfterFailedUnpublish = errors.New("delete issued after failed unpublish")

type entryDestroyRequest struct {
	method         string
	path           string
	version        string
	versionPresent bool
	ifMatch        string
	ifMatchPresent bool
}

type entryDestroyTestAdapter struct {
	delegate  http.Handler
	errorSink *entryFixtureErrorSink
	reject    entryOneShot
	advance   entryOneShot
	after     func() error
	status    int
	errorID   string
	message   string
	recordAll bool

	mu       sync.Mutex
	rejected bool
	requests []entryDestroyRequest
}

func (h *entryDestroyTestAdapter) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodDelete {
		if h.recordAll {
			h.record(request)
		}

		h.delegate.ServeHTTP(responseWriter, request)

		return
	}

	rejected := h.record(request)

	if request.URL.Path == entryTestPublishPath && h.reject.take() {
		h.mu.Lock()
		h.rejected = true
		h.mu.Unlock()

		status := h.status
		if status == 0 {
			status = http.StatusBadRequest
		}

		errorID := h.errorID
		if errorID == "" {
			errorID = "BadRequest"
		}

		message := h.message
		if message == "" {
			message = "injected unpublish failure"
		}

		_ = cmt.WriteContentfulManagementErrorResponse(responseWriter, status, errorID, &message, nil)

		return
	}

	if request.URL.Path == entryTestPublishPath {
		h.mu.Lock()
		h.rejected = false
		h.mu.Unlock()
	}

	if request.URL.Path == entryTestUpdatePath && rejected {
		h.errorSink.record(errEntryDeleteAfterFailedUnpublish)
	}

	h.delegate.ServeHTTP(responseWriter, request)

	if request.URL.Path == entryTestPublishPath && h.advance.take() && h.after != nil {
		h.errorSink.record(h.after())
	}
}

func (h *entryDestroyTestAdapter) record(request *http.Request) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.requests = append(h.requests, entryDestroyRequest{
		method:         request.Method,
		path:           request.URL.Path,
		version:        request.Header.Get("X-Contentful-Version"),
		versionPresent: len(request.Header.Values("X-Contentful-Version")) > 0,
		ifMatch:        request.Header.Get("If-Match"),
		ifMatchPresent: len(request.Header.Values("If-Match")) > 0,
	})

	return h.rejected
}

func (h *entryDestroyTestAdapter) reset() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.requests = nil
}

func (h *entryDestroyTestAdapter) snapshot() []entryDestroyRequest {
	h.mu.Lock()
	defer h.mu.Unlock()

	return append([]entryDestroyRequest(nil), h.requests...)
}

func requireUnconditionalEntryUnpublish(t *testing.T, request entryDestroyRequest) {
	t.Helper()

	require.Equal(t, http.MethodDelete, request.method)
	require.Equal(t, entryTestPublishPath, request.path)
	require.False(t, request.versionPresent, "whole-Entry unpublish has no version precondition")
	require.Empty(t, request.version)
	require.False(t, request.ifMatchPresent, "whole-Entry unpublish has no ETag precondition")
	require.Empty(t, request.ifMatch)
}

func requireUnconditionalEntryDestroy(t *testing.T, requests []entryDestroyRequest) {
	t.Helper()

	require.Len(t, requests, 2, "destroy must unpublish before deleting")
	requireUnconditionalEntryUnpublish(t, requests[0])
	require.Equal(t, http.MethodDelete, requests[1].method)
	require.Equal(t, entryTestUpdatePath, requests[1].path)
	require.False(t, requests[1].versionPresent, "whole-Entry delete has no version precondition")
	require.Empty(t, requests[1].version)
	require.False(t, requests[1].ifMatchPresent, "whole-Entry delete has no ETag precondition")
	require.Empty(t, requests[1].ifMatch)
}

func requireEntryAbsent(t *testing.T, server *cmt.Server) {
	t.Helper()

	response, err := server.Handler().GetEntry(t.Context(), cm.GetEntryParams{
		SpaceID: "space", EnvironmentID: "environment", EntryID: "entry",
	})
	require.NoError(t, err)

	status, ok := response.(cm.StatusCodeResponse)
	require.True(t, ok)
	require.Equal(t, http.StatusNotFound, status.GetStatusCode())
}

func TestEntryResourceDestroyDoesNotConsumePrivateVersion(t *testing.T) {
	t.Parallel()

	malformedPrivate, err := json.Marshal(map[string][]byte{"version": []byte(`"invalid"`)})
	require.NoError(t, err)

	privateStates := map[string][]byte{
		"missing":   []byte(`{}`),
		"malformed": malformedPrivate,
		"zero":      privateVersionBytes(t, 0),
	}

	for name, privateState := range privateStates {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fixture := newEntryAcceptanceFixture(t)
			server, recorder := fixture.server, fixture.recorder
			adapter := &entryDestroyTestAdapter{delegate: server, errorSink: fixture.errorSink, recordAll: true}
			recorder.delegate = adapter

			createdResponse, err := server.Handler().PutEntry(t.Context(), &cm.EntryRequest{}, cm.PutEntryParams{
				SpaceID:                "space",
				EnvironmentID:          "environment",
				EntryID:                "entry",
				XContentfulContentType: cm.NewOptString("article"),
			})
			require.NoError(t, err)

			created, ok := createdResponse.(*cm.EntryStatusCode)
			require.True(t, ok)

			publishedResponse, err := server.Handler().PublishEntry(t.Context(), cm.PublishEntryParams{
				SpaceID:            "space",
				EnvironmentID:      "environment",
				EntryID:            "entry",
				XContentfulVersion: created.Response.Sys.Version,
			})
			require.NoError(t, err)

			_, ok = publishedResponse.(*cm.EntryStatusCode)
			require.True(t, ok)

			testServer := httptest.NewServer(recorder)
			t.Cleanup(testServer.Close)

			providerServer, err := makeTestAccProtoV6ProviderFactories(
				ContentfulProviderOptionsWithHTTPTestServer(testServer)...,
			)["contentful"]()
			require.NoError(t, err)

			providerConfig, err := providerConfigDynamicValue(map[string]any{
				"url":          tftypes.UnknownValue,
				"access_token": tftypes.UnknownValue,
			})
			require.NoError(t, err)

			configureResponse, err := providerServer.ConfigureProvider(t.Context(), &tfprotov6.ConfigureProviderRequest{
				Config: &providerConfig,
			})
			require.NoError(t, err)
			require.Empty(t, configureResponse.Diagnostics)

			stateModel := EntryModel{
				IDIdentityModel:    IDIdentityModel{ID: types.StringValue("space/environment/entry")},
				EntryIdentityModel: NewEntryIdentityModel("space", "environment", "entry"),
				ContentTypeID:      types.StringValue("article"),
				Fields:             NewTypedMap(map[string]TypedMap[jsontypes.Normalized]{}),
				Metadata: NewTypedObject(EntryMetadataValue{
					Concepts: NewTypedList([]types.String{}),
					Tags:     NewTypedList([]types.String{}),
				}),
				PublishedVersion: types.Int64Value(int64(created.Response.Sys.Version)),
				Timeouts:         TimeoutsNull(),
			}
			state := resourceModelDynamicValue(t, EntryResourceSchema(t.Context()), stateModel)
			plannedState := nullResourceDynamicValue(t, EntryResourceSchema(t.Context()))

			applyResponse, err := providerServer.ApplyResourceChange(t.Context(), &tfprotov6.ApplyResourceChangeRequest{
				TypeName:       "contentful_entry",
				PriorState:     &state,
				PlannedState:   &plannedState,
				Config:         &plannedState,
				PlannedPrivate: privateState,
			})
			require.NoError(t, err)
			require.Empty(t, applyResponse.Diagnostics)
			requireUnconditionalEntryDestroy(t, adapter.snapshot())
			requireEntryAbsent(t, server)
		})
	}
}

func TestAccEntryResourceFailedUnpublishStopsBeforeDeleteAndCanRetry(t *testing.T) {
	t.Parallel()
	testAccEntryRejectedUnpublishStopsBeforeDeleteAndCanRetry(t, nil)
}

func TestAccEntryResourceNon400NotPublishedStopsBeforeDeleteAndCanRetry(t *testing.T) {
	t.Parallel()
	testAccEntryRejectedUnpublishStopsBeforeDeleteAndCanRetry(t, func(adapter *entryDestroyTestAdapter) {
		adapter.status = http.StatusConflict
		adapter.errorID = "BadRequest"
		adapter.message = "Not published"
	})
}

func testAccEntryRejectedUnpublishStopsBeforeDeleteAndCanRetry(t *testing.T, configure func(*entryDestroyTestAdapter)) {
	t.Helper()

	fixture := newEntryAcceptanceFixture(t)
	server, recorder := fixture.server, fixture.recorder

	adapter := &entryDestroyTestAdapter{delegate: server, errorSink: fixture.errorSink}
	if configure != nil {
		configure(adapter)
	}

	recorder.delegate = adapter

	ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{Steps: []resource.TestStep{
		{Config: managedEntryConfig("one")},
		{
			PreConfig: func() {
				adapter.reset()
				adapter.reject.arm()
			},
			Config:      "# intentionally empty\n",
			ExpectError: regexp.MustCompile(`Failed to unpublish entry`),
		},
		{
			PreConfig: func() {
				requests := adapter.snapshot()
				require.Len(t, requests, 1, "delete must not be issued after unpublish fails")
				requireUnconditionalEntryUnpublish(t, requests[0])

				entry := getTestEntry(t, server)
				require.True(t, entry.Sys.PublishedVersion.IsSet(), "failed unpublish must leave the Entry published")

				adapter.reset()
			},
			Config: "# intentionally empty\n",
		},
	}})

	requests := adapter.snapshot()
	requireUnconditionalEntryDestroy(t, requests)
	requireEntryAbsent(t, server)
}

func TestAccEntryResourceDestroyRemovesExternalDraftWithoutRefresh(t *testing.T) {
	t.Parallel()

	fixture := newEntryAcceptanceFixture(t)
	server, recorder := fixture.server, fixture.recorder
	adapter := &entryDestroyTestAdapter{delegate: server, errorSink: fixture.errorSink}
	recorder.delegate = adapter

	ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{
			Plan: resource.PlanOptions{NoRefresh: true},
		},
		Steps: []resource.TestStep{
			{Config: managedEntryConfig("one")},
			{
				PreConfig: func() {
					entry := getTestEntry(t, server)
					response, err := server.Handler().PutEntry(t.Context(), &cm.EntryRequest{
						Fields: entry.Fields, Metadata: entry.Metadata,
					}, cm.PutEntryParams{
						SpaceID:                "space",
						EnvironmentID:          "environment",
						EntryID:                "entry",
						XContentfulContentType: cm.NewOptString("article"),
						XContentfulVersion:     cm.NewOptInt(entry.Sys.Version),
					})
					require.NoError(t, err)

					updated, ok := response.(*cm.EntryStatusCode)
					require.True(t, ok)
					require.Greater(t, updated.Response.Sys.Version, entry.Sys.Version)

					adapter.reset()
				},
				Config: "# intentionally empty\n",
			},
		},
	})

	requests := adapter.snapshot()
	requireUnconditionalEntryDestroy(t, requests)
	requireEntryAbsent(t, server)
}

func TestAccEntryResourceDestroyAlreadyAbsentIsIdempotentWithoutRefresh(t *testing.T) {
	t.Parallel()

	fixture := newEntryAcceptanceFixture(t)
	server, recorder := fixture.server, fixture.recorder
	adapter := &entryDestroyTestAdapter{delegate: server, errorSink: fixture.errorSink}
	recorder.delegate = adapter

	ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{
			Plan: resource.PlanOptions{NoRefresh: true},
		},
		Steps: []resource.TestStep{
			{Config: managedEntryConfig("one")},
			{
				PreConfig: func() {
					unpublishResponse, err := server.Handler().UnpublishEntry(t.Context(), cm.UnpublishEntryParams{
						SpaceID: "space", EnvironmentID: "environment", EntryID: "entry",
					})
					require.NoError(t, err)

					_, ok := unpublishResponse.(*cm.Entry)
					require.True(t, ok)

					deleteResponse, err := server.Handler().DeleteEntry(t.Context(), cm.DeleteEntryParams{
						SpaceID: "space", EnvironmentID: "environment", EntryID: "entry",
					})
					require.NoError(t, err)

					_, ok = deleteResponse.(*cm.NoContent)
					require.True(t, ok)

					adapter.reset()
				},
				Config: "# intentionally empty\n",
			},
		},
	})

	requests := adapter.snapshot()
	requireUnconditionalEntryDestroy(t, requests)
}

func TestAccEntryResourceDestroyAlreadyUnpublishedContinuesToDelete(t *testing.T) {
	t.Parallel()

	fixture := newEntryAcceptanceFixture(t)
	server, recorder := fixture.server, fixture.recorder
	adapter := &entryDestroyTestAdapter{delegate: server, errorSink: fixture.errorSink}
	recorder.delegate = adapter

	ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{
			Plan: resource.PlanOptions{NoRefresh: true},
		},
		Steps: []resource.TestStep{
			{Config: managedEntryConfig("one")},
			{
				PreConfig: func() {
					response, err := server.Handler().UnpublishEntry(t.Context(), cm.UnpublishEntryParams{
						SpaceID: "space", EnvironmentID: "environment", EntryID: "entry",
					})
					require.NoError(t, err)

					_, ok := response.(*cm.Entry)
					require.True(t, ok)

					adapter.reset()
				},
				Config: "# intentionally empty\n",
			},
		},
	})

	requests := adapter.snapshot()
	requireUnconditionalEntryDestroy(t, requests)
	requireEntryAbsent(t, server)
}

func TestAccEntryResourceConcurrentDraftAfterUnpublishIsStillDeleted(t *testing.T) {
	t.Parallel()

	fixture := newEntryAcceptanceFixture(t)
	server, recorder := fixture.server, fixture.recorder
	adapter := &entryDestroyTestAdapter{delegate: server, errorSink: fixture.errorSink}
	recorder.delegate = adapter
	adapter.after = func() error {
		entry, err := getEntryFromTestServer(t.Context(), server)
		if err != nil {
			return err
		}

		response, err := server.Handler().PutEntry(t.Context(), &cm.EntryRequest{
			Fields: entry.Fields, Metadata: entry.Metadata,
		}, cm.PutEntryParams{
			SpaceID:                "space",
			EnvironmentID:          "environment",
			EntryID:                "entry",
			XContentfulContentType: cm.NewOptString("article"),
			XContentfulVersion:     cm.NewOptInt(entry.Sys.Version),
		})
		if err != nil {
			return fmt.Errorf("advance Entry after unpublish: %w", err)
		}

		if _, ok := response.(*cm.EntryStatusCode); !ok {
			return fmt.Errorf("%w after unpublish: %T", errUnexpectedEntryResponseType, response)
		}

		return nil
	}

	ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{
			Plan: resource.PlanOptions{NoRefresh: true},
		},
		Steps: []resource.TestStep{
			{Config: managedEntryConfig("one")},
			{
				PreConfig: func() {
					adapter.reset()
					adapter.advance.arm()
				},
				Config: "# intentionally empty\n",
			},
		},
	})

	requests := adapter.snapshot()
	requireUnconditionalEntryDestroy(t, requests)
	requireEntryAbsent(t, server)
}
