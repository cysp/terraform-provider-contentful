package provider

import (
	"context"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func NewPermissionsMapValueFromResponse(ctx context.Context, path path.Path, permissions contentfulManagement.RolePermissions) (basetypes.MapValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	permissionsValuesMap := make(map[string]attr.Value, len(permissions))

	for permission, item := range permissions {
		path := path.AtMapKey(permission)

		permissionActionsListValue, permissionActionsListValueDiags := NewPermissionActionsListValueFromResponse(ctx, path, item)
		diags.Append(permissionActionsListValueDiags...)

		permissionsValuesMap[permission] = permissionActionsListValue
	}

	permissionsMapValue, permissionsListValueDiags := basetypes.NewMapValue(util.NewEmptyListMust(types.String{}.Type(ctx)).Type(ctx), permissionsValuesMap)
	diags.Append(permissionsListValueDiags...)

	return permissionsMapValue, diags
}

func NewPermissionActionsListValueFromResponse(ctx context.Context, path path.Path, item contentfulManagement.RolePermissionsItem) (basetypes.ListValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	switch item.Type {
	case contentfulManagement.StringRolePermissionsItem:
		actionsValues := make([]attr.Value, 1)
		actionsValues[0] = types.StringValue(item.String)

		actionsListValue, actionsListValueDiags := basetypes.NewListValue(types.String{}.Type(ctx), actionsValues)
		diags.Append(actionsListValueDiags...)

		return actionsListValue, diags

	case contentfulManagement.StringArrayRolePermissionsItem:
		actionsListValue, actionsListValueDiags := basetypes.NewListValueFrom(ctx, types.String{}.Type(ctx), item.StringArray)
		diags.Append(actionsListValueDiags...)

		return actionsListValue, diags
	}

	diags.AddAttributeError(path, "unexpected type for permission actions", "")

	return basetypes.ListValue{}, diags
}