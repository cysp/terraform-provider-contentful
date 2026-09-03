package provider

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewPreviewAPIKeyDataSourceModelFromResponse(previewAPIKey cm.PreviewApiKey) PreviewAPIKeyModel {
	model := PreviewAPIKeyModel{
		SpaceID:         types.StringValue(previewAPIKey.Sys.Space.Sys.ID),
		PreviewAPIKeyID: types.StringValue(previewAPIKey.Sys.ID),
	}

	model.Name = types.StringValue(previewAPIKey.Name)
	model.Description = util.OptNilStringToStringValue(previewAPIKey.Description)

	model.AccessToken = types.StringValue(previewAPIKey.AccessToken)

	model.Environments = NewEnvironmentIDsListValueFromEnvironmentLinks(previewAPIKey.Environments)

	return model
}
