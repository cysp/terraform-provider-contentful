package provider

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (model *RoleModel) ToRoleData() (cm.RoleData, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	name, nameDiags := requestRequiredString(model.Name, path.Root("name"))
	diags.Append(nameDiags...)

	description, descriptionDiags := requestNullableString(model.Description, path.Root("description"))
	diags.Append(descriptionDiags...)

	permissions, permissionsDiags := ToRoleDataPermissions(path.Root("permissions"), model.Permissions)
	diags.Append(permissionsDiags...)

	policies, policiesDiags := ToRoleDataPolicies(path.Root("policies"), model.Policies)
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
