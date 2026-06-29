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
)

func TestRoleActionsFromResponse(t *testing.T) {
	t.Parallel()

	assertRoleActionsFromResponse(t, cm.NewStringRolePermissionsItem("read"), NewPermissionActionsListValueFromResponse, false)
	assertRoleActionsFromResponse(t, cm.RolePermissionsItem{}, NewPermissionActionsListValueFromResponse, true)
	assertRoleActionsFromResponse(t, cm.NewStringRolePoliciesItemActions("read"), NewPolicyActionsListValueFromResponse, false)
	assertRoleActionsFromResponse(t, cm.RolePoliciesItemActions{}, NewPolicyActionsListValueFromResponse, true)
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
		assert.True(t, diags.HasError())
		assert.True(t, actual.IsUnknown())

		return
	}

	assert.Empty(t, diags)
	assert.Equal(t, NewTypedList([]types.String{types.StringValue("read")}), actual)
}
