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

	headerIsSecret := header.Secret.Or(false)

	value := WebhookHeaderValue{}

	if existingHeaderValue, existingHeaderValueOk := existingHeaderValue.GetValue(); existingHeaderValueOk {
		value.Value = existingHeaderValue.Value
	}

	if headerValue, ok := header.Value.Get(); ok {
		value.Value = types.StringValue(headerValue)
	} else if !headerIsSecret {
		value.Value = types.StringNull()
	}

	value.Secret = types.BoolValue(headerIsSecret)

	return NewTypedObject(value), diags
}
