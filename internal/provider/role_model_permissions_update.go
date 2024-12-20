//nolint:dupl
package provider

import (
	"context"
	"slices"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/tf"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ToUpdateRoleReqPermissions(ctx context.Context, path path.Path, permissions types.Map) (contentfulManagement.UpdateRoleReqPermissions, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if permissions.IsUnknown() {
		return nil, diags
	}

	permissionsValues := make(map[string]types.List, len(permissions.Elements()))
	diags.Append(permissions.ElementsAs(ctx, &permissionsValues, false)...)

	rolePermissionsItems := make(contentfulManagement.UpdateRoleReqPermissions, len(permissions.Elements()))

	for key, permissionsValueElement := range permissionsValues {
		path := path.AtMapKey(key)

		permissionsItem, permissionsItemDiags := ToUpdateRoleReqPermissionsItem(ctx, path, permissionsValueElement)
		diags.Append(permissionsItemDiags...)

		rolePermissionsItems[key] = permissionsItem
	}

	return rolePermissionsItems, diags
}

func ToUpdateRoleReqPermissionsItem(ctx context.Context, _ path.Path, value types.List) (contentfulManagement.UpdateRoleReqPermissionsItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	actionStrings, acactionStringsDiags := tf.KnownAndPresentStringValues(ctx, value)
	diags.Append(acactionStringsDiags...)

	if slices.Contains(actionStrings, "all") {
		return contentfulManagement.UpdateRoleReqPermissionsItem{
			Type:   contentfulManagement.StringUpdateRoleReqPermissionsItem,
			String: "all",
		}, diags
	}

	return contentfulManagement.UpdateRoleReqPermissionsItem{
		Type:        contentfulManagement.StringArrayUpdateRoleReqPermissionsItem,
		StringArray: actionStrings,
	}, diags
}
