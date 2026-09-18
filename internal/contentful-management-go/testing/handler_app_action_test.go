package cmtesting_test

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	appActionCollectionPath = "/organizations/org/app_definitions/app/actions"
	appActionBody           = `{"name":"Action","category":"Custom","type":"endpoint","url":"https://example.invalid/action","parameters":[]}`
)

func appActionHTTP(t *testing.T, server *cmt.Server, method, path, body string, status int) map[string]any {
	t.Helper()
	request := httptest.NewRequestWithContext(t.Context(), method, path, strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+cmt.ValidAccessToken)
	request.Header.Set("Content-Type", "application/vnd.contentful.management.v1+json")

	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	require.Equal(t, status, response.Code, response.Body.String())

	if status == http.StatusNoContent {
		assert.Empty(t, response.Body.String())

		return nil
	}

	var result map[string]any

	decoder := json.NewDecoder(response.Body)
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&result))

	return result
}

func appActionServer(t *testing.T) *cmt.Server {
	t.Helper()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.SetAppDefinition("org", "app", cm.AppDefinitionData{Name: "App"})

	return server
}

func TestAppActionServerLifecycle(t *testing.T) {
	t.Parallel()
	server := appActionServer(t)
	initial := appActionHTTP(t, server, http.MethodGet, appActionCollectionPath, "", 200)
	assert.Equal(t, json.Number("0"), initial["total"])
	assert.Equal(t, []any{}, initial["items"])

	created := appActionHTTP(t, server, http.MethodPost, appActionCollectionPath, appActionBody, 201)
	sys, ok := created["sys"].(map[string]any)
	require.True(t, ok)
	actionID, ok := sys["id"].(string)
	require.True(t, ok)
	require.NotEmpty(t, actionID)
	assert.Equal(t, "AppAction", sys["type"])
	assert.Equal(t, map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Organization", "id": "org"}}, sys["organization"])
	assert.Equal(t, map[string]any{"sys": map[string]any{"type": "Link", "linkType": "AppDefinition", "id": "app"}}, sys["appDefinition"])
	assert.NotContains(t, sys, "version")

	path := appActionCollectionPath + "/" + actionID
	assert.Equal(t, created, appActionHTTP(t, server, http.MethodGet, path, "", 200))

	// Required input and target fields cannot be cleared by omission.
	for _, invalid := range []string{
		strings.Replace(appActionBody, `,"parameters":[]`, "", 1),
		strings.Replace(appActionBody, `,"url":"https://example.invalid/action"`, "", 1),
		`{"name":"Function","category":"Custom","type":"function-invocation","function":{"sys":{"type":"Link","linkType":"Function","id":""}},"parameters":[]}`,
	} {
		appActionHTTP(t, server, http.MethodPost, appActionCollectionPath, invalid, 422)
		appActionHTTP(t, server, http.MethodPut, path, invalid, 422)
		assert.Equal(t, created, appActionHTTP(t, server, http.MethodGet, path, "", 200))
	}

	function := `{"name":"Function","category":"Custom","type":"function-invocation","function":{"sys":{"type":"Link","linkType":"Function","id":"undeployed"}},"parametersSchema":{"type":"object"},"resultSchema":{"type":"object"},"description":""}`
	updated := appActionHTTP(t, server, http.MethodPut, path, function, 200)
	assert.Equal(t, sys, updated["sys"])
	assert.NotContains(t, updated, "url")
	assert.NotContains(t, updated, "parameters")
	assert.Contains(t, updated, "description")
	assert.Empty(t, updated["description"])
	assert.Equal(t, map[string]any{"type": "object"}, updated["parametersSchema"])
	assert.Equal(t, updated, appActionHTTP(t, server, http.MethodGet, path, "", 200))

	// A complete replacement clears the old target, description, and both schemas.
	restored := appActionHTTP(t, server, http.MethodPut, path, appActionBody, 200)
	assert.Equal(t, created, restored)
	assert.Equal(t, created, appActionHTTP(t, server, http.MethodGet, path, "", 200))
	appActionHTTP(t, server, http.MethodDelete, path, "", 204)

	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
		appActionHTTP(t, server, method, path, appActionBody, 404)
	}

	empty := appActionHTTP(t, server, http.MethodGet, appActionCollectionPath, "", 200)
	assert.Equal(t, []any{}, empty["items"])
	assert.Equal(t, json.Number("0"), empty["total"])
}

func TestAppActionServerScopeAndPagination(t *testing.T) {
	t.Parallel()
	server := appActionServer(t)
	ids := make([]string, 0, 3)

	for range 3 {
		created := appActionHTTP(t, server, http.MethodPost, appActionCollectionPath, appActionBody, 201)
		sys, ok := created["sys"].(map[string]any)
		require.True(t, ok)
		actionID, ok := sys["id"].(string)
		require.True(t, ok)

		ids = append(ids, actionID)
	}

	slices.Sort(ids)
	require.Len(t, slices.Compact(slices.Clone(ids)), 3)

	for index, actionID := range ids {
		page := appActionHTTP(t, server, http.MethodGet, fmt.Sprintf("%s?limit=1&skip=%d", appActionCollectionPath, index), "", 200)
		assert.Equal(t, json.Number("3"), page["total"])
		assert.Equal(t, json.Number(strconv.Itoa(index)), page["skip"])
		assert.Equal(t, json.Number("1"), page["limit"])
		items, ok := page["items"].([]any)
		require.True(t, ok)
		require.Len(t, items, 1)
		assert.Equal(t, appActionHTTP(t, server, http.MethodGet, appActionCollectionPath+"/"+actionID, "", 200), items[0])
	}

	beyond := appActionHTTP(t, server, http.MethodGet, appActionCollectionPath+"?skip=99&limit=1", "", 200)
	assert.Equal(t, []any{}, beyond["items"])
	assert.Equal(t, json.Number("3"), beyond["total"])

	for _, query := range []string{"skip=-1", "limit=0", "limit=1001"} {
		appActionHTTP(t, server, http.MethodGet, appActionCollectionPath+"?"+query, "", 400)
	}

	server.SetAppDefinition("org", "other", cm.AppDefinitionData{Name: "Other"})
	server.SetAppDefinition("other", "foreign", cm.AppDefinitionData{Name: "Foreign"})

	for _, collection := range []string{
		"/organizations/other/app_definitions/app/actions",
		"/organizations/org/app_definitions/missing/actions",
		"/organizations/org/app_definitions/foreign/actions",
	} {
		appActionHTTP(t, server, http.MethodGet, collection, "", 404)
		appActionHTTP(t, server, http.MethodPost, collection, appActionBody, 404)

		for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
			appActionHTTP(t, server, method, collection+"/"+ids[0], appActionBody, 404)
		}
	}

	other := "/organizations/org/app_definitions/other/actions"
	assert.Equal(t, []any{}, appActionHTTP(t, server, http.MethodGet, other, "", 200)["items"])

	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
		appActionHTTP(t, server, method, other+"/"+ids[0], appActionBody, 404)
	}
	// Removing the parent removes its actions; reusing its ID cannot resurrect them.
	appActionHTTP(t, server, http.MethodDelete, "/organizations/org/app_definitions/app", "", 204)
	appActionHTTP(t, server, http.MethodGet, appActionCollectionPath, "", 404)
	server.SetAppDefinition("org", "app", cm.AppDefinitionData{Name: "Recreated"})
	assert.Equal(t, []any{}, appActionHTTP(t, server, http.MethodGet, appActionCollectionPath, "", 200)["items"])
}

func TestAppActionServerRejectsInvalidMutations(t *testing.T) {
	t.Parallel()

	for name, fields := range map[string]map[string]any{
		"empty name":                 {"name": ""},
		"unknown category":           {"category": "unknown"},
		"unknown type":               {"type": "unknown"},
		"HTTP URL":                   {"url": "http://example.invalid"},
		"empty URL":                  {"url": ""},
		"missing Function":           {"type": "function-invocation"},
		"both targets":               {"function": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Function", "id": "function"}}},
		"null parameters":            {"parameters": nil},
		"object parameters":          {"parameters": map[string]any{}},
		"missing parameter members":  {"parameters": []any{map[string]any{"id": "field"}}},
		"unsupported parameter type": {"parameters": []any{map[string]any{"id": "field", "name": "Field", "type": "Secret"}}},
		"both inputs":                {"parametersSchema": map[string]any{"type": "object"}},
		"empty schema":               {"resultSchema": map[string]any{}},
		"null schema":                {"resultSchema": nil},
		"array schema":               {"resultSchema": []any{}},
		"built-in parameters":        {"category": "Entries.v1.0"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			server := appActionServer(t)
			created := appActionHTTP(t, server, http.MethodPost, appActionCollectionPath, appActionBody, 201)
			sys, ok := created["sys"].(map[string]any)
			require.True(t, ok)
			actionID, ok := sys["id"].(string)
			require.True(t, ok)

			path := appActionCollectionPath + "/" + actionID

			var input map[string]any
			require.NoError(t, json.Unmarshal([]byte(appActionBody), &input))
			maps.Copy(input, fields)
			body, err := json.Marshal(input)
			require.NoError(t, err)

			for _, mutation := range []struct{ method, path string }{{http.MethodPost, appActionCollectionPath}, {http.MethodPut, path}} {
				result := appActionHTTP(t, server, mutation.method, mutation.path, string(body), 422)
				sys, ok := result["sys"].(map[string]any)
				require.True(t, ok)
				assert.Equal(t, "ValidationFailed", sys["id"])
			}

			assert.Equal(t, created, appActionHTTP(t, server, http.MethodGet, path, "", 200))
			assert.Equal(t, json.Number("1"), appActionHTTP(t, server, http.MethodGet, appActionCollectionPath, "", 200)["total"])
		})
	}
}

func TestAppActionServerBuiltinParameters(t *testing.T) {
	t.Parallel()

	for category, parameters := range map[string]string{
		"Entries.v1.0":      `[{"id":"entryIds","name":"Entry Ids","description":"Ids of the entries you want to trigger the action for","type":"Symbol","required":true}]`,
		"Notification.v1.0": `[{"id":"message","name":"Message","description":"The message being sent to external messaging service","type":"Symbol","required":true},{"id":"recipient","name":"Recipient","description":"","type":"Symbol","required":true}]`,
	} {
		t.Run(category, func(t *testing.T) {
			t.Parallel()
			server := appActionServer(t)
			body := fmt.Sprintf(`{"name":"Builtin","category":%q,"type":"endpoint","url":"https://example.invalid/action","parametersSchema":{"type":"object"},"resultSchema":{"type":"object"}}`, category)
			created := appActionHTTP(t, server, http.MethodPost, appActionCollectionPath, body, 201)
			actual, err := json.Marshal(created["parameters"])
			require.NoError(t, err)
			assert.JSONEq(t, parameters, string(actual))
			assert.Equal(t, map[string]any{"type": "object"}, created["parametersSchema"])
			assert.Equal(t, map[string]any{"type": "object"}, created["resultSchema"])
			sys, ok := created["sys"].(map[string]any)
			require.True(t, ok)
			actionID, ok := sys["id"].(string)
			require.True(t, ok)
			assert.Equal(t, created, appActionHTTP(t, server, http.MethodGet, appActionCollectionPath+"/"+actionID, "", 200))
		})
	}
}
