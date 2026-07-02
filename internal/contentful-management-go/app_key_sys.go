package contentfulmanagement

func NewAppKeySys(organizationID, appDefinitionID, keyKID string) AppKeySys {
	return AppKeySys{
		ID:            keyKID,
		Type:          AppKeySysTypeAppKey,
		Organization:  NewOrganizationLink(organizationID),
		AppDefinition: NewAppDefinitionLink(appDefinitionID),
	}
}
