package cmtesting_test

import (
	"crypto/rand"
	"encoding/json"
	"maps"
	"net/http"
	"testing"

	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/require"
)

const (
	basicUsername     = "httpBasicUsername"
	basicPassword     = "httpBasicPassword"
	webhookCollection = "/spaces/space/webhook_definitions"
)

func webhookBasicBody(name string, credentials map[string]any) map[string]any {
	body := map[string]any{"name": name, "url": "https://example.invalid/webhook", "topics": []string{"Entry.publish"}, "active": false}
	maps.Copy(body, credentials)

	return body
}

func newWebhookBasicServer(t *testing.T) *cmt.Server {
	t.Helper()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")

	return server
}

func checkWebhookBasicResponse(t *testing.T, response map[string]any, name string, username any, version string) {
	t.Helper()

	_, hasPassword := response[basicPassword]
	require.False(t, hasPassword)

	gotUsername, hasUsername := response[basicUsername]
	require.True(t, hasUsername)
	require.Equal(t, username, gotUsername)
	require.Equal(t, name, response["name"])
	sys, ok := response["sys"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, json.Number(version), sys["version"])
}

func TestWebhookDefinitionBasicAuthUpdates(t *testing.T) {
	t.Parallel()

	server := newWebhookBasicServer(t)

	const path = webhookCollection + "/webhook"

	first, second := rand.Text(), rand.Text()
	pair := map[string]any{basicUsername: "user", basicPassword: first}
	created := webhookHTTP(t, server, http.MethodPut, path, "", webhookBasicBody("Original", pair), http.StatusOK)
	checkWebhookBasicResponse(t, created, "Original", "user", "0")
	require.Equal(t, created, webhookHTTP(t, server, http.MethodGet, path, "", nil, http.StatusOK))

	for _, step := range []struct {
		version      string
		fields       map[string]any
		wantUsername any
		wantVersion  string
	}{
		{version: "0", fields: nil, wantUsername: "user", wantVersion: "1"}, // Rename, omit both.
		{version: "1", fields: pair, wantUsername: "user", wantVersion: "1"},
		{version: "1", fields: map[string]any{basicUsername: "user", basicPassword: second}, wantUsername: "user", wantVersion: "2"},
		{version: "2", fields: map[string]any{basicUsername: nil, basicPassword: nil}, wantUsername: nil, wantVersion: "3"},
		{version: "3", fields: map[string]any{basicUsername: nil}, wantUsername: nil, wantVersion: "3"},
		{version: "3", fields: pair, wantUsername: "user", wantVersion: "4"},
		{version: "4", fields: map[string]any{basicPassword: ""}, wantUsername: "user", wantVersion: "5"},
		{version: "5", fields: map[string]any{basicPassword: nil}, wantUsername: "user", wantVersion: "6"},
		{version: "6", fields: map[string]any{basicUsername: "", basicPassword: ""}, wantUsername: "", wantVersion: "7"},
	} {
		response := webhookHTTP(t, server, http.MethodPut, path, step.version, webhookBasicBody("Renamed", step.fields), http.StatusOK)
		checkWebhookBasicResponse(t, response, "Renamed", step.wantUsername, step.wantVersion)
		require.Equal(t, response, webhookHTTP(t, server, http.MethodGet, path, "", nil, http.StatusOK))
	}
}

func TestWebhookDefinitionBasicAuthValidation(t *testing.T) {
	t.Parallel()

	password := rand.Text()

	const (
		usernameRequiredDetail = "HTTP Basic Auth username is required when password is provided"
		usernameProvidedDetail = "HTTP Basic Auth password is required when username is provided"
		loneNull               = "HTTP Basic Auth username and password must both be provided when setting to null"
	)
	for _, test := range []struct {
		name              string
		fields            map[string]any
		errorName, detail string
	}{
		{name: "username only", fields: map[string]any{basicUsername: "user"}, errorName: basicPassword, detail: usernameProvidedDetail},
		{name: "password only", fields: map[string]any{basicPassword: password}, errorName: basicUsername, detail: usernameRequiredDetail},
		{name: "null password", fields: map[string]any{basicUsername: "user", basicPassword: nil}, errorName: basicPassword, detail: usernameProvidedDetail},
		{name: "null username", fields: map[string]any{basicUsername: nil, basicPassword: password}, errorName: basicUsername, detail: usernameRequiredDetail},
		{name: "empty password", fields: map[string]any{basicUsername: "user", basicPassword: ""}, errorName: basicPassword, detail: usernameProvidedDetail},
		{name: "empty username", fields: map[string]any{basicUsername: "", basicPassword: password}, errorName: basicUsername, detail: usernameRequiredDetail},
		{name: "lone null username", fields: map[string]any{basicUsername: nil}, errorName: basicPassword, detail: loneNull},
		{name: "lone null password", fields: map[string]any{basicPassword: nil}, errorName: basicUsername, detail: loneNull},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			server := newWebhookBasicServer(t)

			const path = webhookCollection + "/webhook"

			pair := map[string]any{basicUsername: "user", basicPassword: password}
			original := webhookHTTP(t, server, http.MethodPut, path, "", webhookBasicBody("Original", pair), http.StatusOK)
			checkWebhookBasicResponse(t, original, "Original", "user", "0")

			rejected := webhookHTTP(t, server, http.MethodPut, path, "0", webhookBasicBody("Rejected", test.fields), http.StatusUnprocessableEntity)
			require.Equal(t, map[string]any{"type": "Error", "id": "ValidationFailed"}, rejected["sys"])
			require.Equal(t, "Validation error", rejected["message"])
			require.Equal(t, map[string]any{"errors": []any{map[string]any{"name": test.errorName, "path": []any{}, "details": test.detail}}}, rejected["details"])
			require.Equal(t, original, webhookHTTP(t, server, http.MethodGet, path, "", nil, http.StatusOK))
			require.Equal(t, original, webhookHTTP(t, server, http.MethodPut, path, "0", webhookBasicBody("Original", pair), http.StatusOK))
		})
	}
}

func TestWebhookDefinitionBasicAuthCreation(t *testing.T) {
	t.Parallel()

	for _, method := range []string{http.MethodPost, http.MethodPut} {
		t.Run(method, func(t *testing.T) {
			t.Parallel()
			server := newWebhookBasicServer(t)

			path := webhookCollection
			if method == http.MethodPut {
				path += "/unused"
			}

			for _, fields := range []map[string]any{{basicUsername: "user"}, {basicPassword: rand.Text()}} {
				response := webhookHTTP(t, server, method, path, "", webhookBasicBody("Rejected", fields), http.StatusUnprocessableEntity)
				require.Equal(t, map[string]any{"type": "Error", "id": "ValidationFailed"}, response["sys"])

				if method == http.MethodPut {
					webhookHTTP(t, server, http.MethodGet, path, "", nil, http.StatusNotFound)
				}
			}

			// A lone null is valid at creation and must read back as an explicit null username.
			response := webhookHTTP(t, server, method, path, "", webhookBasicBody("Created", map[string]any{basicPassword: nil}), http.StatusOK)
			checkWebhookBasicResponse(t, response, "Created", nil, "0")
			sys, ok := response["sys"].(map[string]any)
			require.True(t, ok)
			id, ok := sys["id"].(string)
			require.True(t, ok)
			require.NotEmpty(t, id)
			require.Equal(t, response, webhookHTTP(t, server, http.MethodGet, webhookCollection+"/"+id, "", nil, http.StatusOK))
		})
	}
}
