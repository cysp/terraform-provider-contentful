package contentfulmanagementtestserver

import (
	"fmt"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (ts *ContentfulManagementTestServer) HandlePersonalAccessTokenCreation(personalAccessToken *cm.PersonalAccessToken) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.ServeMux.Handle("/users/me/access_tokens", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			_ = WriteContentfulManagementResponse(w, http.StatusCreated, personalAccessToken)
		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(w)
		}
	}))
}

func (ts *ContentfulManagementTestServer) HandlePersonalAccessToken(personalAccessToken *cm.PersonalAccessToken) {
	personalAccessTokenID := personalAccessToken.Sys.ID

	ts.mu.Lock()
	defer ts.mu.Unlock()

	//nolint:perfsprint
	ts.ServeMux.Handle(fmt.Sprintf("/users/me/access_tokens/%s", personalAccessTokenID), http.HandlerFunc(func(responseWriter http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_ = WriteContentfulManagementResponse(responseWriter, http.StatusOK, personalAccessToken)
		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)
		}
	}))

	ts.ServeMux.Handle(fmt.Sprintf("/users/me/access_tokens/%s/revoked", personalAccessTokenID), http.HandlerFunc(func(responseWriter http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			_ = WriteContentfulManagementResponse(responseWriter, http.StatusOK, personalAccessToken)
		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)
		}
	}))
}
