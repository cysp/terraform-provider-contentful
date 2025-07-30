package contentfulmanagementtestserver

import (
	"net/http"
	"time"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (ts *ContentfulManagementTestServer) setupPersonalAccessTokenHandlers() {
	ts.serveMux.Handle("/users/me/access_tokens", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ts.mu.Lock()
		defer ts.mu.Unlock()

		switch r.Method {
		case http.MethodPost:
			var personalAccessTokenRequestFields cm.PersonalAccessTokenRequestFields

			err := ReadContentfulManagementRequestWithValidation(r, &personalAccessTokenRequestFields, validatePersonalAccessTokenRequestFields)
			if err != nil {
				_ = WriteContentfulManagementErrorBadRequestResponseWithError(w, err)

				return
			}

			personalAccessTokenID := generateResourceID()
			personalAccessToken := NewPersonalAccessTokenFromRequestFields(personalAccessTokenID, personalAccessTokenRequestFields)

			ts.personalAccessTokens[personalAccessToken.Sys.ID] = &personalAccessToken

			personalAccessTokenWithToken := personalAccessToken
			personalAccessTokenWithToken.Token.SetTo(generateResourceID())

			_ = WriteContentfulManagementResponse(w, http.StatusCreated, &personalAccessTokenWithToken)

		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(w)
		}
	}))

	ts.serveMux.Handle("/users/me/access_tokens/{id}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id") //nolint:varnamelen

		if id == NonexistentID {
			_ = WriteContentfulManagementErrorNotFoundResponse(w)

			return
		}

		ts.mu.Lock()
		defer ts.mu.Unlock()

		personalAccessToken := ts.personalAccessTokens[id]

		switch r.Method {
		case http.MethodGet:
			switch personalAccessToken {
			case nil:
				_ = WriteContentfulManagementErrorNotFoundResponse(w)

			default:
				_ = WriteContentfulManagementResponse(w, http.StatusOK, personalAccessToken)
			}

		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(w)
		}
	}))

	ts.serveMux.Handle("/users/me/access_tokens/{id}/revoked", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id") //nolint:varnamelen

		if id == NonexistentID {
			_ = WriteContentfulManagementErrorNotFoundResponse(w)

			return
		}

		ts.mu.Lock()
		defer ts.mu.Unlock()

		personalAccessToken := ts.personalAccessTokens[id]

		switch r.Method {
		case http.MethodPut:
			switch personalAccessToken {
			case nil:
				_ = WriteContentfulManagementErrorNotFoundResponse(w)
			default:
				personalAccessToken.RevokedAt.SetTo(time.Now())
				_ = WriteContentfulManagementResponse(w, http.StatusOK, personalAccessToken)
			}

		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(w)
		}
	}))
}
