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
		path := path.AtListIndex(index)

		policiesValue, actionsListValueDiags := NewPoliciesValueFromResponse(ctx, path, item)
		diags.Append(actionsListValueDiags...)

		policiesValues[index] = policiesValue
	}

	policiesListValue := NewTypedList(policiesValues)

	return policiesListValue, diags
}

func NewPoliciesValueFromResponse(ctx context.Context, path path.Path, item cm.RolePoliciesItem) (TypedObject[RolePolicyValue], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	effect, err := item.Effect.MarshalText()
	effectValue := types.StringNull()

	if err != nil {
		diags.AddAttributeWarning(path.AtName("effect"), "Unsupported policy effect", "Contentful returned an unsupported policy effect. Terraform state retains a null value; a later request conversion will reject it until configured with a supported effect.")
	} else {
		effectValue = types.StringValue(string(effect))
	}

	value := RolePolicyValue{
		Effect: effectValue,
	}

	actionsListValue, actionsListValueDiags := NewPolicyActionsListValueFromResponse(ctx, path.AtName("actions"), item.Actions)
	diags.Append(actionsListValueDiags...)

	value.Actions = actionsListValue

	if item.Constraint != nil {
		constraint, err := util.JxNormalizeOpaqueBytes(item.Constraint, util.JxEncodeOpaqueOptions{EscapeStrings: true})
		if err != nil {
			diags.AddAttributeError(path.AtName("constraint"), "Failed to read policy constraint", err.Error())
		}

		value.Constraint = NewNormalizedJSONValue(constraint)
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
	diags.AddAttributeWarning(path, "Unsupported policy actions", "Contentful returned an unsupported policy action shape. Terraform state retains a known null list; a later request conversion will reject it until configured with a supported shape.")

	return NewTypedListNull[types.String](), diags
}
