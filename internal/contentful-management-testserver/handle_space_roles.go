//nolint:dupl
package contentfulmanagementtestserver

import (
	"fmt"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (ts *ContentfulManagementTestServer) HandleRoleCreation(spaceID string, role *cm.Role) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.ServeMux.Handle(fmt.Sprintf("/spaces/%s/roles", spaceID), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			_ = WriteContentfulManagementResponse(w, http.StatusCreated, role)
		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(w)
		}
	}))
}

func (ts *ContentfulManagementTestServer) HandleRole(spaceID string, role *cm.Role) {
	roleID := role.Sys.ID

	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.ServeMux.Handle(fmt.Sprintf("/spaces/%s/roles/%s", spaceID, roleID), http.HandlerFunc(func(responseWriter http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_ = WriteContentfulManagementResponse(responseWriter, http.StatusOK, role)
		case http.MethodPut:
			_ = WriteContentfulManagementResponse(responseWriter, http.StatusOK, role)
		case http.MethodDelete:
			responseWriter.WriteHeader(http.StatusNoContent)
		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)
		}
	}))
}
