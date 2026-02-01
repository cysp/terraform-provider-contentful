package contentfulmanagement

func NewEnvironmentAliasSys(spaceID, environmentAliasID string) EnvironmentAliasSys {
	return EnvironmentAliasSys{
		Type:    EnvironmentAliasSysTypeEnvironmentAlias,
		ID:      environmentAliasID,
		Version: 1,
		Space:   NewSpaceLink(spaceID),
	}
}
