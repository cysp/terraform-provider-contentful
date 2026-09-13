package provider

import (
	"context"
	"slices"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// appEventSubscriptionResponse projects returned configuration while keeping the
// addressed singleton as the target. Topic projection diagnostics are returned
// separately so a lossy projection cannot accidentally satisfy a known Plan.
func appEventSubscriptionResponse(ctx context.Context, response cm.AppEventSubscription, target AppEventSubscriptionModel) (AppEventSubscriptionModel, diag.Diagnostics, diag.Diagnostics) {
	topicValues := append(make([]string, 0, len(response.Topics)), response.Topics...)
	slices.Sort(topicValues)
	topicValues = slices.Compact(topicValues)

	topics, topicDiags := types.SetValueFrom(ctx, types.StringType, topicValues)
	if len(topics.Elements()) != len(response.Topics) {
		topicDiags.AddAttributeWarning(path.Root("topics"), "Duplicate app event topics in response", "Contentful returned duplicate topics. Terraform can represent each topic only once in the set.")
	}

	data := AppEventSubscriptionModel{
		IDIdentityModel: NewIDIdentityModelFromMultipartID(target.OrganizationID.ValueString(), target.AppDefinitionID.ValueString()),
		OrganizationID:  target.OrganizationID, AppDefinitionID: target.AppDefinitionID,
		Topics: topics, TargetURL: types.StringPointerValue(response.TargetUrl.ValueStringPointer()),
		FilterFunctionID: types.StringNull(), TransformationFunctionID: types.StringNull(), HandlerFunctionID: types.StringNull(), Timeouts: target.Timeouts,
	}
	if functions, ok := response.Functions.Get(); ok {
		for _, role := range []struct {
			link   cm.OptFunctionLink
			target *types.String
		}{
			{functions.Filter, &data.FilterFunctionID}, {functions.Transformation, &data.TransformationFunctionID}, {functions.Handler, &data.HandlerFunctionID},
		} {
			if link, present := role.link.Get(); present {
				*role.target = types.StringValue(link.Sys.ID)
			}
		}
	}

	var identityDiags diag.Diagnostics

	for _, identity := range []struct {
		name, actual string
		expected     types.String
	}{
		{"organization_id", response.Sys.Organization.Sys.ID, target.OrganizationID},
		{"app_definition_id", response.Sys.AppDefinition.Sys.ID, target.AppDefinitionID},
	} {
		if identity.actual != identity.expected.ValueString() {
			identityDiags.AddAttributeError(path.Root(identity.name), "Contentful returned a different app event subscription identity", "The response parent differs from the addressed singleton. Terraform retained the requested identity and returned configuration. Review the subscription before applying again.")
		}
	}

	return data, topicDiags, identityDiags
}

func reconcileAppEventSubscriptionResponse(ctx context.Context, response cm.AppEventSubscription, plan AppEventSubscriptionModel) (AppEventSubscriptionModel, diag.Diagnostics, diag.Diagnostics) {
	data, projectionDiags, identityDiags := appEventSubscriptionResponse(ctx, response, plan)
	reconciler := mutationResponseReconciler{resourceName: "app event subscription", diagnostics: identityDiags}
	reconciler.compareSemantic(path.Root("topics"), "app event topics", "Contentful returned different app event topics", plan.Topics.IsUnknown(), projectionDiags, func() (bool, diag.Diagnostics) { return plan.Topics.Equal(data.Topics), nil })

	for _, field := range []struct {
		name           string
		plan, response types.String
	}{
		{"target_url", plan.TargetURL, data.TargetURL},
		{"filter_function_id", plan.FilterFunctionID, data.FilterFunctionID},
		{"transformation_function_id", plan.TransformationFunctionID, data.TransformationFunctionID},
		{"handler_function_id", plan.HandlerFunctionID, data.HandlerFunctionID},
	} {
		reconciler.compareExact(path.Root(field.name), "Contentful returned a different "+field.name, field.plan, field.response)
	}

	if !plan.ID.IsUnknown() && !plan.ID.IsNull() {
		reconciler.compareExact(path.Root("id"), "App event subscription ID differs from the addressed singleton", plan.ID, data.ID)
	}

	return data, projectionDiags, reconciler.diagnostics
}
