package cmtesting_test

import (
	"context"
	"encoding/json"
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

func requireEntryStatusCode(t *testing.T, response any) *cm.EntryStatusCode {
	t.Helper()

	status, ok := response.(*cm.EntryStatusCode)
	require.True(t, ok)

	return status
}
