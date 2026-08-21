package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (model *EnvironmentAliasModel) ToEnvironmentAliasData(_ context.Context) (cm.EnvironmentAliasData, diag.Diagnostics) {
	targetEnvironmentID, diags := requestRequiredString(model.TargetEnvironmentID, path.Root("target_environment_id"))

	if diags.HasError() {
		return cm.EnvironmentAliasData{}, diags
	}

	environmentAliasRequest := cm.EnvironmentAliasData{
		Environment: cm.NewEnvironmentLink(targetEnvironmentID),
	}

	return environmentAliasRequest, diags
}
