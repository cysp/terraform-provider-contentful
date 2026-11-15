package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewPoliciesListValueFromResponse(ctx context.Context, path path.Path, policies []cm.RolePoliciesItem) ([]*RolePolicyValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	policiesValues := make([]*RolePolicyValue, len(policies))

	for index, item := range policies {
		path := path.AtListIndex(index)

		policiesValue, actionsListValueDiags := NewPoliciesValueFromResponse(ctx, path, item)
		diags.Append(actionsListValueDiags...)

		policiesValues[index] = policiesValue
	}

	return policiesValues, diags
}

func NewPoliciesValueFromResponse(ctx context.Context, path path.Path, item cm.RolePoliciesItem) (*RolePolicyValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	effect, err := item.Effect.MarshalText()
	effectValue := types.StringNull()

	if err != nil {
		diags.AddAttributeWarning(path.AtName("effect"), "Unsupported policy effect", "Contentful returned an unsupported policy effect. Terraform state retains a null value; a later request conversion will reject it until configured with a supported effect.")
	} else {
		effectValue = types.StringValue(string(effect))
	}

	value := &RolePolicyValue{
		Effect: effectValue,
	}

	actionsListValue, actionsListValueDiags := NewPolicyActionsListValueFromResponse(ctx, path.AtName("actions"), item.Actions)
	diags.Append(actionsListValueDiags...)

	value.Actions = actionsListValue

	if item.Constraint != nil {
		constraint, err := parseNormalizedJSONValue(item.Constraint)
		if err != nil {
			diags.AddAttributeError(path.AtName("constraint"), "Failed to read policy constraint", err.Error())
		}

		value.Constraint = constraint
	} else {
		value.Constraint = jsontypes.NewNormalizedNull()
	}

	return value, diags
}

func NewPolicyActionsListValueFromResponse(_ context.Context, path path.Path, actions cm.RolePoliciesItemActions) ([]types.String, diag.Diagnostics) {
	switch actions.Type {
	case cm.StringRolePoliciesItemActions:
		return []types.String{types.StringValue(actions.String)}, nil

	case cm.StringArrayRolePoliciesItemActions:
		values := make([]types.String, len(actions.StringArray))
		for index, value := range actions.StringArray {
			values[index] = types.StringValue(value)
		}

		return values, nil
	}

	diags := diag.Diagnostics{}
	diags.AddAttributeWarning(path, "Unsupported policy actions", "Contentful returned an unsupported policy action shape. Terraform state retains a known null list; a later request conversion will reject it until configured with a supported shape.")

	return nil, diags
}
