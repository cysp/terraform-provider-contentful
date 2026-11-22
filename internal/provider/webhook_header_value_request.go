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

	diags.Append(v.validateValueOmission(path)...)
	value, valueDiags := requestOmittableString(v.Value, path.AtName("value"))
	diags.Append(valueDiags...)

	if diags.HasError() {
		return cm.WebhookDefinitionHeader{}, diags
	}

	return cm.WebhookDefinitionHeader{Key: key, Secret: cm.NewOptBool(secret), Value: value}, diags
}

func (v WebhookHeaderValue) validateValueOmission(headerPath path.Path) diag.Diagnostics {
	if !v.Value.IsNull() || v.Secret.IsUnknown() || v.Secret.ValueBool() {
		return nil
	}

	return diag.Diagnostics{diag.NewAttributeErrorDiagnostic(
		headerPath.AtName("value"), "Missing webhook header value",
		"Supply a value for this ordinary webhook header.",
	)}
}
