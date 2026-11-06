package provider

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ReadWebhookTransformationValueFromResponse(optNilTransformation cm.OptNilWebhookDefinitionTransformation) TypedObject[WebhookTransformationValue] {
	transformation, transformationOk := optNilTransformation.Get()
	if !transformationOk {
		return NewTypedObjectNull[WebhookTransformationValue]()
	}

	bodyValue := jsontypes.NewNormalizedNull()
	if transformation.Body != nil {
		bodyValue = NewNormalizedJSONValue(transformation.Body)
	}

	return NewTypedObject(WebhookTransformationValue{
		Method:               types.StringPointerValue(transformation.Method.ValueStringPointer()),
		ContentType:          types.StringPointerValue(transformation.ContentType.ValueStringPointer()),
		IncludeContentLength: types.BoolPointerValue(transformation.IncludeContentLength.ValueBoolPointer()),
		Body:                 bodyValue,
	})
}
