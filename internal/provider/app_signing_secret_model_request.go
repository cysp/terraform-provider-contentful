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

	return newAppSigningSecretRequest(value, valuePath)
}

func newAppSigningSecretRequest(value string, valuePath path.Path) (cm.AppSigningSecretRequestData, diag.Diagnostics) {
	request := cm.AppSigningSecretRequestData{Value: value}
	// Use a fixed diagnostic so validation errors never disclose the secret.
	if request.Validate() != nil {
		return cm.AppSigningSecretRequestData{}, diag.Diagnostics{
			diag.NewAttributeErrorDiagnostic(valuePath, "Invalid app signing secret value", appSigningValueConstraint+"."),
		}
	}

	return request, nil
}
