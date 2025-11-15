package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type TeamIdentityModel struct {
	OrganizationID types.String `tfsdk:"organization_id"`
	TeamID         types.String `tfsdk:"team_id"`
}

type TeamModel struct {
	IDIdentityModel
	TeamIdentityModel

	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}
