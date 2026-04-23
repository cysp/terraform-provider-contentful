package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type EnvironmentAliasIdentityModel struct {
	SpaceID            types.String `tfsdk:"space_id"`
	EnvironmentAliasID types.String `tfsdk:"environment_alias_id"`
}

type EnvironmentAliasModel struct {
	IDIdentityModel
	EnvironmentAliasIdentityModel

	TargetEnvironmentID types.String `tfsdk:"target_environment_id"`

	Timeouts timeouts.Value `tfsdk:"timeouts"`
}
