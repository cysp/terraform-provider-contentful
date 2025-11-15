package contentfulmanagement

func NewEntrySys(spaceID, environmentID, contentTypeID, entryID string) EntrySys {
	return EntrySys{
		Type:        EntrySysTypeEntry,
		Space:       NewSpaceLink(spaceID),
		Environment: NewEnvironmentLink(environmentID),
		ContentType: NewContentTypeLink(contentTypeID),
		ID:          entryID,
	}
}
