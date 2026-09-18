package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func appActionResponse(response cm.AppAction, target AppActionBaseModel) (AppActionBaseModel, diag.Diagnostics) {
	data := AppActionBaseModel{
		OrganizationID:  types.StringValue(response.Sys.Organization.Sys.ID),
		AppDefinitionID: types.StringValue(response.Sys.AppDefinition.Sys.ID),
		AppActionFields: AppActionFields{
			AppActionID:      types.StringValue(response.Sys.ID),
			Name:             types.StringValue(response.Name),
			Category:         types.StringValue(response.Category),
			Type:             types.StringValue(response.Type),
			Description:      types.StringPointerValue(response.Description.ValueStringPointer()),
			URL:              types.StringPointerValue(response.URL.ValueStringPointer()),
			FunctionID:       types.StringNull(),
			Parameters:       jsontypes.NewNormalizedNull(),
			ParametersSchema: jsontypes.NewNormalizedNull(),
			ResultSchema:     jsontypes.NewNormalizedNull(),
		},
	}

	for _, field := range []struct {
		raw    []byte
		target *jsontypes.Normalized
	}{
		{response.Parameters, &data.Parameters}, {response.ParametersSchema, &data.ParametersSchema}, {response.ResultSchema, &data.ResultSchema},
	} {
		if len(field.raw) > 0 {
			*field.target = NewNormalizedJSONValue(field.raw)
		}
	}

	if function, ok := response.Function.Get(); ok {
		data.FunctionID = types.StringValue(function.Sys.ID)
	}

	var identityDiags diag.Diagnostics

	for _, identity := range []struct {
		name     string
		expected types.String
		actual   *types.String
	}{
		{"organization_id", target.OrganizationID, &data.OrganizationID},
		{"app_definition_id", target.AppDefinitionID, &data.AppDefinitionID},
		{"app_action_id", target.AppActionID, &data.AppActionID},
	} {
		if identity.expected.IsNull() || identity.expected.IsUnknown() {
			continue
		}

		if !identity.expected.Equal(*identity.actual) {
			identityDiags.AddAttributeError(path.Root(identity.name), "Contentful returned a different App Action identity", "Contentful returned a different organization, App Definition, or App Action ID than requested. Check the action in Contentful before applying again.")
		}

		*identity.actual = identity.expected
	}

	data.IDIdentityModel = NewIDIdentityModelFromMultipartID(data.OrganizationID.ValueString(), data.AppDefinitionID.ValueString(), data.AppActionID.ValueString())

	return data, identityDiags
}

func appActionResourceResponse(response cm.AppAction, target AppActionBaseModel) (AppActionBaseModel, diag.Diagnostics) {
	data, diags := appActionResponse(response, target)
	if response.Category != "Custom" {
		data.Parameters = jsontypes.NewNormalizedNull()
	}

	return data, diags
}

func reconcileAppActionResponse(ctx context.Context, response cm.AppAction, plan AppActionModel, target AppActionBaseModel) (AppActionModel, diag.Diagnostics) {
	base, identityDiags := appActionResourceResponse(response, target)
	data := AppActionModel{AppActionBaseModel: base, Timeouts: plan.Timeouts}

	reconciler := mutationResponseReconciler{resourceName: "app action", diagnostics: identityDiags}
	if target.AppActionID.IsNull() || target.AppActionID.IsUnknown() {
		reconciler.compareExact(path.Root("app_action_id"), "Contentful returned a different app_action_id", plan.AppActionID, data.AppActionID)
	}

	for _, field := range []struct {
		name           string
		plan, response types.String
	}{
		{"id", plan.ID, data.ID}, {"name", plan.Name, data.Name}, {"category", plan.Category, data.Category}, {"type", plan.Type, data.Type}, {"description", plan.Description, data.Description}, {"url", plan.URL, data.URL}, {"function_id", plan.FunctionID, data.FunctionID},
	} {
		reconciler.compareExact(path.Root(field.name), "Contentful returned a different "+field.name, field.plan, field.response)
	}

	candidate := data
	for _, field := range []struct {
		name           string
		plan, response jsontypes.Normalized
		target         *jsontypes.Normalized
	}{
		{"parameters", plan.Parameters, data.Parameters, &candidate.Parameters}, {"parameters_schema", plan.ParametersSchema, data.ParametersSchema, &candidate.ParametersSchema}, {"result_schema", plan.ResultSchema, data.ResultSchema, &candidate.ResultSchema},
	} {
		if reconciler.compareSemantic(path.Root(field.name), field.name, "Contentful returned a different "+field.name, field.plan.IsUnknown(), nil, func() (bool, diag.Diagnostics) { return normalizedJSONEquivalent(ctx, field.plan, field.response) }) {
			*field.target = field.plan
		}
	}

	if !reconciler.diagnostics.HasError() {
		data = candidate
	}

	return data, reconciler.diagnostics
}
