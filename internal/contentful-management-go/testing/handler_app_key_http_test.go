package cmtesting_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContentfulManagementServerCreatesSuppliedAppKey(t *testing.T) {
	t.Parallel()

	request := appKeyRequest(t)
	status, responseBody := postAppKeyRequest(t, request)

	assert.Equal(t, http.StatusCreated, status)

	var response cm.AppKey
	require.NoError(t, json.Unmarshal(responseBody, &response))
	assert.Equal(t, request.Jwk.Kid, response.Sys.ID)
	assert.Equal(t, request.Jwk, response.Jwk)
}

func TestContentfulManagementServerRejectsInvalidAppKeyMaterial(t *testing.T) {
	t.Parallel()

	tests := map[string]func(*cm.AppKeyRequestData){
		"encoding": func(request *cm.AppKeyRequestData) {
			request.Jwk.X5c[0] = strings.Repeat("!", len(request.Jwk.X5c[0]))
		},
		"fingerprint": func(request *cm.AppKeyRequestData) {
			request.Jwk.X5t = strings.Repeat("x", len(request.Jwk.X5t))
		},
	}

	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			request := appKeyRequest(t)
			mutate(request)

			status, responseBody := postAppKeyRequest(t, request)
			assert.Equal(t, http.StatusUnprocessableEntity, status)

			var response cm.Error
			require.NoError(t, json.Unmarshal(responseBody, &response))
			assert.Equal(t, "ValidationFailed", response.Sys.ID)
			assert.Equal(t, "Validation error", response.Message.Or(""))
		})
	}
}

func postAppKeyRequest(t *testing.T, payload *cm.AppKeyRequestData) (int, []byte) {
	t.Helper()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)
	server.SetAppDefinition("organization", "app-definition", cm.AppDefinitionData{Name: "App"})

	testServer := httptest.NewServer(server)
	t.Cleanup(testServer.Close)

	requestBody, err := json.Marshal(payload)
	require.NoError(t, err)

	request, err := http.NewRequestWithContext(
		t.Context(),
		http.MethodPost,
		testServer.URL+"/organizations/organization/app_definitions/app-definition/keys",
		bytes.NewReader(requestBody),
	)
	require.NoError(t, err)
	request.Header.Set("Authorization", "Bearer "+cmt.ValidAccessToken)
	request.Header.Set("Content-Type", "application/vnd.contentful.management.v1+json")

	response, err := testServer.Client().Do(request)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, response.Body.Close())
	})

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	return response.StatusCode, body
}
