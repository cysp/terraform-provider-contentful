//nolint:dupl
package provider

import (
	"context"
	"slices"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ToUpdateRoleReqPolicies(ctx context.Context, path path.Path, policies types.List) ([]cm.UpdateRoleReqPoliciesItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if policies.IsUnknown() {
		return nil, diags
	}

	policiesValues := make([]PoliciesValue, len(policies.Elements()))
	diags.Append(policies.ElementsAs(ctx, &policiesValues, false)...)

	rolePoliciesItems := make([]cm.UpdateRoleReqPoliciesItem, len(policies.Elements()))

	for index, policiesValueElement := range policiesValues {
		path := path.AtListIndex(index)

		policiesItem, policiesItemDiags := policiesValueElement.ToUpdateRoleReqPoliciesItem(ctx, path)
		diags.Append(policiesItemDiags...)

		rolePoliciesItems[index] = policiesItem
	}

	return rolePoliciesItems, diags
}

func (m *PoliciesValue) ToUpdateRoleReqPoliciesItem(ctx context.Context, path path.Path) (cm.UpdateRoleReqPoliciesItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	effect := m.Effect.ValueString()

	actions, actionsDiags := ToUpdateRoleReqPoliciesItemActions(ctx, path.AtName("actions"), m.Actions)
	diags.Append(actionsDiags...)

	constraint, constraintDiags := ToOptUpdateRoleReqPoliciesItemConstraint(ctx, path.AtName("constraint"), m.Constraint)
	diags.Append(constraintDiags...)

	return cm.UpdateRoleReqPoliciesItem{
		Effect:     cm.UpdateRoleReqPoliciesItemEffect(effect),
		Actions:    actions,
		Constraint: constraint,
	}, diags
}

func ToUpdateRoleReqPoliciesItemActions(ctx context.Context, _ path.Path, actions types.List) (cm.UpdateRoleReqPoliciesItemActions, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	actionsStrings := make([]string, len(actions.Elements()))
	diags.Append(actions.ElementsAs(ctx, &actionsStrings, false)...)

	if slices.Contains(actionsStrings, "all") {
		return cm.UpdateRoleReqPoliciesItemActions{
			Type:   cm.StringUpdateRoleReqPoliciesItemActions,
			String: "all",
		}, diags
	}

	return cm.UpdateRoleReqPoliciesItemActions{
		Type:        cm.StringArrayUpdateRoleReqPoliciesItemActions,
		StringArray: actionsStrings,
	}, diags
}

func ToOptUpdateRoleReqPoliciesItemConstraint(_ context.Context, _ path.Path, constraint types.String) (jx.Raw, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if constraint.IsNull() {
		return nil, diags
	}

	constraintValue := []byte(constraint.ValueString())

	return constraintValue, diags
}
