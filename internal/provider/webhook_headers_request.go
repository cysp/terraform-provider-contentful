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

		header, headerDiags := headersValue.Value().ToWebhookDefinitionHeader(ctx, path.AtMapKey(key), key)
		diags.Append(headerDiags...)

		headers[index] = header
	}

	return headers, diags
}
