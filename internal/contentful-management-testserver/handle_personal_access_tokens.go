package contentfulmanagementtestserver

import (
	"net/http"
	"slices"
	"time"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (ts *ContentfulManagementTestServer) setupPersonalAccessTokenHandlers() {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	// ts.personalAccessTokenIDsToCreate = make([]string, 0)
	ts.personalAccessTokens = make(map[string]*cm.PersonalAccessToken)

	ts.serveMux.Handle("/users/me/access_tokens", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ts.mu.Lock()
		defer ts.mu.Unlock()

		switch r.Method {
		case http.MethodPost:
			var personalAccessTokenRequestFields cm.PersonalAccessTokenRequestFields
			if err := ReadContentfulManagementRequest(r, &personalAccessTokenRequestFields); err != nil {
				_ = WriteContentfulManagementErrorBadRequestResponseWithError(w, err)

				return
			}

			personalAccessTokenID := ts.generateResourceID()

			personalAccessToken := cm.PersonalAccessToken{
				Sys: cm.PersonalAccessTokenSys{
					Type: cm.PersonalAccessTokenSysTypePersonalAccessToken,
					ID:   personalAccessTokenID,
				},
				Name:   personalAccessTokenRequestFields.Name,
				Scopes: personalAccessTokenRequestFields.Scopes,
			}

			personalAccessToken.Token.SetTo(personalAccessTokenID)

			if err := ts.validatePersonalAccessToken(&personalAccessToken); err != nil {
				_ = WriteContentfulManagementErrorBadRequestResponseWithError(w, err)
			}

			ts.personalAccessTokens[personalAccessToken.Sys.ID] = &personalAccessToken

			_ = WriteContentfulManagementResponse(w, http.StatusCreated, &personalAccessToken)

		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(w)
		}
	}))

	ts.serveMux.Handle("/users/me/access_tokens/{id}", http.HandlerFunc(func(responseWriter http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		//nolint:gocritic
		switch r.Method {
		case http.MethodGet:
			if personalAccessToken, exists := ts.personalAccessTokens[id]; exists {
				_ = WriteContentfulManagementResponse(responseWriter, http.StatusOK, personalAccessToken)

				return
			}
		}

		_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)
	}))

	ts.serveMux.Handle("/users/me/access_tokens/{id}/revoked", http.HandlerFunc(func(responseWriter http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		//nolint:gocritic
		switch r.Method {
		case http.MethodPut:
			if personalAccessToken, exists := ts.personalAccessTokens[id]; exists {
				personalAccessToken.RevokedAt = cm.NewOptNilDateTime(time.Now())

				_ = WriteContentfulManagementResponse(responseWriter, http.StatusOK, personalAccessToken)

				return
			}
		}

		_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)
	}))
}

type personalAccessTokenValidationError struct {
	field       string
	description string
}

func (e personalAccessTokenValidationError) Error() string {
	return e.field + ": " + e.description
}

func (ts *ContentfulManagementTestServer) validatePersonalAccessToken(pat *cm.PersonalAccessToken) error {
	if pat.Name == "" {
		return personalAccessTokenValidationError{
			field:       "name",
			description: "Name is required",
		}
	}

	personalAccessTokenScopeValues := []string{"content_management_manage", "content_management_read"}

	for _, scope := range pat.Scopes {
		if scope == "" {
			return personalAccessTokenValidationError{
				field:       "scopes",
				description: "Scope must not be empty",
			}
		} else if !slices.Contains(personalAccessTokenScopeValues, scope) {
			return personalAccessTokenValidationError{
				field:       "scopes",
				description: "Scope is not valid",
			}
		}
	}

	return nil
}
