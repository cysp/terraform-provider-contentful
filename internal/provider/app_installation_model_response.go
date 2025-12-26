package provider

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewAppInstallationResourceModelFromResponse(appInstallation cm.AppInstallation, marketplace types.Set) (AppInstallationModel, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	spaceID := appInstallation.Sys.Space.Sys.ID
	environmentID := appInstallation.Sys.Environment.Sys.ID
	appDefinitionID := appInstallation.Sys.AppDefinition.Sys.ID

	model := AppInstallationModel{
		IDIdentityModel: NewIDIdentityModelFromMultipartID(spaceID, environmentID, appDefinitionID),
		AppInstallationIdentityModel: AppInstallationIdentityModel{
			SpaceID:         types.StringValue(spaceID),
			EnvironmentID:   types.StringValue(environmentID),
			AppDefinitionID: types.StringValue(appDefinitionID),
		},
	}

	if appInstallation.Parameters != nil {
		constraint, err := util.JxNormalizeOpaqueBytes(appInstallation.Parameters, util.JxEncodeOpaqueOptions{EscapeStrings: true})
		if err != nil {
			diags.AddAttributeError(path.Root("parameters"), "Failed to read parameters", err.Error())
		}

		model.Parameters = NewNormalizedJSONTypesNormalizedValue(constraint)
	} else {
		model.Parameters = jsontypes.NewNormalizedNull()
	}

	model.Marketplace = marketplace

	return model, diags
}
