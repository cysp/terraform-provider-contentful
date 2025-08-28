package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func ReadHeaderValueMapFromResponse(ctx context.Context, path path.Path, headers []cm.WebhookDefinitionHeader, existingHeaderValues map[string]WebhookHeaderValue) (TypedMap[WebhookHeaderValue], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	headersValues := make(map[string]WebhookHeaderValue, len(headers))

	for _, header := range headers {
		var value WebhookHeaderValue
		if existingHeader, found := existingHeaderValues[header.Key]; found {
			value = existingHeader
		} else {
			value = NewWebhookHeaderValueKnown()
		}

		diags.Append(value.ReadFromResponse(ctx, path.AtMapKey(header.Key), header)...)
		headersValues[header.Key] = value
	}

	headersList := NewTypedMap(headersValues)

	return headersList, diags
}
