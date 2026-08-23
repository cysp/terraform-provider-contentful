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

const appSigningSecretPath = "/organizations/organization/app_definitions/app-definition/signing_secret"

func TestContentfulManagementServerAppSigningSecretLifecycle(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)
	server.SetAppDefinition("organization", "app-definition", cm.AppDefinitionData{Name: "App"})

	testServer := httptest.NewServer(server)
	t.Cleanup(testServer.Close)

	initialValue := strings.Repeat("a", 55) + "Ab09+/=_-"
	replacementValue := strings.Repeat("b", 55) + "Zy87-/=_+"

	status, body := appSigningSecretHTTPRequest(t, testServer, http.MethodPut, initialValue)
	assert.Equal(t, http.StatusCreated, status)
	assertAppSigningSecretResponse(t, body, initialValue)

	status, body = appSigningSecretHTTPRequest(t, testServer, http.MethodPut, replacementValue)
	assert.Equal(t, http.StatusOK, status)
	assertAppSigningSecretResponse(t, body, replacementValue)

	status, body = appSigningSecretHTTPRequest(t, testServer, http.MethodGet, "")
	assert.Equal(t, http.StatusOK, status)
	assertAppSigningSecretResponse(t, body, replacementValue)

	status, body = appSigningSecretHTTPRequest(t, testServer, http.MethodDelete, "")
	assert.Equal(t, http.StatusNoContent, status)
	assert.Empty(t, body)
}

func TestContentfulManagementServerRejectsInvalidAppSigningSecretsWithoutEchoingThem(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"too short":         strings.Repeat("s", 63),
		"too long":          strings.Repeat("s", 65),
		"invalid character": strings.Repeat("s", 63) + "!",
	}

	for name, value := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer()
			require.NoError(t, err)
			server.SetAppDefinition("organization", "app-definition", cm.AppDefinitionData{Name: "App"})

			testServer := httptest.NewServer(server)
			t.Cleanup(testServer.Close)

			status, body := appSigningSecretHTTPRequest(t, testServer, http.MethodPut, value)
			assert.Equal(t, http.StatusBadRequest, status)
			assert.NotContains(t, string(body), value)
		})
	}
}

func appSigningSecretHTTPRequest(t *testing.T, server *httptest.Server, method, value string) (int, []byte) {
	t.Helper()

	var requestBody io.Reader

	if value != "" {
		payload, err := json.Marshal(cm.AppSigningSecretRequestData{Value: value})
		require.NoError(t, err)

		requestBody = bytes.NewReader(payload)
	}

	request, err := http.NewRequestWithContext(t.Context(), method, server.URL+appSigningSecretPath, requestBody)
	require.NoError(t, err)
	request.Header.Set("Authorization", "Bearer "+cmt.ValidAccessToken)

	if requestBody != nil {
		request.Header.Set("Content-Type", "application/vnd.contentful.management.v1+json")
	}

	response, err := server.Client().Do(request)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, response.Body.Close())
	})

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	return response.StatusCode, body
}

func assertAppSigningSecretResponse(t *testing.T, body []byte, submittedValue string) {
	t.Helper()

	var rawResponse struct {
		Sys map[string]json.RawMessage `json:"sys"`
	}
	require.NoError(t, json.Unmarshal(body, &rawResponse))
	assert.NotContains(t, rawResponse.Sys, "createdAt")
	assert.NotContains(t, rawResponse.Sys, "updatedAt")
	assert.NotContains(t, rawResponse.Sys, "createdBy")
	assert.NotContains(t, rawResponse.Sys, "updatedBy")

	var response cm.AppSigningSecret
	require.NoError(t, json.Unmarshal(body, &response))
	require.NoError(t, response.Validate())

	assert.Len(t, response.RedactedValue, 4)
	assert.Equal(t, submittedValue[len(submittedValue)-4:], response.RedactedValue)
	assert.NotEqual(t, submittedValue, response.RedactedValue)
	assert.Equal(t, "organization", response.Sys.Organization.Sys.ID)
	assert.Equal(t, "app-definition", response.Sys.AppDefinition.Sys.ID)
}
