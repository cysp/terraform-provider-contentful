package cmtesting

import (
	"cmp"
	"context"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"slices"
	"strconv"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) GetRoles(_ context.Context, params cm.GetRolesParams) (cm.GetRolesRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.environments.Get(params.SpaceID, "master") == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("Space not found"), nil), nil
	}

	values := ts.roles.List(params.SpaceID)
	// The fixture store is a map; choose a repeatable response order for tests.
	slices.SortFunc(values, func(a, b *cm.Role) int { return cmp.Compare(a.Sys.ID, b.Sys.ID) })

	if params.PageNext.IsSet() && params.PagePrev.IsSet() {
		return NewContentfulManagementErrorStatusCodeBadRequest(new("Conflicting pagination parameters"), nil), nil
	}

	limit64 := params.Limit.Or(100) //nolint:mnd
	if limit64 < 1 || limit64 > math.MaxInt {
		return NewContentfulManagementErrorStatusCodeBadRequest(new("Invalid pagination parameters"), nil), nil
	}

	limit := int(limit64)
	offset := 0

	cursor := params.PageNext
	if params.PagePrev.IsSet() {
		cursor = params.PagePrev
	}

	if cursor.IsSet() {
		var err error

		offset, err = strconv.Atoi(cursor.Value)
		if err != nil || offset < 0 {
			return NewContentfulManagementErrorStatusCodeBadRequest(new("Invalid cursor"), nil), nil
		}
	}

	start := min(offset, len(values))
	end := start + min(limit, len(values)-start)

	items := make([]cm.Role, 0, end-start)
	for _, value := range values[start:end] {
		items = append(items, *value)
	}

	pages := cm.RoleCollectionPages{}

	link := func(parameter string, offset int) string {
		return fmt.Sprintf("/spaces/%s/roles?%s=%d&limit=%d", url.PathEscape(params.SpaceID), parameter, offset, limit)
	}
	if end < len(values) {
		pages.Next = cm.NewOptString(link("pageNext", end))
	}

	if start > 0 {
		pages.Prev = cm.NewOptString(link("pagePrev", max(0, start-limit)))
	}

	return &cm.RoleCollection{
		Sys:   cm.RoleCollectionSys{Type: cm.RoleCollectionSysTypeArray},
		Limit: cm.NewOptInt(limit),
		Items: items,
		Pages: cm.NewOptRoleCollectionPages(pages),
	}, nil
}

//nolint:ireturn
func (ts *Handler) CreateRole(_ context.Context, req *cm.RoleData, params cm.CreateRoleParams) (cm.CreateRoleRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.environments.Get(params.SpaceID, "master") == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("Space not found"), nil), nil
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

	role := ts.roles.Get(params.SpaceID, params.RoleID)
	if role == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("Role not found"), nil), nil
	}

	return role, nil
}

//nolint:ireturn
func (ts *Handler) UpdateRole(_ context.Context, req *cm.RoleData, params cm.UpdateRoleParams) (cm.UpdateRoleRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.environments.Get(params.SpaceID, "master") == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("Space not found"), nil), nil
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

	if params.XContentfulVersion != role.Sys.Version {
		return NewContentfulManagementErrorStatusCodeVersionMismatch(nil, nil), nil
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

	role := ts.roles.Get(params.SpaceID, params.RoleID)
	if role == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("Role not found"), nil), nil
	}

	ts.roles.Delete(params.SpaceID, params.RoleID)

	return &cm.NoContent{}, nil
}
