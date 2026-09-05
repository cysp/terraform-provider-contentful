package cmtesting

import (
	"cmp"
	"context"
	"net/http"
	"slices"
	"strings"

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
	for ts.locales.Get(params.SpaceID, params.EnvironmentID, localeID) != nil {
		localeID = generateResourceID()
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
		return compareLocalesByOrder(a, b, params.Order)
	})

	skip := params.Skip.Or(0)
	if skip < 0 {
		return NewContentfulManagementErrorStatusCodeInvalidQuery(new(`The value provided for "skip" is invalid. Please provide a value larger than or equal to 0`), nil), nil
	}

	limit := params.Limit.Or(100) //nolint:mnd
	if limit < 0 || limit > 1000 {
		return NewContentfulManagementErrorStatusCodeInvalidQuery(new(`The value provided for "limit" is invalid. Please provide a value between 0 and 1000`), nil), nil
	}

	start := min(skip, int64(len(items)))
	end := min(start+limit, int64(len(items)))

	return &cm.LocaleCollection{
		Sys: cm.LocaleCollectionSys{
			Type: cm.LocaleCollectionSysTypeArray,
		},
		Total: cm.NewOptInt(len(locales)),
		Skip:  cm.NewOptInt(int(start)),
		Limit: cm.NewOptInt(int(limit)),
		Items: items[start:end],
	}, nil
}

func compareLocalesByOrder(leftLocale, rightLocale cm.Locale, order []string) int {
	for _, field := range order {
		field, descending := strings.CutPrefix(field, "-")

		result, supported := compareLocaleOrderField(leftLocale, rightLocale, field)
		if !supported {
			continue
		}

		if descending {
			result = -result
		}

		if result != 0 {
			return result
		}
	}

	return cmp.Compare(leftLocale.Sys.ID, rightLocale.Sys.ID)
}

func compareLocaleOrderField(leftLocale, rightLocale cm.Locale, field string) (int, bool) {
	switch field {
	case "sys.id":
		return cmp.Compare(leftLocale.Sys.ID, rightLocale.Sys.ID), true
	case "name":
		return cmp.Compare(leftLocale.Name, rightLocale.Name), true
	case "code":
		return cmp.Compare(leftLocale.Code, rightLocale.Code), true
	case "fallbackCode":
		return cmp.Compare(leftLocale.FallbackCode.Or(""), rightLocale.FallbackCode.Or("")), true
	default:
		return 0, false
	}
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
