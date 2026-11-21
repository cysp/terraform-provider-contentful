package provider

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (v WebhookHeaderValue) ToWebhookDefinitionHeader(
	path path.Path,
	key string,
	prior TypedObject[WebhookHeaderValue],
) (cm.WebhookDefinitionHeader, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	secret, secretDiags := requestRequiredBool(v.Secret, path.AtName("secret"))
	diags.Append(secretDiags...)

	if diags.HasError() {
		return cm.WebhookDefinitionHeader{}, diags
	}

	diags.Append(v.validateValueOmission(path, prior)...)
	value, valueDiags := requestOmittableString(v.Value, path.AtName("value"))
	diags.Append(valueDiags...)

	if diags.HasError() {
		return cm.WebhookDefinitionHeader{}, diags
	}

	return cm.WebhookDefinitionHeader{Key: key, Secret: cm.NewOptBool(secret), Value: value}, diags
}

// A null value asks Contentful to keep the secret at this exact key.
// An unknown flag can permit omission only if there is a prior secret to retain.
func (v WebhookHeaderValue) validateValueOmission(headerPath path.Path, prior TypedObject[WebhookHeaderValue]) diag.Diagnostics {
	if !v.Value.IsNull() {
		return nil
	}

	previous, present := prior.GetValue()
	if present && previous.Secret.ValueBool() && (v.Secret.ValueBool() || v.Secret.IsUnknown()) {
		return nil
	}

	return diag.Diagnostics{diag.NewAttributeErrorDiagnostic(
		headerPath.AtName("value"), "Missing webhook header value",
		"Supply a value for this header. You can omit value only to retain an existing secret header with the exact same name and secret = true.",
	)}
}
