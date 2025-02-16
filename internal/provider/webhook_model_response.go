package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (model *WebhookModel) ReadFromResponse(ctx context.Context, webhookDefinition *cm.WebhookDefinition) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model.WebhookId = types.StringValue(webhookDefinition.Sys.ID)

	model.Name = types.StringValue(webhookDefinition.Name)

	model.Url = types.StringValue(webhookDefinition.URL)

	topicsList, topicsListDiags := types.ListValueFrom(ctx, types.String{}.Type(ctx), webhookDefinition.Topics)
	diags.Append(topicsListDiags...)

	model.Topics = topicsList

	filtersList, filtersListDiags := ReadWebhookFiltersListValueFromResponse(ctx, path.Root("filters"), webhookDefinition.Filters)
	diags.Append(filtersListDiags...)

	model.Filters = filtersList

	model.HttpBasicUsername = types.StringPointerValue(webhookDefinition.HttpBasicUsername.ValueStringPointer())
	model.HttpBasicPassword = types.StringPointerValue(webhookDefinition.HttpBasicPassword.ValueStringPointer())

	headersList, headersListDiags := ReadHeadersListValueFromResponse(ctx, path.Root("headers"), model.Headers, webhookDefinition.Headers)
	diags.Append(headersListDiags...)

	model.Headers = headersList

	transformationValue, transformationValueDiags := ReadWebhookTransformationValueFromResponse(ctx, path.Root("transformation"), webhookDefinition.Transformation)
	diags.Append(transformationValueDiags...)

	model.Transformation = transformationValue

	model.Active = util.OptBoolToBoolValue(webhookDefinition.Active)

	return diags
}

func ReadWebhookTransformationValueFromResponse(ctx context.Context, _ path.Path, optNilTransformation cm.OptNilWebhookDefinitionTransformation) (TransformationValue, diag.Diagnostics) {
	transformation, transformationOk := optNilTransformation.Get()
	if !transformationOk {
		return NewTransformationValueNull(), nil
	}

	diags := diag.Diagnostics{}

	methodValue := types.StringPointerValue(transformation.Method.ValueStringPointer())
	contentTypeValue := types.StringPointerValue(transformation.ContentType.ValueStringPointer())
	includeContentLengthValue := types.BoolPointerValue(transformation.IncludeContentLength.ValueBoolPointer())

	bodyValue := types.StringNull()
	if transformation.Body != nil {
		bodyValue = types.StringValue(transformation.Body.String())
	}

	//nolint:contextcheck
	value, valueDiags := NewTransformationValue(TransformationValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"method":                 methodValue,
		"content_type":           contentTypeValue,
		"include_content_length": includeContentLengthValue,
		"body":                   bodyValue,
	})
	diags.Append(valueDiags...)

	return value, diags
}
