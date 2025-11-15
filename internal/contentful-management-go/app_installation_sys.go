package contentfulmanagement

func NewAppInstallationSys(spaceID, environmentID, appDefinitionID string) AppInstallationSys {
	return AppInstallationSys{
		Type:          AppInstallationSysTypeAppInstallation,
		Space:         NewSpaceLink(spaceID),
		Environment:   NewEnvironmentLink(environmentID),
		AppDefinition: NewAppDefinitionLink(appDefinitionID),
	}
}
