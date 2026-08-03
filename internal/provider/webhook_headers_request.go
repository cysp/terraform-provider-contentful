package provider

import (
	"context"
	"slices"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func ToWebhookDefinitionHeaders(ctx context.Context, path path.Path, model TypedMap[TypedObject[WebhookHeaderValue]]) (cm.WebhookDefinitionHeaders, diag.Diagnostics) {
	if model.IsNull() || model.IsUnknown() {
		return nil, nil
	}

	diags := diag.Diagnostics{}

	headers := make(cm.WebhookDefinitionHeaders, len(model.Elements()))

	headersValues := model.Elements()

	headersKeys := make([]string, len(headersValues))

	index := 0

	for key := range headersValues {
		headersKeys[index] = key
		index++
	}

	slices.Sort(headersKeys)

	for index, key := range headersKeys {
		headersValue := headersValues[key]
		headerValue, ok := headersValue.GetValue()
		headerPath := path.AtMapKey(key)

		if !ok {
			if headersValue.IsUnknown() {
				diags.AddAttributeError(
					headerPath,
					"Unexpected unknown webhook header",
					"The webhook header must be known before it can be sent to Contentful.",
				)
			} else {
				diags.AddAttributeError(
					headerPath,
					"Unexpected null webhook header",
					"Null webhook headers cannot be sent to Contentful.",
				)
			}

			continue
		}

		header, headerDiags := headerValue.ToWebhookDefinitionHeader(ctx, headerPath, key)
		diags.Append(headerDiags...)

		if !headerDiags.HasError() {
			headers[index] = header
		}
	}

	if diags.HasError() {
		return nil, diags
	}

	return headers, diags
}
