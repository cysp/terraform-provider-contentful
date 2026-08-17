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

func TestRoleActionsFromResponse(t *testing.T) {
	t.Parallel()

	assertRoleActionsFromResponse(t, cm.NewStringRolePermissionsItem("read"), NewPermissionActionsListValueFromResponse, false)
	assertRoleActionsFromResponse(t, cm.RolePermissionsItem{}, NewPermissionActionsListValueFromResponse, true)
	assertRoleActionsFromResponse(t, cm.NewStringRolePoliciesItemActions("read"), NewPolicyActionsListValueFromResponse, false)
	assertRoleActionsFromResponse(t, cm.RolePoliciesItemActions{}, NewPolicyActionsListValueFromResponse, true)
}

func TestRolePolicyResponsePreservesUnknownEffect(t *testing.T) {
	t.Parallel()

	valuePath := path.Root("policies").AtListIndex(0)
	actual, diags := NewPoliciesValueFromResponse(t.Context(), valuePath, cm.RolePoliciesItem{
		Effect:  cm.RolePoliciesItemEffect("future"),
		Actions: cm.NewStringRolePoliciesItemActions("read"),
	})

	assert.False(t, diags.HasError())
	require.Len(t, diags.Warnings(), 1)
	diagnostic, ok := diags.Warnings()[0].(diag.DiagnosticWithPath)
	require.True(t, ok)
	assert.Equal(t, valuePath.AtName("effect"), diagnostic.Path())
	assert.True(t, actual.Value().Effect.IsNull())
}

func assertRoleActionsFromResponse[T any](
	t *testing.T,
	input T,
	convert func(context.Context, path.Path, T) (TypedList[types.String], diag.Diagnostics),
	expectError bool,
) {
	t.Helper()

	actual, diags := convert(t.Context(), path.Root("actions"), input)

	if expectError {
		assert.False(t, diags.HasError())
		require.Len(t, diags.Warnings(), 1)
		diagnostic, ok := diags.Warnings()[0].(diag.DiagnosticWithPath)
		require.True(t, ok)
		assert.Equal(t, path.Root("actions"), diagnostic.Path())
		assert.True(t, actual.IsNull())

		return
	}

	assert.Empty(t, diags)
	assert.Equal(t, NewTypedList([]types.String{types.StringValue("read")}), actual)
}
