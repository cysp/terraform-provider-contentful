package contentfulmanagement

func NewLocaleSys(spaceID, environmentID, localeID string) LocaleSys {
	return LocaleSys{
		Space:       NewSpaceLink(spaceID),
		Environment: NewEnvironmentLink(environmentID),
		Type:        LocaleSysTypeLocale,
		ID:          localeID,
		Version:     1,
	}
}
