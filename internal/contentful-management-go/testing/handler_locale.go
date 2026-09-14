package cmtesting

import (
	"context"
	"net/http"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) CreateLocale(_ context.Context, req *cm.LocaleData, params cm.CreateLocaleParams) (cm.CreateLocaleRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.environments.Get(params.SpaceID, params.EnvironmentID) == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("Environment not found"), nil), nil
	}

	for _, locale := range ts.locales.List(params.SpaceID, params.EnvironmentID) {
		if locale.Code == req.Code {
			return NewContentfulManagementErrorStatusCodeValidationFailed(new("A locale with this code already exists"), nil), nil
		}
	}

	localeID := generateResourceID()

	newLocale := NewLocaleFromData(params.SpaceID, params.EnvironmentID, localeID, *req, false)
	ts.locales.Set(params.SpaceID, params.EnvironmentID, localeID, &newLocale)

	return &cm.LocaleStatusCode{
		StatusCode: http.StatusCreated,
		Response:   newLocale,
	}, nil
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

	if params.XContentfulVersion != locale.Sys.Version.Or(0) {
		return NewContentfulManagementErrorStatusCodeVersionMismatch(nil, nil), nil
	}

	updateLocaleFromData(locale, *req)

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
