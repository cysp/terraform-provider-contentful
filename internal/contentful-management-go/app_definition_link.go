package contentfulmanagement

func NewAppDefinitionLink(id string) AppDefinitionLink {
	return AppDefinitionLink{
		Sys: NewAppDefinitionLinkSys(id),
	}
}

func NewAppDefinitionLinkSys(id string) AppDefinitionLinkSys {
	return AppDefinitionLinkSys{
		Type:     AppDefinitionLinkSysTypeLink,
		LinkType: AppDefinitionLinkSysLinkTypeAppDefinition,
		ID:       id,
	}
}
