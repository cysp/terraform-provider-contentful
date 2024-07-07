package provider

import (
	"context"
	"fmt"
	"slices"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func NewHeadersValueKnown() HeadersValue {
	return HeadersValue{
		state: attr.ValueStateKnown,
	}
}

func NewHeadersValueKnownFromAttributes(_ context.Context, attributes map[string]attr.Value) (HeadersValue, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value, valueOk := attributes["value"].(basetypes.StringValue)
	if !valueOk {
		diags.AddError("Invalid data", fmt.Sprintf("expected value to be of type String, got %T", attributes["value"]))
	}

	secret, secretOk := attributes["secret"].(basetypes.BoolValue)
	if !secretOk {
		diags.AddError("Invalid data", fmt.Sprintf("expected secret to be of type Bool, got %T", attributes["secret"]))
	}

	return HeadersValue{
		Value:  value,
		Secret: secret,
		state:  attr.ValueStateKnown,
	}, diags
}

func NewHeadersValueKnownFromAttributesMust(ctx context.Context, attributes map[string]attr.Value) HeadersValue {
	value, diags := NewHeadersValueKnownFromAttributes(ctx, attributes)
	if diags.HasError() {
		panic(diags)
	}

	return value
}

func ReadHeadersListValueFromResponse(ctx context.Context, path path.Path, model types.Map, headers []contentfulManagement.WebhookDefinitionHeader) (types.Map, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	existingHeaderValues := make(map[string]HeadersValue, len(model.Elements()))
	if !model.IsNull() && !model.IsUnknown() {
		diags.Append(model.ElementsAs(ctx, &existingHeaderValues, false)...)
	}

	headersValues := make(map[string]attr.Value, len(headers))

	for _, header := range headers {
		var value HeadersValue
		if existingHeader, found := existingHeaderValues[header.Key]; found {
			value = existingHeader
		} else {
			value = NewHeadersValueKnown()
		}

		diags.Append(value.ReadFromResponse(ctx, path.AtMapKey(header.Key), header)...)
		headersValues[header.Key] = value
	}

	headersList, headersListDiags := types.MapValue(HeadersValue{}.Type(ctx), headersValues)
	diags.Append(headersListDiags...)

	return headersList, diags
}

func (model *HeadersValue) ReadFromResponse(_ context.Context, _ path.Path, header contentfulManagement.WebhookDefinitionHeader) diag.Diagnostics {
	diags := diag.Diagnostics{}

	headerIsSecret := header.Secret.Or(false)

	if value, ok := header.Value.Get(); ok {
		model.Value = types.StringValue(value)
	} else if !headerIsSecret {
		model.Value = types.StringNull()
	}

	model.Secret = types.BoolValue(headerIsSecret)

	return diags
}

func ToWebhookDefinitionHeaders(ctx context.Context, path path.Path, model types.Map) (contentfulManagement.WebhookDefinitionHeaders, diag.Diagnostics) {
	if model.IsNull() || model.IsUnknown() {
		return nil, nil
	}

	diags := diag.Diagnostics{}

	headers := make(contentfulManagement.WebhookDefinitionHeaders, len(model.Elements()))

	headersValues := make(map[string]HeadersValue, len(model.Elements()))
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

func (model *HeadersValue) ToWebhookDefinitionHeader(_ context.Context, path path.Path, key string) (contentfulManagement.WebhookDefinitionHeader, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	header := contentfulManagement.WebhookDefinitionHeader{
		Key: key,
	}

	if model.Value.IsNull() || model.Value.IsUnknown() {
		diags.AddAttributeError(path.AtName("value"), "Value is required", "")
	}

	header.Value = contentfulManagement.NewOptPointerString(model.Value.ValueStringPointer())

	header.Secret = contentfulManagement.NewOptBool(model.Secret.ValueBool())

	return header, diags
}
