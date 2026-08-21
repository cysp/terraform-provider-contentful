package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (m *AppSigningSecretModel) ToAppSigningSecretRequest(_ context.Context, modelPath path.Path) (cm.AppSigningSecretRequestData, diag.Diagnostics) {
	value, diags := requestRequiredString(m.Value, modelPath.AtName("value"))

	if diags.HasError() {
		return cm.AppSigningSecretRequestData{}, diags
	}

	req := cm.AppSigningSecretRequestData{
		Value: value,
	}

	return req, diags
}
