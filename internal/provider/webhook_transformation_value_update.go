//nolint:dupl
package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func ToOptNilUpdateWebhookDefinitionReqTransformation(_ context.Context, _ path.Path, value WebhookTransformationValue) (cm.OptNilUpdateWebhookDefinitionReqTransformation, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	optNilTransformation := cm.OptNilUpdateWebhookDefinitionReqTransformation{}

	switch {
	case value.IsUnknown():
		//
	case value.IsNull():
		optNilTransformation.SetToNull()
	default:
		transformation := cm.UpdateWebhookDefinitionReqTransformation{}

		transformation.Method = cm.NewOptPointerString(value.Method.ValueStringPointer())
		transformation.ContentType = cm.NewOptPointerString(value.ContentType.ValueStringPointer())
		transformation.IncludeContentLength = cm.NewOptPointerBool(value.IncludeContentLength.ValueBoolPointer())

		//nolint:revive
		if value.Body.IsUnknown() {
		} else {
			bodyStringPointer := value.Body.ValueStringPointer()
			if bodyStringPointer != nil {
				transformation.Body = []byte(*bodyStringPointer)
			} else {
				transformation.Body = nil
			}
		}

		optNilTransformation.SetTo(transformation)
	}

	return optNilTransformation, diags
}
