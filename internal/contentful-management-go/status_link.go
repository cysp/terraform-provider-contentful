package contentfulmanagement

func NewStatusLink(id string) StatusLink {
	return StatusLink{
		Sys: NewStatusLinkSys(id),
	}
}

func NewStatusLinkSys(id string) StatusLinkSys {
	return StatusLinkSys{
		Type:     StatusLinkSysTypeLink,
		LinkType: StatusLinkSysLinkTypeStatus,
		ID:       id,
	}
}
