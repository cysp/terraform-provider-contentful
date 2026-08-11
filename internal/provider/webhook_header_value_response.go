package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewWebhookHeaderValueFromResponse(_ context.Context, _ path.Path, header cm.WebhookDefinitionHeader, existingHeaderValue TypedObject[WebhookHeaderValue]) (TypedObject[WebhookHeaderValue], diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value := WebhookHeaderValue{}
	existingValueIsNull := false

	if existingHeaderValue, existingHeaderValueOk := existingHeaderValue.GetValue(); existingHeaderValueOk {
		value.Value = existingHeaderValue.Value
		value.Secret = existingHeaderValue.Secret
		existingValueIsNull = existingHeaderValue.Value.IsNull()
	}

	headerIsSecret := false
	if headerSecret, ok := header.Secret.Get(); ok {
		headerIsSecret = headerSecret
	} else if !value.Secret.IsNull() && !value.Secret.IsUnknown() {
		headerIsSecret = value.Secret.ValueBool()
	}

	if existingValueIsNull && !value.Secret.IsNull() && !value.Secret.IsUnknown() && value.Secret.ValueBool() {
		headerIsSecret = true
	}

	if headerValue, ok := header.Value.Get(); ok && !existingValueIsNull {
		value.Value = types.StringValue(headerValue)
	} else if !headerIsSecret {
		value.Value = types.StringNull()
	}

	value.Secret = types.BoolValue(headerIsSecret)

	return NewTypedObject(value), diags
}
