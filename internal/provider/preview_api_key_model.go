package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (model *PreviewApiKeyModel) ReadFromResponse(ctx context.Context, previewAPIKey *cm.PreviewApiKey) diag.Diagnostics {
	diags := diag.Diagnostics{}

	model.PreviewApiKeyId = types.StringValue(previewAPIKey.Sys.ID)

	model.Name = types.StringValue(previewAPIKey.Name)
	model.Description = util.OptNilStringToStringValue(previewAPIKey.Description)

	model.AccessToken = types.StringValue(previewAPIKey.AccessToken)

	environmentsList, environmentsListDiags := NewEnvironmentIDsListValueFromEnvironmentLinks(ctx, path.Root("environments"), previewAPIKey.Environments)
	diags.Append(environmentsListDiags...)

	model.Environments = environmentsList

	return diags
}
