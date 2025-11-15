package testing

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewTeamSpaceMembershipFromFields(spaceID, teamSpaceMembershipID, teamID string, req cm.TeamSpaceMembershipData) cm.TeamSpaceMembership {
	teamSpaceMembership := cm.TeamSpaceMembership{
		Sys: cm.NewTeamSpaceMembershipSys(spaceID, teamSpaceMembershipID, teamID),
	}

	UpdateTeamSpaceMembershipFromFields(&teamSpaceMembership, spaceID, teamSpaceMembershipID, teamID, req)

	return teamSpaceMembership
}

func UpdateTeamSpaceMembershipFromFields(teamSpaceMembership *cm.TeamSpaceMembership, spaceID, teamSpaceMembershipID, teamID string, req cm.TeamSpaceMembershipData) {
	teamSpaceMembership.Sys.ID = teamSpaceMembershipID
	teamSpaceMembership.Sys.Space.Sys.ID = spaceID
	teamSpaceMembership.Sys.Team.Sys.ID = teamID

	teamSpaceMembership.Admin = req.Admin
	teamSpaceMembership.Roles = req.Roles
}
