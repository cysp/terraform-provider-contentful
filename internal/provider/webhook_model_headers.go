package provider

import (
	"context"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewHeadersValueKnown() HeadersValue {
	return HeadersValue{
		state: attr.ValueStateKnown,
	}
}

func ReadHeadersListValueFromResponse(ctx context.Context, path path.Path, model types.List, headers []contentfulManagement.WebhookDefinitionHeader) (types.List, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	headersValues := make([]attr.Value, len(headers))

	existingHeaderValues := make([]HeadersValue, len(headers))
	diags.Append(model.ElementsAs(ctx, &existingHeaderValues, false)...)

	existingHeaderValuesByKey := make(map[string]HeadersValue, len(existingHeaderValues))
	for _, value := range existingHeaderValues {
		existingHeaderValuesByKey[value.Key.ValueString()] = value
	}

	for index, header := range headers {
		var value HeadersValue
		if existingHeader, found := existingHeaderValuesByKey[header.Key]; found {
			value = existingHeader
		} else {
			value = NewHeadersValueKnown()
		}

		diags.Append(value.ReadFromResponse(ctx, path.AtListIndex(index), header)...)
		headersValues[index] = value
	}

	headersList, headersListDiags := types.ListValue(HeadersValue{}.Type(ctx), headersValues)
	diags.Append(headersListDiags...)

	return headersList, diags
}

func (model *HeadersValue) ReadFromResponse(_ context.Context, _ path.Path, header contentfulManagement.WebhookDefinitionHeader) diag.Diagnostics {
	diags := diag.Diagnostics{}

	headerIsSecret := header.Secret.Or(false)

	model.Key = types.StringValue(header.Key)

	if value, ok := header.Value.Get(); ok {
		model.Value = types.StringValue(value)
	} else if !headerIsSecret {
		model.Value = types.StringNull()
	}

	model.Secret = types.BoolValue(headerIsSecret)

	return diags
}

func ToWebhookDefinitionHeaders(ctx context.Context, path path.Path, model types.List) (contentfulManagement.WebhookDefinitionHeaders, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	headers := make(contentfulManagement.WebhookDefinitionHeaders, len(model.Elements()))

	headersValues := make([]HeadersValue, len(model.Elements()))
	diags.Append(model.ElementsAs(ctx, &headersValues, false)...)

	for index, headersValue := range headersValues {
		header, headerDiags := headersValue.ToWebhookDefinitionHeader(ctx, path.AtListIndex(index))
		diags.Append(headerDiags...)

		headers[index] = header
	}

	return headers, diags
}

func (model *HeadersValue) ToWebhookDefinitionHeader(_ context.Context, path path.Path) (contentfulManagement.WebhookDefinitionHeader, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	header := contentfulManagement.WebhookDefinitionHeader{
		Key: model.Key.ValueString(),
	}

	if model.Value.IsNull() || model.Value.IsUnknown() {
		diags.AddAttributeError(path.AtName("value"), "Value is required", "")
	}

	header.Value = contentfulManagement.NewOptPointerString(model.Value.ValueStringPointer())

	header.Secret = contentfulManagement.NewOptBool(model.Secret.ValueBool())

	return header, diags
}
