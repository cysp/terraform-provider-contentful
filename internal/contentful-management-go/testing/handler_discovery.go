package cmtesting

import (
	"cmp"
	"context"
	"slices"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

// SetSpace seeds a synthetic fixture; it performs no CMA mutation.
func (s *Server) SetSpace(space cm.Space) {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()

	s.h.spaces[space.Sys.ID] = &space
}

func (s *Server) SetLocale(locale cm.Locale) {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()

	s.h.locales.Set(locale.Sys.Space.Sys.ID, locale.Sys.Environment.Sys.ID, locale.Sys.ID, &locale)
}

func (s *Server) SetEnvironmentAlias(alias cm.EnvironmentAlias) {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()

	s.h.environmentAliases.Set(alias.Sys.Space.Sys.ID, alias.Sys.ID, &alias)
}

//nolint:ireturn
func (h *Handler) GetSpace(_ context.Context, params cm.GetSpaceParams) (cm.GetSpaceRes, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if space := h.spaces[params.SpaceID]; space != nil {
		return space, nil
	}

	return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
}

func (h *Handler) resolveEnvironmentAlias(spaceID, environmentID string) string {
	if alias := h.environmentAliases.Get(spaceID, environmentID); alias != nil {
		return alias.Environment.Sys.ID
	}

	return environmentID
}

//nolint:ireturn
func (h *Handler) GetLocale(_ context.Context, params cm.GetLocaleParams) (cm.GetLocaleRes, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	locale := h.locales.Get(params.SpaceID, h.resolveEnvironmentAlias(params.SpaceID, params.EnvironmentID), params.LocaleID)
	if locale == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	result := projectLocaleResponse(*locale, params.EnvironmentID)

	return &result, nil
}

//nolint:ireturn
func (h *Handler) GetSpaces(_ context.Context, params cm.GetSpacesParams) (cm.GetSpacesRes, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	values := make([]*cm.Space, 0, len(h.spaces))
	for _, space := range h.spaces {
		if organization, ok := params.XContentfulOrganization.Get(); !ok || space.Sys.Organization.Sys.ID == organization {
			values = append(values, space)
		}
	}

	slices.SortFunc(values, func(a, b *cm.Space) int { return cmp.Compare(a.Sys.ID, b.Sys.ID) })

	skip, limit := params.Skip.Or(0), params.Limit.Or(100) //nolint:mnd
	if skip < 0 || limit < 1 {
		return NewContentfulManagementErrorStatusCodeBadRequest(new("Invalid pagination parameters"), nil), nil
	}

	start := min(skip, int64(len(values)))
	end := start + min(limit, int64(len(values))-start)

	items := make([]cm.Space, 0, end-start)
	for _, value := range values[start:end] {
		item := *value
		items = append(items, item)
	}

	return &cm.SpaceCollection{Sys: cm.SpaceCollectionSys{Type: cm.SpaceCollectionSysTypeArray}, Skip: cm.NewOptInt(int(skip)), Limit: cm.NewOptInt(int(limit)), Total: cm.NewOptInt(len(values)), Items: items}, nil
}

//nolint:ireturn,dupl // Keep generated endpoint interfaces and entity collections explicit.
func (h *Handler) GetEnvironments(_ context.Context, params cm.GetEnvironmentsParams) (cm.GetEnvironmentsRes, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.environments.Get(params.SpaceID, "master") == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	values := h.environments.List(params.SpaceID)
	slices.SortFunc(values, func(a, b *cm.Environment) int { return cmp.Compare(a.Sys.ID, b.Sys.ID) })

	skip, limit := params.Skip.Or(0), params.Limit.Or(100) //nolint:mnd
	if skip < 0 || limit < 1 {
		return NewContentfulManagementErrorStatusCodeBadRequest(new("Invalid pagination parameters"), nil), nil
	}

	start := min(skip, int64(len(values)))
	end := start + min(limit, int64(len(values))-start)

	items := make([]cm.Environment, 0, end-start)
	for _, value := range values[start:end] {
		item := *value
		items = append(items, item)
	}

	return &cm.EnvironmentCollection{Sys: cm.EnvironmentCollectionSys{Type: cm.EnvironmentCollectionSysTypeArray}, Skip: cm.NewOptInt(int(skip)), Limit: cm.NewOptInt(int(limit)), Total: cm.NewOptInt(len(values)), Items: items}, nil
}

//nolint:ireturn,dupl // Keep generated endpoint interfaces and entity collections explicit.
func (h *Handler) GetEnvironmentAliases(_ context.Context, params cm.GetEnvironmentAliasesParams) (cm.GetEnvironmentAliasesRes, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.environments.Get(params.SpaceID, "master") == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	values := h.environmentAliases.List(params.SpaceID)
	slices.SortFunc(values, func(a, b *cm.EnvironmentAlias) int { return cmp.Compare(a.Sys.ID, b.Sys.ID) })

	skip, limit := params.Skip.Or(0), params.Limit.Or(100) //nolint:mnd
	if skip < 0 || limit < 1 {
		return NewContentfulManagementErrorStatusCodeBadRequest(new("Invalid pagination parameters"), nil), nil
	}

	start := min(skip, int64(len(values)))
	end := start + min(limit, int64(len(values))-start)

	items := make([]cm.EnvironmentAlias, 0, end-start)
	for _, value := range values[start:end] {
		item := *value
		items = append(items, item)
	}

	return &cm.EnvironmentAliasCollection{Sys: cm.EnvironmentAliasCollectionSys{Type: cm.EnvironmentAliasCollectionSysTypeArray}, Skip: cm.NewOptInt(int(skip)), Limit: cm.NewOptInt(int(limit)), Total: cm.NewOptInt(len(values)), Items: items}, nil
}

//nolint:ireturn
func (h *Handler) GetLocales(_ context.Context, params cm.GetLocalesParams) (cm.GetLocalesRes, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	environmentID := h.resolveEnvironmentAlias(params.SpaceID, params.EnvironmentID)
	if h.environments.Get(params.SpaceID, environmentID) == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(nil, nil), nil
	}

	values := h.locales.List(params.SpaceID, environmentID)
	slices.SortFunc(values, func(a, b *cm.Locale) int { return cmp.Compare(a.Sys.ID, b.Sys.ID) })

	skip, limit := params.Skip.Or(0), params.Limit.Or(100) //nolint:mnd
	if skip < 0 || limit < 1 {
		return NewContentfulManagementErrorStatusCodeBadRequest(new("Invalid pagination parameters"), nil), nil
	}

	start := min(skip, int64(len(values)))
	end := start + min(limit, int64(len(values))-start)

	items := make([]cm.Locale, 0, end-start)
	for _, value := range values[start:end] {
		item := projectLocaleResponse(*value, params.EnvironmentID)
		items = append(items, item)
	}

	return &cm.LocaleCollection{Sys: cm.LocaleCollectionSys{Type: cm.LocaleCollectionSysTypeArray}, Skip: cm.NewOptInt(int(skip)), Limit: cm.NewOptInt(int(limit)), Total: cm.NewOptInt(len(values)), Items: items}, nil
}
