package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (m *AppSigningSecretModel) ToAppSigningSecretRequest(_ context.Context, valuePath path.Path) (cm.AppSigningSecretRequestData, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	value, valueDiags := requestRequiredString(m.Value, valuePath.AtName("value"))
	diags.Append(valueDiags...)

	req := cm.AppSigningSecretRequestData{
		Value: value,
	}

	return req, diags
}

func AppSigningSecretModelWithWriteOnlySecrets(plan, config AppSigningSecretModel) (AppSigningSecretModel, WriteOnlySecretValues, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	values := WriteOnlySecretValues{}
	model := plan

	value, usedWriteOnly, valueDiags := resolveStringSecret(
		config.Value,
		config.ValueWO,
		path.Root("value"),
		path.Root("value_wo"),
	)
	diags.Append(valueDiags...)

	model.Value = value
	model.ValueWO = types.StringNull()

	if usedWriteOnly {
		values.Add(path.Root("value_wo"), config.ValueWO)
	}

	return model, values, diags
}
