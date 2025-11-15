package contentfulmanagement

func NewPreviewAPIKeySys(spaceID, previewAPIKeyID string) PreviewApiKeySys {
	return PreviewApiKeySys{
		Type:  PreviewApiKeySysTypePreviewApiKey,
		ID:    previewAPIKeyID,
		Space: NewSpaceLink(spaceID),
	}
}
