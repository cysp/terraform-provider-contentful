package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func ReadHeaderValueMapFromResponse(ctx context.Context, path path.Path, headers []cm.WebhookDefinitionHeader, reference TypedMap[TypedObject[WebhookHeaderValue]]) (TypedMap[TypedObject[WebhookHeaderValue]], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	// CMA returns an empty array for both an omitted and an explicitly empty map.
	if len(headers) == 0 && reference.IsNull() {
		return reference, diags
	}

	headersValues := make(map[string]TypedObject[WebhookHeaderValue], len(headers))

	for _, header := range headers {
		fallbackHeader := reference.Elements()[header.Key]

		value, valueDiags := NewWebhookHeaderValueFromResponse(ctx, path.AtMapKey(header.Key), header, fallbackHeader)
		diags.Append(valueDiags...)

		headersValues[header.Key] = value
	}

	headersList := NewTypedMap(headersValues)

	return headersList, diags
}
