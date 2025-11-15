package contentfulmanagement

func NewAPIKeySys(spaceID string, apiKeyID string) ApiKeySys {
	return ApiKeySys{
		Type:  ApiKeySysTypeApiKey,
		Space: NewSpaceLink(spaceID),
		ID:    apiKeyID,
	}
}
