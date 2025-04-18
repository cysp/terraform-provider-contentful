package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewPermissionsMapValueFromResponse(ctx context.Context, path path.Path, permissions cm.RolePermissions) (types.Map, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	permissionsValuesMap := make(map[string]attr.Value, len(permissions))

	for permission, item := range permissions {
		path := path.AtMapKey(permission)

		permissionActionsListValue, permissionActionsListValueDiags := NewPermissionActionsListValueFromResponse(ctx, path, item)
		diags.Append(permissionActionsListValueDiags...)

		permissionsValuesMap[permission] = permissionActionsListValue
	}

	permissionsMapValue, permissionsListValueDiags := types.MapValue(types.ListType{ElemType: types.StringType}, permissionsValuesMap)
	diags.Append(permissionsListValueDiags...)

	return permissionsMapValue, diags
}

func NewPermissionActionsListValueFromResponse(ctx context.Context, path path.Path, item cm.RolePermissionsItem) (types.List, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	switch item.Type {
	case cm.StringRolePermissionsItem:
		actionsValues := make([]attr.Value, 1)
		actionsValues[0] = types.StringValue(item.String)

		actionsListValue, actionsListValueDiags := types.ListValue(types.String{}.Type(ctx), actionsValues)
		diags.Append(actionsListValueDiags...)

		return actionsListValue, diags

	case cm.StringArrayRolePermissionsItem:
		actionsListValue, actionsListValueDiags := types.ListValueFrom(ctx, types.String{}.Type(ctx), item.StringArray)
		diags.Append(actionsListValueDiags...)

		return actionsListValue, diags
	}

	diags.AddAttributeError(path, "unexpected type for permission actions", "")

	return types.List{}, diags
}
