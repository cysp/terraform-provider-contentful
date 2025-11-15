package contentfulmanagement

func NewPreviewAPIKeyLink(id string) PreviewAPIKeyLink {
	return PreviewAPIKeyLink{
		Sys: NewPreviewAPIKeyLinkSys(id),
	}
}

func NewPreviewAPIKeyLinkSys(id string) PreviewAPIKeyLinkSys {
	return PreviewAPIKeyLinkSys{
		Type:     PreviewAPIKeyLinkSysTypeLink,
		LinkType: PreviewAPIKeyLinkSysLinkTypePreviewApiKey,
		ID:       id,
	}
}
