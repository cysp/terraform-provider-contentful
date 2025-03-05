package contentfulmanagementtestserver

import (
	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func NewPersonalAccessTokenFromRequestFields(personalAccessTokenID string, personalAccessTokenFields cm.PersonalAccessTokenRequestFields) cm.PersonalAccessToken {
	personalAccessToken := cm.PersonalAccessToken{
		Sys: NewPersonalAccessTokenSys(personalAccessTokenID),
	}

	UpdatePersonalAccessTokenFromRequestFields(&personalAccessToken, personalAccessTokenFields)

	return personalAccessToken
}

func NewPersonalAccessTokenSys(personalAccessTokenID string) cm.PersonalAccessTokenSys {
	return cm.PersonalAccessTokenSys{
		Type: cm.PersonalAccessTokenSysTypePersonalAccessToken,
		ID:   personalAccessTokenID,
	}
}

func UpdatePersonalAccessTokenFromRequestFields(personalAccessToken *cm.PersonalAccessToken, personalAccessTokenFields cm.PersonalAccessTokenRequestFields) {
	personalAccessToken.Name = personalAccessTokenFields.Name
	personalAccessToken.Scopes = personalAccessTokenFields.Scopes
}
