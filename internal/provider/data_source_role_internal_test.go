package provider

import (
	"net/http"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoleDataSourceResponseOrderWarningPaths(t *testing.T) {
	t.Parallel()

	// Unknown effects fail decoding; typed values exercise projector warning paths.
	firstRole := cm.Role{Sys: cm.NewRoleSys("space", "item-a"), Name: "First", Permissions: cm.RolePermissions{}, Policies: []cm.RolePoliciesItem{}}
	warningRole := cm.Role{Sys: cm.NewRoleSys("space", "item-b"), Name: "Second", Permissions: cm.RolePermissions{}, Policies: []cm.RolePoliciesItem{{Effect: cm.RolePoliciesItemEffect("future"), Actions: cm.NewStringRolePoliciesItemActions("all")}}}

	_, singularDiagnostics := newRoleDataSourceItem(t.Context(), warningRole, path.Empty())
	require.Len(t, singularDiagnostics.Warnings(), 1)
	assertDiagnosticPath(t, singularDiagnostics.Warnings()[0], path.Root("policies").AtListIndex(0).AtName("effect"))

	items, diagnostics := projectRoles(t.Context(), []cm.Role{warningRole, firstRole})
	require.False(t, diagnostics.HasError(), diagnostics)
	require.Len(t, diagnostics.Warnings(), 1)
	assertDiagnosticPath(t, diagnostics.Warnings()[0], path.Root("roles").AtListIndex(0).AtName("policies").AtListIndex(0).AtName("effect"))
	require.Len(t, items, 2)
	assert.Equal(t, "item-b", items[0].RoleID.ValueString())
	assert.Equal(t, "item-a", items[1].RoleID.ValueString())
	assert.True(t, items[0].Policies.Elements()[0].Value().Effect.IsNull())
}

//nolint:forcetypeassert // Fixture mutations target independently defined object shapes.
func TestRoleDataSourceDecoderRejectsUnsupportedEffect(t *testing.T) {
	t.Parallel()
	body := mutateTestJSON(discoveryFixture(t, "role"), func(document map[string]any) { document["policies"].([]any)[0].(map[string]any)["effect"] = "future" })
	response := discoveryReadTest(t.Context(), t, NewRolesDataSource, map[string]any{"space_id": "space"}, roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return discoveryHTTPResponse(request, 200, discoveryPage(0, 100, 1, body)), nil
	}))
	require.True(t, response.Diagnostics.HasError())
	assert.Contains(t, response.Diagnostics[0].Detail(), "invalid value: future")
	assert.True(t, response.State.Raw.IsNull())
}

func TestRoleDataSourceProjectionWarningsFromTypedIrregularResponse(t *testing.T) {
	t.Parallel()

	role := cm.Role{
		Sys:         cm.NewRoleSys("space", "item"),
		Name:        "irregular",
		Permissions: cm.RolePermissions{"ContentModel": cm.RolePermissionsItem{}},
		Policies: []cm.RolePoliciesItem{{
			Effect:  cm.RolePoliciesItemEffect("future"),
			Actions: cm.RolePoliciesItemActions{},
		}},
	}
	root := path.Root("roles").AtListIndex(2)
	item, diagnostics := newRoleDataSourceItem(t.Context(), role, root)
	require.False(t, diagnostics.HasError(), diagnostics)
	require.Len(t, diagnostics.Warnings(), 3)
	assertDiagnosticPath(t, diagnostics.Warnings()[0], root.AtName("permissions").AtMapKey("ContentModel"))
	assertDiagnosticPath(t, diagnostics.Warnings()[1], root.AtName("policies").AtListIndex(0).AtName("effect"))
	assertDiagnosticPath(t, diagnostics.Warnings()[2], root.AtName("policies").AtListIndex(0).AtName("actions"))
	assert.True(t, item.Permissions.Elements()["ContentModel"].IsNull())
	assert.True(t, item.Policies.Elements()[0].Value().Actions.IsNull())
	assert.True(t, item.Policies.Elements()[0].Value().Effect.IsNull())
	assert.True(t, item.Description.IsNull())
	assert.True(t, item.Policies.Elements()[0].Value().Constraint.IsNull())
}

//nolint:forcetypeassert // Fixture mutations target independently defined object shapes.
func TestRolesDataSourceRejectsMismatchedSpaceWithRoleID(t *testing.T) {
	t.Parallel()

	body := mutateTestJSON(discoveryFixture(t, "role"), func(document map[string]any) {
		document["sys"].(map[string]any)["space"].(map[string]any)["sys"].(map[string]any)["id"] = "other-space"
	})
	response := discoveryReadTest(t.Context(), t, NewRolesDataSource, map[string]any{"space_id": "space"}, roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return discoveryHTTPResponse(request, 200, discoveryPage(0, 100, 1, body)), nil
	}))
	require.True(t, response.Diagnostics.HasError())
	assert.Contains(t, response.Diagnostics[0].Detail(), `Role "item-b"`)
	assert.True(t, response.State.Raw.IsNull())
}

func TestRolesDataSourceStableDuplicateIDs(t *testing.T) {
	t.Parallel()

	body := discoveryFixture(t, "role")
	first := mutateTestJSON(body, func(document map[string]any) { document["name"] = "First arrival" })
	second := mutateTestJSON(body, func(document map[string]any) { document["name"] = "Second arrival" })

	response := discoveryReadTest(t.Context(), t, NewRolesDataSource, map[string]any{"space_id": "space"}, roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.URL.Query().Get("skip") == "0" {
			return discoveryHTTPResponse(request, 200, discoveryPage(0, 1, 2, first)), nil
		}

		return discoveryHTTPResponse(request, 200, discoveryPage(1, 1, 2, second)), nil
	}))
	require.False(t, response.Diagnostics.HasError(), response.Diagnostics)

	var data RolesDataSourceModel
	require.False(t, response.State.Get(t.Context(), &data).HasError())
	require.Len(t, data.Roles, 2)
	assert.Equal(t, "First arrival", data.Roles[0].Name.ValueString())
	assert.Equal(t, "Second arrival", data.Roles[1].Name.ValueString())
}

func assertDiagnosticPath(t *testing.T, diagnostic diag.Diagnostic, expected path.Path) {
	t.Helper()

	withPath, ok := diagnostic.(diag.DiagnosticWithPath)
	require.True(t, ok)
	assert.Equal(t, expected, withPath.Path())
}
