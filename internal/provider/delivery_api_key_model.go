package provider

import (
	"context"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (model *DeliveryApiKeyModel) ToCreateDeliveryAPIKeyReq(ctx context.Context) (contentfulManagement.CreateDeliveryApiKeyReq, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	req := contentfulManagement.CreateDeliveryApiKeyReq{
		Name:        model.Name.ValueString(),
		Description: util.StringValueToOptNilString(model.Description),
	}

	environments, environmentsDiags := ToEnvironmentLinks(ctx, path.Root("environments"), model.Environments)
	diags.Append(environmentsDiags...)

	req.Environments = environments

	return req, diags
}

func (model *DeliveryApiKeyModel) ToUpdateDeliveryAPIKeyReq(ctx context.Context) (contentfulManagement.UpdateDeliveryApiKeyReq, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	req := contentfulManagement.UpdateDeliveryApiKeyReq{
		Name:        model.Name.ValueString(),
		Description: util.StringValueToOptNilString(model.Description),
	}

	environments, environmentsDiags := ToEnvironmentLinks(ctx, path.Root("environments"), model.Environments)
	diags.Append(environmentsDiags...)

	req.Environments = environments

	return req, diags
}

func (model *DeliveryApiKeyModel) ReadFromResponse(ctx context.Context, apiKey *contentfulManagement.ApiKey) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model.ApiKeyId = types.StringValue(apiKey.Sys.ID)

	model.Name = types.StringValue(apiKey.Name)
	model.Description = util.OptNilStringToStringValue(apiKey.Description)

	model.AccessToken = types.StringValue(apiKey.AccessToken)

	environmentsList, environmentsListDiags := NewEnvironmentIDsListValueFromEnvironmentLinks(ctx, path.Root("environments"), apiKey.Environments)
	diags.Append(environmentsListDiags...)

	model.Environments = environmentsList

	if previewAPIKey, ok := apiKey.PreviewAPIKey.Get(); ok {
		model.PreviewApiKeyId = types.StringValue(previewAPIKey.Sys.ID)
	} else {
		model.PreviewApiKeyId = types.StringNull()
	}

	return diags
}
