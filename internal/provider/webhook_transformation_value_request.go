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
		diags.AddAttributeError(
			valuePath,
			"Unexpected unknown webhook transformation",
			"The webhook transformation must be known before it can be sent to Contentful.",
		)
	case value.IsNull():
		optNilTransformation.SetToNull()
	default:
		value := value.Value()

		method, methodDiags := requestOptionalString(value.Method, valuePath.AtName("method"))
		diags.Append(methodDiags...)

		contentType, contentTypeDiags := requestOptionalString(value.ContentType, valuePath.AtName("content_type"))
		diags.Append(contentTypeDiags...)

		includeContentLength, includeContentLengthDiags := requestOptionalBool(
			value.IncludeContentLength,
			valuePath.AtName("include_content_length"),
		)
		diags.Append(includeContentLengthDiags...)

		var body []byte

		switch {
		case value.Body.IsUnknown():
			diags.AddAttributeError(
				valuePath.AtName("body"),
				"Unexpected unknown webhook transformation body",
				"The webhook transformation body must be known before it can be sent to Contentful.",
			)
		case value.Body.IsNull():
		default:
			body = []byte(value.Body.ValueString())
		}

		if diags.HasError() {
			return cm.OptNilWebhookDefinitionDataTransformation{}, diags
		}

		optNilTransformation.SetTo(cm.WebhookDefinitionDataTransformation{
			Method:               method,
			ContentType:          contentType,
			IncludeContentLength: includeContentLength,
			Body:                 body,
		})
	}

	return optNilTransformation, diags
}
