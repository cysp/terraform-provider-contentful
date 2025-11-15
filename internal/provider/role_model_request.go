package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (model *RoleModel) ToRoleData(ctx context.Context) (cm.RoleData, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	request := cm.RoleData{
		Name:        model.Name.ValueString(),
		Description: cm.NewOptNilPointerString(model.Description.ValueStringPointer()),
	}

	permissions, permissionsDiags := ToRoleDataPermissions(ctx, path.Root("permissions"), model.Permissions)
	diags.Append(permissionsDiags...)

	request.Permissions = permissions

	policies, policiesDiags := ToRoleDataPolicies(ctx, path.Root("policies"), model.Policies)
	diags.Append(policiesDiags...)

	request.Policies = policies

	return request, diags
}
