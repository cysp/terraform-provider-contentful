package provider

import (
	"slices"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ToRoleDataPolicies(path path.Path, policies []*RolePolicyValue) ([]cm.RoleDataPoliciesItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if policies == nil {
		diags.AddAttributeError(
			path,
			"Unexpected null policies",
			"Policies are required.",
		)

		return nil, diags
	}

	rolePoliciesItems := make([]cm.RoleDataPoliciesItem, 0, len(policies))

	for index, policiesValueElement := range policies {
		itemPath := path.AtListIndex(index)

		policiesItem, policiesItemDiags := ToRoleDataPoliciesItem(itemPath, policiesValueElement)
		diags.Append(policiesItemDiags...)

		if !policiesItemDiags.HasError() {
			rolePoliciesItems = append(rolePoliciesItems, policiesItem)
		}
	}

	if diags.HasError() {
		return nil, diags
	}

	return rolePoliciesItems, diags
}

func ToRoleDataPoliciesItem(path path.Path, policy *RolePolicyValue) (cm.RoleDataPoliciesItem, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if policy == nil {
		diags.AddAttributeError(
			path,
			"Unexpected null policy",
			"Null policies cannot be sent to Contentful.",
		)

		return cm.RoleDataPoliciesItem{}, diags
	}

	effect, effectDiags := requestRequiredString(policy.Effect, path.AtName("effect"))
	diags.Append(effectDiags...)

	actions, actionsDiags := ToRoleDataPoliciesItemActions(path.AtName("actions"), policy.Actions)
	diags.Append(actionsDiags...)

	constraint, constraintDiags := ToOptRoleDataPoliciesItemConstraint(path.AtName("constraint"), policy.Constraint)
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

func ToRoleDataPoliciesItemActions(path path.Path, actions []types.String) (cm.RoleDataPoliciesItemActions, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if actions == nil {
		diags.AddAttributeError(
			path,
			"Unexpected null policy actions",
			"Policy actions are required.",
		)

		return cm.RoleDataPoliciesItemActions{}, diags
	}

	actionsStrings, actionDiags := knownStringListElements(path, actions)
	diags.Append(actionDiags...)

	if actionDiags.HasError() {
		return cm.RoleDataPoliciesItemActions{}, diags
	}

	diags.Append(validateRolePolicyActions(path, len(actionsStrings), actionsStrings)...)

	if diags.HasError() {
		return cm.RoleDataPoliciesItemActions{}, diags
	}

	if len(actionsStrings) == 1 && actionsStrings[0] == "all" {
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

func validateRolePolicyActions(path path.Path, actionCount int, actions []string) diag.Diagnostics {
	diags := diag.Diagnostics{}

	if actionCount != 1 && slices.Contains(actions, "all") {
		diags.AddAttributeError(
			path,
			"Invalid policy actions",
			`"all" must be specified by itself. Remove "all" or the other policy actions from this list.`,
		)
	}

	return diags
}

func ToOptRoleDataPoliciesItemConstraint(path path.Path, constraint jsontypes.Normalized) (jx.Raw, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if constraint.IsNull() {
		return nil, diags
	}

	if constraint.IsUnknown() {
		diags.AddAttributeError(
			path,
			"Unexpected unknown policy constraint",
			"The policy constraint must be known before it can be sent to Contentful.",
		)

		return nil, diags
	}

	constraintValue := []byte(constraint.ValueString())

	return constraintValue, diags
}
