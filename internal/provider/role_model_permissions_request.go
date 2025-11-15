package provider

import (
	"context"
	"slices"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ToRoleDataPermissions(ctx context.Context, path path.Path, permissions TypedMap[TypedList[types.String]]) (cm.RoleDataPermissions, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if permissions.IsUnknown() {
		return nil, diags
	}

	permissionsValues := permissions.Elements()

	rolePermissionsItems := make(cm.RoleDataPermissions, len(permissions.Elements()))

	for key, permissionsValueElement := range permissionsValues {
		path := path.AtMapKey(key)

		permissionsItem, permissionsItemDiags := ToRoleDataPermissionsItem(ctx, path, permissionsValueElement)
		diags.Append(permissionsItemDiags...)

		rolePermissionsItems[key] = permissionsItem
	}

	return rolePermissionsItems, diags
}

func ToRoleDataPermissionsItem(ctx context.Context, _ path.Path, value TypedList[types.String]) (cm.RoleDataPermissionsItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	actionStrings := make([]string, len(value.Elements()))
	diags.Append(tfsdk.ValueAs(ctx, value, &actionStrings)...)

	if slices.Contains(actionStrings, "all") {
		return cm.RoleDataPermissionsItem{
			Type:   cm.StringRoleDataPermissionsItem,
			String: "all",
		}, diags
	}

	return cm.RoleDataPermissionsItem{
		Type:        cm.StringArrayRoleDataPermissionsItem,
		StringArray: actionStrings,
	}, diags
}
