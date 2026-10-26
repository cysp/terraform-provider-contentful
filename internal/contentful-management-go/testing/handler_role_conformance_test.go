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

func TestRoleUpdateDoesNotCreateMissingRole(t *testing.T) {
	t.Parallel()

	for _, deleted := range []bool{false, true} {
		name := "absent"
		if deleted {
			name = "deleted"
		}

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer()
			require.NoError(t, err)
			server.RegisterSpaceEnvironment("space", "master")

			if deleted {
				server.SetRole("space", "role", cm.RoleData{Name: "Original"})
				roleHTTP(t, server, http.MethodDelete, "", http.StatusNoContent)
			}

			body := testJSON(map[string]any{
				"name": "Replacement", "permissions": map[string]any{}, "policies": []any{},
			})
			response := roleHTTP(t, server, http.MethodPut, body, http.StatusNotFound)
			assert.Equal(t, map[string]any{"type": "Error", "id": "NotFound"}, response["sys"])
			roleHTTP(t, server, http.MethodGet, "", http.StatusNotFound)
		})
	}
}

func TestRoleUpdateChangesExistingRole(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")
	server.SetRole("space", "role", cm.RoleData{Name: "Original"})

	body := testJSON(map[string]any{
		"name": "Updated", "permissions": map[string]any{}, "policies": []any{},
	})
	response := roleHTTP(t, server, http.MethodPut, body, http.StatusOK)
	assert.Equal(t, "Updated", response["name"])
	assert.Equal(t, response, roleHTTP(t, server, http.MethodGet, "", http.StatusOK))
}

func roleHTTP(t *testing.T, server *cmt.Server, method, body string, status int) map[string]any {
	t.Helper()

	request := httptest.NewRequestWithContext(t.Context(), method, "/spaces/space/roles/role", strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+cmt.ValidAccessToken)
	request.Header.Set("Content-Type", "application/vnd.contentful.management.v1+json")
	request.Header.Set("X-Contentful-Version", "0")

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
