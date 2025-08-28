package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewPermissionsMapValueFromResponse(ctx context.Context, path path.Path, permissions cm.RolePermissions) (TypedMap[TypedList[types.String]], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	permissionsValuesMap := make(map[string]TypedList[types.String], len(permissions))

	for permission, item := range permissions {
		path := path.AtMapKey(permission)

		permissionActionsListValue, permissionActionsListValueDiags := NewPermissionActionsListValueFromResponse(ctx, path, item)
		diags.Append(permissionActionsListValueDiags...)

		permissionsValuesMap[permission] = permissionActionsListValue
	}

	permissionsMapValue, permissionsListValueDiags := NewTypedMap(ctx, permissionsValuesMap)
	diags.Append(permissionsListValueDiags...)

	return permissionsMapValue, diags
}

func NewPermissionActionsListValueFromResponse(_ context.Context, path path.Path, item cm.RolePermissionsItem) (TypedList[types.String], diag.Diagnostics) {
	switch item.Type {
	case cm.StringRolePermissionsItem:
		actionsValues := make([]types.String, 1)
		actionsValues[0] = types.StringValue(item.String)

		actionsListValue := NewTypedList(actionsValues)

		return actionsListValue, nil

	case cm.StringArrayRolePermissionsItem:
		return NewTypedListFromStringSlice(item.StringArray), nil
	}

	diags := diag.Diagnostics{}
	diags.AddAttributeError(path, "unexpected type for permission actions", "")

	return NewTypedListUnknown[types.String](), diags
}
