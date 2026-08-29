package cmtesting_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublishEntryRequiresExactVersion(t *testing.T) {
	t.Parallel()

	handler := cmt.NewHandler()
	handler.RegisterSpaceEnvironment("space", "environment", "ready")

	createdResponse, err := handler.PutEntry(context.Background(), &cm.EntryRequest{}, cm.PutEntryParams{
		SpaceID:                "space",
		EnvironmentID:          "environment",
		EntryID:                "entry",
		XContentfulContentType: cm.NewOptString("article"),
	})
	require.NoError(t, err)

	createdStatus, ok := createdResponse.(*cm.EntryStatusCode)
	require.True(t, ok)
	assert.Equal(t, http.StatusCreated, createdStatus.StatusCode)
	assert.Equal(t, 1, createdStatus.Response.Sys.Version)

	staleResponse, err := handler.PublishEntry(context.Background(), cm.PublishEntryParams{
		SpaceID:            "space",
		EnvironmentID:      "environment",
		EntryID:            "entry",
		XContentfulVersion: 0,
	})
	require.NoError(t, err)
	requireContentfulConflictWithNonemptyMessage(t, staleResponse, cm.ErrorSysIDVersionMismatch)

	publishedResponse, err := handler.PublishEntry(context.Background(), cm.PublishEntryParams{
		SpaceID:            "space",
		EnvironmentID:      "environment",
		EntryID:            "entry",
		XContentfulVersion: createdStatus.Response.Sys.Version,
	})
	require.NoError(t, err)

	publishedStatus, ok := publishedResponse.(*cm.EntryStatusCode)
	require.True(t, ok)
	assert.Equal(t, http.StatusOK, publishedStatus.StatusCode)
	assert.Equal(t, 2, publishedStatus.Response.Sys.Version)
	assert.Equal(t, 1, publishedStatus.Response.Sys.PublishedVersion.Or(0))
}

func TestUnpublishEntryAdvancesVersionAndReturnsEntry(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")
	handler := server.Handler()

	createdResponse, err := handler.PutEntry(t.Context(), &cm.EntryRequest{}, cm.PutEntryParams{
		SpaceID:                "space",
		EnvironmentID:          "environment",
		EntryID:                "entry",
		XContentfulContentType: cm.NewOptString("article"),
	})
	require.NoError(t, err)
	created := requireEntryStatusCode(t, createdResponse)

	publishedResponse, err := handler.PublishEntry(t.Context(), cm.PublishEntryParams{
		SpaceID:            "space",
		EnvironmentID:      "environment",
		EntryID:            "entry",
		XContentfulVersion: created.Response.Sys.Version,
	})
	require.NoError(t, err)
	published := requireEntryStatusCode(t, publishedResponse)

	draftResponse, err := handler.PutEntry(t.Context(), &cm.EntryRequest{}, cm.PutEntryParams{
		SpaceID:            "space",
		EnvironmentID:      "environment",
		EntryID:            "entry",
		XContentfulVersion: cm.NewOptInt(published.Response.Sys.Version),
	})
	require.NoError(t, err)
	draft := requireEntryStatusCode(t, draftResponse)

	testServer := httptest.NewServer(server)
	t.Cleanup(testServer.Close)

	request, err := http.NewRequestWithContext(
		t.Context(),
		http.MethodDelete,
		testServer.URL+"/spaces/space/environments/environment/entries/entry/published",
		nil,
	)
	require.NoError(t, err)
	request.Header.Set("Authorization", "Bearer "+cmt.ValidAccessToken)

	unpublishResponse, err := testServer.Client().Do(request)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, unpublishResponse.Body.Close()) })
	require.Equal(t, http.StatusOK, unpublishResponse.StatusCode)

	var unpublished cm.Entry
	require.NoError(t, json.NewDecoder(unpublishResponse.Body).Decode(&unpublished))
	assert.Equal(t, draft.Response.Sys.Version+1, unpublished.Sys.Version)
	assert.False(t, unpublished.Sys.PublishedVersion.IsSet())

	getResponse, err := handler.GetEntry(t.Context(), cm.GetEntryParams{
		SpaceID: "space", EnvironmentID: "environment", EntryID: "entry",
	})
	require.NoError(t, err)

	got, ok := getResponse.(*cm.Entry)
	require.True(t, ok)
	assert.Equal(t, unpublished.Sys.Version, got.Sys.Version)
	assert.False(t, got.Sys.PublishedVersion.IsSet())
}

func TestDeletePublishedEntryReturnsBadRequestAndPreservesEntry(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")
	handler := server.Handler()

	createdResponse, err := handler.PutEntry(t.Context(), &cm.EntryRequest{}, cm.PutEntryParams{
		SpaceID:                "space",
		EnvironmentID:          "environment",
		EntryID:                "entry",
		XContentfulContentType: cm.NewOptString("article"),
	})
	require.NoError(t, err)
	created := requireEntryStatusCode(t, createdResponse)

	_, err = handler.PublishEntry(t.Context(), cm.PublishEntryParams{
		SpaceID:            "space",
		EnvironmentID:      "environment",
		EntryID:            "entry",
		XContentfulVersion: created.Response.Sys.Version,
	})
	require.NoError(t, err)

	testServer := httptest.NewServer(server)
	t.Cleanup(testServer.Close)
	client, err := cm.NewClient(
		testServer.URL,
		cm.NewAccessTokenSecuritySource(cmt.ValidAccessToken),
		cm.WithClient(testServer.Client()),
	)
	require.NoError(t, err)

	deleteResponse, err := client.DeleteEntry(t.Context(), cm.DeleteEntryParams{
		SpaceID: "space", EnvironmentID: "environment", EntryID: "entry",
	})
	require.NoError(t, err)
	requireContentfulError(t, deleteResponse, http.StatusBadRequest, "BadRequest", "Cannot delete published")

	getResponse, err := client.GetEntry(t.Context(), cm.GetEntryParams{
		SpaceID: "space", EnvironmentID: "environment", EntryID: "entry",
	})
	require.NoError(t, err)

	_, present := getResponse.(*cm.Entry)
	require.True(t, present, "a rejected published Entry deletion must preserve the Entry")
}

func TestUnpublishUnpublishedEntryReturnsBadRequestWithoutMutation(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")
	handler := server.Handler()

	createdResponse, err := handler.PutEntry(t.Context(), &cm.EntryRequest{}, cm.PutEntryParams{
		SpaceID:                "space",
		EnvironmentID:          "environment",
		EntryID:                "entry",
		XContentfulContentType: cm.NewOptString("article"),
	})
	require.NoError(t, err)
	created := requireEntryStatusCode(t, createdResponse)

	testServer := httptest.NewServer(server)
	t.Cleanup(testServer.Close)
	client, err := cm.NewClient(
		testServer.URL,
		cm.NewAccessTokenSecuritySource(cmt.ValidAccessToken),
		cm.WithClient(testServer.Client()),
	)
	require.NoError(t, err)

	unpublishResponse, err := client.UnpublishEntry(t.Context(), cm.UnpublishEntryParams{
		SpaceID: "space", EnvironmentID: "environment", EntryID: "entry",
	})
	require.NoError(t, err)
	requireContentfulError(t, unpublishResponse, http.StatusBadRequest, "BadRequest", "Not published")

	getResponse, err := client.GetEntry(t.Context(), cm.GetEntryParams{
		SpaceID: "space", EnvironmentID: "environment", EntryID: "entry",
	})
	require.NoError(t, err)

	entry, present := getResponse.(*cm.Entry)
	require.True(t, present)
	require.Equal(t, created.Response.Sys.Version, entry.Sys.Version, "a rejected unpublish must not advance the Entry version")
}

func TestMissingEntryDestroyResponsesMatchCMA(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	testServer := httptest.NewServer(server)
	t.Cleanup(testServer.Close)
	client, err := cm.NewClient(
		testServer.URL,
		cm.NewAccessTokenSecuritySource(cmt.ValidAccessToken),
		cm.WithClient(testServer.Client()),
	)
	require.NoError(t, err)

	deleteResponse, err := client.DeleteEntry(t.Context(), cm.DeleteEntryParams{
		SpaceID: "space", EnvironmentID: "environment", EntryID: "missing",
	})
	require.NoError(t, err)
	requireContentfulError(t, deleteResponse, http.StatusNotFound, "NotFound", "The resource could not be found.")

	unpublishResponse, err := client.UnpublishEntry(t.Context(), cm.UnpublishEntryParams{
		SpaceID: "space", EnvironmentID: "environment", EntryID: "missing",
	})
	require.NoError(t, err)
	requireContentfulError(t, unpublishResponse, http.StatusNotFound, "NotFound", "The resource could not be found.")
}

func TestEntryUnpublishIgnoresVersionHeader(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		version *string
		ifMatch *string
	}{
		"omitted":       {},
		"stale version": {version: new("2")},
		"stale ETag":    {ifMatch: new(`"stale-etag"`)},
		"zero":          {version: new("0")},
	}

	for name, headers := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer()
			require.NoError(t, err)
			server.RegisterSpaceEnvironment("space", "environment")
			handler := server.Handler()

			createdResponse, err := handler.PutEntry(t.Context(), &cm.EntryRequest{}, cm.PutEntryParams{
				SpaceID:                "space",
				EnvironmentID:          "environment",
				EntryID:                "entry",
				XContentfulContentType: cm.NewOptString("article"),
			})
			require.NoError(t, err)
			created := requireEntryStatusCode(t, createdResponse)

			publishedResponse, err := handler.PublishEntry(t.Context(), cm.PublishEntryParams{
				SpaceID:            "space",
				EnvironmentID:      "environment",
				EntryID:            "entry",
				XContentfulVersion: created.Response.Sys.Version,
			})
			require.NoError(t, err)
			published := requireEntryStatusCode(t, publishedResponse)

			draftResponse, err := handler.PutEntry(t.Context(), &cm.EntryRequest{}, cm.PutEntryParams{
				SpaceID:                "space",
				EnvironmentID:          "environment",
				EntryID:                "entry",
				XContentfulVersion:     cm.NewOptInt(published.Response.Sys.Version),
				XContentfulContentType: cm.NewOptString("article"),
			})
			require.NoError(t, err)
			draft := requireEntryStatusCode(t, draftResponse)
			require.Equal(t, 3, draft.Response.Sys.Version)

			testServer := httptest.NewServer(server)
			t.Cleanup(testServer.Close)

			unpublished := deleteRawEntry(t, testServer, "/published", headers.version, headers.ifMatch)
			require.Equal(t, http.StatusOK, unpublished.status)
			require.Equal(t, 4, unpublished.entry.Sys.Version)
		})
	}
}

func TestEntryDeleteIgnoresVersionHeader(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		version *string
		ifMatch *string
	}{
		"omitted":       {},
		"stale version": {version: new("1")},
		"stale ETag":    {ifMatch: new(`"stale-etag"`)},
		"zero":          {version: new("0")},
	}

	for name, headers := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer()
			require.NoError(t, err)
			server.RegisterSpaceEnvironment("space", "environment")
			handler := server.Handler()

			createdResponse, err := handler.PutEntry(t.Context(), &cm.EntryRequest{}, cm.PutEntryParams{
				SpaceID:                "space",
				EnvironmentID:          "environment",
				EntryID:                "entry",
				XContentfulContentType: cm.NewOptString("article"),
			})
			require.NoError(t, err)
			created := requireEntryStatusCode(t, createdResponse)

			updatedResponse, err := handler.PutEntry(t.Context(), &cm.EntryRequest{}, cm.PutEntryParams{
				SpaceID:                "space",
				EnvironmentID:          "environment",
				EntryID:                "entry",
				XContentfulVersion:     cm.NewOptInt(created.Response.Sys.Version),
				XContentfulContentType: cm.NewOptString("article"),
			})
			require.NoError(t, err)
			updated := requireEntryStatusCode(t, updatedResponse)
			require.Equal(t, 2, updated.Response.Sys.Version)

			testServer := httptest.NewServer(server)
			t.Cleanup(testServer.Close)

			deleted := deleteRawEntry(t, testServer, "", headers.version, headers.ifMatch)
			require.Equal(t, http.StatusNoContent, deleted.status)

			getResponse, err := handler.GetEntry(t.Context(), cm.GetEntryParams{
				SpaceID: "space", EnvironmentID: "environment", EntryID: "entry",
			})
			require.NoError(t, err)

			status, ok := getResponse.(cm.StatusCodeResponse)
			require.True(t, ok)
			require.Equal(t, http.StatusNotFound, status.GetStatusCode())
		})
	}
}

func deleteRawEntry(t *testing.T, server *httptest.Server, suffix string, version, ifMatch *string) rawEntryResponse {
	t.Helper()

	request, err := http.NewRequestWithContext(
		t.Context(),
		http.MethodDelete,
		server.URL+"/spaces/space/environments/environment/entries/entry"+suffix,
		nil,
	)
	require.NoError(t, err)
	request.Header.Set("Authorization", "Bearer "+cmt.ValidAccessToken)

	if version != nil {
		request.Header.Set("X-Contentful-Version", *version)
	}

	if ifMatch != nil {
		request.Header.Set("If-Match", *ifMatch)
	}

	response, err := server.Client().Do(request)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, response.Body.Close()) })

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	result := rawEntryResponse{status: response.StatusCode, body: body}
	if len(body) > 0 && response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices {
		require.NoError(t, json.Unmarshal(body, &result.entry))
	}

	return result
}

func requireEntryStatusCode(t *testing.T, response any) *cm.EntryStatusCode {
	t.Helper()

	status, ok := response.(*cm.EntryStatusCode)
	require.True(t, ok)

	return status
}
