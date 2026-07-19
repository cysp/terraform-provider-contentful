package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (v WebhookHeaderValue) ToWebhookDefinitionHeader(_ context.Context, path path.Path, key string) (cm.WebhookDefinitionHeader, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	header := cm.WebhookDefinitionHeader{
		Key: key,
	}

	value, valueDiags := KnownStringValue(v.Value, path.AtName("value"))
	diags.Append(valueDiags...)

	secret, secretDiags := KnownBoolValue(v.Secret, path.AtName("secret"))
	diags.Append(secretDiags...)

	if !diags.HasError() {
		header.Value = cm.NewOptString(value)
		header.Secret = cm.NewOptBool(secret)
	}

	return header, diags
}
