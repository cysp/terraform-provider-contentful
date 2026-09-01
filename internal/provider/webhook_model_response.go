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

// ReconcileWebhookMutationResponse projects the complete mutation
// response and restores the planned filters representation only after proving
// lossless semantic equivalence.
func ReconcileWebhookMutationResponse(ctx context.Context, webhookDefinition cm.WebhookDefinition, plan WebhookModel) (WebhookModel, diag.Diagnostics, diag.Diagnostics) {
	state, responseDiags, filtersDiags := newWebhookResourceModelFromResponse(ctx, webhookDefinition, plan.Headers.Elements())
	if plan.Filters.IsUnknown() {
		return state, responseDiags, nil
	}

	consistencyDiags := diag.Diagnostics{}

	switch {
	case len(filtersDiags) != 0:
		consistencyDiags.AddAttributeError(path.Root("filters"), "Provider cannot fully represent webhook filters", "Contentful accepted the request, but the returned webhook filters contain values this provider cannot fully represent. Terraform retained the representable response values but cannot verify that they match the value Terraform applied. Review the webhook in Contentful before applying again.")
	case webhookFiltersEquivalent(plan.Filters, state.Filters):
		state.Filters = plan.Filters
	default:
		consistencyDiags.AddAttributeError(path.Root("filters"), "Contentful returned different webhook filters", "Contentful accepted the request but returned webhook filters that differ from the value Terraform applied. Terraform retained the returned value in state rather than substituting the planned value. Review the webhook and configuration before applying again.")
	}

	return state, responseDiags, consistencyDiags
}
