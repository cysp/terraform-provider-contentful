package cmtesting

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const webhookCollectionPath = "/spaces/space/webhook_definitions"

func webhookDefinitionRequest(t *testing.T, server *Server, method, path, version string, headers []map[string]any, status int) string {
	t.Helper()

	var body io.Reader
	if headers != nil {
		body = strings.NewReader(testJSON(map[string]any{
			"name": "Webhook", "url": "https://example.invalid/webhook", "topics": []string{"Entry.publish"}, "active": false, "headers": headers,
		}))
	}

	req := httptest.NewRequestWithContext(t.Context(), method, path, body)
	req.Header.Set("Authorization", "Bearer "+ValidAccessToken)
	req.Header.Set("Content-Type", "application/vnd.contentful.management.v1+json")

	if version != "" {
		req.Header.Set("X-Contentful-Version", version)
	}

	resp := httptest.NewRecorder()
	server.ServeHTTP(resp, req)
	require.Equal(t, status, resp.Code, resp.Body.String())

	return resp.Body.String()
}

func TestWebhookDefinitionPOSTRequiresHeaderValues(t *testing.T) {
	t.Parallel()

	server, err := NewContentfulManagementServer()
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")

	headers := []map[string]any{
		{"key": "X-Empty", "secret": true, "value": ""},
		{"key": "X-Secret", "secret": true},
		{"key": "X-Ordinary"},
	}
	response := webhookDefinitionRequest(t, server, http.MethodPost, webhookCollectionPath, "", headers, http.StatusUnprocessableEntity)
	require.JSONEq(t, testJSON(map[string]any{
		"sys": map[string]string{"type": "Error", "id": "ValidationFailed"}, "message": "Validation error",
		"details": map[string]any{"errors": []map[string]any{
			{"name": "invalid_type", "path": []any{"headers", 1, "value"}, "details": `Invalid input: expected string, received undefined at "headers.1.value"`},
			{"name": "invalid_type", "path": []any{"headers", 2, "value"}, "details": `Invalid input: expected string, received undefined at "headers.2.value"`},
		}},
	}), response)
	// POST assigns the webhook ID, so inspect the collection for a partial write.
	require.Empty(t, server.h.webhookDefinitions.List("space"))

	webhookDefinitionRequest(t, server, http.MethodPost, webhookCollectionPath, "", headers[:1], http.StatusOK)
	stored := server.h.webhookDefinitions.List("space")
	require.Len(t, stored, 1)
	require.Len(t, stored[0].Headers, 1)
	value, present := stored[0].Headers[0].Value.Get()
	require.True(t, present)
	require.Empty(t, value)
}

func TestWebhookDefinitionRejectedPOSTLeavesCollectionEmpty(t *testing.T) {
	t.Parallel()

	server, err := NewContentfulManagementServer()
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")

	headers := []map[string]any{{"key": "X-Review", "value": "first"}, {"key": "X-Review", "value": "second"}}
	webhookDefinitionRequest(t, server, http.MethodPost, webhookCollectionPath, "", headers, http.StatusUnprocessableEntity)
	require.Empty(t, server.h.webhookDefinitions.List("space"))
}

func TestWebhookDefinitionPUTAfterDeletionRequiresHeaderValues(t *testing.T) {
	t.Parallel()

	server, err := NewContentfulManagementServer(WithRateLimitPerSecond(100))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")

	path := webhookCollectionPath + "/webhook"
	header := map[string]any{"key": "X-Secret", "secret": true, "value": ""}
	headers := []map[string]any{header}
	webhookDefinitionRequest(t, server, http.MethodPut, path, "", headers, http.StatusOK)
	stored, ok := server.StoredWebhookDefinition("space", "webhook")
	require.True(t, ok)
	require.Zero(t, stored.Sys.Version)
	webhookDefinitionRequest(t, server, http.MethodDelete, path, "", nil, http.StatusNoContent)

	delete(header, "value")

	for _, secret := range []bool{true, false} {
		header["secret"] = secret
		response := webhookDefinitionRequest(t, server, http.MethodPut, path, "0", headers, http.StatusInternalServerError)
		require.JSONEq(t, testJSON(map[string]any{
			"sys": map[string]string{"type": "Error", "id": "ServerError"}, "message": "Failed to create webhook definition/s",
		}), response)
		webhookDefinitionRequest(t, server, http.MethodGet, path, "", nil, http.StatusNotFound)
	}

	headers = []map[string]any{header, {"key": "X-Ordinary"}}
	response := webhookDefinitionRequest(t, server, http.MethodPut, path, "0", headers, http.StatusUnprocessableEntity)
	require.JSONEq(t, testJSON(map[string]any{
		"sys": map[string]string{"type": "Error", "id": "ValidationFailed"}, "message": "Validation error",
		"details": map[string]any{"errors": []map[string]any{
			{"name": "invalid_union", "path": []any{"headers", 1}, "details": `Invalid input at "headers.1"`},
		}},
	}), response)
	webhookDefinitionRequest(t, server, http.MethodGet, path, "", nil, http.StatusNotFound)
}
