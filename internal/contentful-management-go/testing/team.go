package testing

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewTeamFromFields(organizationID, teamID string, req cm.TeamData) cm.Team {
	team := cm.Team{
		Sys: cm.NewTeamSys(organizationID, teamID),
	}

	UpdateTeamFromFields(&team, organizationID, teamID, req)

	return team
}

func UpdateTeamFromFields(team *cm.Team, organizationID, teamID string, req cm.TeamData) {
	team.Sys.ID = teamID
	team.Sys.Organization.Sys.ID = organizationID

	team.Name = req.Name
	team.Description = req.Description
}
