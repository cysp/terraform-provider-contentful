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

func TestAppDefinitionEndpointsRejectAnotherOrganization(t *testing.T) {
	t.Parallel()

	for _, endpoint := range []struct {
		name, suffix string
		body         map[string]any
	}{
		{name: "app definition", body: map[string]any{"name": "Changed"}},
		{name: "resource provider", suffix: "/resource_provider", body: map[string]any{
			"sys": map[string]any{"id": "provider-a"}, "type": "function",
			"function": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Function", "id": "changed"}},
		}},
		{name: "signing secret", suffix: "/signing_secret", body: map[string]any{"value": strings.Repeat("b", 64)}},
		{name: "resource type", suffix: "/resource_provider/resource_types/provider-a:article", body: map[string]any{
			"name": "Changed", "defaultFieldMapping": map[string]any{"title": "{ /changed }"},
		}},
	} {
		for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
			t.Run(endpoint.name+"/"+method, func(t *testing.T) {
				t.Parallel()
				server := appDefinitionScopeServer(t)
				path := "/organizations/org/app_definitions/app-a" + endpoint.suffix
				before := appDefinitionScopeHTTP(t, server, http.MethodGet, path, "", http.StatusOK)

				body := ""
				if method == http.MethodPut {
					body = testJSON(endpoint.body)
				}

				wrongOrganization := "/organizations/another-org/app_definitions/app-a" + endpoint.suffix
				appDefinitionScopeHTTP(t, server, method, wrongOrganization, body, http.StatusNotFound)
				assert.Equal(t, before, appDefinitionScopeHTTP(t, server, http.MethodGet, path, "", http.StatusOK))
			})
		}
	}
}

func TestResourceTypeCannotBeReadOrDeletedThroughAnotherAppDefinition(t *testing.T) {
	t.Parallel()

	for _, method := range []string{http.MethodGet, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			t.Parallel()
			server := appDefinitionScopeServer(t)
			originalPath := "/organizations/org/app_definitions/app-a/resource_provider/resource_types/provider-a:article"
			before := appDefinitionScopeHTTP(t, server, http.MethodGet, originalPath, "", http.StatusOK)

			appDefinitionScopeHTTP(t, server, method, "/organizations/org/app_definitions/app-b/resource_provider/resource_types/provider-a:article", "", http.StatusNotFound)
			assert.Equal(t, before, appDefinitionScopeHTTP(t, server, http.MethodGet, originalPath, "", http.StatusOK))
		})
	}
}

func TestAppDefinitionChildrenRequireTheirParent(t *testing.T) {
	t.Parallel()

	for _, suffix := range []string{"/resource_provider", "/signing_secret", "/resource_provider/resource_types/provider-a:article"} {
		t.Run(suffix, func(t *testing.T) {
			t.Parallel()
			server := appDefinitionScopeServer(t)
			parentPath := "/organizations/org/app_definitions/app-a"
			appDefinitionScopeHTTP(t, server, http.MethodDelete, parentPath, "", http.StatusNoContent)
			appDefinitionScopeHTTP(t, server, http.MethodGet, parentPath+suffix, "", http.StatusNotFound)
			appDefinitionScopeHTTP(t, server, http.MethodDelete, parentPath+suffix, "", http.StatusNotFound)
		})
	}
}

func appDefinitionScopeServer(t *testing.T) *cmt.Server {
	t.Helper()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)

	for _, parent := range []struct{ app, provider string }{{"app-a", "provider-a"}, {"app-b", "provider-b"}} {
		server.SetAppDefinition("org", parent.app, cm.AppDefinitionData{Name: parent.app})
		server.SetResourceProvider("org", parent.app, cm.ResourceProviderRequest{
			Sys: cm.NewResourceProviderRequestSys(parent.provider), Type: cm.ResourceProviderRequestTypeFunction, Function: cm.NewFunctionLink("function-" + parent.app),
		})
	}

	server.SetAppSigningSecret("org", "app-a", cm.AppSigningSecretRequestData{Value: strings.Repeat("a", 64)})
	appDefinitionScopeHTTP(t, server, http.MethodPut, "/organizations/org/app_definitions/app-a/resource_provider/resource_types/provider-a:article", testJSON(map[string]any{
		"name": "Article", "defaultFieldMapping": map[string]any{"title": "{ /name }"},
	}), http.StatusCreated)

	return server
}

func appDefinitionScopeHTTP(t *testing.T, server *cmt.Server, method, path, body string, status int) map[string]any {
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
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &result))

	return result
}
