package provider

import (
	"slices"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ToRoleDataPermissions(path path.Path, permissions TypedMap[TypedList[types.String]]) (cm.RoleDataPermissions, diag.Diagnostics) {
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

		permissionsItem, permissionsItemDiags := ToRoleDataPermissionsItem(itemPath, permissionsValues[key])
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

//nolint:dupl // Permission values and policy actions require distinct domain terminology and diagnostics.
func ToRoleDataPermissionsItem(path path.Path, value TypedList[types.String]) (cm.RoleDataPermissionsItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	switch {
	case value.IsUnknown():
		diags.AddAttributeError(
			path,
			"Unexpected unknown permission values",
			"Permission values must be known before they can be sent to Contentful.",
		)

		return cm.RoleDataPermissionsItem{}, diags
	case value.IsNull():
		diags.AddAttributeError(
			path,
			"Unexpected null permission values",
			"Permission values cannot be null.",
		)

		return cm.RoleDataPermissionsItem{}, diags
	}

	permissionValues, valueDiags := knownStringListElements(path, value.Elements())
	diags.Append(valueDiags...)

	if valueDiags.HasError() {
		return cm.RoleDataPermissionsItem{}, diags
	}

	diags.Append(validateRolePermissionValues(path, len(permissionValues), permissionValues)...)

	if diags.HasError() {
		return cm.RoleDataPermissionsItem{}, diags
	}

	if len(permissionValues) == 1 && permissionValues[0] == "all" {
		return cm.RoleDataPermissionsItem{
			Type:   cm.StringRoleDataPermissionsItem,
			String: "all",
		}, diags
	}

	return cm.RoleDataPermissionsItem{
		Type:        cm.StringArrayRoleDataPermissionsItem,
		StringArray: permissionValues,
	}, diags
}

func validateRolePermissionValues(path path.Path, valueCount int, values []string) diag.Diagnostics {
	diags := diag.Diagnostics{}

	if valueCount != 1 && slices.Contains(values, "all") {
		diags.AddAttributeError(
			path,
			"Invalid permission values",
			`"all" must be specified by itself. Remove "all" or the other permission values from this list.`,
		)
	}

	return diags
}
