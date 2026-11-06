package provider

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider/util"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// NewAppInstallationResourceModelFromResponse uses the effective Plan after a
// successful mutation, or prior State during Read, as the reference for values
// that Contentful does not return in full.
func NewAppInstallationResourceModelFromResponse(appInstallation cm.AppInstallation, reference AppInstallationModel) (AppInstallationModel, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	spaceID := appInstallation.Sys.Space.Sys.ID
	environmentID := appInstallation.Sys.Environment.Sys.ID
	appDefinitionID := appInstallation.Sys.AppDefinition.Sys.ID

	model := AppInstallationModel{
		IDIdentityModel: NewIDIdentityModelFromMultipartID(spaceID, environmentID, appDefinitionID),
		SpaceID:         types.StringValue(spaceID),
		EnvironmentID:   types.StringValue(environmentID),
		AppDefinitionID: types.StringValue(appDefinitionID),
		Marketplace:     reference.Marketplace,
		Timeouts:        reference.Timeouts,
	}

	if appInstallation.Parameters != nil {
		constraint, err := util.JxNormalizeOpaqueBytes(appInstallation.Parameters, util.JxEncodeOpaqueOptions{EscapeStrings: true})
		if err != nil {
			diags.AddAttributeError(path.Root("parameters"), "Failed to read parameters", err.Error())
		}

		model.Parameters = NewNormalizedJSONValue(constraint)
	} else {
		model.Parameters = jsontypes.NewNormalizedNull()
	}

	if diags.HasError() {
		return model, diags
	}

	parameters, parameterDiags := reconcileAppInstallationParameters(model.Parameters, reference.Parameters)
	diags.Append(parameterDiags...)

	model.Parameters = parameters

	return model, diags
}
