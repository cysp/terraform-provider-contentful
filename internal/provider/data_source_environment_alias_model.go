package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type EnvironmentAliasDataSourceItemModel struct {
	EnvironmentAliasID  types.String `tfsdk:"environment_alias_id"`
	TargetEnvironmentID types.String `tfsdk:"target_environment_id"`
}
type EnvironmentAliasDataSourceModel struct {
	IDIdentityModel
	EnvironmentAliasDataSourceItemModel

	SpaceID  types.String   `tfsdk:"space_id"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}
type EnvironmentAliasesDataSourceModel struct {
	IDIdentityModel

	SpaceID            types.String                          `tfsdk:"space_id"`
	EnvironmentAliases []EnvironmentAliasDataSourceItemModel `tfsdk:"environment_aliases"`
	Timeouts           timeouts.Value                        `tfsdk:"timeouts"`
}
