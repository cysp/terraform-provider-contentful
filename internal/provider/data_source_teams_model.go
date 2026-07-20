package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type TeamsDataSourceModel struct {
	ID             types.String               `tfsdk:"id"`
	OrganizationID types.String               `tfsdk:"organization_id"`
	Teams          []TeamsDataSourceTeamModel `tfsdk:"teams"`
	Timeouts       timeouts.Value             `tfsdk:"timeouts"`
}

type TeamsDataSourceTeamModel struct {
	TeamID      types.String `tfsdk:"team_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}
