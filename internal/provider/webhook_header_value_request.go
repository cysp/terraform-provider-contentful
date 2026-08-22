package provider

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (v WebhookHeaderValue) ToWebhookDefinitionHeader(
	path path.Path,
	key string,
	configuredHeaders TypedMap[TypedObject[WebhookHeaderValue]],
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

	if v.Value.IsNull() && secret && configuredHeaders.IsNull() {
		// Contentful redacts secret values. For response-owned imported
		// headers, sending secret=true without value asks CMA to retain the
		// existing secret during an update.
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
