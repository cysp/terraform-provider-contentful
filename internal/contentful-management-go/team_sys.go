package contentfulmanagement

func NewTeamSys(organizationID, teamID string) TeamSys {
	return TeamSys{
		Type:         TeamSysTypeTeam,
		Organization: NewOrganizationLink(organizationID),
		ID:           teamID,
	}
}
