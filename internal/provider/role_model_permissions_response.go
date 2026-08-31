package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewPermissionsMapValueFromResponse(ctx context.Context, path path.Path, permissions cm.RolePermissions) (TypedMap[TypedList[types.String]], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	permissionsValuesMap := make(map[string]TypedList[types.String], len(permissions))

	for permission, item := range permissions {
		path := path.AtMapKey(permission)

		permissionValuesListValue, permissionValuesListValueDiags := NewPermissionValuesListValueFromResponse(ctx, path, item)
		diags.Append(permissionValuesListValueDiags...)

		permissionsValuesMap[permission] = permissionValuesListValue
	}

	permissionsMapValue := NewTypedMap(permissionsValuesMap)

	return permissionsMapValue, diags
}

func NewPermissionValuesListValueFromResponse(_ context.Context, path path.Path, item cm.RolePermissionsItem) (TypedList[types.String], diag.Diagnostics) {
	switch item.Type {
	case cm.StringRolePermissionsItem:
		permissionValues := make([]types.String, 1)
		permissionValues[0] = types.StringValue(item.String)

		permissionValuesListValue := NewTypedList(permissionValues)

		return permissionValuesListValue, nil

	case cm.StringArrayRolePermissionsItem:
		return NewTypedListFromStringSlice(item.StringArray), nil
	}

	diags := diag.Diagnostics{}
	diags.AddAttributeWarning(path, "Unsupported permission value", "Contentful returned an unsupported permission value representation. Terraform state retains a known null list; a later request conversion will reject it until configured with a supported representation.")

	return NewTypedListNull[types.String](), diags
}
