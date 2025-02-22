package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ReadWebhookTransformationValueFromResponse(ctx context.Context, _ path.Path, optNilTransformation cm.OptNilWebhookDefinitionTransformation) (WebhookTransformationValue, diag.Diagnostics) {
	transformation, transformationOk := optNilTransformation.Get()
	if !transformationOk {
		return NewWebhookTransformationValueNull(), nil
	}

	diags := diag.Diagnostics{}

	methodValue := types.StringPointerValue(transformation.Method.ValueStringPointer())
	contentTypeValue := types.StringPointerValue(transformation.ContentType.ValueStringPointer())
	includeContentLengthValue := types.BoolPointerValue(transformation.IncludeContentLength.ValueBoolPointer())

	bodyValue := jsontypes.NewNormalizedNull()
	if transformation.Body != nil {
		bodyValue = jsontypes.NewNormalizedValue(transformation.Body.String())
	}

	value, valueDiags := NewWebhookTransformationValueKnownFromAttributes(ctx, map[string]attr.Value{
		"method":                 methodValue,
		"content_type":           contentTypeValue,
		"include_content_length": includeContentLengthValue,
		"body":                   bodyValue,
	})
	diags.Append(valueDiags...)

	return value, diags
}
