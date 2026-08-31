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

// NewRoleResourceModelForMutationState starts with the complete response
// projection. It restores exact known Plan representations only when both
// permissions and policies have been proven semantically equivalent; any lossy
// or contradictory owned value leaves the complete response as recovery state
// with attribute-scoped diagnostics. Unknown Plan values resolve from the
// response. Read projects remote state without reconciliation.
func NewRoleResourceModelForMutationState(ctx context.Context, role cm.Role, appliedPlan RoleModel) (RoleModel, diag.Diagnostics, diag.Diagnostics) {
	mutationState, responseDiags, permissionsDiags, policiesDiags := newRoleResourceModelFromResponse(ctx, role)
	consistencyDiags := diag.Diagnostics{}

	candidateState := mutationState
	mismatch := false

	if !appliedPlan.Permissions.IsUnknown() {
		switch {
		case len(permissionsDiags) != 0:
			consistencyDiags.AddAttributeError(path.Root("permissions"), "Unexpected Contentful role response", "The permissions response could not be projected without loss, so equivalence with the Terraform plan could not be established.")

			mismatch = true
		case rolePermissionsEquivalent(appliedPlan.Permissions, mutationState.Permissions):
			candidateState.Permissions = appliedPlan.Permissions
		default:
			consistencyDiags.AddAttributeError(path.Root("permissions"), "Unexpected Contentful role response", "The permissions response differed meaningfully from the Terraform plan.")

			mismatch = true
		}
	}

	if !appliedPlan.Policies.IsUnknown() {
		policiesEquivalent, comparisonDiags := rolePoliciesEquivalent(ctx, appliedPlan.Policies, mutationState.Policies)
		switch {
		case len(policiesDiags) != 0:
			consistencyDiags.AddAttributeError(path.Root("policies"), "Unexpected Contentful role response", "The policies response could not be projected without loss, so equivalence with the Terraform plan could not be established.")

			mismatch = true
		case comparisonDiags.HasError():
			consistencyDiags.AddAttributeError(path.Root("policies"), "Unexpected Contentful role response", "The policies response could not be compared semantically with the Terraform plan.")

			mismatch = true
		case policiesEquivalent:
			candidateState.Policies = appliedPlan.Policies
		default:
			consistencyDiags.AddAttributeError(path.Root("policies"), "Unexpected Contentful role response", "The policies response differed meaningfully from the Terraform plan.")

			mismatch = true
		}
	}

	if !mismatch {
		mutationState = candidateState
	}

	return mutationState, responseDiags, consistencyDiags
}
