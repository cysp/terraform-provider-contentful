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

func ToRoleDataPolicies(ctx context.Context, path path.Path, policies TypedList[TypedObject[RolePolicyValue]]) ([]cm.RoleDataPoliciesItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if policies.IsUnknown() {
		return nil, diags
	}

	policiesValues := policies.Elements()

	rolePoliciesItems := make([]cm.RoleDataPoliciesItem, len(policiesValues))

	for index, policiesValueElement := range policiesValues {
		path := path.AtListIndex(index)

		policiesItem, policiesItemDiags := ToRoleDataPoliciesItem(ctx, path, policiesValueElement)
		diags.Append(policiesItemDiags...)

		rolePoliciesItems[index] = policiesItem
	}

	return rolePoliciesItems, diags
}

func ToRoleDataPoliciesItem(ctx context.Context, path path.Path, policy TypedObject[RolePolicyValue]) (cm.RoleDataPoliciesItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	effect := policy.Value().Effect.ValueString()

	actions, actionsDiags := ToRoleDataPoliciesItemActions(ctx, path.AtName("actions"), policy.Value().Actions)
	diags.Append(actionsDiags...)

	constraint, constraintDiags := ToOptRoleDataPoliciesItemConstraint(ctx, path.AtName("constraint"), policy.Value().Constraint)
	diags.Append(constraintDiags...)

	return cm.RoleDataPoliciesItem{
		Effect:     cm.RoleDataPoliciesItemEffect(effect),
		Actions:    actions,
		Constraint: constraint,
	}, diags
}

func ToRoleDataPoliciesItemActions(ctx context.Context, _ path.Path, actions TypedList[types.String]) (cm.RoleDataPoliciesItemActions, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	actionsStrings := make([]string, len(actions.Elements()))
	diags.Append(tfsdk.ValueAs(ctx, actions, &actionsStrings)...)

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

func ToOptRoleDataPoliciesItemConstraint(_ context.Context, _ path.Path, constraint jsontypes.Normalized) (jx.Raw, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if constraint.IsNull() {
		return nil, diags
	}

	constraintValue := []byte(constraint.ValueString())

	return constraintValue, diags
}
