package cmtesting_test

import (
	"crypto/rand"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
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
			name:   "create missing topics and header value",
			method: http.MethodPost,
			path:   "/spaces/space/webhook_definitions",
			body:   `{"name":"Webhook","url":"https://example.invalid/webhook","active":false,"headers":[{"key":"X","secret":true}]}`,
			want: `{"sys":{"type":"Error","id":"ValidationFailed"},"message":"Validation error","details":{"errors":[
				{"name":"invalid_type","path":["headers",0,"value"],"details":"Invalid input: expected string, received undefined at \"headers.0.value\""},
				{"name":"invalid_type","path":["topics"],"details":"Invalid input: expected array, received undefined at \"topics\""}
			]}}`,
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

		result := webhookHTTP(t, server, method, path, version, body, status)
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

	// The existing exact name keeps its secret flag. A flag-only PUT with
	// matching bytes also keeps its version.
	header["secret"] = false

	require.Equal(t, rotated, request(http.MethodPut, path, "1", body, http.StatusOK))

	header["secret"] = true
	delete(header, "value")
	require.Equal(t, rotated, request(http.MethodPut, path, "1", body, http.StatusOK))

	stored, present := server.StoredWebhookDefinition("space", webhookID)
	require.True(t, present)
	require.Len(t, stored.Headers, 1)
	require.Equal(t, cm.NewOptBool(true), stored.Headers[0].Secret)
	require.Equal(t, cm.NewOptString("rotated"), stored.Headers[0].Value)
}

func TestWebhookDefinitionRejectsMissingValueForNewSecretHeader(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")
	server.SetWebhookDefinition("space", "webhook", cm.WebhookDefinitionData{
		Name: "Original", URL: "https://example.invalid/webhook", Topics: []string{"Entry.publish"},
		Headers: cm.WebhookDefinitionRequestHeaders{
			{Key: "X-Token", Secret: cm.NewOptBool(true), Value: cm.NewOptNilString("stored")},
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

func webhookHTTP(t *testing.T, server *cmt.Server, method, path, version string, body map[string]any, status int) map[string]any {
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
	require.Equal(t, status, resp.Code)

	return testJSONObject(resp.Body.String())
}

func TestWebhookDefinitionHeaderValidationOrder(t *testing.T) {
	t.Parallel()

	const (
		existing = webhookCollection + "/existing"
		unused   = webhookCollection + "/unused"
	)

	header := func(key string) map[string]any { return map[string]any{"key": key, "value": "readable"} }
	duplicate := func(key string) map[string]any {
		return map[string]any{"name": "headers", "path": []any{"key"}, "details": "Header key " + key + " is duplicated"}
	}
	duplicates := []map[string]any{header("A"), header("B"), header("B"), header("A"), header("B")}
	missingPassword := map[string]any{"name": "httpBasicPassword", "path": []any{}, "details": "HTTP Basic Auth password is required when username is provided"}

	for _, route := range []struct{ name, method, path, version string }{
		{name: "POST", method: http.MethodPost, path: webhookCollection},
		{name: "create PUT", method: http.MethodPut, path: unused},
		{name: "update PUT", method: http.MethodPut, path: existing, version: "0"},
		{name: "version conflict", method: http.MethodPut, path: existing, version: "7"},
	} {
		t.Run(route.name, func(t *testing.T) {
			t.Parallel()

			shape := map[string]any{"name": "invalid_union", "path": []any{"headers", json.Number("5")}, "details": `Invalid input at "headers.5"`}
			if route.method == http.MethodPost {
				shape = map[string]any{"name": "invalid_type", "path": []any{"headers", json.Number("5"), "value"}, "details": `Invalid input: expected string, received undefined at "headers.5.value"`}
			}

			for _, test := range []struct {
				name    string
				headers []map[string]any
				fields  map[string]any
				errors  []any
			}{
				{name: "repeated names", headers: duplicates, errors: []any{duplicate("B"), duplicate("A"), duplicate("B")}},
				{name: "credentials after names", headers: duplicates, fields: map[string]any{basicUsername: "new"}, errors: []any{duplicate("B"), duplicate("A"), duplicate("B"), missingPassword}},
				{name: "shape before names and credentials", headers: append(slices.Clone(duplicates), map[string]any{"key": "Missing"}), fields: map[string]any{basicUsername: "new"}, errors: []any{shape}},
				{name: "known incomplete pair", headers: []map[string]any{header("A")}, fields: map[string]any{basicUsername: "new"}, errors: []any{missingPassword}},
			} {
				t.Run(test.name, func(t *testing.T) {
					server := newWebhookBasicServer(t)
					seed := webhookBasicBody("Original", map[string]any{basicUsername: "user", basicPassword: rand.Text()})
					original := webhookHTTP(t, server, http.MethodPut, existing, "", seed, http.StatusOK)
					body := webhookBasicBody("Rejected", test.fields)
					body["headers"] = test.headers
					got := webhookHTTP(t, server, route.method, route.path, route.version, body, http.StatusUnprocessableEntity)
					require.Equal(t, "Validation error", got["message"])
					require.Equal(t, map[string]any{"type": "Error", "id": "ValidationFailed"}, got["sys"])
					require.Equal(t, map[string]any{"errors": test.errors}, got["details"])
					require.Equal(t, original, webhookHTTP(t, server, http.MethodGet, existing, "", nil, http.StatusOK))
					webhookHTTP(t, server, http.MethodGet, unused, "", nil, http.StatusNotFound)
					require.Equal(t, original, webhookHTTP(t, server, http.MethodPut, existing, "0", seed, http.StatusOK))
				})
			}
		})
	}
}

func TestWebhookDefinitionPUTRetainsHeaderWithExplicitFalse(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name    string
		missing bool
	}{{name: "existing keys"}, {name: "missing key", missing: true}} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			server := newWebhookBasicServer(t)

			const path = webhookCollection + "/retained"

			secret := rand.Text()
			body := webhookBasicBody("Original", nil)
			body["headers"] = []map[string]any{
				{"key": "X-A", "secret": true, "value": secret},
				{"key": "x-a", "secret": false, "value": "visible"},
			}
			initial := webhookHTTP(t, server, http.MethodPut, path, "", body, http.StatusOK)
			require.Equal(t, []any{map[string]any{"key": "X-A", "secret": true}, map[string]any{"key": "x-a", "secret": false, "value": "visible"}}, initial["headers"])

			keep := webhookBasicBody("Original", nil)

			keep["headers"] = []map[string]any{{"key": "X-A", "secret": false}, {"key": "x-a", "secret": false}}
			if test.missing {
				keep["name"] = "Rejected"
				keep["headers"] = []map[string]any{{"key": "X-A", "secret": true, "value": rand.Text()}, {"key": "Missing", "secret": false}}
				result := webhookHTTP(t, server, http.MethodPut, path, "0", keep, http.StatusUnprocessableEntity)
				require.Equal(t, map[string]any{"errors": "Header Missing is missing a value and cannot be set"}, result["details"])
			} else {
				require.Equal(t, initial, webhookHTTP(t, server, http.MethodPut, path, "0", keep, http.StatusOK))
			}

			require.Equal(t, initial, webhookHTTP(t, server, http.MethodGet, path, "", nil, http.StatusOK))
			require.Equal(t, initial, webhookHTTP(t, server, http.MethodPut, path, "0", body, http.StatusOK))
		})
	}
}

func TestWebhookDefinitionNullHeaderValue(t *testing.T) {
	t.Parallel()

	const path = webhookCollection + "/hook"

	header := map[string]any{"key": "X-Secret", "secret": true, "value": nil}
	body := webhookBasicBody("Webhook", map[string]any{"headers": []any{header}})

	t.Run("POST", func(t *testing.T) {
		t.Parallel()
		server := newWebhookBasicServer(t)
		got := webhookHTTP(t, server, http.MethodPost, webhookCollection, "", body, http.StatusUnprocessableEntity)
		require.Equal(t, map[string]any{"type": "Error", "id": "ValidationFailed"}, got["sys"])
		require.Equal(t, map[string]any{"errors": []any{map[string]any{"name": "invalid_type", "path": []any{"headers", json.Number("0"), "value"}, "details": `Invalid input: expected string, received null at "headers.0.value"`}}}, got["details"])
	})
	t.Run("create PUT", func(t *testing.T) {
		t.Parallel()
		server := newWebhookBasicServer(t)
		got := webhookHTTP(t, server, http.MethodPut, path, "", body, http.StatusInternalServerError)
		require.Equal(t, map[string]any{"type": "Error", "id": "ServerError"}, got["sys"])
		webhookHTTP(t, server, http.MethodGet, path, "", nil, http.StatusNotFound)
	})
	t.Run("update PUT", func(t *testing.T) {
		t.Parallel()
		server := newWebhookBasicServer(t)
		managed := map[string]any{"key": "X-Secret", "secret": true, "value": rand.Text()}
		seed := webhookBasicBody("Webhook", map[string]any{"headers": []any{managed}})
		original := webhookHTTP(t, server, http.MethodPut, path, "", seed, http.StatusOK)
		require.Equal(t, []any{map[string]any{"key": "X-Secret", "secret": true}}, original["headers"])
		require.Equal(t, original, webhookHTTP(t, server, http.MethodPut, path, "0", body, http.StatusOK))
		require.Equal(t, original, webhookHTTP(t, server, http.MethodGet, path, "", nil, http.StatusOK))
		require.Equal(t, original, webhookHTTP(t, server, http.MethodPut, path, "0", seed, http.StatusOK))

		managed["value"] = rand.Text()
		got := webhookHTTP(t, server, http.MethodPut, path, "0", seed, http.StatusOK)
		sys, ok := got["sys"].(map[string]any)
		require.True(t, ok)
		require.Equal(t, json.Number("1"), sys["version"])
	})
}

func TestWebhookDefinitionHeaderRejectsNullResponseValue(t *testing.T) {
	t.Parallel()

	var header cm.WebhookDefinitionHeader
	require.Error(t, json.Unmarshal([]byte(`{"key":"X","secret":true,"value":null}`), &header))
}

func TestWebhookDefinitionCreateRequiresHeaderValues(t *testing.T) {
	t.Parallel()

	for _, header := range []map[string]any{
		{"key": "X-Test", "secret": true},
		{"key": "X-Test", "secret": true, "value": nil},
	} {
		body := webhookBasicBody("Create", map[string]any{"headers": []any{header}})
		raw := []byte(testJSON(body))

		var create cm.WebhookDefinitionCreateData
		require.Error(t, json.Unmarshal(raw, &create))

		var update cm.WebhookDefinitionData
		require.NoError(t, json.Unmarshal(raw, &update))
	}
}
