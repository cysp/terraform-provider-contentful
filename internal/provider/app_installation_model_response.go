package provider

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// NewAppInstallationResourceModelFromResponse uses the effective Plan after a
// successful mutation, or prior State during Read, as the reference for values
// that Contentful does not return in full.
func NewAppInstallationResourceModelFromResponse(appInstallation cm.AppInstallation, reference AppInstallationModel) (AppInstallationModel, diag.Diagnostics) {
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

	parameters, diags := decodeInstallationParameters(path.Root("parameters"), appInstallation.Parameters)
	model.Parameters = reconcileAppInstallationParameters(parameters, reference.Parameters)

	return model, diags
}
