package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (m *AppSigningSecretModel) ToAppSigningSecretRequest(_ context.Context, modelPath path.Path) (cm.AppSigningSecretRequestData, diag.Diagnostics) {
	selected, valuePath, diags := resolveSigningSecretValue(m.Value, m.ValueWO, modelPath)
	if diags.HasError() {
		return cm.AppSigningSecretRequestData{}, diags
	}

	value, valueDiags := requestRequiredString(selected, valuePath)
	diags.Append(valueDiags...)

	if diags.HasError() {
		return cm.AppSigningSecretRequestData{}, diags
	}

	diags.Append(validateAppSigningSecretValue(value, valuePath)...)

	req := cm.AppSigningSecretRequestData{
		Value: value,
	}

	return req, diags
}
