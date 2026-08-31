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

// NewWebhookResourceModelForMutationState starts with the complete response
// projection and uses fallback headers only for values omitted from secret or
// redacted responses. It restores the exact known Plan filter representation
// only after semantic equality is proven; lossy or contradictory projections
// remain recovery state with a consistency diagnostic. Unknown Plan filters
// resolve from the response. Read projects remote state without reconciliation.
func NewWebhookResourceModelForMutationState(ctx context.Context, webhookDefinition cm.WebhookDefinition, appliedPlan WebhookModel) (WebhookModel, diag.Diagnostics, diag.Diagnostics) {
	mutationState, responseDiags, filtersDiags := newWebhookResourceModelFromResponse(ctx, webhookDefinition, appliedPlan.Headers.Elements())
	if appliedPlan.Filters.IsUnknown() {
		return mutationState, responseDiags, nil
	}

	consistencyDiags := diag.Diagnostics{}

	switch {
	case len(filtersDiags) != 0:
		consistencyDiags.AddAttributeError(path.Root("filters"), "Unexpected Contentful webhook response", "The filters response could not be projected without loss, so equivalence with the Terraform plan could not be established.")
	case webhookFiltersEquivalent(appliedPlan.Filters, mutationState.Filters):
		mutationState.Filters = appliedPlan.Filters
	default:
		consistencyDiags.AddAttributeError(path.Root("filters"), "Unexpected Contentful webhook response", "The filters response differed meaningfully from the Terraform plan.")
	}

	return mutationState, responseDiags, consistencyDiags
}
