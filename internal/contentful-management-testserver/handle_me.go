package contentfulmanagementtestserver

import (
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (ts *ContentfulManagementTestServer) HandleUser(user *cm.User) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.ServeMux.Handle("/users/me", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_ = WriteContentfulManagementResponse(w, http.StatusOK, user)
		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(w)
		}
	}))
}
