package cmtesting

import (
	"cmp"
	"context"
	"net/http"
	"slices"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) GetRoles(_ context.Context, params cm.GetRolesParams) (cm.GetRolesRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.environments.Get(params.SpaceID, "master") == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(new("Space not found"), nil)), nil
	}

	values := ts.roles.List(params.SpaceID)
	// The fixture store is a map; choose a repeatable response order for tests.
	slices.SortFunc(values, func(a, b *cm.Role) int { return cmp.Compare(a.Sys.ID, b.Sys.ID) })

	skip, limit := params.Skip.Or(0), params.Limit.Or(100) //nolint:mnd
	if skip < 0 || limit < 1 {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeBadRequest(new("Invalid pagination parameters"), nil)), nil
	}

	start := min(skip, int64(len(values)))
	end := start + min(limit, int64(len(values))-start)

	items := make([]cm.Role, 0, end-start)
	for _, value := range values[start:end] {
		items = append(items, *value)
	}

	return snapshotResponse(&cm.RoleCollection{Sys: cm.RoleCollectionSys{Type: cm.RoleCollectionSysTypeArray}, Skip: cm.NewOptInt(int(skip)), Limit: cm.NewOptInt(int(limit)), Total: cm.NewOptInt(len(values)), Items: items}), nil
}

//nolint:ireturn
func (ts *Handler) CreateRole(_ context.Context, req *cm.RoleData, params cm.CreateRoleParams) (cm.CreateRoleRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.environments.Get(params.SpaceID, "master") == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(new("Space not found"), nil)), nil
	}

	role := NewRoleFromFields(params.SpaceID, generateResourceID(), *req)
	ts.roles.Set(params.SpaceID, role.Sys.ID, &role)

	return snapshotResponse(&cm.RoleStatusCode{
		StatusCode: http.StatusCreated,
		Response:   role,
	}), nil
}

//nolint:ireturn
func (ts *Handler) GetRole(_ context.Context, params cm.GetRoleParams) (cm.GetRoleRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	role := ts.roles.Get(params.SpaceID, params.RoleID)
	if role == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(new("Role not found"), nil)), nil
	}

	return snapshotResponse(role), nil
}

//nolint:ireturn
func (ts *Handler) UpdateRole(_ context.Context, req *cm.RoleData, params cm.UpdateRoleParams) (cm.UpdateRoleRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.environments.Get(params.SpaceID, "master") == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(new("Space not found"), nil)), nil
	}

	role := ts.roles.Get(params.SpaceID, params.RoleID)
	if role == nil {
		newRole := NewRoleFromFields(params.SpaceID, params.RoleID, *req)
		ts.roles.Set(params.SpaceID, newRole.Sys.ID, &newRole)

		return snapshotResponse(&cm.RoleStatusCode{
			StatusCode: http.StatusCreated,
			Response:   newRole,
		}), nil
	}

	if params.XContentfulVersion != role.Sys.Version {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeVersionMismatch(nil, nil)), nil
	}

	UpdateRoleFromFields(role, *req)

	return snapshotResponse(&cm.RoleStatusCode{
		StatusCode: http.StatusOK,
		Response:   *role,
	}), nil
}

//nolint:ireturn
func (ts *Handler) DeleteRole(_ context.Context, params cm.DeleteRoleParams) (cm.DeleteRoleRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	role := ts.roles.Get(params.SpaceID, params.RoleID)
	if role == nil {
		return snapshotResponse(NewContentfulManagementErrorStatusCodeNotFound(new("Role not found"), nil)), nil
	}

	ts.roles.Delete(params.SpaceID, params.RoleID)

	return snapshotResponse(&cm.NoContent{}), nil
}
