package cmtesting_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
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
				roleHTTP(t, server, http.MethodDelete, "/spaces/space/roles/role", "", 0, http.StatusNoContent)
			}

			body := testJSON(map[string]any{
				"name": "Replacement", "permissions": map[string]any{}, "policies": []any{},
			})
			response := roleHTTP(t, server, http.MethodPut, "/spaces/space/roles/role", body, 0, http.StatusNotFound)
			assert.Equal(t, map[string]any{"type": "Error", "id": "NotFound"}, response["sys"])
			roleHTTP(t, server, http.MethodGet, "/spaces/space/roles/role", "", 0, http.StatusNotFound)
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
	response := roleHTTP(t, server, http.MethodPut, "/spaces/space/roles/role", body, 0, http.StatusOK)
	assert.Equal(t, "Updated", response["name"])
	assert.Equal(t, response, roleHTTP(t, server, http.MethodGet, "/spaces/space/roles/role", "", 0, http.StatusOK))
}

func TestRoleDescriptionAndVersionLifecycle(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")
	created := roleHTTP(t, server, http.MethodPost, "/spaces/space/roles", testJSON(map[string]any{
		"name": "Original", "description": "Original description", "permissions": map[string]any{}, "policies": []any{},
	}), 0, http.StatusCreated)
	assert.Equal(t, "Original description", created["description"])
	sys, ok := created["sys"].(map[string]any)
	require.True(t, ok)
	roleID, ok := sys["id"].(string)
	require.True(t, ok)
	require.NotEmpty(t, roleID)
	assert.Equal(t, json.Number("0"), sys["version"])

	path := "/spaces/space/roles/" + roleID
	assert.Equal(t, created, roleHTTP(t, server, http.MethodGet, path, "", 0, http.StatusOK))

	for version, description := range []any{"Changed description", "", nil} {
		updated := roleHTTP(t, server, http.MethodPut, path, testJSON(map[string]any{
			"name": "Updated", "description": description, "permissions": map[string]any{}, "policies": []any{},
		}), version, http.StatusOK)
		assert.Contains(t, updated, "description")
		assert.Equal(t, description, updated["description"])
		sys, ok := updated["sys"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, json.Number(strconv.Itoa(version+1)), sys["version"])

		stale := roleHTTP(t, server, http.MethodPut, path, testJSON(map[string]any{
			"name": "Stale", "description": "Stale description", "permissions": map[string]any{}, "policies": []any{},
		}), version, http.StatusConflict)
		assert.Equal(t, map[string]any{"type": "Error", "id": "VersionMismatch"}, stale["sys"])
		assert.Equal(t, updated, roleHTTP(t, server, http.MethodGet, path, "", 0, http.StatusOK))
	}

	updated := roleHTTP(t, server, http.MethodPut, path, testJSON(map[string]any{
		"name": "Updated", "permissions": map[string]any{}, "policies": []any{},
	}), 3, http.StatusOK)
	assert.NotContains(t, updated, "description")
	assert.Equal(t, updated, roleHTTP(t, server, http.MethodGet, path, "", 0, http.StatusOK))
}

func roleHTTP(t *testing.T, server *cmt.Server, method, path, body string, version, status int) map[string]any {
	t.Helper()

	request := httptest.NewRequestWithContext(t.Context(), method, path, strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+cmt.ValidAccessToken)
	request.Header.Set("Content-Type", "application/vnd.contentful.management.v1+json")
	request.Header.Set("X-Contentful-Version", strconv.Itoa(version))

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
