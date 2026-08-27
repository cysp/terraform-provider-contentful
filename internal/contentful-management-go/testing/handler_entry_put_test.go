package cmtesting_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/require"
)

const entryPutContentType = "application/vnd.contentful.management.v1+json"

type rawEntryObservation struct {
	Sys struct {
		Version int `json:"version"`
	} `json:"sys"`
	Fields map[string]json.RawMessage `json:"fields"`
}

func TestEntrySpecifiedIDPutVersionHeaderMatrix(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	testServer := httptest.NewServer(server)
	t.Cleanup(testServer.Close)

	createdWithoutVersion := putRawEntry(t, testServer, "create-no-version", "", `{"fields":{"marker":{"en-US":"created without version"}}}`)
	require.Equal(t, http.StatusCreated, createdWithoutVersion.status)
	require.Equal(t, 1, createdWithoutVersion.entry.Sys.Version)
	require.JSONEq(t, `{"en-US":"created without version"}`, string(createdWithoutVersion.entry.Fields["marker"]))

	createdWithVersion := putRawEntry(t, testServer, "create-with-version", "1", `{"fields":{"marker":{"en-US":"created with version"}}}`)
	require.Equal(t, http.StatusCreated, createdWithVersion.status)
	require.Equal(t, 1, createdWithVersion.entry.Sys.Version)
	require.JSONEq(t, `{"en-US":"created with version"}`, string(createdWithVersion.entry.Fields["marker"]))

	beforeCollision := getRawEntry(t, testServer)
	collision := putRawEntry(t, testServer, "create-no-version", "", `{"fields":{"marker":{"en-US":"must not replace"}}}`)
	require.JSONEq(t, string(beforeCollision.body), string(getRawEntry(t, testServer).body), "an absent version must reject an existing Entry without mutation")
	require.Equal(t, http.StatusConflict, collision.status)

	exactUpdate := putRawEntry(t, testServer, "create-no-version", "1", `{"fields":{"marker":{"en-US":"updated exactly"}}}`)
	require.Equal(t, http.StatusOK, exactUpdate.status)
	require.Equal(t, 2, exactUpdate.entry.Sys.Version)
	require.JSONEq(t, `{"en-US":"updated exactly"}`, string(exactUpdate.entry.Fields["marker"]))

	beforeStale := getRawEntry(t, testServer)
	stale := putRawEntry(t, testServer, "create-no-version", "1", `{"fields":{"marker":{"en-US":"stale replacement"}}}`)
	require.Equal(t, http.StatusConflict, stale.status)
	require.JSONEq(t, string(beforeStale.body), string(getRawEntry(t, testServer).body), "a stale version must not mutate an Entry")
}

type rawEntryResponse struct {
	status int
	body   []byte
	entry  rawEntryObservation
}

func putRawEntry(t *testing.T, server *httptest.Server, entryID, version, body string) rawEntryResponse {
	t.Helper()

	request, err := http.NewRequestWithContext(
		t.Context(),
		http.MethodPut,
		server.URL+"/spaces/space/environments/environment/entries/"+entryID,
		bytes.NewBufferString(body),
	)
	require.NoError(t, err)
	request.Header.Set("Authorization", "Bearer "+cmt.ValidAccessToken)
	request.Header.Set("Content-Type", entryPutContentType)
	request.Header.Set("X-Contentful-Content-Type", "article")

	if version != "" {
		request.Header.Set("X-Contentful-Version", version)
	}

	response, err := server.Client().Do(request)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, response.Body.Close()) })

	responseBody, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	result := rawEntryResponse{status: response.StatusCode, body: responseBody}
	if response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices {
		require.NoError(t, json.Unmarshal(responseBody, &result.entry))
	}

	return result
}

func getRawEntry(t *testing.T, server *httptest.Server) rawEntryResponse {
	t.Helper()

	request, err := http.NewRequestWithContext(
		t.Context(),
		http.MethodGet,
		server.URL+"/spaces/space/environments/environment/entries/create-no-version",
		nil,
	)
	require.NoError(t, err)
	request.Header.Set("Authorization", "Bearer "+cmt.ValidAccessToken)

	response, err := server.Client().Do(request)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, response.Body.Close()) })
	require.Equal(t, http.StatusOK, response.StatusCode)

	responseBody, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	var entry rawEntryObservation
	require.NoError(t, json.Unmarshal(responseBody, &entry))

	return rawEntryResponse{status: response.StatusCode, body: responseBody, entry: entry}
}
