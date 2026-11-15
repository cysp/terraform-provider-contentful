package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewPermissionsMapValueFromResponse(ctx context.Context, path path.Path, permissions cm.RolePermissions) (map[string][]types.String, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	permissionsValuesMap := make(map[string][]types.String, len(permissions))

	for permission, item := range permissions {
		path := path.AtMapKey(permission)

		permissionValuesListValue, permissionValuesListValueDiags := NewPermissionValuesListValueFromResponse(ctx, path, item)
		diags.Append(permissionValuesListValueDiags...)

		permissionsValuesMap[permission] = permissionValuesListValue
	}

	return permissionsValuesMap, diags
}

func NewPermissionValuesListValueFromResponse(_ context.Context, path path.Path, item cm.RolePermissionsItem) ([]types.String, diag.Diagnostics) {
	switch item.Type {
	case cm.StringRolePermissionsItem:
		return []types.String{types.StringValue(item.String)}, nil

	case cm.StringArrayRolePermissionsItem:
		values := make([]types.String, len(item.StringArray))
		for index, value := range item.StringArray {
			values[index] = types.StringValue(value)
		}

		return values, nil
	}

	diags := diag.Diagnostics{}
	diags.AddAttributeWarning(path, "Unsupported permission value", "Contentful returned an unsupported permission value representation. Terraform state retains a known null list; a later request conversion will reject it until configured with a supported representation.")

	return nil, diags
}
