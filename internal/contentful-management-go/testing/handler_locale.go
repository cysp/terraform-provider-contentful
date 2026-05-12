//nolint:dupl
package cmtesting

import (
	"cmp"
	"context"
	"net/http"
	"slices"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) CreateLocale(_ context.Context, req *cm.LocaleData, params cm.CreateLocaleParams) (cm.CreateLocaleRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.environments.Get(params.SpaceID, params.EnvironmentID) == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("Environment not found"), nil), nil
	}

	localeID := req.Code
	if ts.locales.Get(params.SpaceID, params.EnvironmentID, localeID) != nil {
		return NewContentfulManagementErrorStatusCodeVersionMismatch(nil, nil), nil
	}

	newLocale := NewLocaleFromData(params.SpaceID, params.EnvironmentID, localeID, *req, false)
	ts.locales.Set(params.SpaceID, params.EnvironmentID, localeID, &newLocale)

	return &cm.LocaleStatusCode{
		StatusCode: http.StatusCreated,
		Response:   newLocale,
	}, nil
}

//nolint:ireturn
func (ts *Handler) GetLocales(_ context.Context, params cm.GetLocalesParams) (cm.GetLocalesRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.environments.Get(params.SpaceID, params.EnvironmentID) == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("Environment not found"), nil), nil
	}

	locales := ts.locales.List(params.SpaceID, params.EnvironmentID)

	items := make([]cm.Locale, 0, len(locales))
	for _, locale := range locales {
		items = append(items, *locale)
	}

	slices.SortFunc(items, func(a, b cm.Locale) int {
		return cmp.Compare(a.Sys.ID, b.Sys.ID)
	})

	skip := max(params.Skip.Or(0), 0)
	limit := max(params.Limit.Or(100), 0) //nolint:mnd
	start := min(skip, int64(len(items)))
	end := min(start+limit, int64(len(items)))

	return &cm.LocaleCollection{
		Sys: cm.LocaleCollectionSys{
			Type: cm.LocaleCollectionSysTypeArray,
		},
		Total: cm.NewOptInt(len(locales)),
		Items: items[start:end],
	}, nil
}

//nolint:ireturn
func (ts *Handler) GetLocale(_ context.Context, params cm.GetLocaleParams) (cm.GetLocaleRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	locale := ts.locales.Get(params.SpaceID, params.EnvironmentID, params.LocaleID)
	if locale == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("Locale not found"), nil), nil
	}

	return locale, nil
}

//nolint:ireturn
func (ts *Handler) PutLocale(_ context.Context, req *cm.LocaleData, params cm.PutLocaleParams) (cm.PutLocaleRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.environments.Get(params.SpaceID, params.EnvironmentID) == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("Environment not found"), nil), nil
	}

	locale := ts.locales.Get(params.SpaceID, params.EnvironmentID, params.LocaleID)
	if locale == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("Locale not found"), nil), nil
	}

	if params.XContentfulVersion != locale.Sys.Version {
		return NewContentfulManagementErrorStatusCodeVersionMismatch(nil, nil), nil
	}

	UpdateLocaleFromData(locale, *req)

	return &cm.LocaleStatusCode{
		StatusCode: http.StatusOK,
		Response:   *locale,
	}, nil
}

//nolint:ireturn
func (ts *Handler) DeleteLocale(_ context.Context, params cm.DeleteLocaleParams) (cm.DeleteLocaleRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	locale := ts.locales.Get(params.SpaceID, params.EnvironmentID, params.LocaleID)
	if locale == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("Locale not found"), nil), nil
	}

	ts.locales.Delete(params.SpaceID, params.EnvironmentID, params.LocaleID)

	return &cm.NoContent{}, nil
}
