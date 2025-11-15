package contentfulmanagement

func NewResourceTypeSys(organizationID, appDefinitionID, resourceProviderID, resourceTypeID string) ResourceTypeSys {
	return ResourceTypeSys{
		Type:             ResourceTypeSysTypeResourceType,
		Organization:     NewOrganizationLink(organizationID),
		AppDefinition:    NewAppDefinitionLink(appDefinitionID),
		ResourceProvider: NewResourceProviderLink(resourceProviderID),
		ID:               resourceTypeID,
	}
}
