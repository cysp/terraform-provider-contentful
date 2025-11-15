package contentfulmanagement

func NewResourceProviderSys(organizationID, appDefinitionID, resourceProviderID string) ResourceProviderSys {
	return ResourceProviderSys{
		Type:          ResourceProviderSysTypeResourceProvider,
		Organization:  NewOrganizationLink(organizationID),
		AppDefinition: NewAppDefinitionLink(appDefinitionID),
		ID:            resourceProviderID,
	}
}
