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

func ToCreateRoleReqPolicies(ctx context.Context, path path.Path, policies types.List) ([]cm.CreateRoleReqPoliciesItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if policies.IsUnknown() {
		return nil, diags
	}

	policiesValues := make([]PoliciesValue, len(policies.Elements()))
	diags.Append(policies.ElementsAs(ctx, &policiesValues, false)...)

	rolePoliciesItems := make([]cm.CreateRoleReqPoliciesItem, len(policies.Elements()))

	for index, policiesValueElement := range policiesValues {
		path := path.AtListIndex(index)

		policiesItem, policiesItemDiags := policiesValueElement.ToCreateRoleReqPoliciesItem(ctx, path)
		diags.Append(policiesItemDiags...)

		rolePoliciesItems[index] = policiesItem
	}

	return rolePoliciesItems, diags
}

func (m *PoliciesValue) ToCreateRoleReqPoliciesItem(ctx context.Context, path path.Path) (cm.CreateRoleReqPoliciesItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	effect := m.Effect.ValueString()

	actions, actionsDiags := ToCreateRoleReqPoliciesItemActions(ctx, path.AtName("actions"), m.Actions)
	diags.Append(actionsDiags...)

	constraint, constraintDiags := ToOptCreateRoleReqPoliciesItemConstraint(ctx, path.AtName("constraint"), m.Constraint)
	diags.Append(constraintDiags...)

	return cm.CreateRoleReqPoliciesItem{
		Effect:     cm.CreateRoleReqPoliciesItemEffect(effect),
		Actions:    actions,
		Constraint: constraint,
	}, diags
}

func ToCreateRoleReqPoliciesItemActions(ctx context.Context, _ path.Path, actions types.List) (cm.CreateRoleReqPoliciesItemActions, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	actionsStrings := make([]string, len(actions.Elements()))
	diags.Append(actions.ElementsAs(ctx, &actionsStrings, false)...)

	if slices.Contains(actionsStrings, "all") {
		return cm.CreateRoleReqPoliciesItemActions{
			Type:   cm.StringCreateRoleReqPoliciesItemActions,
			String: "all",
		}, diags
	}

	return cm.CreateRoleReqPoliciesItemActions{
		Type:        cm.StringArrayCreateRoleReqPoliciesItemActions,
		StringArray: actionsStrings,
	}, diags
}

func ToOptCreateRoleReqPoliciesItemConstraint(_ context.Context, _ path.Path, constraint types.String) (jx.Raw, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if constraint.IsNull() {
		return nil, diags
	}

	constraintValue := []byte(constraint.ValueString())

	return constraintValue, diags
}
