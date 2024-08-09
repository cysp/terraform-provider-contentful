package datasource_preview_api_key

import (
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cysp/terraform-provider-contentful/internal/provider/util"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (model *PreviewApiKeyModel) ReadFromResponse(previewAPIKey *contentfulManagement.PreviewApiKey) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model.PreviewApiKeyId = types.StringValue(previewAPIKey.Sys.ID)

	model.Name = types.StringValue(previewAPIKey.Name)
	model.Description = util.OptNilStringToStringValue(previewAPIKey.Description)

	model.AccessToken = types.StringValue(previewAPIKey.AccessToken)

	// model.Environments = make([]EnvironmentValue, len(previewApiKey.Environments))

	return diags
}
