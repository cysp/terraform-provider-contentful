package contentfulmanagement

func NewEnvironmentSys(spaceID, environmentID string) EnvironmentSys {
	return EnvironmentSys{
		Type:    EnvironmentSysTypeEnvironment,
		ID:      environmentID,
		Version: 1,
		Space:   NewSpaceLink(spaceID),
		Status:  NewStatusLink("ready"),
	}
}
