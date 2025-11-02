package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func (model *EnvironmentModel) ToEnvironmentData(_ context.Context) (cm.EnvironmentData, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	environmentFields := cm.EnvironmentData{
		Name: model.Name.ValueString(),
	}

	return environmentFields, diags
}
