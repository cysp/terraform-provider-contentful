package contentfulmanagement

func NewContentTypeLink(id string) ContentTypeLink {
	return ContentTypeLink{
		Sys: NewContentTypeLinkSys(id),
	}
}

func NewContentTypeLinkSys(id string) ContentTypeLinkSys {
	return ContentTypeLinkSys{
		Type:     ContentTypeLinkSysTypeLink,
		LinkType: ContentTypeLinkSysLinkTypeContentType,
		ID:       id,
	}
}
