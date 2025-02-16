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

func ToUpdateRoleReqPermissions(ctx context.Context, path path.Path, permissions types.Map) (cm.UpdateRoleReqPermissions, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if permissions.IsUnknown() {
		return nil, diags
	}

	permissionsValues := make(map[string]types.List, len(permissions.Elements()))
	diags.Append(permissions.ElementsAs(ctx, &permissionsValues, false)...)

	rolePermissionsItems := make(cm.UpdateRoleReqPermissions, len(permissions.Elements()))

	for key, permissionsValueElement := range permissionsValues {
		path := path.AtMapKey(key)

		permissionsItem, permissionsItemDiags := ToUpdateRoleReqPermissionsItem(ctx, path, permissionsValueElement)
		diags.Append(permissionsItemDiags...)

		rolePermissionsItems[key] = permissionsItem
	}

	return rolePermissionsItems, diags
}

func ToUpdateRoleReqPermissionsItem(ctx context.Context, _ path.Path, value types.List) (cm.UpdateRoleReqPermissionsItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	actionStrings := make([]string, len(value.Elements()))
	diags.Append(value.ElementsAs(ctx, &actionStrings, false)...)

	if slices.Contains(actionStrings, "all") {
		return cm.UpdateRoleReqPermissionsItem{
			Type:   cm.StringUpdateRoleReqPermissionsItem,
			String: "all",
		}, diags
	}

	return cm.UpdateRoleReqPermissionsItem{
		Type:        cm.StringArrayUpdateRoleReqPermissionsItem,
		StringArray: actionStrings,
	}, diags
}
