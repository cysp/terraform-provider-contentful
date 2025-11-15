package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewPoliciesListValueFromResponse(ctx context.Context, path path.Path, policies []cm.RolePoliciesItem) (TypedList[TypedObject[RolePolicyValue]], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	policiesValues := make([]TypedObject[RolePolicyValue], len(policies))

	for index, item := range policies {
		policyPath := path.AtListIndex(index)

		policiesValue, actionsListValueDiags := NewPoliciesValueFromResponse(ctx, policyPath, item)
		diags.Append(actionsListValueDiags...)

		policiesValues[index] = policiesValue
	}

	policiesListValue := NewTypedList(policiesValues)

	return policiesListValue, diags
}

func NewPoliciesValueFromResponse(ctx context.Context, path path.Path, item cm.RolePoliciesItem) (TypedObject[RolePolicyValue], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	effect, err := item.Effect.MarshalText()
	if err != nil {
		diags.AddAttributeError(path.AtName("effect"), "Failed to read policy effect", err.Error())
	}

	value := RolePolicyValue{
		Effect: types.StringValue(string(effect)),
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

	object := NewTypedObject(value)

	return object, diags
}

func NewPolicyActionsListValueFromResponse(_ context.Context, path path.Path, actions cm.RolePoliciesItemActions) (TypedList[types.String], diag.Diagnostics) {
	switch actions.Type {
	case cm.StringRolePoliciesItemActions:
		actionsValues := make([]types.String, 1)
		actionsValues[0] = types.StringValue(actions.String)

		actionsListValue := NewTypedList(actionsValues)

		return actionsListValue, nil

	case cm.StringArrayRolePoliciesItemActions:
		return NewTypedListFromStringSlice(actions.StringArray), nil
	}

	diags := diag.Diagnostics{}
	diags.AddAttributeError(path, "unexpected type for policy actions", "")

	return NewTypedListUnknown[types.String](), diags
}
