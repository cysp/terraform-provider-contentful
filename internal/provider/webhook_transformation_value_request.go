package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func ToOptNilWebhookDefinitionDataTransformation(_ context.Context, valuePath path.Path, value TypedObject[WebhookTransformationValue]) (cm.OptNilWebhookDefinitionDataTransformation, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	optNilTransformation := cm.OptNilWebhookDefinitionDataTransformation{}

	switch {
	case value.IsUnknown():
		diags.AddAttributeError(valuePath, "Unexpected unknown webhook transformation", "The webhook transformation must be known before it can be sent to Contentful.")
	case value.IsNull():
		optNilTransformation.SetToNull()
	default:
		transformation := cm.WebhookDefinitionDataTransformation{}

		value := value.Value()

		if value.Method.IsUnknown() {
			diags.AddAttributeError(valuePath.AtName("method"), "Unexpected unknown transformation method", "The transformation method must be known before it can be sent to Contentful.")
		} else {
			transformation.Method = cm.NewOptPointerString(value.Method.ValueStringPointer())
		}

		if value.ContentType.IsUnknown() {
			diags.AddAttributeError(valuePath.AtName("content_type"), "Unexpected unknown transformation content type", "The transformation content type must be known before it can be sent to Contentful.")
		} else {
			transformation.ContentType = cm.NewOptPointerString(value.ContentType.ValueStringPointer())
		}

		if value.IncludeContentLength.IsUnknown() {
			diags.AddAttributeError(valuePath.AtName("include_content_length"), "Unexpected unknown content-length setting", "The content-length setting must be known before it can be sent to Contentful.")
		} else {
			transformation.IncludeContentLength = cm.NewOptPointerBool(value.IncludeContentLength.ValueBoolPointer())
		}

		if value.Body.IsUnknown() {
			diags.AddAttributeError(valuePath.AtName("body"), "Unexpected unknown transformation body", "The transformation body must be known before it can be sent to Contentful.")
		} else {
			bodyStringPointer := value.Body.ValueStringPointer()
			if bodyStringPointer != nil {
				transformation.Body = []byte(*bodyStringPointer)
			} else {
				transformation.Body = nil
			}
		}

		if !diags.HasError() {
			optNilTransformation.SetTo(transformation)
		}
	}

	return optNilTransformation, diags
}
