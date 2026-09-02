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

func TestEntrySpecifiedIDPutHeaderMatrix(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	testServer := httptest.NewServer(server)
	t.Cleanup(testServer.Close)

	for name, test := range map[string]struct {
		entryID     string
		contentType string
		version     string
		wantStatus  int
	}{
		"content type without version": {entryID: "content-type-no-version", contentType: "article", wantStatus: http.StatusCreated},
		"content type with version":    {entryID: "content-type-version", contentType: "article", version: "7", wantStatus: http.StatusCreated},
		"no content type or version":   {entryID: "no-content-type-no-version", wantStatus: http.StatusBadRequest},
		"version without content type": {entryID: "version-no-content-type", version: "7", wantStatus: http.StatusBadRequest},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result := putRawEntry(
				t, testServer, test.entryID, test.contentType, test.version,
				`{"fields":{"marker":{"en-US":"absent target"}}}`,
			)
			require.Equal(t, test.wantStatus, result.status)

			stored := getRawEntry(t, testServer, test.entryID)
			if test.wantStatus == http.StatusCreated {
				require.Equal(t, 1, result.entry.Sys.Version)
				require.JSONEq(t, `{"en-US":"absent target"}`, string(result.entry.Fields["marker"]))
				require.Equal(t, http.StatusOK, stored.status)
				require.Equal(t, 1, stored.entry.Sys.Version)
				require.JSONEq(t, `{"en-US":"absent target"}`, string(stored.entry.Fields["marker"]))

				return
			}

			require.JSONEq(t, `{
				"sys":{"type":"Error","id":"BadRequest"},
				"message":"You should provide a content type in X-Contentful-Content-Type request header."
			}`, string(result.body))
			require.Equal(t, http.StatusNotFound, stored.status, "a rejected absent-target PUT must not store an Entry")
		})
	}

	for name, test := range map[string]struct {
		entryID     string
		contentType string
	}{
		"with content type":    {entryID: "existing-with-content-type", contentType: "article"},
		"without content type": {entryID: "existing-without-content-type"},
	} {
		t.Run("existing exact update "+name, func(t *testing.T) {
			created := putRawEntry(
				t, testServer, test.entryID, "article", "",
				`{"fields":{"marker":{"en-US":"created"}}}`,
			)
			require.Equal(t, http.StatusCreated, created.status)
			require.Equal(t, 1, created.entry.Sys.Version)

			missingVersion := putRawEntry(
				t, testServer, test.entryID, test.contentType, "",
				`{"fields":{"marker":{"en-US":"must not replace"}}}`,
			)
			require.Equal(t, http.StatusConflict, missingVersion.status)
			require.JSONEq(t, `{
				"sys":{"type":"Error","id":"VersionMismatch"},
				"message":"Version mismatch"
			}`, string(missingVersion.body))

			storedAfterMissingVersion := getRawEntry(t, testServer, test.entryID)
			require.Equal(t, http.StatusOK, storedAfterMissingVersion.status)
			require.Equal(t, 1, storedAfterMissingVersion.entry.Sys.Version, "an absent version must not mutate an existing Entry")
			require.JSONEq(t, `{"en-US":"created"}`, string(storedAfterMissingVersion.entry.Fields["marker"]))

			result := putRawEntry(
				t, testServer, test.entryID, test.contentType, "1",
				`{"fields":{"marker":{"en-US":"updated exactly"}}}`,
			)
			require.Equal(t, http.StatusOK, result.status)
			require.Equal(t, 2, result.entry.Sys.Version)
			require.JSONEq(t, `{"en-US":"updated exactly"}`, string(result.entry.Fields["marker"]))

			stored := getRawEntry(t, testServer, test.entryID)
			require.Equal(t, http.StatusOK, stored.status)
			require.Equal(t, 2, stored.entry.Sys.Version)
			require.JSONEq(t, `{"en-US":"updated exactly"}`, string(stored.entry.Fields["marker"]))

			stale := putRawEntry(
				t, testServer, test.entryID, test.contentType, "1",
				`{"fields":{"marker":{"en-US":"must not replace"}}}`,
			)
			require.Equal(t, http.StatusConflict, stale.status)
			require.JSONEq(t, `{
				"sys":{"type":"Error","id":"VersionMismatch"},
				"message":"Version mismatch"
			}`, string(stale.body))

			storedAfterStale := getRawEntry(t, testServer, test.entryID)
			require.Equal(t, http.StatusOK, storedAfterStale.status)
			require.Equal(t, 2, storedAfterStale.entry.Sys.Version, "a stale version must not mutate an Entry")
			require.JSONEq(t, `{"en-US":"updated exactly"}`, string(storedAfterStale.entry.Fields["marker"]))
		})
	}
}

type rawEntryResponse struct {
	status int
	body   []byte
	entry  rawEntryObservation
}

func putRawEntry(t *testing.T, server *httptest.Server, entryID, contentType, version, body string) rawEntryResponse {
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

	if contentType != "" {
		request.Header.Set("X-Contentful-Content-Type", contentType)
	}

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

func getRawEntry(t *testing.T, server *httptest.Server, entryID string) rawEntryResponse {
	t.Helper()

	request, err := http.NewRequestWithContext(
		t.Context(),
		http.MethodGet,
		server.URL+"/spaces/space/environments/environment/entries/"+entryID,
		nil,
	)
	require.NoError(t, err)
	request.Header.Set("Authorization", "Bearer "+cmt.ValidAccessToken)

	response, err := server.Client().Do(request)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, response.Body.Close()) })

	responseBody, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	result := rawEntryResponse{status: response.StatusCode, body: responseBody}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return result
	}

	var entry rawEntryObservation
	require.NoError(t, json.Unmarshal(responseBody, &entry))

	result.entry = entry

	return result
}
