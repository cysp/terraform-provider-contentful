package provider

import (
	"context"
	"strings"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewWebhookResourceModelFromResponse(ctx context.Context, webhookDefinition cm.WebhookDefinition, existingHeaderValues map[string]WebhookHeaderValue) (WebhookModel, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	spaceID := webhookDefinition.Sys.Space.Sys.ID
	webhookID := webhookDefinition.Sys.ID

	model := WebhookModel{
		ID:        types.StringValue(strings.Join([]string{spaceID, webhookID}, "/")),
		SpaceID:   types.StringValue(spaceID),
		WebhookID: types.StringValue(webhookID),
	}

	model.Name = types.StringValue(webhookDefinition.Name)

	model.URL = types.StringValue(webhookDefinition.URL)

	topicsList, topicsListDiags := NewTypedListFromStringSlice(ctx, webhookDefinition.Topics)
	diags.Append(topicsListDiags...)

	model.Topics = topicsList

	filtersList, filtersListDiags := ReadWebhookFiltersListValueFromResponse(ctx, path.Root("filters"), webhookDefinition.Filters)
	diags.Append(filtersListDiags...)

	model.Filters = filtersList

	model.HTTPBasicUsername = types.StringPointerValue(webhookDefinition.HttpBasicUsername.ValueStringPointer())
	model.HTTPBasicPassword = types.StringPointerValue(webhookDefinition.HttpBasicPassword.ValueStringPointer())

	headersList, headersListDiags := ReadHeaderValueMapFromResponse(ctx, path.Root("headers"), webhookDefinition.Headers, existingHeaderValues)
	diags.Append(headersListDiags...)

	model.Headers = headersList

	transformationValue, transformationValueDiags := ReadWebhookTransformationValueFromResponse(ctx, path.Root("transformation"), webhookDefinition.Transformation)
	diags.Append(transformationValueDiags...)

	model.Transformation = transformationValue

	model.Active = util.OptBoolToBoolValue(webhookDefinition.Active)

	return model, diags
}
