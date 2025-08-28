package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func ReadHeaderValueMapFromResponse(ctx context.Context, path path.Path, headers []cm.WebhookDefinitionHeader, existingHeaderValues map[string]TypedObject[WebhookHeaderValue]) (TypedMap[TypedObject[WebhookHeaderValue]], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	headersValues := make(map[string]TypedObject[WebhookHeaderValue], len(headers))

	for _, header := range headers {
		existingHeader := existingHeaderValues[header.Key]

		value, valueDiags := NewWebhookHeaderValueFromResponse(ctx, path.AtMapKey(header.Key), header, existingHeader)
		diags.Append(valueDiags...)

		headersValues[header.Key] = value
	}

	headersList := NewTypedMap(headersValues)

	return headersList, diags
}
