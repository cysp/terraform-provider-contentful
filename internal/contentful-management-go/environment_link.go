package contentfulmanagement

func NewEnvironmentLink(id string) EnvironmentLink {
	return EnvironmentLink{
		Sys: NewEnvironmentLinkSys(id),
	}
}

func NewEnvironmentLinkSys(id string) EnvironmentLinkSys {
	return EnvironmentLinkSys{
		Type:     EnvironmentLinkSysTypeLink,
		LinkType: EnvironmentLinkSysLinkTypeEnvironment,
		ID:       id,
	}
}
