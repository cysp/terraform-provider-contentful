package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewRoleResourceModelFromResponse(ctx context.Context, role cm.Role) (RoleModel, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	spaceID := role.Sys.Space.Sys.ID
	roleID := role.Sys.ID

	model := RoleModel{
		IDIdentityModel: NewIDIdentityModelFromMultipartID(spaceID, roleID),
		RoleIdentityModel: RoleIdentityModel{
			SpaceID: types.StringValue(spaceID),
			RoleID:  types.StringValue(roleID),
		},
	}

	model.Name = types.StringValue(role.Name)
	model.Description = types.StringPointerValue(role.Description.ValueStringPointer())

	permissionsMapValue, permissionsMapValueDiags := NewPermissionsMapValueFromResponse(ctx, path.Root("permissions"), role.Permissions)
	diags.Append(permissionsMapValueDiags...)

	model.Permissions = permissionsMapValue

	policiesListValue, policiesListValueDiags := NewPoliciesListValueFromResponse(ctx, path.Root("policies"), role.Policies)
	diags.Append(policiesListValueDiags...)

	model.Policies = policiesListValue

	return model, diags
}

// NewRoleResourceModelForMutationState starts with the response projection and
// restores known plan-owned permissions and policies. The response resolves
// null or unknown plan values, which are never copied into state. Read skips
// this reconciliation.
func NewRoleResourceModelForMutationState(ctx context.Context, role cm.Role, appliedPlan RoleModel) (RoleModel, diag.Diagnostics) {
	mutationState, diags := NewRoleResourceModelFromResponse(ctx, role)

	if !appliedPlan.Permissions.IsNull() && !appliedPlan.Permissions.IsUnknown() {
		mutationState.Permissions = appliedPlan.Permissions
	}

	if !appliedPlan.Policies.IsNull() && !appliedPlan.Policies.IsUnknown() {
		mutationState.Policies = appliedPlan.Policies
	}

	return mutationState, diags
}
