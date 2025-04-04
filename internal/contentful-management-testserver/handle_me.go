package contentfulmanagementtestserver

import (
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (ts *ContentfulManagementTestServer) setupUserHandler() {
	ts.serveMux.Handle("/users/me", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ts.mu.Lock()
		defer ts.mu.Unlock()

		switch r.Method {
		case http.MethodGet:
			switch ts.me {
			case nil:
				_ = WriteContentfulManagementErrorNotFoundResponse(w)
			default:
				_ = WriteContentfulManagementResponse(w, http.StatusOK, ts.me)
			}
		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(w)
		}
	}))
}

func (ts *ContentfulManagementTestServer) SetMe(me *cm.User) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.me = me
}
