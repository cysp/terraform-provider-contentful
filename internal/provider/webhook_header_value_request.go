package provider

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (v WebhookHeaderValue) ToWebhookDefinitionHeader(
	path path.Path,
	key string,
) (cm.WebhookDefinitionHeader, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	secret, secretDiags := requestRequiredBool(v.Secret, path.AtName("secret"))
	diags.Append(secretDiags...)

	if diags.HasError() {
		return cm.WebhookDefinitionHeader{}, diags
	}

	header := cm.WebhookDefinitionHeader{
		Key:    key,
		Secret: cm.NewOptBool(secret),
	}

	if v.Value.IsNull() && secret {
		// The lifecycle fills new write-only values and checks that a
		// value-less header can retain an existing secret before sending.
		return header, diags
	}

	value, valueDiags := requestRequiredString(v.Value, path.AtName("value"))
	diags.Append(valueDiags...)

	if diags.HasError() {
		return cm.WebhookDefinitionHeader{}, diags
	}

	header.Value = cm.NewOptString(value)

	return header, diags
}
