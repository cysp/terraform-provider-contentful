package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewPermissionsMapValueFromResponse(ctx context.Context, path path.Path, permissions cm.RolePermissions) (map[string][]types.String, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	permissionsValuesMap := make(map[string][]types.String, len(permissions))

	for permission, item := range permissions {
		path := path.AtMapKey(permission)

		permissionActionsListValue, permissionActionsListValueDiags := NewPermissionActionsListValueFromResponse(ctx, path, item)
		diags.Append(permissionActionsListValueDiags...)

		permissionsValuesMap[permission] = permissionActionsListValue
	}

	return permissionsValuesMap, diags
}

func NewPermissionActionsListValueFromResponse(ctx context.Context, path path.Path, item cm.RolePermissionsItem) ([]types.String, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	switch item.Type {
	case cm.StringRolePermissionsItem:
		actionsValues := make([]types.String, 1)
		actionsValues[0] = types.StringValue(item.String)

		return actionsValues, diags

	case cm.StringArrayRolePermissionsItem:
		actionsValues := make([]types.String, len(item.StringArray))
		for i, action := range item.StringArray {
			actionsValues[i] = types.StringValue(action)
		}

		return actionsValues, diags
	}

	diags.AddAttributeError(path, "unexpected type for permission actions", "")

	return nil, diags
}
