package contentfulmanagement

func NewPersonalAccessTokenSys(personalAccessTokenID string) PersonalAccessTokenSys {
	return PersonalAccessTokenSys{
		Type: PersonalAccessTokenSysTypePersonalAccessToken,
		ID:   personalAccessTokenID,
	}
}
