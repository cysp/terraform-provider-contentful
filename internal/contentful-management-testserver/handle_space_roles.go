//nolint:dupl
package contentfulmanagementtestserver

import (
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (ts *ContentfulManagementTestServer) setupSpaceRoleHandlers() {
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

			role := NewRoleFromFields(spaceID, generateResourceID(), roleFields)

			ts.roles.Set(spaceID, role.Sys.ID, &role)

			_ = WriteContentfulManagementResponse(w, http.StatusCreated, &role)

		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(w)
		}
	}))

	ts.serveMux.Handle("/spaces/{spaceID}/roles/{roleID}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		spaceID := r.PathValue("spaceID")
		roleID := r.PathValue("roleID")

		if spaceID == NonexistentID || roleID == NonexistentID {
			_ = WriteContentfulManagementErrorNotFoundResponse(w)

			return
		}

		ts.mu.Lock()
		defer ts.mu.Unlock()

		role := ts.roles.Get(spaceID, roleID)

		switch r.Method {
		case http.MethodGet:
			switch role {
			case nil:
				_ = WriteContentfulManagementErrorNotFoundResponse(w)
			default:
				_ = WriteContentfulManagementResponse(w, http.StatusOK, role)
			}

		case http.MethodPut:
			var roleFields cm.RoleFields
			if err := ReadContentfulManagementRequest(r, &roleFields); err != nil {
				_ = WriteContentfulManagementErrorBadRequestResponseWithError(w, err)

				return
			}

			switch role {
			case nil:
				role := NewRoleFromFields(spaceID, roleID, roleFields)
				ts.roles.Set(spaceID, role.Sys.ID, &role)
				_ = WriteContentfulManagementResponse(w, http.StatusCreated, &role)
			default:
				UpdateRoleFromFields(role, roleFields)
				_ = WriteContentfulManagementResponse(w, http.StatusOK, role)
			}

		case http.MethodDelete:
			switch role {
			case nil:
				_ = WriteContentfulManagementErrorNotFoundResponse(w)
			default:
				ts.roles.Delete(spaceID, roleID)
				w.WriteHeader(http.StatusNoContent)
			}

		default:
			_ = WriteContentfulManagementErrorNotFoundResponse(w)
		}
	}))
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
