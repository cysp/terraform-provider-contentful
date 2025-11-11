package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewTeamSpaceMembershipResourceModelFromResponse(_ context.Context, response cm.TeamSpaceMembership) (TeamSpaceMembershipModel, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	spaceID := response.Sys.Space.Sys.ID
	teamSpaceMembershipID := response.Sys.ID
	teamID := response.Sys.Team.Sys.ID

	model := TeamSpaceMembershipModel{
		IDIdentityModel: NewIDIdentityModelFromMultipartID([]string{spaceID, teamSpaceMembershipID}),
		TeamSpaceMembershipIdentityModel: TeamSpaceMembershipIdentityModel{
			SpaceID:               types.StringValue(spaceID),
			TeamSpaceMembershipID: types.StringValue(teamSpaceMembershipID),
		},
		TeamID: types.StringValue(teamID),
		Admin:  types.BoolValue(response.Admin),
	}

	if response.Roles != nil {
		roles := make([]types.String, 0, len(response.Roles))
		for _, roleLink := range response.Roles {
			roles = append(roles, types.StringValue(roleLink.Sys.ID))
		}

		model.Roles = roles
	}

	return model, diags
}
