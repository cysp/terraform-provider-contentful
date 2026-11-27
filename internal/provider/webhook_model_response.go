package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewWebhookResourceModelFromResponse(ctx context.Context, webhookDefinition cm.WebhookDefinition, fallback WebhookModel) (WebhookModel, diag.Diagnostics) {
	var filterDiags diag.Diagnostics

	spaceID := webhookDefinition.Sys.Space.Sys.ID
	webhookID := webhookDefinition.Sys.ID

	model := WebhookModel{
		IDIdentityModel: NewIDIdentityModelFromMultipartID(spaceID, webhookID),
		SpaceID:         types.StringValue(spaceID),
		WebhookID:       types.StringValue(webhookID),
		Version:         types.Int64Value(int64(webhookDefinition.Sys.Version)),
	}

	model.Name = types.StringValue(webhookDefinition.Name)

	model.URL = types.StringValue(webhookDefinition.URL)

	model.Topics = NewTypedListFromStringSlice(webhookDefinition.Topics)

	model.Filters, filterDiags = ReadWebhookFiltersListValueFromResponse(ctx, path.Root("filters"), webhookDefinition.Filters)

	model.HTTPBasicUsername = types.StringPointerValue(webhookDefinition.HttpBasicUsername.ValueStringPointer())

	model.HTTPBasicPassword = types.StringPointerValue(webhookDefinition.HttpBasicPassword.ValueStringPointer())
	if webhookDefinition.HttpBasicPassword.IsEmpty() && !model.HTTPBasicUsername.IsNull() &&
		model.HTTPBasicUsername.Equal(fallback.HTTPBasicUsername) && !fallback.HTTPBasicPassword.IsUnknown() {
		model.HTTPBasicPassword = fallback.HTTPBasicPassword
	}

	model.Headers = webhookResponseHeaders(webhookDefinition.Headers, fallback.Headers)

	model.Transformation = ReadWebhookTransformationValueFromResponse(webhookDefinition.Transformation)

	model.Active = util.OptBoolToBoolValue(webhookDefinition.Active)

	return model, filterDiags
}

// ReconcileWebhookMutationResponse projects the complete mutation response and
// restores known effective Plan values only after proving that every
// API-backed value is losslessly, semantically equivalent. Endpoint identity
// always remains the requested Terraform target. Password and secret-header
// response omissions retain their established narrow fallbacks.
func ReconcileWebhookMutationResponse(ctx context.Context, webhookDefinition cm.WebhookDefinition, plan WebhookModel, endpointID types.String) (WebhookModel, diag.Diagnostics, diag.Diagnostics, bool) {
	state, filterDiags := NewWebhookResourceModelFromResponse(ctx, webhookDefinition, plan)
	reconciler := mutationResponseReconciler{resourceName: "webhook"}
	reconciler.pinIdentity(path.Root("space_id"), plan.SpaceID, state.SpaceID, &state.SpaceID)
	reconciler.pinCreatedIdentity(path.Root("webhook_id"), endpointID, plan.WebhookID, state.WebhookID, &state.WebhookID)

	if reconciler.identityMismatch {
		state.Version = types.Int64Null()
	}

	state.IDIdentityModel = NewIDIdentityModelFromMultipartID(state.SpaceID.ValueString(), state.WebhookID.ValueString())

	if !plan.ID.IsNull() && !plan.ID.IsUnknown() && !plan.ID.Equal(state.ID) {
		reconciler.diagnostics.AddAttributeError(path.Root("id"), "Webhook identity is inconsistent with its endpoint", "The planned legacy ID differs from the requested Webhook endpoint identity. Terraform retained the endpoint identity as the resource target and the remaining returned values as recovery state. Review or re-import the Webhook before applying again.")
	}

	candidateState := state
	if reconciler.compareExact(path.Root("name"), "Contentful returned a different webhook name", plan.Name, state.Name) {
		candidateState.Name = plan.Name
	}

	if reconciler.compareExact(path.Root("url"), "Contentful returned a different webhook URL", plan.URL, state.URL) {
		candidateState.URL = plan.URL
	}

	if reconciler.compareExact(path.Root("topics"), "Contentful returned different webhook topics", plan.Topics, state.Topics) {
		candidateState.Topics = plan.Topics
	}

	if reconciler.compareExact(path.Root("http_basic_password"), "Contentful returned a different webhook Basic password", plan.HTTPBasicPassword, state.HTTPBasicPassword) {
		candidateState.HTTPBasicPassword = plan.HTTPBasicPassword
	}

	if reconciler.compareExact(path.Root("http_basic_username"), "Contentful returned a different webhook Basic username", plan.HTTPBasicUsername, state.HTTPBasicUsername) {
		candidateState.HTTPBasicUsername = plan.HTTPBasicUsername
	}

	if reconciler.compareSemantic(
		path.Root("filters"), "webhook filters", "Contentful returned different webhook filters",
		plan.Filters.IsUnknown(), filterDiags,
		func() (bool, diag.Diagnostics) { return webhookFiltersEquivalent(plan.Filters, state.Filters), nil },
	) {
		candidateState.Filters = plan.Filters
	}

	if reconciler.compareExact(path.Root("headers"), "Contentful returned different webhook headers", plan.Headers, state.Headers) {
		candidateState.Headers = plan.Headers
	}

	if reconciler.compareSemantic(
		path.Root("transformation"), "webhook transformation", "Contentful returned a different webhook transformation",
		plan.Transformation.IsUnknown(), diag.Diagnostics{},
		func() (bool, diag.Diagnostics) {
			return webhookTransformationEquivalent(ctx, plan.Transformation, state.Transformation)
		},
	) {
		candidateState.Transformation = plan.Transformation
	}

	if reconciler.compareExact(path.Root("active"), "Contentful returned a different webhook active value", plan.Active, state.Active) {
		candidateState.Active = plan.Active
	}

	if !reconciler.diagnostics.HasError() {
		state = candidateState
	}

	return state, filterDiags, reconciler.diagnostics, !reconciler.identityMismatch
}

func webhookFiltersEquivalent(planned, response TypedList[TypedObject[WebhookFilterValue]]) bool {
	if planned.Equal(response) {
		return true
	}

	if planned.IsNull() || planned.IsUnknown() || response.IsNull() || response.IsUnknown() {
		return false
	}

	equivalent, _ := unorderedElementsEquivalent(planned.Elements(), response.Elements(), func(plannedElement, responseElement TypedObject[WebhookFilterValue]) (bool, diag.Diagnostics) {
		return webhookFilterEquivalent(plannedElement, responseElement), nil
	})

	return equivalent
}

func webhookFilterEquivalent(planned, response TypedObject[WebhookFilterValue]) bool {
	if planned.Equal(response) {
		return true
	}

	plannedValue, plannedKnown := planned.GetValue()

	responseValue, responseKnown := response.GetValue()
	if !plannedKnown || !responseKnown {
		return false
	}

	return webhookFilterNotEquivalent(plannedValue.Not, responseValue.Not) &&
		plannedValue.Equals.Equal(responseValue.Equals) &&
		webhookFilterInEquivalent(plannedValue.In, responseValue.In) &&
		plannedValue.Regexp.Equal(responseValue.Regexp)
}

func webhookFilterNotEquivalent(planned, response TypedObject[WebhookFilterNotValue]) bool {
	if planned.Equal(response) {
		return true
	}

	plannedValue, plannedKnown := planned.GetValue()

	responseValue, responseKnown := response.GetValue()
	if !plannedKnown || !responseKnown {
		return false
	}

	return plannedValue.Equals.Equal(responseValue.Equals) &&
		webhookFilterInEquivalent(plannedValue.In, responseValue.In) &&
		plannedValue.Regexp.Equal(responseValue.Regexp)
}

func webhookFilterInEquivalent(planned, response TypedObject[WebhookFilterInValue]) bool {
	if planned.Equal(response) {
		return true
	}

	plannedValue, plannedKnown := planned.GetValue()

	responseValue, responseKnown := response.GetValue()
	if !plannedKnown || !responseKnown {
		return false
	}

	return plannedValue.Doc.Equal(responseValue.Doc) && unorderedStringListsEquivalent(plannedValue.Values, responseValue.Values)
}

func webhookTransformationEquivalent(ctx context.Context, planned, response TypedObject[WebhookTransformationValue]) (bool, diag.Diagnostics) {
	if planned.Equal(response) {
		return true, nil
	}

	plannedValue, plannedKnown := planned.GetValue()

	responseValue, responseKnown := response.GetValue()
	if !plannedKnown || !responseKnown || !plannedValue.Method.Equal(responseValue.Method) || !plannedValue.ContentType.Equal(responseValue.ContentType) || !plannedValue.IncludeContentLength.Equal(responseValue.IncludeContentLength) {
		return false, nil
	}

	return normalizedJSONEquivalent(ctx, plannedValue.Body, responseValue.Body)
}
