package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

func (model *DeliveryAPIKeyModel) ToAPIKeyRequestFields(ctx context.Context) (cm.ApiKeyRequestData, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	req := cm.ApiKeyRequestData{
		Name:        model.Name.ValueString(),
		Description: util.StringValueToOptNilString(model.Description),
	}

	environments, environmentsDiags := ToEnvironmentLinks(ctx, path.Root("environments"), model.Environments)
	diags.Append(environmentsDiags...)

	req.Environments = environments

	return req, diags
}
