package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AppInstallationModel struct {
	ID              types.String         `tfsdk:"id"`
	SpaceID         types.String         `tfsdk:"space_id"`
	EnvironmentID   types.String         `tfsdk:"environment_id"`
	AppDefinitionID types.String         `tfsdk:"app_definition_id"`
	Marketplace     types.Set            `tfsdk:"marketplace"`
	Parameters      jsontypes.Normalized `tfsdk:"parameters"`
}
