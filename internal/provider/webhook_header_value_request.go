package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (v *WebhookHeaderValue) ToWebhookDefinitionHeader(_ context.Context, path path.Path, key string) (cm.WebhookDefinitionHeader, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	header := cm.WebhookDefinitionHeader{
		Key: key,
	}

	if v.Value.IsNull() || v.Value.IsUnknown() {
		diags.AddAttributeError(path.AtName("value"), "Value is required", "")
	}

	header.Value = cm.NewOptPointerString(v.Value.ValueStringPointer())

	header.Secret = cm.NewOptBool(v.Secret.ValueBool())

	return header, diags
}
