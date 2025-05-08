package provider

import (
	"context"
	"slices"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ToRoleFieldsPermissions(ctx context.Context, path path.Path, permissions map[string][]types.String) (cm.RoleFieldsPermissions, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if permissions == nil {
		return nil, diags
	}

	rolePermissionsItems := make(cm.RoleFieldsPermissions, len(permissions))

	for key, permissionsValueElement := range permissions {
		path := path.AtMapKey(key)

		permissionsItem, permissionsItemDiags := ToRoleFieldsPermissionsItem(ctx, path, permissionsValueElement)
		diags.Append(permissionsItemDiags...)

		rolePermissionsItems[key] = permissionsItem
	}

	return rolePermissionsItems, diags
}

func ToRoleFieldsPermissionsItem(ctx context.Context, _ path.Path, value []types.String) (cm.RoleFieldsPermissionsItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	actionStrings := make([]string, len(value))
	for i, action := range value {
		actionStrings[i] = action.ValueString()
	}

	if slices.Contains(actionStrings, "all") {
		return cm.RoleFieldsPermissionsItem{
			Type:   cm.StringRoleFieldsPermissionsItem,
			String: "all",
		}, diags
	}

	return cm.RoleFieldsPermissionsItem{
		Type:        cm.StringArrayRoleFieldsPermissionsItem,
		StringArray: actionStrings,
	}, diags
}
