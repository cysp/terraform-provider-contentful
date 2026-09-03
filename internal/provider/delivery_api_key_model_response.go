package provider

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewDeliveryAPIKeyResourceModelFromResponse(apiKey cm.ApiKey) DeliveryAPIKeyModel {
	spaceID := apiKey.Sys.Space.Sys.ID
	apiKeyID := apiKey.Sys.ID

	model := DeliveryAPIKeyModel{
		IDIdentityModel: NewIDIdentityModelFromMultipartID(spaceID, apiKeyID),
		DeliveryAPIKeyIdentityModel: DeliveryAPIKeyIdentityModel{
			SpaceID:  types.StringValue(spaceID),
			APIKeyID: types.StringValue(apiKeyID),
		},
	}

	model.Name = types.StringValue(apiKey.Name)
	model.Description = util.OptNilStringToStringValue(apiKey.Description)

	model.AccessToken = types.StringValue(apiKey.AccessToken)

	model.Environments = NewEnvironmentIDsListValueFromEnvironmentLinks(apiKey.Environments)

	if previewAPIKey, ok := apiKey.PreviewAPIKey.Get(); ok {
		model.PreviewAPIKeyID = types.StringValue(previewAPIKey.Sys.ID)
	} else {
		model.PreviewAPIKeyID = types.StringNull()
	}

	return model
}
