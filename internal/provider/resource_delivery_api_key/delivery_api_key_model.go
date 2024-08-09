package resource_delivery_api_key

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
)

func (model *DeliveryApiKeyModel) ToPostAPIKeyReq() (contentfulManagement.PostApiKeyReq, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	req := contentfulManagement.PostApiKeyReq{
		Name:        model.Name.ValueString(),
		Description: util.StringValueToOptNilString(model.Description),
		// Environments: ,
	}

	return req, diags
}

func (model *DeliveryApiKeyModel) ToPutAPIKeyReq() (contentfulManagement.PutApiKeyReq, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	req := contentfulManagement.PutApiKeyReq{
		Name:        model.Name.ValueString(),
		Description: util.StringValueToOptNilString(model.Description),
		// Environments: ,
	}

	return req, diags
}

func (model *DeliveryApiKeyModel) ReadFromResponse(apiKey *contentfulManagement.ApiKey) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model.ApiKeyId = types.StringValue(apiKey.Sys.ID)

	model.Name = types.StringValue(apiKey.Name)
	model.Description = util.OptNilStringToStringValue(apiKey.Description)

	model.AccessToken = types.StringValue(apiKey.AccessToken)

	// model.Environments = make([]EnvironmentValue, len(previewApiKey.Environments))

	if previewAPIKey, ok := apiKey.PreviewAPIKey.Get(); ok {
		model.PreviewApiKeyId = types.StringValue(previewAPIKey.Sys.ID)
	} else {
		model.PreviewApiKeyId = types.StringNull()
	}

	return diags
}
