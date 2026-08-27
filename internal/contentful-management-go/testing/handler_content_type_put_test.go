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

const contentTypePutContentType = "application/vnd.contentful.management.v1+json"

type rawContentTypeObservation struct {
	Sys struct {
		Version          int  `json:"version"`
		PublishedVersion *int `json:"publishedVersion"`
	} `json:"sys"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	DisplayField string `json:"displayField"`
	Fields       []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"fields"`
}

func TestContentTypeSpecifiedIDPutVersionHeaderMatrix(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	testServer := httptest.NewServer(server)
	t.Cleanup(testServer.Close)

	createdWithoutVersion := putRawContentType(t, testServer, "create-no-version", "", `{"name":"Created without version","description":"canonical create","displayField":"title","fields":[{"id":"title","name":"Title","type":"Symbol"}]}`)
	require.Equal(t, http.StatusCreated, createdWithoutVersion.status)
	require.Equal(t, 1, createdWithoutVersion.contentType.Sys.Version)
	require.Nil(t, createdWithoutVersion.contentType.Sys.PublishedVersion)
	require.Equal(t, "Created without version", createdWithoutVersion.contentType.Name)

	createdWithVersion := putRawContentType(t, testServer, "create-with-version", "1", `{"name":"Created with version","description":"compatibility create","displayField":"title","fields":[{"id":"title","name":"Title","type":"Symbol"}]}`)
	require.Equal(t, http.StatusCreated, createdWithVersion.status)
	require.Equal(t, 1, createdWithVersion.contentType.Sys.Version)
	require.Nil(t, createdWithVersion.contentType.Sys.PublishedVersion)
	require.Equal(t, "Created with version", createdWithVersion.contentType.Name)

	beforeCollision := getRawContentType(t, testServer)
	collision := putRawContentType(t, testServer, "create-no-version", "", `{"name":"Must not replace","description":"collision","displayField":"replacement","fields":[]}`)
	require.Equal(t, http.StatusConflict, collision.status)
	require.JSONEq(t, string(beforeCollision.body), string(getRawContentType(t, testServer).body), "an absent version must reject an existing Content Type without mutation")

	exactUpdate := putRawContentType(t, testServer, "create-no-version", "1", `{"name":"Updated exactly","description":"exact update","displayField":"slug","fields":[{"id":"slug","name":"Slug","type":"Symbol"}]}`)
	require.Equal(t, http.StatusOK, exactUpdate.status)
	require.Equal(t, 2, exactUpdate.contentType.Sys.Version)
	require.Nil(t, exactUpdate.contentType.Sys.PublishedVersion)
	require.Equal(t, "Updated exactly", exactUpdate.contentType.Name)
}

type rawContentTypeResponse struct {
	status      int
	body        []byte
	contentType rawContentTypeObservation
}

func putRawContentType(t *testing.T, server *httptest.Server, contentTypeID, version, body string) rawContentTypeResponse {
	t.Helper()

	request, err := http.NewRequestWithContext(
		t.Context(),
		http.MethodPut,
		server.URL+"/spaces/space/environments/environment/content_types/"+contentTypeID,
		bytes.NewBufferString(body),
	)
	require.NoError(t, err)
	request.Header.Set("Authorization", "Bearer "+cmt.ValidAccessToken)
	request.Header.Set("Content-Type", contentTypePutContentType)

	if version != "" {
		request.Header.Set("X-Contentful-Version", version)
	}

	status, responseBody := executeRawContentTypeRequest(t, server, request)

	result := rawContentTypeResponse{status: status, body: responseBody}
	if status >= http.StatusOK && status < http.StatusMultipleChoices {
		require.NoError(t, json.Unmarshal(responseBody, &result.contentType))
	}

	return result
}

func getRawContentType(t *testing.T, server *httptest.Server) rawContentTypeResponse {
	t.Helper()

	request, err := http.NewRequestWithContext(
		t.Context(),
		http.MethodGet,
		server.URL+"/spaces/space/environments/environment/content_types/create-no-version",
		nil,
	)
	require.NoError(t, err)
	request.Header.Set("Authorization", "Bearer "+cmt.ValidAccessToken)

	status, responseBody := executeRawContentTypeRequest(t, server, request)
	require.Equal(t, http.StatusOK, status)

	var contentType rawContentTypeObservation
	require.NoError(t, json.Unmarshal(responseBody, &contentType))

	return rawContentTypeResponse{status: status, body: responseBody, contentType: contentType}
}

func executeRawContentTypeRequest(t *testing.T, server *httptest.Server, request *http.Request) (int, []byte) {
	t.Helper()

	response, err := server.Client().Do(request)
	require.NoError(t, err)

	responseBody, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())

	return response.StatusCode, responseBody
}
