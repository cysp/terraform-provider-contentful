package contentfulmanagementtestserver

import (
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (ts *ContentfulManagementTestServer) setupSpaceRoleHandlers() {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.serveMux.Handle("/spaces/{spaceID}/roles", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		spaceID := r.PathValue("spaceID")

		if spaceID == NonexistentID {
			_ = WriteContentfulManagementErrorNotFoundResponse(w)

			return
		}

		ts.mu.Lock()
		defer ts.mu.Unlock()

		switch r.Method {
		case http.MethodPost:
			var roleFields cm.RoleFields
			if err := ReadContentfulManagementRequest(r, &roleFields); err != nil {
				_ = WriteContentfulManagementErrorBadRequestResponseWithError(w, err)

				return
			}

			role := NewRoleFromFields(spaceID, ts.generateResourceID(), roleFields)

			ts.roles.Set(spaceID, role.Sys.ID, &role)

			_ = WriteContentfulManagementResponse(w, http.StatusCreated, &role)

		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(w)
		}
	}))

	ts.serveMux.Handle("/spaces/{spaceID}/roles/{roleID}", http.HandlerFunc(func(responseWriter http.ResponseWriter, r *http.Request) {
		spaceID := r.PathValue("spaceID")
		roleID := r.PathValue("roleID")

		if spaceID == NonexistentID || roleID == NonexistentID {
			_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)

			return
		}

		ts.mu.Lock()
		defer ts.mu.Unlock()

		role, exists := ts.roles.Get(spaceID, roleID)

		switch r.Method {
		case http.MethodGet:
			if !exists {
				_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)

				return
			}

			_ = WriteContentfulManagementResponse(responseWriter, http.StatusOK, role)

		case http.MethodPut:
			var roleFields cm.RoleFields
			_ = ReadContentfulManagementRequest(r, &roleFields)

			if exists {
				UpdateRoleFromFields(role, roleFields)

				_ = WriteContentfulManagementResponse(responseWriter, http.StatusOK, role)
			} else {
				role := NewRoleFromFields(spaceID, roleID, roleFields)

				ts.roles.Set(spaceID, role.Sys.ID, &role)

				_ = WriteContentfulManagementResponse(responseWriter, http.StatusCreated, &role)
			}

		case http.MethodDelete:
			if !exists {
				_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)

				return
			}

			ts.roles.Delete(spaceID, roleID)

			responseWriter.WriteHeader(http.StatusNoContent)

		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)
		}
	}))
}

func (ts *ContentfulManagementTestServer) GetRole(spaceID, roleID string) (*cm.Role, bool) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	return ts.roles.Get(spaceID, roleID)
}

func (ts *ContentfulManagementTestServer) SetRole(spaceID, roleID string, roleFields cm.RoleFields) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	role := NewRoleFromFields(spaceID, roleID, roleFields)

	ts.roles.Set(spaceID, role.Sys.ID, &role)
}

func (ts *ContentfulManagementTestServer) DeleteRole(spaceID, roleID string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.roles.Delete(spaceID, roleID)
}
