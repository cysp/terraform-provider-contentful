package contentfulmanagement

func NewContentTypeSys(spaceID, environmentID, contentTypeID string) ContentTypeSys {
	return ContentTypeSys{
		Type:        ContentTypeSysTypeContentType,
		Space:       NewSpaceLink(spaceID),
		Environment: NewEnvironmentLink(environmentID),
		ID:          contentTypeID,
	}
}
