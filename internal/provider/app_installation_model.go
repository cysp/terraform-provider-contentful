package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AppInstallationIdentityModel struct {
	SpaceID         types.String `tfsdk:"space_id"`
	EnvironmentID   types.String `tfsdk:"environment_id"`
	AppDefinitionID types.String `tfsdk:"app_definition_id"`
}

type AppInstallationModel struct {
	IDIdentityModel
	AppInstallationIdentityModel

	Marketplace types.Set            `tfsdk:"marketplace"`
	Parameters  jsontypes.Normalized `tfsdk:"parameters"`
}
