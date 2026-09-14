package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SpaceDataSourceItemModel struct {
	SpaceID        types.String `tfsdk:"space_id"`
	Name           types.String `tfsdk:"name"`
	OrganizationID types.String `tfsdk:"organization_id"`
}
type SpaceDataSourceModel struct {
	IDIdentityModel
	SpaceDataSourceItemModel

	Timeouts timeouts.Value `tfsdk:"timeouts"`
}
type SpacesDataSourceModel struct {
	IDIdentityModel

	OrganizationID types.String               `tfsdk:"organization_id"`
	Spaces         []SpaceDataSourceItemModel `tfsdk:"spaces"`
	Timeouts       timeouts.Value             `tfsdk:"timeouts"`
}
