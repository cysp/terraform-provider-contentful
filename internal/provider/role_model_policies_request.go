package provider

import (
	"context"
	"slices"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ToRoleDataPolicies(ctx context.Context, path path.Path, policies TypedList[TypedObject[RolePolicyValue]]) ([]cm.RoleDataPoliciesItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if policies.IsNull() || policies.IsUnknown() {
		if policies.IsUnknown() {
			diags.AddAttributeError(path, "Unexpected unknown policies", "Policies must be known before they can be sent to Contentful.")
		} else {
			diags.AddAttributeError(path, "Unexpected null policies", "Policies are required.")
		}

		return nil, diags
	}

	rolePoliciesItems, policyDiags := ConvertKnownObjectListElements(ctx, path, policies.Elements(), ToRoleDataPoliciesItem)
	diags.Append(policyDiags...)

	return rolePoliciesItems, diags
}

func ToRoleDataPoliciesItem(ctx context.Context, path path.Path, policy RolePolicyValue) (cm.RoleDataPoliciesItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	effect, effectDiags := KnownStringValue(policy.Effect, path.AtName("effect"))
	diags.Append(effectDiags...)

	actions, actionsDiags := ToRoleDataPoliciesItemActions(ctx, path.AtName("actions"), policy.Actions)
	diags.Append(actionsDiags...)

	constraint, constraintDiags := ToOptRoleDataPoliciesItemConstraint(ctx, path.AtName("constraint"), policy.Constraint)
	diags.Append(constraintDiags...)

	if diags.HasError() {
		return cm.RoleDataPoliciesItem{}, diags
	}

	return cm.RoleDataPoliciesItem{
		Effect:     cm.RoleDataPoliciesItemEffect(effect),
		Actions:    actions,
		Constraint: constraint,
	}, diags
}

func ToRoleDataPoliciesItemActions(ctx context.Context, path path.Path, actions TypedList[types.String]) (cm.RoleDataPoliciesItemActions, diag.Diagnostics) {
	actionsStrings, diags := KnownStringListValues(
		ctx,
		actions,
		path,
		"Unexpected unknown policy actions",
		"Policy actions must be known before they can be sent to Contentful.",
		"Unexpected null policy actions",
		"Policy actions are required.",
	)

	if diags.HasError() {
		return cm.RoleDataPoliciesItemActions{}, diags
	}

	if slices.Contains(actionsStrings, "all") {
		return cm.RoleDataPoliciesItemActions{
			Type:   cm.StringRoleDataPoliciesItemActions,
			String: "all",
		}, diags
	}

	return cm.RoleDataPoliciesItemActions{
		Type:        cm.StringArrayRoleDataPoliciesItemActions,
		StringArray: actionsStrings,
	}, diags
}

func ToOptRoleDataPoliciesItemConstraint(_ context.Context, path path.Path, constraint jsontypes.Normalized) (jx.Raw, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if constraint.IsNull() {
		return nil, diags
	}

	if constraint.IsUnknown() {
		diags.AddAttributeError(path, "Unexpected unknown constraint", "The policy constraint must be known before it can be sent to Contentful.")

		return nil, diags
	}

	constraintValue := []byte(constraint.ValueString())

	return constraintValue, diags
}
