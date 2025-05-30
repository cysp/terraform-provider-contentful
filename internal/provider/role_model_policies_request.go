package provider

import (
	"context"
	"slices"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ToRoleFieldsPolicies(ctx context.Context, path path.Path, policies TypedList[RolePolicyValue]) ([]cm.RoleFieldsPoliciesItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if policies.IsUnknown() {
		return nil, diags
	}

	policiesValues := policies.Elements()

	rolePoliciesItems := make([]cm.RoleFieldsPoliciesItem, len(policiesValues))

	for index, policiesValueElement := range policiesValues {
		path := path.AtListIndex(index)

		policiesItem, policiesItemDiags := ToRoleFieldsPoliciesItem(ctx, path, policiesValueElement)
		diags.Append(policiesItemDiags...)

		rolePoliciesItems[index] = policiesItem
	}

	return rolePoliciesItems, diags
}

func ToRoleFieldsPoliciesItem(ctx context.Context, path path.Path, policy RolePolicyValue) (cm.RoleFieldsPoliciesItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	effect := policy.Effect.ValueString()

	actions, actionsDiags := ToRoleFieldsPoliciesItemActions(ctx, path.AtName("actions"), policy.Actions)
	diags.Append(actionsDiags...)

	constraint, constraintDiags := ToOptRoleFieldsPoliciesItemConstraint(ctx, path.AtName("constraint"), policy.Constraint)
	diags.Append(constraintDiags...)

	return cm.RoleFieldsPoliciesItem{
		Effect:     cm.RoleFieldsPoliciesItemEffect(effect),
		Actions:    actions,
		Constraint: constraint,
	}, diags
}

func ToRoleFieldsPoliciesItemActions(ctx context.Context, _ path.Path, actions TypedList[types.String]) (cm.RoleFieldsPoliciesItemActions, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	actionsStrings := make([]string, len(actions.Elements()))
	diags.Append(tfsdk.ValueAs(ctx, actions, &actionsStrings)...)

	if slices.Contains(actionsStrings, "all") {
		return cm.RoleFieldsPoliciesItemActions{
			Type:   cm.StringRoleFieldsPoliciesItemActions,
			String: "all",
		}, diags
	}

	return cm.RoleFieldsPoliciesItemActions{
		Type:        cm.StringArrayRoleFieldsPoliciesItemActions,
		StringArray: actionsStrings,
	}, diags
}

func ToOptRoleFieldsPoliciesItemConstraint(_ context.Context, _ path.Path, constraint jsontypes.Normalized) (jx.Raw, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if constraint.IsNull() {
		return nil, diags
	}

	constraintValue := []byte(constraint.ValueString())

	return constraintValue, diags
}
