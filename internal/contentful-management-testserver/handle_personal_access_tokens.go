package contentfulmanagementtestserver

import (
	"net/http"
	"time"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (ts *ContentfulManagementTestServer) setupPersonalAccessTokenHandlers() {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.personalAccessTokenIDsToCreate = make([]string, 0)
	ts.personalAccessTokens = make(map[string]*cm.PersonalAccessToken)

	ts.serveMux.Handle("/users/me/access_tokens", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ts.mu.Lock()
		defer ts.mu.Unlock()

		switch r.Method {
		case http.MethodPost:
			var personalAccessToken cm.PersonalAccessToken
			_ = ReadContentfulManagementRequest(r, &personalAccessToken)

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
