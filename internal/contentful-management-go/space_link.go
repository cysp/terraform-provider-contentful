package contentfulmanagement

func NewSpaceLink(id string) SpaceLink {
	return SpaceLink{
		Sys: NewSpaceLinkSys(id),
	}
}

func NewSpaceLinkSys(id string) SpaceLinkSys {
	return SpaceLinkSys{
		Type:     SpaceLinkSysTypeLink,
		LinkType: SpaceLinkSysLinkTypeSpace,
		ID:       id,
	}
}
