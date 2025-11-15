package contentfulmanagement

func NewTeamSpaceMembershipSys(spaceID, teamSpaceMembershipID, teamID string) TeamSpaceMembershipSys {
	return TeamSpaceMembershipSys{
		Type:  TeamSpaceMembershipSysTypeTeamSpaceMembership,
		Space: NewSpaceLink(spaceID),
		Team:  NewTeamLink(teamID),
		ID:    teamSpaceMembershipID,
	}
}
