package cmtesting_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const appInstallationPath = "/spaces/space/environments/environment/app_installations/app"

func appInstallationServer(t *testing.T, declarations ...cm.AppDefinitionParameter) *cmt.Server {
	t.Helper()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")
	server.SetAppDefinition("org", "app", cm.AppDefinitionData{Name: "App", Parameters: cm.NewOptAppDefinitionParameters(cm.AppDefinitionParameters{Installation: declarations})})

	return server
}

func appInstallationHTTP(t *testing.T, server *cmt.Server, method, body string, status int) map[string]any {
	t.Helper()
	request := httptest.NewRequestWithContext(t.Context(), method, appInstallationPath, strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+cmt.ValidAccessToken)
	request.Header.Set("Content-Type", "application/vnd.contentful.management.v1+json")
	// The live endpoint ignored both stale and impossible version preconditions.
	if method == http.MethodPut {
		request.Header.Set("X-Contentful-Version", "2147483647")
	}

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

	if status == http.StatusNotFound {
		assert.Equal(t, map[string]any{"sys": map[string]any{"type": "Error", "id": "NotFound"}, "message": "The resource could not be found.", "details": "AppInstallation does not exist."}, result)
	}

	return result
}

// Literal expectations below come from the paired PAT/App Identity observations,
// not from the mock's mask or request conversion implementation.
//
//nolint:paralleltest,tparallel // Each step depends on the preceding installation mutation.
func TestAppInstallationServerSecretLifecycle(t *testing.T) {
	t.Parallel()
	server := appInstallationServer(t,
		cm.AppDefinitionParameter{ID: "secret", Name: "Secret", Type: "Secret"},
		cm.AppDefinitionParameter{ID: "otherSecret", Name: "Other secret", Type: "Secret"},
		cm.AppDefinitionParameter{ID: "ordinary", Name: "Ordinary", Type: "Symbol"},
	)
	appInstallationHTTP(t, server, http.MethodGet, "", 404)
	appInstallationHTTP(t, server, http.MethodDelete, "", 404)

	for _, step := range []struct {
		name                            string
		request, response, read, stored map[string]any
	}{
		{"create plaintext", map[string]any{"secret": "a🔑é", "ordinary": "***"}, map[string]any{"secret": "a🔑é", "ordinary": "***"}, map[string]any{"secret": "****", "ordinary": "***"}, map[string]any{"secret": "a🔑é", "ordinary": "***"}},
		{"replay current mask", map[string]any{"secret": "****", "ordinary": "changed"}, map[string]any{"secret": "****", "ordinary": "changed"}, map[string]any{"secret": "****", "ordinary": "changed"}, map[string]any{"secret": "a🔑é", "ordinary": "changed"}},
		{"different mask is literal", map[string]any{"secret": "***"}, map[string]any{"secret": "***"}, map[string]any{"secret": "***"}, map[string]any{"secret": "***"}},
		{"rotate both secrets", map[string]any{"secret": "alpha", "otherSecret": "bravo"}, map[string]any{"secret": "*****", "otherSecret": "*****"}, map[string]any{"secret": "*****", "otherSecret": "*****"}, map[string]any{"secret": "alpha", "otherSecret": "bravo"}},
		{"same length rotation", map[string]any{"secret": "delta"}, map[string]any{"secret": "*****"}, map[string]any{"secret": "*****"}, map[string]any{"secret": "delta"}},
		{"remove secret", map[string]any{"ordinary": "one"}, map[string]any{"ordinary": "one"}, map[string]any{"ordinary": "one"}, map[string]any{"ordinary": "one"}},
		{"mask with absent key is literal", map[string]any{"secret": "*******"}, map[string]any{"secret": "*******"}, map[string]any{"secret": "*******"}, map[string]any{"secret": "*******"}},
		{"empty secret", map[string]any{"secret": ""}, map[string]any{"secret": ""}, map[string]any{"secret": ""}, map[string]any{"secret": ""}},
		{"empty object", map[string]any{}, map[string]any{}, map[string]any{}, map[string]any{}},
	} {
		t.Run(step.name, func(t *testing.T) {
			response := appInstallationHTTP(t, server, http.MethodPut, testJSON(map[string]any{"parameters": step.request}), 200)
			assert.Equal(t, step.response, response["parameters"])
			sys, ok := response["sys"].(map[string]any)
			require.True(t, ok)
			assert.NotContains(t, sys, "version")
			assert.Equal(t, step.read, appInstallationHTTP(t, server, http.MethodGet, "", 200)["parameters"])
			stored, present := server.StoredAppInstallation("space", "environment", "app")
			require.True(t, present)
			assert.JSONEq(t, testJSON(step.stored), string(stored.Parameters))
		})
	}

	assert.NotContains(t, appInstallationHTTP(t, server, http.MethodPut, testJSON(map[string]any{}), 200), "parameters")
	assert.NotContains(t, appInstallationHTTP(t, server, http.MethodGet, "", 200), "parameters")
	appInstallationHTTP(t, server, http.MethodDelete, "", 204)
	appInstallationHTTP(t, server, http.MethodDelete, "", 404)
	appInstallationHTTP(t, server, http.MethodGet, "", 404)
	// Deleting and reinstalling restores the first-write plaintext representation.
	response := appInstallationHTTP(t, server, http.MethodPut, testJSON(map[string]any{"parameters": map[string]any{"secret": "again"}}), 200)
	assert.Equal(t, map[string]any{"secret": "again"}, response["parameters"])
}

func TestAppInstallationServerValidation(t *testing.T) {
	t.Parallel()

	for name, parameters := range map[string]any{
		"null object": nil, "array object": []any{}, "string object": "wrong", "number object": 42, "bool object": true,
		"missing required": map[string]any{}, "null secret": map[string]any{"secret": nil},
		"number secret": map[string]any{"secret": 42}, "object secret": map[string]any{"secret": map[string]any{"redacted": true}},
		"unexpected key": map[string]any{"secret": "valid", "extra": "wrong"},
		"number symbol":  map[string]any{"secret": "valid", "ordinary": 42},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			server := appInstallationServer(t, cm.AppDefinitionParameter{ID: "secret", Name: "Secret", Type: "Secret", Required: cm.NewOptBool(true)}, cm.AppDefinitionParameter{ID: "ordinary", Name: "Ordinary", Type: "Symbol"})
			body := testJSON(map[string]any{"parameters": parameters})
			invalid := appInstallationHTTP(t, server, http.MethodPut, body, 422)
			assert.Equal(t, map[string]any{"type": "Error", "id": "ValidationFailed"}, invalid["sys"])
			assert.Equal(t, "Validation error", invalid["message"])
			appInstallationHTTP(t, server, http.MethodGet, "", 404)
			appInstallationHTTP(t, server, http.MethodPut, testJSON(map[string]any{"parameters": map[string]any{"secret": "initial"}}), 200)
			appInstallationHTTP(t, server, http.MethodPut, body, 422)
			assert.Equal(t, map[string]any{"secret": "*******"}, appInstallationHTTP(t, server, http.MethodGet, "", 200)["parameters"])
			stored, present := server.StoredAppInstallation("space", "environment", "app")
			require.True(t, present)
			assert.JSONEq(t, testJSON(map[string]any{"secret": "initial"}), string(stored.Parameters))
		})
	}
}

func TestAppInstallationServerRequiredAndVisibility(t *testing.T) {
	t.Parallel()

	declaration := cm.AppDefinitionParameter{ID: "secret", Name: "Secret", Type: "Secret", Required: cm.NewOptBool(true)}
	server := appInstallationServer(t, declaration)
	appInstallationHTTP(t, server, http.MethodPut, testJSON(map[string]any{"parameters": map[string]any{"secret": ""}}), 200)
	assert.NotContains(t, appInstallationHTTP(t, server, http.MethodPut, testJSON(map[string]any{}), 200), "parameters")
	assert.NotContains(t, appInstallationHTTP(t, server, http.MethodGet, "", 200), "parameters")
	appInstallationHTTP(t, server, http.MethodPut, testJSON(map[string]any{"parameters": map[string]any{"secret": "hidden"}}), 200)

	for _, step := range []struct{ kind, expected string }{
		{"Symbol", "hidden"}, {"Secret", "******"}, {"", "hidden"},
	} {
		var declarations []cm.AppDefinitionParameter

		if step.kind != "" {
			declaration.Type = step.kind
			declarations = []cm.AppDefinitionParameter{declaration}
		}

		server.SetAppDefinition("org", "app", cm.AppDefinitionData{Name: "App", Parameters: cm.NewOptAppDefinitionParameters(cm.AppDefinitionParameters{Installation: declarations})})
		assert.Equal(t, map[string]any{"secret": step.expected}, appInstallationHTTP(t, server, http.MethodGet, "", 200)["parameters"])
	}
}

func TestAppInstallationServerFreeFormParameters(t *testing.T) {
	t.Parallel()

	for _, declarationsPresent := range []bool{false, true} {
		name := "no declarations"
		if declarationsPresent {
			name = "empty declarations"
		}

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server := appInstallationServer(t, []cm.AppDefinitionParameter{}...)
			if !declarationsPresent {
				server.SetAppDefinition("org", "app", cm.AppDefinitionData{Name: "App"})
			}

			parameters := map[string]any{"string": "plain", "number": json.Number("42"), "boolean": true, "null": nil, "array": []any{json.Number("1"), "two"}, "object": map[string]any{"x": "y"}}
			body := testJSON(map[string]any{"parameters": parameters})
			assert.Equal(t, parameters, appInstallationHTTP(t, server, http.MethodPut, body, 200)["parameters"])
			assert.Equal(t, parameters, appInstallationHTTP(t, server, http.MethodGet, "", 200)["parameters"])
			assert.Equal(t, parameters, appInstallationHTTP(t, server, http.MethodPut, body, 200)["parameters"])
		})
	}
}
