package contentfulmanagement

func NewTagLink(id string) TagLink {
	return TagLink{
		Sys: NewTagLinkSys(id),
	}
}

func NewTagLinkSys(id string) TagLinkSys {
	return TagLinkSys{
		Type:     TagLinkSysTypeLink,
		LinkType: TagLinkSysLinkTypeTag,
		ID:       id,
	}
}
