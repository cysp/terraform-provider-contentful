package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewWebhookResourceModelFromResponse(ctx context.Context, webhookDefinition cm.WebhookDefinition, existingHeaderValues TypedMap[TypedObject[WebhookHeaderValue]]) (WebhookModel, diag.Diagnostics) {
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

	headersList, headersListDiags := ReadHeaderValueMapFromResponse(ctx, path.Root("headers"), webhookDefinition.Headers, existingHeaderValues)
	diags.Append(headersListDiags...)

	model.Headers = headersList
	model.HeaderValuesWO = NewTypedMapNull[types.String]()

	transformationValue, transformationValueDiags := ReadWebhookTransformationValueFromResponse(ctx, path.Root("transformation"), webhookDefinition.Transformation)
	diags.Append(transformationValueDiags...)

	model.Transformation = transformationValue

	model.Active = util.OptBoolToBoolValue(webhookDefinition.Active)

	return model, diags
}

// NewWebhookResourceModelForMutationState starts with the response projection,
// reconciles known planned filters (including null), and uses fallback headers
// to preserve values omitted from secret or redacted responses. The response
// resolves unknown filters, which are never copied into state. Read skips
// filter reconciliation so it can expose remote drift.
func NewWebhookResourceModelForMutationState(ctx context.Context, webhookDefinition cm.WebhookDefinition, appliedPlan WebhookModel) (WebhookModel, diag.Diagnostics) {
	mutationState, diags := NewWebhookResourceModelFromResponse(ctx, webhookDefinition, appliedPlan.Headers)
	if !appliedPlan.Filters.IsUnknown() {
		mutationState.Filters = appliedPlan.Filters
	}

	return mutationState, diags
}
