package contentfulmanagement

func NewPreviewEnvironmentSys(spaceID, previewEnvironmentID string) PreviewEnvironmentSys {
	return PreviewEnvironmentSys{
		Type:  PreviewEnvironmentSysTypePreviewEnvironment,
		ID:    previewEnvironmentID,
		Space: NewSpaceLink(spaceID),
	}
}
