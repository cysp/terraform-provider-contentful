package cmtesting

import (
	"cmp"
	"context"
	"encoding/base64"
	"fmt"
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

	skip, limit := params.Skip.Or(0), params.Limit.Or(100) //nolint:mnd
	if params.PageNext.IsSet() && params.PagePrev.IsSet() || params.Skip.IsSet() && (params.PageNext.IsSet() || params.PagePrev.IsSet()) {
		return NewContentfulManagementErrorStatusCodeBadRequest(new("Conflicting pagination parameters"), nil), nil
	}

	cursor := params.PageNext
	if params.PagePrev.IsSet() {
		cursor = params.PagePrev
	}

	if cursor.IsSet() {
		decoded, err := base64.RawURLEncoding.DecodeString(cursor.Value)
		if err != nil {
			return NewContentfulManagementErrorStatusCodeBadRequest(new("Invalid cursor"), nil), nil
		}

		skip, err = strconv.ParseInt(string(decoded), 10, 64)
		if err != nil {
			return NewContentfulManagementErrorStatusCodeBadRequest(new("Invalid cursor"), nil), nil
		}
	}

	if skip < 0 || limit < 1 {
		return NewContentfulManagementErrorStatusCodeBadRequest(new("Invalid pagination parameters"), nil), nil
	}

	start := min(skip, int64(len(values)))
	end := start + min(limit, int64(len(values))-start)

	items := make([]cm.Role, 0, end-start)
	for _, value := range values[start:end] {
		items = append(items, *value)
	}

	collection := &cm.RoleCollection{Sys: cm.RoleCollectionSys{Type: cm.RoleCollectionSysTypeArray}, Limit: cm.NewOptInt(int(limit)), Items: items}
	// Explicit skip selects the legacy fixture; otherwise model cursor responses.
	if params.Skip.IsSet() {
		collection.Skip = cm.NewOptInt(int(skip))
		collection.Total = cm.NewOptInt(len(values))
	} else {
		pages := cm.RoleCollectionPages{}

		link := func(parameter string, offset int64) string {
			token := base64.RawURLEncoding.EncodeToString([]byte(strconv.FormatInt(offset, 10)))

			return fmt.Sprintf("/spaces/%s/roles?%s=%s&limit=%d", url.PathEscape(params.SpaceID), parameter, token, limit)
		}
		if end < int64(len(values)) {
			pages.Next = cm.NewOptString(link("pageNext", end))
		}

		if start > 0 {
			pages.Prev = cm.NewOptString(link("pagePrev", max(0, start-limit)))
		}

		collection.Pages = cm.NewOptRoleCollectionPages(pages)
	}

	return collection, nil
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
