package cmtesting_test

import (
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

func TestWebhookSigningSecretMockWireContract(t *testing.T) {
	t.Parallel()

	mock, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	mock.RegisterSpaceEnvironment("space", "master")
	server := httptest.NewServer(mock)
	t.Cleanup(server.Close)

	for _, step := range []struct {
		method, space, body string
		status              int
		expected            string
	}{
		{"GET", "space", "", 404, `"id":"NotFound"`},
		{"PUT", "missing", `{"value":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaAb09+/=_-"}`, 404, `"id":"NotFound"`},
		{"PUT", "space", `{"value":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaAb09+/=_-"}`, 201, `{"sys":{"type":"WebhookSigningSecret","space":{"sys":{"type":"Link","linkType":"Space","id":"space"}}},"redactedValue":"/=_-"}`},
		{"PUT", "space", `{"value":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbZy87-/=_+"}`, 200, `{"sys":{"type":"WebhookSigningSecret","space":{"sys":{"type":"Link","linkType":"Space","id":"space"}}},"redactedValue":"/=_+"}`},
		{"GET", "space", "", 200, `{"sys":{"type":"WebhookSigningSecret","space":{"sys":{"type":"Link","linkType":"Space","id":"space"}}},"redactedValue":"/=_+"}`},
		{"GET", "missing", "", 404, `"id":"NotFound"`},
		{"DELETE", "missing", "", 404, `"id":"NotFound"`},
		{"DELETE", "space", "", 204, ""},
		{"GET", "space", "", 404, `"id":"NotFound"`},
		{"DELETE", "space", "", 404, `"id":"NotFound"`},
		{"PUT", "space", `{"value":"DO_NOT_ECHO_INVALID_SECRET!"}`, 422, `"id":"ValidationFailed"`},
	} {
		req, requestErr := http.NewRequestWithContext(t.Context(), step.method, server.URL+"/spaces/"+step.space+"/webhook_settings/signing_secret", strings.NewReader(step.body))
		require.NoError(t, requestErr)
		req.Header.Set("Authorization", "Bearer "+cmt.ValidAccessToken)

		if step.body != "" {
			req.Header.Set("Content-Type", "application/vnd.contentful.management.v1+json")
		}

		resp, requestErr := server.Client().Do(req)
		require.NoError(t, requestErr)

		body, readErr := io.ReadAll(resp.Body)
		require.NoError(t, readErr)
		require.NoError(t, resp.Body.Close())
		assert.Equal(t, step.status, resp.StatusCode)

		if step.status == 200 || step.status == 201 {
			assert.JSONEq(t, step.expected, string(body))
		} else if step.expected != "" {
			assert.Contains(t, string(body), step.expected)
		}

		if step.status == 204 {
			assert.Empty(t, body)
		}

		assert.NotContains(t, string(body), "DO_NOT_ECHO_INVALID_SECRET")
		assert.NotContains(t, string(body), "aaaaaaaaaaaaaaaa")
		assert.NotContains(t, string(body), "bbbbbbbbbbbbbbbb")
	}
}

func TestWebhookSigningSecretMockRequestErrors(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, body, errorID, message string
		status                       int
	}{
		{"empty body", "", "ValidationFailed", "Validation error", 422},
		{"malformed JSON", "{", "BadRequest", "Invalid request payload JSON format", 400},
		{"missing value", `{}`, "ValidationFailed", "Validation error", 422},
		{"null value", `{"value":null}`, "ValidationFailed", "Validation error", 422},
		{"empty value", `{"value":""}`, "ValidationFailed", "Validation error", 422},
		{"short value", `{"value":"` + strings.Repeat("a", 63) + `"}`, "ValidationFailed", "Validation error", 422},
		{"long value", `{"value":"` + strings.Repeat("a", 65) + `"}`, "ValidationFailed", "Validation error", 422},
		{"invalid ASCII", `{"value":"` + strings.Repeat("a", 63) + `!"}`, "ValidationFailed", "Validation error", 422},
		{"non-ASCII", `{"value":"` + strings.Repeat("é", 64) + `"}`, "ValidationFailed", "Validation error", 422},
		{"number", `{"value":42}`, "ValidationFailed", "Validation error", 422},
		{"array", `{"value":[]}`, "ValidationFailed", "Validation error", 422},
		{"object", `{"value":{}}`, "ValidationFailed", "Validation error", 422},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
			require.NoError(t, err)
			server.RegisterSpaceEnvironment("space", "master")

			request := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/spaces/space/webhook_settings/signing_secret", strings.NewReader(test.body))
			request.Header.Set("Authorization", "Bearer "+cmt.ValidAccessToken)
			request.Header.Set("Content-Type", "application/vnd.contentful.management.v1+json")

			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)
			require.Equal(t, test.status, response.Code)

			var failure struct {
				Sys struct {
					Type string `json:"type"`
					ID   string `json:"id"`
				} `json:"sys"`
				Message string `json:"message"`
			}
			require.NoError(t, json.Unmarshal(response.Body.Bytes(), &failure))
			assert.Equal(t, "Error", failure.Sys.Type)
			assert.Equal(t, test.errorID, failure.Sys.ID)
			assert.Equal(t, test.message, failure.Message)

			read, readErr := server.Handler().GetWebhookSigningSecret(t.Context(), cm.GetWebhookSigningSecretParams{SpaceID: "space"})
			require.NoError(t, readErr)

			missing, ok := read.(*cm.ErrorStatusCode)
			require.True(t, ok)
			assert.Equal(t, http.StatusNotFound, missing.StatusCode)
		})
	}
}
