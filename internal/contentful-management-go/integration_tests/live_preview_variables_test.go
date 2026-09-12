package integration_tests_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const livePreviewVariablesResponse = `{"sys":{"space":{"sys":{"type":"Link","linkType":"Space","id":"space"}},"environment":{"sys":{"type":"Link","linkType":"Environment","id":"alias"}},"version":8},"variables":{"global":"value","localized":{"en-US":null},"empty":{},"null":null}}`

func TestLivePreviewVariablesWireContract(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		method         string
		version        int
		status         int
		contentType    string
		response       string
		errorID        string
		service        bool
		serviceMessage string
		decodeError    bool
	}{
		"get ordinary JSON":                   {method: http.MethodGet, status: http.StatusOK, contentType: "application/json; charset=utf-8", response: livePreviewVariablesResponse},
		"create version zero":                 {method: http.MethodPut, version: 0, status: http.StatusOK, contentType: "application/json", response: livePreviewVariablesResponse},
		"update exact version":                {method: http.MethodPut, version: 7, status: http.StatusOK, contentType: "application/json", response: livePreviewVariablesResponse},
		"delete empty response":               {method: http.MethodDelete, status: http.StatusNoContent},
		"CMA not found":                       {method: http.MethodGet, status: http.StatusNotFound, contentType: "application/vnd.contentful.management.v1+json", response: `{"sys":{"type":"Error","id":"NotFound"},"message":"The resource could not be found."}`, errorID: "NotFound"},
		"version conflict":                    {method: http.MethodPut, version: 7, status: http.StatusConflict, contentType: "application/vnd.contentful.management.v1+json", response: `{"sys":{"type":"Error","id":"VersionMismatch"},"message":"The given version value is not the current one"}`, errorID: "VersionMismatch"},
		"validation without optional fields":  {method: http.MethodPut, status: http.StatusUnprocessableEntity, contentType: "application/json", response: `{"sys":{"type":"Error","id":"ValidationFailed"},"message":"Validation error","details":{"errors":[{"name":"required","details":"Required property"}]}}`, errorID: "ValidationFailed"},
		"enterprise feature unavailable":      {method: http.MethodGet, status: http.StatusForbidden, contentType: "application/vnd.contentful.management.v1+json", response: `{"statusCode":403,"error":"Forbidden","message":"previewLocalization is not enabled"}`, service: true, serviceMessage: "previewLocalization is not enabled"},
		"service error preserves HTTP status": {method: http.MethodGet, status: http.StatusNotFound, contentType: "application/vnd.contentful.management.v1+json", response: `{"statusCode":500,"error":"Not Found","message":"Not Found"}`, service: true, serviceMessage: "Not Found"},
		"missing variables":                   {method: http.MethodGet, status: http.StatusOK, contentType: "application/json", response: `{"sys":{"space":{"sys":{"type":"Link","linkType":"Space","id":"space"}},"environment":{"sys":{"type":"Link","linkType":"Environment","id":"alias"}},"version":8}}`, decodeError: true},
		"malformed response":                  {method: http.MethodGet, status: http.StatusOK, contentType: "application/json", response: `{`, decodeError: true},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var requests atomic.Int64

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				assert.Equal(t, test.method, r.Method)
				assert.Equal(t, "/spaces/space/environments/alias/live_preview/variables", r.URL.Path)
				assert.Empty(t, r.URL.RawQuery)
				assert.Empty(t, r.Header.Get("If-Match"))

				body, err := io.ReadAll(r.Body)
				assert.NoError(t, err)

				if r.Method == http.MethodPut {
					assert.Equal(t, "application/vnd.contentful.management.v1+json", r.Header.Get("Content-Type"))

					if test.version == 0 {
						assert.Equal(t, "0", r.Header.Get("X-Contentful-Version"))
					} else {
						assert.Equal(t, "7", r.Header.Get("X-Contentful-Version"))
					}

					assert.JSONEq(t, `{"variables":{"global":"value","localized":{"en-US":null},"empty":{},"null":null}}`, string(body))
				} else {
					assert.Empty(t, r.Header.Get("X-Contentful-Version"))
					assert.Empty(t, body)
				}

				if test.contentType != "" {
					w.Header().Set("Content-Type", test.contentType)
				}

				w.WriteHeader(test.status)
				_, writeErr := io.WriteString(w, test.response)
				assert.NoError(t, writeErr)
			}))
			t.Cleanup(server.Close)

			client := testContentfulManagementClient(t, server.URL, cmt.ValidAccessToken)

			var (
				response any
				err      error
			)

			switch test.method {
			case http.MethodGet:
				response, err = client.GetLivePreviewVariables(t.Context(), cm.GetLivePreviewVariablesParams{SpaceID: "space", EnvironmentID: "alias"})
			case http.MethodPut:
				response, err = client.PutLivePreviewVariables(t.Context(), &cm.LivePreviewVariablesData{Variables: []byte(`{"global":"value","localized":{"en-US":null},"empty":{},"null":null}`)}, cm.PutLivePreviewVariablesParams{SpaceID: "space", EnvironmentID: "alias", XContentfulVersion: test.version})
			case http.MethodDelete:
				response, err = client.DeleteLivePreviewVariables(t.Context(), cm.DeleteLivePreviewVariablesParams{SpaceID: "space", EnvironmentID: "alias"})
			}

			require.EqualValues(t, 1, requests.Load())

			if test.decodeError {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)

			switch test.status {
			case http.StatusOK:
				variables, ok := response.(*cm.LivePreviewVariables)
				require.True(t, ok)
				require.Equal(t, 8, variables.Sys.Version)
				require.Equal(t, "alias", variables.Sys.Environment.Sys.ID)
				require.JSONEq(t, `{"global":"value","localized":{"en-US":null},"empty":{},"null":null}`, string(variables.Variables))
			case http.StatusNoContent:
				require.IsType(t, &cm.NoContent{}, response)
			default:
				failure, ok := response.(*cm.LivePreviewVariablesErrorStatusCode)
				require.True(t, ok)
				require.Equal(t, test.status, failure.GetStatusCode())
				cmaError, isCMAError := failure.GetError()
				require.Equal(t, !test.service, isCMAError)

				if isCMAError {
					require.Equal(t, test.errorID, cmaError.Sys.ID)
				} else {
					serviceError, isServiceError := failure.Response.GetLivePreviewVariablesServiceError()
					require.True(t, isServiceError)

					require.Equal(t, test.serviceMessage, serviceError.Message)
				}
			}
		})
	}
}

func TestLivePreviewVariablesMockLifecycle(t *testing.T) {
	t.Parallel()

	server, testServer := testContentfulManagementHTTPTestServer(t, cmt.WithRateLimitPerSecond(100))
	server.RegisterSpaceEnvironment("space", "environment")

	client := testContentfulManagementClient(t, testServer.URL, cmt.ValidAccessToken)
	getParams := cm.GetLivePreviewVariablesParams{SpaceID: "space", EnvironmentID: "environment"}
	putParams := cm.PutLivePreviewVariablesParams{SpaceID: "space", EnvironmentID: "environment"}
	deleteParams := cm.DeleteLivePreviewVariablesParams{SpaceID: "space", EnvironmentID: "environment"}

	response, err := client.GetLivePreviewVariables(t.Context(), getParams)
	require.NoError(t, err)
	require.IsType(t, &cm.LivePreviewVariablesErrorStatusCode{}, response)

	created, err := client.PutLivePreviewVariables(t.Context(), &cm.LivePreviewVariablesData{Variables: []byte(`{"remove":"value","localized":{"en-US":"value"}}`)}, putParams)
	require.NoError(t, err)

	createdDocument, ok := created.(*cm.LivePreviewVariables)
	require.True(t, ok)
	require.Equal(t, 1, createdDocument.Sys.Version)

	conflict, err := client.PutLivePreviewVariables(t.Context(), &cm.LivePreviewVariablesData{Variables: []byte(`{}`)}, putParams)
	require.NoError(t, err)

	conflictDocument, ok := conflict.(*cm.LivePreviewVariablesErrorStatusCode)
	require.True(t, ok)
	require.Equal(t, http.StatusConflict, conflictDocument.StatusCode)

	putParams.XContentfulVersion = 1
	updated, err := client.PutLivePreviewVariables(t.Context(), &cm.LivePreviewVariablesData{Variables: []byte(`{"localized":{},"null":null,"empty":""}`)}, putParams)
	require.NoError(t, err)

	updatedDocument, ok := updated.(*cm.LivePreviewVariables)
	require.True(t, ok)
	require.Equal(t, 2, updatedDocument.Sys.Version)
	read, err := client.GetLivePreviewVariables(t.Context(), getParams)
	require.NoError(t, err)

	readDocument, ok := read.(*cm.LivePreviewVariables)
	require.True(t, ok)
	require.JSONEq(t, `{"localized":{},"null":null,"empty":""}`, string(readDocument.Variables))

	putParams.XContentfulVersion = 2
	empty, err := client.PutLivePreviewVariables(t.Context(), &cm.LivePreviewVariablesData{Variables: []byte(`{}`)}, putParams)
	require.NoError(t, err)

	emptyDocument, ok := empty.(*cm.LivePreviewVariables)
	require.True(t, ok)
	require.Equal(t, 3, emptyDocument.Sys.Version)
	read, err = client.GetLivePreviewVariables(t.Context(), getParams)
	require.NoError(t, err)

	readDocument, ok = read.(*cm.LivePreviewVariables)
	require.True(t, ok)
	require.JSONEq(t, `{}`, string(readDocument.Variables))

	for range 2 {
		deleted, deleteErr := client.DeleteLivePreviewVariables(t.Context(), deleteParams)
		require.NoError(t, deleteErr)
		require.IsType(t, &cm.NoContent{}, deleted)
	}
	// An absent document accepts the old version and restarts at version 1.
	recreated, err := client.PutLivePreviewVariables(t.Context(), &cm.LivePreviewVariablesData{Variables: []byte(`{}`)}, putParams)
	require.NoError(t, err)

	recreatedDocument, ok := recreated.(*cm.LivePreviewVariables)
	require.True(t, ok)
	require.Equal(t, 1, recreatedDocument.Sys.Version)

	// Version 1 from the previous lifetime also authorizes a write to the recreated document.
	putParams.XContentfulVersion = 1
	reused, err := client.PutLivePreviewVariables(t.Context(), &cm.LivePreviewVariablesData{Variables: []byte(`{"reused":"version"}`)}, putParams)
	require.NoError(t, err)

	reusedDocument, ok := reused.(*cm.LivePreviewVariables)
	require.True(t, ok)
	require.Equal(t, 2, reusedDocument.Sys.Version)
	require.JSONEq(t, `{"reused":"version"}`, string(reusedDocument.Variables))

	_, err = client.DeleteEnvironment(t.Context(), cm.DeleteEnvironmentParams{SpaceID: "space", EnvironmentID: "environment"})
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	read, err = client.GetLivePreviewVariables(t.Context(), getParams)
	require.NoError(t, err)
	require.IsType(t, &cm.LivePreviewVariablesErrorStatusCode{}, read)
}
