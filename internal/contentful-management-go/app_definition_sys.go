package contentfulmanagement

func NewAppDefinitionSys(organizationID, appDefinitionID string) AppDefinitionSys {
	return AppDefinitionSys{
		Type:         AppDefinitionSysTypeAppDefinition,
		Organization: NewOrganizationLink(organizationID),
		ID:           appDefinitionID,
	}
}
