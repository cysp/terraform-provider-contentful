package provider

import (
	"context"
	"slices"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func ToWebhookDefinitionHeaders(ctx context.Context, path path.Path, model TypedMap[TypedObject[WebhookHeaderValue]]) (cm.WebhookDefinitionHeaders, diag.Diagnostics) {
	if model.IsNull() {
		return nil, nil
	}

	diags := diag.Diagnostics{}

	if model.IsUnknown() {
		diags.AddAttributeError(path, "Unexpected unknown headers", "Webhook headers must be known before they can be sent to Contentful.")

		return nil, diags
	}

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
		headerPath := path.AtMapKey(key)

		value, valueDiags := KnownObjectValue(headersValue, headerPath)
		diags.Append(valueDiags...)

		if valueDiags.HasError() {
			continue
		}

		header, headerDiags := value.ToWebhookDefinitionHeader(ctx, headerPath, key)
		diags.Append(headerDiags...)

		headers[index] = header
	}

	return headers, diags
}
