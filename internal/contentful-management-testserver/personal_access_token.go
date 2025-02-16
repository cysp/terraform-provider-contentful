package contentfulmanagementtestserver

import (
	"slices"

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

type personalAccessTokenValidationError struct {
	field       string
	description string
}

func (e personalAccessTokenValidationError) Error() string {
	return e.field + ": " + e.description
}

func validatePersonalAccessTokenRequestFields(pat *cm.PersonalAccessTokenRequestFields) error {
	if pat.Name == "" {
		return personalAccessTokenValidationError{
			field:       "name",
			description: "Name is required",
		}
	}

	personalAccessTokenScopeValues := []string{"content_management_manage", "content_management_read"}

	for _, scope := range pat.Scopes {
		if !slices.Contains(personalAccessTokenScopeValues, scope) {
			return personalAccessTokenValidationError{
				field:       "scopes",
				description: "Scope is not valid",
			}
		}
	}

	return nil
}
