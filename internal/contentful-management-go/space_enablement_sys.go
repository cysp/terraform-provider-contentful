package contentfulmanagement

func NewSpaceEnablementSys(spaceID string) SpaceEnablementSys {
	return SpaceEnablementSys{
		Type:  SpaceEnablementSysTypeSpaceEnablement,
		Space: NewSpaceLink(spaceID),
	}
}
