//nolint:dupl
package provider

import (
	"context"
	"slices"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ToCreateRoleReqPermissions(ctx context.Context, path path.Path, permissions types.Map) (cm.CreateRoleReqPermissions, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if permissions.IsUnknown() {
		return nil, diags
	}

	permissionsValues := make(map[string]types.List, len(permissions.Elements()))
	diags.Append(permissions.ElementsAs(ctx, &permissionsValues, false)...)

	rolePermissionsItems := make(cm.CreateRoleReqPermissions, len(permissions.Elements()))

	for key, permissionsValueElement := range permissionsValues {
		path := path.AtMapKey(key)

		permissionsItem, permissionsItemDiags := ToCreateRoleReqPermissionsItem(ctx, path, permissionsValueElement)
		diags.Append(permissionsItemDiags...)

		rolePermissionsItems[key] = permissionsItem
	}

	return rolePermissionsItems, diags
}

func ToCreateRoleReqPermissionsItem(ctx context.Context, _ path.Path, value types.List) (cm.CreateRoleReqPermissionsItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	actionStrings := make([]string, len(value.Elements()))
	diags.Append(value.ElementsAs(ctx, &actionStrings, false)...)

	if slices.Contains(actionStrings, "all") {
		return cm.CreateRoleReqPermissionsItem{
			Type:   cm.StringCreateRoleReqPermissionsItem,
			String: "all",
		}, diags
	}

	return cm.CreateRoleReqPermissionsItem{
		Type:        cm.StringArrayCreateRoleReqPermissionsItem,
		StringArray: actionStrings,
	}, diags
}
