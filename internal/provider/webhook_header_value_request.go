package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (v WebhookHeaderValue) ToWebhookDefinitionHeader(_ context.Context, path path.Path, key string) (cm.WebhookDefinitionHeader, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value, valueDiags := requestRequiredString(v.Value, path.AtName("value"))
	diags.Append(valueDiags...)

	secret, secretDiags := requestRequiredBool(v.Secret, path.AtName("secret"))
	diags.Append(secretDiags...)

	if diags.HasError() {
		return cm.WebhookDefinitionHeader{}, diags
	}

	return cm.WebhookDefinitionHeader{
		Key:    key,
		Value:  cm.NewOptString(value),
		Secret: cm.NewOptBool(secret),
	}, diags
}
