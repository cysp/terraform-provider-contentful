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

// ReconcileRoleMutationResponse projects the complete mutation response
// and restores the planned permissions and policies only after proving both are
// losslessly, semantically equivalent.
func ReconcileRoleMutationResponse(ctx context.Context, role cm.Role, plan RoleModel) (RoleModel, diag.Diagnostics, diag.Diagnostics) {
	state, responseDiags, permissionsDiags, policiesDiags := newRoleResourceModelFromResponse(ctx, role)
	consistencyDiags := diag.Diagnostics{}

	candidateState := state
	mismatch := false

	if !plan.Permissions.IsUnknown() {
		switch {
		case len(permissionsDiags) != 0:
			consistencyDiags.AddAttributeError(path.Root("permissions"), "Provider cannot fully represent role permissions", "Contentful accepted the request, but the returned role permissions contain values this provider cannot fully represent. Terraform retained the representable response values but cannot verify that they match the value Terraform applied. Review the role in Contentful before applying again.")

			mismatch = true
		case rolePermissionsEquivalent(plan.Permissions, state.Permissions):
			candidateState.Permissions = plan.Permissions
		default:
			consistencyDiags.AddAttributeError(path.Root("permissions"), "Contentful returned different role permissions", "Contentful accepted the request but returned role permissions that differ from the value Terraform applied. Terraform retained the returned value in state rather than substituting the planned value. Review the role and configuration before applying again.")

			mismatch = true
		}
	}

	if !plan.Policies.IsUnknown() {
		if len(policiesDiags) != 0 {
			consistencyDiags.AddAttributeError(path.Root("policies"), "Provider cannot fully represent role policies", "Contentful accepted the request, but the returned role policies contain values this provider cannot fully represent. Terraform retained the representable response values but cannot verify that they match the value Terraform applied. Review the role in Contentful before applying again.")

			mismatch = true
		} else if policiesEquivalent, comparisonDiags := rolePoliciesEquivalent(ctx, plan.Policies, state.Policies); comparisonDiags.HasError() {
			consistencyDiags.AddAttributeError(path.Root("policies"), "Role policies could not be compared", "Contentful accepted the request, but the provider could not compare the returned role policies with the value Terraform applied. Terraform retained the Contentful response. Review the role in Contentful before applying again.")

			mismatch = true
		} else if policiesEquivalent {
			candidateState.Policies = plan.Policies
		} else {
			consistencyDiags.AddAttributeError(path.Root("policies"), "Contentful returned different role policies", "Contentful accepted the request but returned role policies that differ from the value Terraform applied. Terraform retained the returned value in state rather than substituting the planned value. Review the role and configuration before applying again.")

			mismatch = true
		}
	}

	if !mismatch {
		state = candidateState
	}

	return state, responseDiags, consistencyDiags
}
