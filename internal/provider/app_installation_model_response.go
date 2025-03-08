package provider

import (
	"strings"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (model *AppInstallationModel) ReadFromResponse(appInstallation *cm.AppInstallation) diag.Diagnostics {
	diags := diag.Diagnostics{}

	spaceID := appInstallation.Sys.Space.Sys.ID
	environmentID := appInstallation.Sys.Environment.Sys.ID
	appDefinitionID := appInstallation.Sys.AppDefinition.Sys.ID

	model.ID = types.StringValue(strings.Join([]string{spaceID, environmentID, appDefinitionID}, "/"))
	model.SpaceID = types.StringValue(spaceID)
	model.EnvironmentID = types.StringValue(environmentID)
	model.AppDefinitionID = types.StringValue(appDefinitionID)

	if appInstallation.Parameters != nil {
		constraint, err := util.JxNormalizeOpaqueBytes(appInstallation.Parameters, util.JxEncodeOpaqueOptions{EscapeStrings: true})
		if err != nil {
			diags.AddAttributeError(path.Root("parameters"), "Failed to read parameters", err.Error())
		}

		model.Parameters = jsontypes.NewNormalizedValue(string(constraint))
	} else {
		model.Parameters = jsontypes.NewNormalizedNull()
	}

	return diags
}
