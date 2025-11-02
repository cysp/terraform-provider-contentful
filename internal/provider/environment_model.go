package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type EnvironmentIdentityModel struct {
	SpaceID       types.String `tfsdk:"space_id"`
	EnvironmentID types.String `tfsdk:"environment_id"`
}

type EnvironmentModel struct {
	IDIdentityModel
	EnvironmentIdentityModel

	Name                types.String `tfsdk:"name"`
	SourceEnvironmentID types.String `tfsdk:"source_environment_id"`
}
