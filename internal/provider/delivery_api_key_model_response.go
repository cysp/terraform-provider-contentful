package provider

import (
	"context"
	"strings"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewDeliveryAPIKeyResourceModelFromResponse(ctx context.Context, apiKey cm.ApiKey) (DeliveryAPIKeyModel, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	spaceID := apiKey.Sys.Space.Sys.ID
	apiKeyID := apiKey.Sys.ID

	model := DeliveryAPIKeyModel{
		ID:       types.StringValue(strings.Join([]string{spaceID, apiKeyID}, "/")),
		SpaceID:  types.StringValue(spaceID),
		APIKeyID: types.StringValue(apiKeyID),
	}

	model.Name = types.StringValue(apiKey.Name)
	model.Description = util.OptNilStringToStringValue(apiKey.Description)

	model.AccessToken = types.StringValue(apiKey.AccessToken)

	environmentsList, environmentsListDiags := NewEnvironmentIDsListValueFromEnvironmentLinks(ctx, path.Root("environments"), apiKey.Environments)
	diags.Append(environmentsListDiags...)

	model.Environments = environmentsList

	if previewAPIKey, ok := apiKey.PreviewAPIKey.Get(); ok {
		model.PreviewAPIKeyID = types.StringValue(previewAPIKey.Sys.ID)
	} else {
		model.PreviewAPIKeyID = types.StringNull()
	}

	return model, diags
}
