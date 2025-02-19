package provider

import (
	"context"
	"slices"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ReadHeadersListValueFromResponse(ctx context.Context, path path.Path, model types.Map, headers []cm.WebhookDefinitionHeader) (types.Map, diag.Diagnostics) {
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

func (v *WebhookHeaderValue) ReadFromResponse(_ context.Context, _ path.Path, header cm.WebhookDefinitionHeader) diag.Diagnostics {
	diags := diag.Diagnostics{}

	headerIsSecret := header.Secret.Or(false)

	if value, ok := header.Value.Get(); ok {
		v.Value = types.StringValue(value)
	} else if !headerIsSecret {
		v.Value = types.StringNull()
	}

	v.Secret = types.BoolValue(headerIsSecret)

	return diags
}

func ToWebhookDefinitionHeaders(ctx context.Context, path path.Path, model types.Map) (cm.WebhookDefinitionHeaders, diag.Diagnostics) {
	if model.IsNull() || model.IsUnknown() {
		return nil, nil
	}

	diags := diag.Diagnostics{}

	headers := make(cm.WebhookDefinitionHeaders, len(model.Elements()))

	headersValues := make(map[string]WebhookHeaderValue, len(model.Elements()))
	diags.Append(model.ElementsAs(ctx, &headersValues, false)...)

	headersKeys := make([]string, len(headersValues))

	index := 0

	for key := range headersValues {
		headersKeys[index] = key
		index++
	}

	slices.Sort(headersKeys)

	for index, key := range headersKeys {
		headersValue := headersValues[key]

		header, headerDiags := headersValue.ToWebhookDefinitionHeader(ctx, path.AtMapKey(key), key)
		diags.Append(headerDiags...)

		headers[index] = header
	}

	return headers, diags
}

func (v *WebhookHeaderValue) ToWebhookDefinitionHeader(_ context.Context, path path.Path, key string) (cm.WebhookDefinitionHeader, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	header := cm.WebhookDefinitionHeader{
		Key: key,
	}

	if v.Value.IsNull() || v.Value.IsUnknown() {
		diags.AddAttributeError(path.AtName("value"), "Value is required", "")
	}

	header.Value = cm.NewOptPointerString(v.Value.ValueStringPointer())

	header.Secret = cm.NewOptBool(v.Secret.ValueBool())

	return header, diags
}
