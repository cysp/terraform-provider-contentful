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

func TestContentfulManagementServerRejectsMissingOrEmptyWebhookTopics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
		path   string
		body   string
		want   string
	}{
		{
			name:   "create missing",
			method: http.MethodPost,
			path:   "/spaces/space/webhook_definitions",
			body:   testJSON(map[string]any{"name": "Webhook", "url": "https://example.com/webhook"}),
			want: testJSON(map[string]any{
				"sys":     map[string]any{"type": "Error", "id": "ValidationFailed"},
				"message": "Validation error",
				"details": map[string]any{
					"errors": []any{
						map[string]any{
							"name":    "invalid_type",
							"path":    []any{"topics"},
							"details": "Invalid input: expected array, received undefined at \"topics\"",
						},
					},
				},
			}),
		},
		{
			name:   "create empty",
			method: http.MethodPost,
			path:   "/spaces/space/webhook_definitions",
			body:   testJSON(map[string]any{"name": "Webhook", "url": "https://example.com/webhook", "topics": []any{}}),
			want: testJSON(map[string]any{
				"sys":     map[string]any{"type": "Error", "id": "ValidationFailed"},
				"message": "Validation error",
				"details": map[string]any{
					"errors": []any{map[string]any{"name": "topics", "path": []any{}, "details": "Topics cannot be empty"}},
				},
			}),
		},
		{
			name:   "update missing",
			method: http.MethodPut,
			path:   "/spaces/space/webhook_definitions/webhook",
			body:   testJSON(map[string]any{"name": "Webhook", "url": "https://example.com/webhook"}),
			want: testJSON(map[string]any{
				"sys":     map[string]any{"type": "Error", "id": "ValidationFailed"},
				"message": "Validation error",
				"details": map[string]any{
					"errors": []any{
						map[string]any{
							"name":    "invalid_type",
							"path":    []any{"topics"},
							"details": "Invalid input: expected array, received undefined at \"topics\"",
						},
					},
				},
			}),
		},
		{
			name:   "update empty",
			method: http.MethodPut,
			path:   "/spaces/space/webhook_definitions/webhook",
			body:   testJSON(map[string]any{"name": "Webhook", "url": "https://example.com/webhook", "topics": []any{}}),
			want: testJSON(map[string]any{
				"sys":     map[string]any{"type": "Error", "id": "ValidationFailed"},
				"message": "Validation error",
				"details": map[string]any{
					"errors": []any{map[string]any{"name": "topics", "path": []any{}, "details": "Topics cannot be empty"}},
				},
			}),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(100))
			require.NoError(t, err)
			server.RegisterSpaceEnvironment("space", "master")

			if test.method == http.MethodPut {
				server.SetWebhookDefinition("space", "webhook", cm.WebhookDefinitionData{
					Name:   "Existing webhook",
					URL:    "https://example.com/webhook",
					Topics: []string{"Entry.publish"},
				})
			}

			testServer := httptest.NewServer(server)
			t.Cleanup(testServer.Close)

			request, err := http.NewRequestWithContext(t.Context(), test.method, testServer.URL+test.path, strings.NewReader(test.body))
			require.NoError(t, err)
			request.Header.Set("Authorization", "Bearer CFPAT-12345")
			request.Header.Set("Content-Type", "application/vnd.contentful.management.v1+json")

			if test.method == http.MethodPut {
				request.Header.Set("X-Contentful-Version", "0")
			}

			response, err := testServer.Client().Do(request)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, response.Body.Close()) })

			body, err := io.ReadAll(response.Body)
			require.NoError(t, err)

			assert.Equal(t, http.StatusUnprocessableEntity, response.StatusCode)
			assert.JSONEq(t, test.want, string(body))

			if test.method == http.MethodPut {
				stored, ok := server.StoredWebhookDefinition("space", "webhook")
				require.True(t, ok)
				assert.Equal(t, "Existing webhook", stored.Name)
				assert.Equal(t, []string{"Entry.publish"}, stored.Topics)
			}
		})
	}
}

func TestWebhookDefinitionHTTPVersions(t *testing.T) {
	t.Parallel()

	const collection = "/spaces/space/webhook_definitions"

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")

	request := func(method, path, version string, body map[string]any, status int) map[string]any {
		t.Helper()

		var payload io.Reader
		if body != nil {
			payload = strings.NewReader(testJSON(body))
		}

		req := httptest.NewRequestWithContext(t.Context(), method, path, payload)
		req.Header.Set("Authorization", "Bearer "+cmt.ValidAccessToken)
		req.Header.Set("Content-Type", "application/vnd.contentful.management.v1+json")

		if version != "" {
			req.Header.Set("X-Contentful-Version", version)
		}

		resp := httptest.NewRecorder()
		server.ServeHTTP(resp, req)
		require.Equal(t, status, resp.Code, resp.Body.String())

		var result map[string]any

		decoder := json.NewDecoder(resp.Body)
		decoder.UseNumber()
		require.NoError(t, decoder.Decode(&result))
		sys, ok := result["sys"].(map[string]any)
		require.True(t, ok)

		return sys
	}

	header := map[string]any{"key": "X-Token", "secret": true, "value": "first"}
	body := map[string]any{"name": "Webhook", "url": "https://example.invalid/webhook", "topics": []string{"Entry.publish"}, "active": false, "headers": []any{header}}
	// PUT creates at an unused ID, even with a nonzero version precondition.
	for _, version := range []string{"", "7"} {
		id := "webhook" + version
		sys := request(http.MethodPut, collection+"/"+id, version, body, http.StatusOK)
		require.Equal(t, id, sys["id"])
		require.Equal(t, json.Number("0"), sys["version"])
		require.Equal(t, sys, request(http.MethodGet, collection+"/"+id, "", nil, http.StatusOK))
	}

	created := request(http.MethodPost, collection, "", body, http.StatusOK)
	require.Equal(t, json.Number("0"), created["version"])
	webhookID, ok := created["id"].(string)
	require.True(t, ok)
	require.NotEmpty(t, webhookID)
	path := collection + "/" + webhookID
	require.Equal(t, created, request(http.MethodGet, path, "", nil, http.StatusOK))
	require.Equal(t, "Conflict", request(http.MethodPut, path, "", body, http.StatusConflict)["id"])
	require.Equal(t, created, request(http.MethodPut, path, "0", body, http.StatusOK))

	delete(header, "value")
	require.Equal(t, created, request(http.MethodPut, path, "0", body, http.StatusOK))

	header["value"] = "rotated"
	rotated := request(http.MethodPut, path, "0", body, http.StatusOK)
	require.Equal(t, json.Number("1"), rotated["version"])
	require.Equal(t, "Conflict", request(http.MethodPut, path, "0", body, http.StatusConflict)["id"])
	require.Equal(t, rotated, request(http.MethodGet, path, "", nil, http.StatusOK))

	header["secret"] = false
	ordinary := request(http.MethodPut, path, "1", body, http.StatusOK)
	require.Equal(t, json.Number("2"), ordinary["version"])

	header["secret"] = true
	delete(header, "value")
	require.Equal(t, ordinary, request(http.MethodPut, path, "2", body, http.StatusOK))

	stored, present := server.StoredWebhookDefinition("space", webhookID)
	require.True(t, present)
	require.Len(t, stored.Headers, 1)
	require.Equal(t, cm.NewOptBool(false), stored.Headers[0].Secret)
	require.Equal(t, cm.NewOptString("rotated"), stored.Headers[0].Value)
}

func TestWebhookDefinitionRejectsMissingValueForNewSecretHeader(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")
	server.SetWebhookDefinition("space", "webhook", cm.WebhookDefinitionData{
		Name: "Original", URL: "https://example.invalid/webhook", Topics: []string{"Entry.publish"},
		Headers: cm.WebhookDefinitionHeaders{
			{Key: "X-Token", Secret: cm.NewOptBool(true), Value: cm.NewOptString("stored")},
		},
	})

	before, ok := server.StoredWebhookDefinition("space", "webhook")
	require.True(t, ok)

	body := testJSON(map[string]any{
		"name": "Changed", "url": "https://example.invalid/webhook", "topics": []string{"Entry.publish"},
		"headers": []map[string]any{
			{"key": "X-Token", "secret": true, "value": "rotated"},
			{"key": "x-token", "secret": true},
		},
	})
	for _, test := range []struct {
		version string
		status  int
		errorID string
	}{
		{"7", http.StatusConflict, "Conflict"},
		{"0", http.StatusUnprocessableEntity, "ValidationFailed"},
	} {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/spaces/space/webhook_definitions/webhook", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+cmt.ValidAccessToken)
		req.Header.Set("Content-Type", "application/vnd.contentful.management.v1+json")
		req.Header.Set("X-Contentful-Version", test.version)

		resp := httptest.NewRecorder()
		server.ServeHTTP(resp, req)

		require.Equal(t, test.status, resp.Code, resp.Body.String())
		result := testJSONObject(resp.Body.String())
		require.Equal(t, map[string]any{"type": "Error", "id": test.errorID}, result["sys"])

		if test.status == http.StatusUnprocessableEntity {
			require.Equal(t, "Validation error", result["message"])
			require.Equal(t, map[string]any{"errors": "Header x-token is missing a value and cannot be set"}, result["details"])
		}

		after, ok := server.StoredWebhookDefinition("space", "webhook")
		require.True(t, ok)
		require.Equal(t, before, after)
	}
}
