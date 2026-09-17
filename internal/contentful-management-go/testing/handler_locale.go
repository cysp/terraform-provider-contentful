package cmtesting

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

//nolint:ireturn
func (ts *Handler) CreateLocale(_ context.Context, req *cm.LocaleData, params cm.CreateLocaleParams) (cm.CreateLocaleRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	environmentID := ts.resolveEnvironmentAlias(params.SpaceID, params.EnvironmentID)

	if ts.environments.Get(params.SpaceID, environmentID) == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("Environment not found"), nil), nil
	}

	if strings.TrimSpace(req.Name) == "" {
		return NewContentfulManagementErrorStatusCodeValidationFailed(new("Validation error"), nil), nil
	}

	if failure := validateLocaleData(ts.locales.List(params.SpaceID, environmentID), "", req); failure != nil {
		return failure, nil
	}

	localeID := generateResourceID()

	newLocale := NewLocaleFromData(params.SpaceID, environmentID, localeID, *req, false)
	ts.locales.Set(params.SpaceID, environmentID, localeID, &newLocale)

	return &cm.LocaleStatusCode{
		StatusCode: http.StatusCreated,
		Response:   projectLocaleResponse(newLocale, params.EnvironmentID),
	}, nil
}

//nolint:ireturn
func (ts *Handler) PutLocale(_ context.Context, req *cm.LocaleData, params cm.PutLocaleParams) (cm.PutLocaleRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	environmentID := ts.resolveEnvironmentAlias(params.SpaceID, params.EnvironmentID)

	if ts.environments.Get(params.SpaceID, environmentID) == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("Environment not found"), nil), nil
	}

	locale := ts.locales.Get(params.SpaceID, environmentID, params.LocaleID)
	if locale == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("Locale not found"), nil), nil
	}

	if !localeCodePattern.MatchString(req.Code) {
		return NewContentfulManagementErrorStatusCodeValidationFailed(new("Validation error"), nil), nil
	}

	// CMA accepts identical PUTs with stale or future versions.
	if localeDataEqual(locale, req) {
		return &cm.LocaleStatusCode{StatusCode: http.StatusOK, Response: projectLocaleResponse(*locale, params.EnvironmentID)}, nil
	}

	if params.XContentfulVersion != locale.Sys.Version.Or(0) {
		return NewContentfulManagementErrorStatusCodeVersionMismatch(nil, nil), nil
	}

	locales := ts.locales.List(params.SpaceID, environmentID)
	if failure := validateLocaleData(locales, params.LocaleID, req); failure != nil {
		return failure, nil
	}

	if locale.Default && req.FallbackCode.ValueStringPointer() != nil {
		return NewContentfulManagementErrorStatusCodeValidationFailed(new("The default locale cannot have a fallback"), nil), nil
	}

	// CMA permits delivery-disabled default fallback targets; see docs/research/locales.md.
	if localeUsedAsFallback(locales, locale.Code) && (req.Code != locale.Code || (!locale.Default && !req.ContentDeliveryApi)) {
		return NewContentfulManagementErrorStatusCodeValidationFailed(new("A fallback locale's code and delivery availability cannot be changed"), nil), nil
	}

	if req.Code != locale.Code {
		ts.changeLocaleContent(params.SpaceID, environmentID, locale.Code, req.Code)
	}

	updateLocaleFromData(locale, *req)

	return &cm.LocaleStatusCode{
		StatusCode: http.StatusOK,
		Response:   projectLocaleResponse(*locale, params.EnvironmentID),
	}, nil
}

//nolint:ireturn
func (ts *Handler) DeleteLocale(_ context.Context, params cm.DeleteLocaleParams) (cm.DeleteLocaleRes, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	environmentID := ts.resolveEnvironmentAlias(params.SpaceID, params.EnvironmentID)

	if ts.environments.Get(params.SpaceID, environmentID) == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("Environment not found"), nil), nil
	}

	locale := ts.locales.Get(params.SpaceID, environmentID, params.LocaleID)
	if locale == nil {
		return NewContentfulManagementErrorStatusCodeNotFound(new("Locale not found"), nil), nil
	}

	if locale.Default {
		return NewContentfulManagementErrorStatusCode(http.StatusInternalServerError, "ServerError", nil, nil), nil
	}

	if localeUsedAsFallback(ts.locales.List(params.SpaceID, environmentID), locale.Code) {
		return NewContentfulManagementErrorStatusCodeValidationFailed(new("A fallback locale cannot be deleted"), nil), nil
	}

	ts.changeLocaleContent(params.SpaceID, environmentID, locale.Code, "")
	ts.locales.Delete(params.SpaceID, environmentID, params.LocaleID)

	return &cm.NoContent{}, nil
}

var localeCodePattern = regexp.MustCompile(`^[a-zA-Z0-9-]{2,11}$`)

func validateLocaleData(locales []*cm.Locale, localeID string, data *cm.LocaleData) *cm.ErrorStatusCode {
	if !localeCodePattern.MatchString(data.Code) {
		return NewContentfulManagementErrorStatusCodeValidationFailed(new("Validation error"), nil)
	}

	byCode := make(map[string]*cm.Locale, len(locales))

	for _, locale := range locales {
		if locale.Sys.ID == localeID {
			continue
		}

		if locale.Code == data.Code {
			if localeID != "" {
				return NewContentfulManagementErrorStatusCode(http.StatusInternalServerError, "ServerError", nil, nil)
			}

			return NewContentfulManagementErrorStatusCodeValidationFailed(new("A locale with this code already exists"), nil)
		}

		byCode[locale.Code] = locale
	}

	visited := map[string]struct{}{data.Code: {}}

	for fallback := data.FallbackCode.ValueStringPointer(); fallback != nil; {
		if _, found := visited[*fallback]; found {
			return NewContentfulManagementErrorStatusCodeValidationFailed(new("A locale fallback chain cannot contain cycles"), nil)
		}

		visited[*fallback] = struct{}{}

		locale, found := byCode[*fallback]
		if !found {
			return NewContentfulManagementErrorStatusCodeValidationFailed(new("A fallback locale must exist"), nil)
		}

		if !locale.Default && !locale.ContentDeliveryApi {
			return NewContentfulManagementErrorStatusCodeValidationFailed(new("A fallback locale must be enabled for delivery"), nil)
		}

		fallback = locale.FallbackCode.ValueStringPointer()
	}

	return nil
}

func localeUsedAsFallback(locales []*cm.Locale, code string) bool {
	for _, locale := range locales {
		if fallback, ok := locale.FallbackCode.Get(); ok && fallback == code {
			return true
		}
	}

	return false
}
