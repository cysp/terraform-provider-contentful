package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ReadHeaderValueMapFromResponse(ctx context.Context, path path.Path, model types.Map, headers []cm.WebhookDefinitionHeader) (types.Map, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	existingHeaderValues := make(map[string]WebhookHeaderValue, len(model.Elements()))
	if !model.IsNull() && !model.IsUnknown() {
		diags.Append(model.ElementsAs(ctx, &existingHeaderValues, false)...)
	}

	headersValues := make(map[string]attr.Value, len(headers))

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

	headersList, headersListDiags := types.MapValue(WebhookHeaderValue{}.Type(ctx), headersValues)
	diags.Append(headersListDiags...)

	return headersList, diags
}
