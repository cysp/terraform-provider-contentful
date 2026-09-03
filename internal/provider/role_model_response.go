package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewRoleResourceModelFromResponse(ctx context.Context, role cm.Role) (RoleModel, diag.Diagnostics) {
	model, diags, _, _ := newRoleResourceModelFromResponse(ctx, role)

	return model, diags
}

func newRoleResourceModelFromResponse(ctx context.Context, role cm.Role) (RoleModel, diag.Diagnostics, diag.Diagnostics, diag.Diagnostics) {
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

	return model, diags, permissionsMapValueDiags, policiesListValueDiags
}

// ReconcileRoleMutationResponse projects the complete mutation response and
// restores known effective Plan values only after proving that every
// API-backed value is losslessly, semantically equivalent. Endpoint identity
// always remains the requested Terraform target.
func ReconcileRoleMutationResponse(ctx context.Context, role cm.Role, plan RoleModel) (RoleModel, diag.Diagnostics, diag.Diagnostics) {
	state, responseDiags, permissionsDiags, policiesDiags := newRoleResourceModelFromResponse(ctx, role)
	reconciler := mutationResponseReconciler{resourceName: "role"}
	reconciler.pinIdentity(path.Root("space_id"), plan.SpaceID, state.SpaceID, &state.SpaceID)
	reconciler.pinIdentity(path.Root("role_id"), plan.RoleID, state.RoleID, &state.RoleID)

	state.IDIdentityModel = NewIDIdentityModelFromMultipartID(state.SpaceID.ValueString(), state.RoleID.ValueString())

	if !plan.ID.IsNull() && !plan.ID.IsUnknown() && !plan.ID.Equal(state.ID) {
		reconciler.diagnostics.AddAttributeError(path.Root("id"), "Role identity is inconsistent with its endpoint", "The planned legacy ID differs from the requested Role endpoint identity. Terraform retained the endpoint identity as the resource target and the remaining returned values as recovery state. Review or re-import the Role before applying again.")
	}

	candidateState := state
	if reconciler.compareExact(path.Root("name"), "Contentful returned a different role name", plan.Name, state.Name) {
		candidateState.Name = plan.Name
	}

	if reconciler.compareExact(path.Root("description"), "Contentful returned a different role description", plan.Description, state.Description) {
		candidateState.Description = plan.Description
	}

	if reconciler.compareSemantic(
		path.Root("permissions"), "role permissions", "Contentful returned different role permissions",
		plan.Permissions.IsUnknown(), permissionsDiags,
		func() (bool, diag.Diagnostics) {
			return rolePermissionsEquivalent(plan.Permissions, state.Permissions), nil
		},
	) {
		candidateState.Permissions = plan.Permissions
	}

	if reconciler.compareSemantic(
		path.Root("policies"), "role policies", "Contentful returned different role policies",
		plan.Policies.IsUnknown(), policiesDiags,
		func() (bool, diag.Diagnostics) { return rolePoliciesEquivalent(ctx, plan.Policies, state.Policies) },
	) {
		candidateState.Policies = plan.Policies
	}

	if !reconciler.diagnostics.HasError() {
		state = candidateState
	}

	return state, responseDiags, reconciler.diagnostics
}
