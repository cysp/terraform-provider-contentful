package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (model *RoleModel) ToRoleData(ctx context.Context) (cm.RoleData, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	var name string

	switch {
	case model.Name.IsUnknown():
		diags.AddAttributeError(
			path.Root("name"),
			"Unexpected unknown role name",
			"The role name must be known before it can be sent to Contentful.",
		)
	case model.Name.IsNull():
		diags.AddAttributeError(
			path.Root("name"),
			"Unexpected null role name",
			"The role name is required.",
		)
	default:
		name = model.Name.ValueString()
	}

	description := cm.NewOptNilStringNull()

	if model.Description.IsUnknown() {
		diags.AddAttributeError(
			path.Root("description"),
			"Unexpected unknown role description",
			"The role description must be known before it can be sent to Contentful.",
		)
	} else if !model.Description.IsNull() {
		description = cm.NewOptNilString(model.Description.ValueString())
	}

	permissions, permissionsDiags := ToRoleDataPermissions(ctx, path.Root("permissions"), model.Permissions)
	diags.Append(permissionsDiags...)

	policies, policiesDiags := ToRoleDataPolicies(ctx, path.Root("policies"), model.Policies)
	diags.Append(policiesDiags...)

	if diags.HasError() {
		return cm.RoleData{}, diags
	}

	return cm.RoleData{
		Name:        name,
		Description: description,
		Permissions: permissions,
		Policies:    policies,
	}, diags
}
