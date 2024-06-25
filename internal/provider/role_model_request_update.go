package provider

import (
	"context"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (model *RoleModel) ToUpdateRoleReq(ctx context.Context) (contentfulManagement.UpdateRoleReq, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	request := contentfulManagement.UpdateRoleReq{
		Name:        model.Name.ValueString(),
		Description: contentfulManagement.NewOptNilPointerString(model.Description.ValueStringPointer()),
	}

	permissions, permissionsDiags := ToUpdateRoleReqPermissions(ctx, path.Root("permissions"), model.Permissions)
	diags.Append(permissionsDiags...)

	request.Permissions = permissions

	policies, policiesDiags := ToUpdateRoleReqPolicies(ctx, path.Root("policies"), model.Policies)
	diags.Append(policiesDiags...)

	request.Policies = policies

	return request, diags
}
