package contentfulmanagement

func NewExtensionSys(spaceID, environmentID, extensionID string) ExtensionSys {
	return ExtensionSys{
		Type:        ExtensionSysTypeExtension,
		Space:       NewSpaceLink(spaceID),
		Environment: NewEnvironmentLink(environmentID),
		ID:          extensionID,
	}
}
