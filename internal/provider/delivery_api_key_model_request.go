package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (model *DeliveryAPIKeyModel) ToAPIKeyRequestFields(ctx context.Context, config DeliveryAPIKeyModel) (cm.ApiKeyRequestData, diag.Diagnostics) {
	diags := rejectUnknownConfigurationOwnedRequestValue(model.Environments, config.Environments, path.Root("environments"))

	name, nameDiags := requestRequiredString(model.Name, path.Root("name"))
	diags.Append(nameDiags...)

	description, descriptionDiags := requestNullableString(model.Description, path.Root("description"))
	diags.Append(descriptionDiags...)

	environments, environmentsDiags := ToEnvironmentLinks(ctx, path.Root("environments"), model.Environments)
	diags.Append(environmentsDiags...)

	if diags.HasError() {
		return cm.ApiKeyRequestData{}, diags
	}

	return cm.ApiKeyRequestData{
		Name:         name,
		Description:  description,
		Environments: environments,
	}, diags
}
