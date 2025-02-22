package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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
