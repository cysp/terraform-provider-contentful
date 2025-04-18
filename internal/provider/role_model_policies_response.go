package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewPoliciesListValueFromResponse(ctx context.Context, path path.Path, policies []cm.RolePoliciesItem) (types.List, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	policiesValues := make([]RolePolicyValue, len(policies))

	for index, item := range policies {
		path := path.AtListIndex(index)

		policiesValue, actionsListValueDiags := NewPoliciesValueFromResponse(ctx, path, item)
		diags.Append(actionsListValueDiags...)

		policiesValues[index] = policiesValue
	}

	policiesListValue, policiesListValueDiags := types.ListValueFrom(ctx, RolePolicyValue{}.Type(ctx), policiesValues)
	diags.Append(policiesListValueDiags...)

	return policiesListValue, diags
}

func NewPoliciesValueFromResponse(ctx context.Context, path path.Path, item cm.RolePoliciesItem) (RolePolicyValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	effect, err := item.Effect.MarshalText()
	if err != nil {
		diags.AddAttributeError(path.AtName("effect"), "Failed to read policy effect", err.Error())
	}

	value := RolePolicyValue{
		Effect: types.StringValue(string(effect)),
		state:  attr.ValueStateKnown,
	}

	actionsListValue, actionsListValueDiags := NewPolicyActionsListValueFromResponse(ctx, path.AtName("actions"), item.Actions)
	diags.Append(actionsListValueDiags...)

	value.Actions = actionsListValue

	if item.Constraint != nil {
		constraint, err := util.JxNormalizeOpaqueBytes(item.Constraint, util.JxEncodeOpaqueOptions{EscapeStrings: true})
		if err != nil {
			diags.AddAttributeError(path.AtName("constraint"), "Failed to read policy constraint", err.Error())
		}

		value.Constraint = jsontypes.NewNormalizedValue(string(constraint))
	} else {
		value.Constraint = jsontypes.NewNormalizedNull()
	}

	return value, diags
}

func NewPolicyActionsListValueFromResponse(ctx context.Context, path path.Path, actions cm.RolePoliciesItemActions) (types.List, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	switch actions.Type {
	case cm.StringRolePoliciesItemActions:
		actionsValues := make([]attr.Value, 1)
		actionsValues[0] = types.StringValue(actions.String)

		actionsListValue, actionsListValueDiags := types.ListValueFrom(ctx, types.String{}.Type(ctx), actionsValues)
		diags.Append(actionsListValueDiags...)

		return actionsListValue, diags

	case cm.StringArrayRolePoliciesItemActions:
		actionsListValue, actionsListValueDiags := types.ListValueFrom(ctx, types.String{}.Type(ctx), actions.StringArray)
		diags.Append(actionsListValueDiags...)

		return actionsListValue, diags
	}

	diags.AddAttributeError(path, "unexpected type for policy actions", "")

	return types.List{}, diags
}
