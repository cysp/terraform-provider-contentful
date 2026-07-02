package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func ReadHeaderValueMapFromResponse(ctx context.Context, path path.Path, headers []cm.WebhookDefinitionHeader, existingHeaderValues TypedMap[TypedObject[WebhookHeaderValue]]) (TypedMap[TypedObject[WebhookHeaderValue]], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if len(headers) == 0 && existingHeaderValues.IsNull() {
		return NewTypedMapNull[TypedObject[WebhookHeaderValue]](), diags
	}

	headersValues := make(map[string]TypedObject[WebhookHeaderValue], len(headers))

	for _, header := range headers {
		existingHeader := existingHeaderValues.Elements()[header.Key]

		value, valueDiags := NewWebhookHeaderValueFromResponse(ctx, path.AtMapKey(header.Key), header, existingHeader)
		diags.Append(valueDiags...)

		headersValues[header.Key] = value
	}

	headersList := NewTypedMap(headersValues)

	return headersList, diags
}
