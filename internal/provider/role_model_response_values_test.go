package provider_test

import (
	"context"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRolePermissionValuesAndPolicyActionsFromResponse(t *testing.T) {
	t.Parallel()

	permissionPath := path.Root("permissions").AtMapKey("ContentModel")
	assertRoleStringListFromResponse(t, permissionPath, cm.NewStringRolePermissionsItem("all"), NewPermissionValuesListValueFromResponse, false)
	assertRoleStringListFromResponse(t, permissionPath, cm.RolePermissionsItem{}, NewPermissionValuesListValueFromResponse, true)

	policyActionsPath := path.Root("policies").AtListIndex(0).AtName("actions")
	assertRoleStringListFromResponse(t, policyActionsPath, cm.NewStringRolePoliciesItemActions("all"), NewPolicyActionsListValueFromResponse, false)
	assertRoleStringListFromResponse(t, policyActionsPath, cm.RolePoliciesItemActions{}, NewPolicyActionsListValueFromResponse, true)
}

func TestRolePolicyResponsePreservesUnknownEffect(t *testing.T) {
	t.Parallel()

	valuePath := path.Root("policies").AtListIndex(0)
	actual, diags := NewPoliciesValueFromResponse(t.Context(), valuePath, cm.RolePoliciesItem{
		Effect:  cm.RolePoliciesItemEffect("future"),
		Actions: cm.NewStringRolePoliciesItemActions("all"),
	})

	assert.False(t, diags.HasError())
	require.Len(t, diags.Warnings(), 1)
	diagnostic, ok := diags.Warnings()[0].(diag.DiagnosticWithPath)
	require.True(t, ok)
	assert.Equal(t, valuePath.AtName("effect"), diagnostic.Path())
	assert.True(t, actual.Value().Effect.IsNull())
}

func assertRoleStringListFromResponse[T any](
	t *testing.T,
	valuePath path.Path,
	input T,
	convert func(context.Context, path.Path, T) (TypedList[types.String], diag.Diagnostics),
	expectWarning bool,
) {
	t.Helper()

	actual, diags := convert(t.Context(), valuePath, input)

	if expectWarning {
		assert.False(t, diags.HasError())
		require.Len(t, diags.Warnings(), 1)
		diagnostic, ok := diags.Warnings()[0].(diag.DiagnosticWithPath)
		require.True(t, ok)
		assert.Equal(t, valuePath, diagnostic.Path())
		assert.True(t, actual.IsNull())

		return
	}

	assert.Empty(t, diags)
	assert.Equal(t, NewTypedList([]types.String{types.StringValue("all")}), actual)
}
