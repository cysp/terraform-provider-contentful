//nolint:dupl
package testing

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) CreateRole(_ context.Context, req *cm.RoleData, params cm.CreateRoleParams) (cm.CreateRoleRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.SpaceID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	role := NewRoleFromFields(params.SpaceID, generateResourceID(), *req)
	ts.roles.Set(params.SpaceID, role.Sys.ID, &role)

	return &cm.RoleStatusCode{
		StatusCode: http.StatusCreated,
		Response:   role,
	}, nil
}

//nolint:ireturn
func (ts *Handler) GetRole(_ context.Context, params cm.GetRoleParams) (cm.GetRoleRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.SpaceID == NonexistentID || params.RoleID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	role := ts.roles.Get(params.SpaceID, params.RoleID)
	if role == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("Role not found"), nil), nil
	}

	return role, nil
}

//nolint:ireturn
func (ts *Handler) UpdateRole(_ context.Context, req *cm.RoleData, params cm.UpdateRoleParams) (cm.UpdateRoleRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.SpaceID == NonexistentID || params.RoleID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	role := ts.roles.Get(params.SpaceID, params.RoleID)
	if role == nil {
		newRole := NewRoleFromFields(params.SpaceID, params.RoleID, *req)
		ts.roles.Set(params.SpaceID, newRole.Sys.ID, &newRole)

		return &cm.RoleStatusCode{
			StatusCode: http.StatusCreated,
			Response:   newRole,
		}, nil
	}

	UpdateRoleFromFields(role, *req)

	return &cm.RoleStatusCode{
		StatusCode: http.StatusOK,
		Response:   *role,
	}, nil
}

//nolint:ireturn
func (ts *Handler) DeleteRole(_ context.Context, params cm.DeleteRoleParams) (cm.DeleteRoleRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if params.SpaceID == NonexistentID || params.RoleID == NonexistentID {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	role := ts.roles.Get(params.SpaceID, params.RoleID)
	if role == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(pointerTo("Role not found"), nil), nil
	}

	ts.roles.Delete(params.SpaceID, params.RoleID)

	return &cm.NoContent{}, nil
}
