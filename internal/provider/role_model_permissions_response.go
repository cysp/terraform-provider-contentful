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

		permissionActionsListValue, permissionActionsListValueDiags := NewPermissionActionsListValueFromResponse(ctx, path, item)
		diags.Append(permissionActionsListValueDiags...)

		permissionsValuesMap[permission] = permissionActionsListValue
	}

	permissionsMapValue := NewTypedMap(permissionsValuesMap)

	return permissionsMapValue, diags
}

func NewPermissionActionsListValueFromResponse(_ context.Context, path path.Path, item cm.RolePermissionsItem) (TypedList[types.String], diag.Diagnostics) {
	switch item.Type {
	case cm.StringRolePermissionsItem:
		actionsValues := make([]types.String, 1)
		actionsValues[0] = types.StringValue(item.String)

		actionsListValue := NewTypedList(actionsValues)

		return actionsListValue, nil

	case cm.StringArrayRolePermissionsItem:
		return NewTypedListFromStringSlice(item.StringArray), nil
	}

	diags := diag.Diagnostics{}
	diags.AddAttributeWarning(path, "Unsupported permission actions", "Contentful returned an unsupported permission action shape. Terraform state retains a known null list; a later request conversion will reject it until configured with a supported shape.")

	return NewTypedListNull[types.String](), diags
}
