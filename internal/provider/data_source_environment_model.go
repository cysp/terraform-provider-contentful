package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type EnvironmentDataSourceItemModel struct {
	EnvironmentID        types.String `tfsdk:"environment_id"`
	Name                 types.String `tfsdk:"name"`
	Status               types.String `tfsdk:"status"`
	AliasedEnvironmentID types.String `tfsdk:"aliased_environment_id"`
}
type EnvironmentDataSourceModel struct {
	IDIdentityModel
	EnvironmentDataSourceItemModel

	SpaceID  types.String   `tfsdk:"space_id"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}
type EnvironmentsDataSourceModel struct {
	IDIdentityModel

	SpaceID      types.String                     `tfsdk:"space_id"`
	Environments []EnvironmentDataSourceItemModel `tfsdk:"environments"`
	Timeouts     timeouts.Value                   `tfsdk:"timeouts"`
}
