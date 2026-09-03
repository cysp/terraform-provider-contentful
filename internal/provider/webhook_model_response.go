package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewWebhookResourceModelFromResponse(ctx context.Context, webhookDefinition cm.WebhookDefinition, fallbackHeaderValues map[string]TypedObject[WebhookHeaderValue]) (WebhookModel, diag.Diagnostics) {
	model, diags, _ := newWebhookResourceModelFromResponse(ctx, webhookDefinition, fallbackHeaderValues)

	return model, diags
}

func newWebhookResourceModelFromResponse(ctx context.Context, webhookDefinition cm.WebhookDefinition, fallbackHeaderValues map[string]TypedObject[WebhookHeaderValue]) (WebhookModel, diag.Diagnostics, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	spaceID := webhookDefinition.Sys.Space.Sys.ID
	webhookID := webhookDefinition.Sys.ID

	model := WebhookModel{
		IDIdentityModel: NewIDIdentityModelFromMultipartID(spaceID, webhookID),
		WebhookIdentityModel: WebhookIdentityModel{
			SpaceID:   types.StringValue(spaceID),
			WebhookID: types.StringValue(webhookID),
		},
	}

	model.Name = types.StringValue(webhookDefinition.Name)

	model.URL = types.StringValue(webhookDefinition.URL)

	model.Topics = NewTypedListFromStringSlice(webhookDefinition.Topics)

	filtersList, filtersListDiags := ReadWebhookFiltersListValueFromResponse(ctx, path.Root("filters"), webhookDefinition.Filters)
	diags.Append(filtersListDiags...)

	model.Filters = filtersList

	model.HTTPBasicUsername = types.StringPointerValue(webhookDefinition.HttpBasicUsername.ValueStringPointer())
	model.HTTPBasicPassword = types.StringPointerValue(webhookDefinition.HttpBasicPassword.ValueStringPointer())

	headersList, headersListDiags := ReadHeaderValueMapFromResponse(ctx, path.Root("headers"), webhookDefinition.Headers, fallbackHeaderValues)
	diags.Append(headersListDiags...)

	model.Headers = headersList

	transformationValue, transformationValueDiags := ReadWebhookTransformationValueFromResponse(ctx, path.Root("transformation"), webhookDefinition.Transformation)
	diags.Append(transformationValueDiags...)

	model.Transformation = transformationValue

	model.Active = util.OptBoolToBoolValue(webhookDefinition.Active)

	return model, diags, filtersListDiags
}

// ReconcileWebhookMutationResponse projects the complete mutation response and
// restores known effective Plan values only after proving that every
// API-backed value is losslessly, semantically equivalent. Endpoint identity
// always remains the requested Terraform target. Password and secret-header
// response omissions retain their established narrow fallbacks.
func ReconcileWebhookMutationResponse(ctx context.Context, webhookDefinition cm.WebhookDefinition, plan WebhookModel) (WebhookModel, diag.Diagnostics, diag.Diagnostics) {
	state, responseDiags, filtersDiags := newWebhookResourceModelFromResponse(ctx, webhookDefinition, plan.Headers.Elements())
	reconciler := mutationResponseReconciler{resourceName: "webhook"}
	reconciler.pinIdentity(path.Root("space_id"), plan.SpaceID, state.SpaceID, &state.SpaceID)
	reconciler.pinIdentity(path.Root("webhook_id"), plan.WebhookID, state.WebhookID, &state.WebhookID)

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

	if !plan.HTTPBasicPassword.IsUnknown() {
		switch {
		case webhookDefinition.HttpBasicPassword.IsEmpty() && !plan.HTTPBasicPassword.IsNull():
			state.HTTPBasicPassword = plan.HTTPBasicPassword
			candidateState.HTTPBasicPassword = plan.HTTPBasicPassword
		case !webhookDefinition.HttpBasicPassword.IsEmpty() && !plan.HTTPBasicPassword.Equal(state.HTTPBasicPassword):
			reconciler.diagnostics.AddAttributeError(path.Root("http_basic_password"), "Contentful returned a different webhook Basic password", "Contentful accepted the request but returned a webhook Basic password that differs from the value Terraform applied. Terraform retained the returned value in state. Review the webhook and configuration before applying again.")
		default:
			candidateState.HTTPBasicPassword = plan.HTTPBasicPassword
		}
	}

	if reconciler.compareExact(path.Root("http_basic_username"), "Contentful returned a different webhook Basic username", plan.HTTPBasicUsername, state.HTTPBasicUsername) {
		candidateState.HTTPBasicUsername = plan.HTTPBasicUsername
	}

	if reconciler.compareSemantic(
		path.Root("filters"), "webhook filters", "Contentful returned different webhook filters",
		plan.Filters.IsUnknown(), filtersDiags,
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

	return state, responseDiags, reconciler.diagnostics
}
