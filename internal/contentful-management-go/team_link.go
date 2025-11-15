package contentfulmanagement

func NewTeamLink(id string) TeamLink {
	return TeamLink{
		Sys: NewTeamLinkSys(id),
	}
}

func NewTeamLinkSys(id string) TeamLinkSys {
	return TeamLinkSys{
		Type:     TeamLinkSysTypeLink,
		LinkType: TeamLinkSysLinkTypeTeam,
		ID:       id,
	}
}
