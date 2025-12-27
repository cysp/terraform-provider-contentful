package contentfulmanagement

func NewTagSys(spaceID, environmentID, tagID string) TagSys {
	return TagSys{
		Type:        TagSysTypeTag,
		Space:       NewSpaceLink(spaceID),
		Environment: NewEnvironmentLink(environmentID),
		ID:          tagID,
		Visibility:  "private",
	}
}
