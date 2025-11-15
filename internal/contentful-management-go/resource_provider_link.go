package contentfulmanagement

func NewResourceProviderLink(id string) ResourceProviderLink {
	return ResourceProviderLink{
		Sys: NewResourceProviderLinkSys(id),
	}
}

func NewResourceProviderLinkSys(id string) ResourceProviderLinkSys {
	return ResourceProviderLinkSys{
		Type:     ResourceProviderLinkSysTypeLink,
		LinkType: ResourceProviderLinkSysLinkTypeResourceProvider,
		ID:       id,
	}
}
