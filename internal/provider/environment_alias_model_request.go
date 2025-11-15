package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func (model *EnvironmentAliasModel) ToEnvironmentAliasData(_ context.Context) (cm.EnvironmentAliasData, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	environmentAliasRequest := cm.EnvironmentAliasData{
		Environment: cm.NewEnvironmentLink(model.TargetEnvironmentID.ValueString()),
	}

	return environmentAliasRequest, diags
}
