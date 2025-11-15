package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type TeamSpaceMembershipIdentityModel struct {
	SpaceID               types.String `tfsdk:"space_id"`
	TeamSpaceMembershipID types.String `tfsdk:"team_space_membership_id"`
}

type TeamSpaceMembershipModel struct {
	IDIdentityModel
	TeamSpaceMembershipIdentityModel

	TeamID types.String   `tfsdk:"team_id"`
	Admin  types.Bool     `tfsdk:"admin"`
	Roles  []types.String `tfsdk:"roles"`
}
