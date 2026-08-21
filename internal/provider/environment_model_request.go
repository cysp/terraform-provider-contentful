package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (model *EnvironmentModel) ToEnvironmentData(_ context.Context) (cm.EnvironmentData, diag.Diagnostics) {
	name, diags := requestRequiredString(model.Name, path.Root("name"))

	if diags.HasError() {
		return cm.EnvironmentData{}, diags
	}

	environmentFields := cm.EnvironmentData{
		Name: name,
	}

	return environmentFields, diags
}

func (model *EnvironmentModel) ToSourceEnvironmentHeader() (cm.OptString, diag.Diagnostics) {
	if model.SourceEnvironmentID.IsUnknown() {
		diags := diag.Diagnostics{}
		diags.AddAttributeError(
			path.Root("source_environment_id"),
			"Unexpected unknown source environment ID",
			"The source environment ID must be known before it can be sent to Contentful.",
		)

		return cm.OptString{}, diags
	}

	if model.SourceEnvironmentID.IsNull() || model.SourceEnvironmentID.ValueString() == "" {
		return cm.OptString{}, nil
	}

	return cm.NewOptString(model.SourceEnvironmentID.ValueString()), nil
}
