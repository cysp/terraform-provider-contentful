package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (m *AppSigningSecretModel) ToAppSigningSecretRequest(_ context.Context, _ path.Path) (cm.AppSigningSecretRequestData, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	req := cm.AppSigningSecretRequestData{
		Value: m.Value.ValueString(),
	}

	return req, diags
}
