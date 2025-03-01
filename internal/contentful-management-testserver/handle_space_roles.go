package contentfulmanagementtestserver

import (
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (ts *ContentfulManagementTestServer) setupSpaceRoleHandlers() {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.serveMux.Handle("/spaces/{spaceID}/roles", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ts.mu.Lock()
		defer ts.mu.Unlock()

		spaceID := r.PathValue("spaceID")

		switch r.Method {
		case http.MethodPost:
			var role cm.Role
			_ = ReadContentfulManagementRequest(r, &role)

			id := ts.generateResourceID()

			role.Sys = cm.RoleSys{
				ID:   id,
				Type: cm.RoleSysTypeRole,
			}

			ts.roles.Set(spaceID, id, &role)

			_ = WriteContentfulManagementResponse(w, http.StatusCreated, &role)

		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(w)
		}
	}))

	ts.serveMux.Handle("/spaces/{spaceID}/roles/{roleID}", http.HandlerFunc(func(responseWriter http.ResponseWriter, r *http.Request) {
		ts.mu.Lock()
		defer ts.mu.Unlock()

		spaceID := r.PathValue("spaceID")
		roleID := r.PathValue("roleID")

		role, exists := ts.roles.Get(spaceID, roleID)
		if !exists {
			_ = WriteContentfulManagementErrorNotFoundResponse(responseWriter)

			return
		}

		switch r.Method {
		case http.MethodGet:
			_ = WriteContentfulManagementResponse(responseWriter, http.StatusOK, role)

		case http.MethodPut:
			var updatedRole cm.Role
			_ = ReadContentfulManagementRequest(r, &updatedRole)

			updatedRole.Sys = cm.RoleSys{
				ID:   roleID,
				Type: cm.RoleSysTypeRole,
			}

			ts.roles.Set(spaceID, roleID, &updatedRole)

			_ = WriteContentfulManagementResponse(responseWriter, http.StatusOK, &updatedRole)

		case http.MethodDelete:
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

func (ts *ContentfulManagementTestServer) SetRole(spaceID string, role *cm.Role) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.roles.Set(spaceID, role.Sys.ID, role)
}

func (ts *ContentfulManagementTestServer) DeleteRole(spaceID, roleID string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.roles.Delete(spaceID, roleID)
}
