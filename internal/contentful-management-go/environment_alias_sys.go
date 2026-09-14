package contentfulmanagement

// EnvironmentAliasSysTypeEnvironmentAlias preserves the established name when ogen numbers the two equivalent type spellings.
const EnvironmentAliasSysTypeEnvironmentAlias EnvironmentAliasSysType = "EnvironmentAlias"

func NewEnvironmentAliasSys(spaceID, environmentAliasID string) EnvironmentAliasSys {
	return EnvironmentAliasSys{
		Type:    EnvironmentAliasSysTypeEnvironmentAlias,
		ID:      environmentAliasID,
		Version: 1,
		Space:   NewSpaceLink(spaceID),
	}
}
