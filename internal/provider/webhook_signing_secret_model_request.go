package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (m *WebhookSigningSecretModel) ToWebhookSigningSecretRequest(_ context.Context, modelPath path.Path) (cm.WebhookSigningSecretRequestData, diag.Diagnostics) {
	value, diags := requestRequiredString(m.Value, modelPath.AtName("value"))
	if diags.HasError() {
		return cm.WebhookSigningSecretRequestData{}, diags
	}

	diags.Append(validateWebhookSigningSecretValue(value, modelPath.AtName("value"))...)

	return cm.WebhookSigningSecretRequestData{Value: value}, diags
}

func webhookSigningSecretSpaceID(value types.String) (string, diag.Diagnostics) {
	spaceID, diags := requestRequiredString(value, path.Root("space_id"))
	if !diags.HasError() && spaceID == "" {
		diags.AddAttributeError(path.Root("space_id"), "Invalid space ID", "The space ID must not be empty.")
	}

	return spaceID, diags
}
