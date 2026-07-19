package provider

import (
	"context"
	"slices"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ToRoleDataPermissions(ctx context.Context, path path.Path, permissions TypedMap[TypedList[types.String]]) (cm.RoleDataPermissions, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if permissions.IsNull() || permissions.IsUnknown() {
		if permissions.IsUnknown() {
			diags.AddAttributeError(path, "Unexpected unknown permissions", "Permissions must be known before they can be sent to Contentful.")
		} else {
			diags.AddAttributeError(path, "Unexpected null permissions", "Permissions are required.")
		}

		return nil, diags
	}

	permissionsValues := permissions.Elements()

	rolePermissionsItems := make(cm.RoleDataPermissions, len(permissions.Elements()))

	for key, permissionsValueElement := range permissionsValues {
		path := path.AtMapKey(key)

		permissionsItem, permissionsItemDiags := ToRoleDataPermissionsItem(ctx, path, permissionsValueElement)
		diags.Append(permissionsItemDiags...)

		if !permissionsItemDiags.HasError() {
			rolePermissionsItems[key] = permissionsItem
		}
	}

	if diags.HasError() {
		return nil, diags
	}

	return rolePermissionsItems, diags
}

func ToRoleDataPermissionsItem(ctx context.Context, path path.Path, value TypedList[types.String]) (cm.RoleDataPermissionsItem, diag.Diagnostics) {
	actionStrings, diags := KnownStringListValues(
		ctx,
		value,
		path,
		"Unexpected unknown permission actions",
		"Permission actions must be known before they can be sent to Contentful.",
		"Unexpected null permission actions",
		"Permission actions cannot be null.",
	)

	if diags.HasError() {
		return cm.RoleDataPermissionsItem{}, diags
	}

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
