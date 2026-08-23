package contentfulmanagement

func NewAppSigningSecretSys(organizationID, appDefinitionID string) AppSigningSecretSys {
	return AppSigningSecretSys{
		Type:          AppSigningSecretSysTypeAppSigningSecret,
		Organization:  NewOrganizationLink(organizationID),
		AppDefinition: NewAppDefinitionLink(appDefinitionID),
	}
}
