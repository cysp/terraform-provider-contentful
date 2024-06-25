package provider

import (
	"context"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func NewPoliciesListValueFromResponse(ctx context.Context, path path.Path, policies []contentfulManagement.RolePoliciesItem) (basetypes.ListValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	policiesValues := make([]PoliciesValue, len(policies))

	for index, item := range policies {
		path := path.AtListIndex(index)

		policiesValue, actionsListValueDiags := NewPoliciesValueFromResponse(ctx, path, item)
		diags.Append(actionsListValueDiags...)

		policiesValues[index] = policiesValue
	}

	policiesListValue, policiesListValueDiags := basetypes.NewListValueFrom(ctx, PoliciesValue{}.Type(ctx), policiesValues)
	diags.Append(policiesListValueDiags...)

	return policiesListValue, diags
}

func NewPoliciesValueFromResponse(ctx context.Context, path path.Path, item contentfulManagement.RolePoliciesItem) (PoliciesValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	effect, err := item.Effect.MarshalText()
	if err != nil {
		diags.AddAttributeError(path.AtName("effect"), "Failed to read policy effect", err.Error())
	}

	value := PoliciesValue{
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

		value.Constraint = types.StringValue(string(constraint))
	} else {
		value.Constraint = types.StringNull()
	}

	return value, diags
}

func NewPolicyActionsListValueFromResponse(ctx context.Context, path path.Path, actions contentfulManagement.RolePoliciesItemActions) (basetypes.ListValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	switch actions.Type {
	case contentfulManagement.StringRolePoliciesItemActions:
		actionsValues := make([]attr.Value, 1)
		actionsValues[0] = types.StringValue(actions.String)

		actionsListValue, actionsListValueDiags := basetypes.NewListValue(types.String{}.Type(ctx), actionsValues)
		diags.Append(actionsListValueDiags...)

		return actionsListValue, diags

	case contentfulManagement.StringArrayRolePoliciesItemActions:
		actionsListValue, actionsListValueDiags := basetypes.NewListValueFrom(ctx, types.String{}.Type(ctx), actions.StringArray)
		diags.Append(actionsListValueDiags...)

		return actionsListValue, diags
	}

	diags.AddAttributeError(path, "unexpected type for policy actions", "")

	return basetypes.ListValue{}, diags
}
