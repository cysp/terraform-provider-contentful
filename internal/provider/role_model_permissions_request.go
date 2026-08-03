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

	switch {
	case permissions.IsUnknown():
		diags.AddAttributeError(
			path,
			"Unexpected unknown permissions",
			"Permissions must be known before they can be sent to Contentful.",
		)

		return nil, diags
	case permissions.IsNull():
		diags.AddAttributeError(
			path,
			"Unexpected null permissions",
			"Permissions are required.",
		)

		return nil, diags
	}

	permissionsValues := permissions.Elements()

	keys := make([]string, 0, len(permissionsValues))
	for key := range permissionsValues {
		keys = append(keys, key)
	}

	slices.Sort(keys)

	rolePermissionsItems := make(cm.RoleDataPermissions, len(permissionsValues))

	for _, key := range keys {
		itemPath := path.AtMapKey(key)

		permissionsItem, permissionsItemDiags := ToRoleDataPermissionsItem(ctx, itemPath, permissionsValues[key])
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

func ToRoleDataPermissionsItem(_ context.Context, path path.Path, value TypedList[types.String]) (cm.RoleDataPermissionsItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	switch {
	case value.IsUnknown():
		diags.AddAttributeError(
			path,
			"Unexpected unknown permission actions",
			"Permission actions must be known before they can be sent to Contentful.",
		)

		return cm.RoleDataPermissionsItem{}, diags
	case value.IsNull():
		diags.AddAttributeError(
			path,
			"Unexpected null permission actions",
			"Permission actions cannot be null.",
		)

		return cm.RoleDataPermissionsItem{}, diags
	}

	actionStrings, actionDiags := knownStringListElements(path, value.Elements())
	diags.Append(actionDiags...)

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
