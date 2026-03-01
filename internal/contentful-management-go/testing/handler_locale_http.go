package cmtesting

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
)

func (h *Handler) HandleLocaleHTTP(w http.ResponseWriter, r *http.Request, sec *SecurityHandler) bool {
	spaceID, localeID, ok := localeRoutePath(r.URL.Path)
	if !ok {
		return false
	}

	if !authorizeLocaleRoute(w, r, sec) {
		return true
	}

	switch r.Method {
	case http.MethodGet:
		h.handleGetLocale(w, spaceID, localeID)
	case http.MethodPut:
		h.handlePutLocale(w, r, spaceID, localeID)
	case http.MethodDelete:
		h.handleDeleteLocale(w, r, spaceID, localeID)
	default:
		w.Header().Set("Allow", strings.Join([]string{http.MethodGet, http.MethodPut, http.MethodDelete}, ", "))
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

	return true
}

func localeRoutePath(routePath string) (spaceID, localeID string, ok bool) {
	trimmedPath := strings.Trim(routePath, "/")
	parts := strings.Split(trimmedPath, "/")
	if len(parts) != 4 || parts[0] != "spaces" || parts[2] != "locales" {
		return "", "", false
	}

	if parts[1] == "" || parts[3] == "" {
		return "", "", false
	}

	return parts[1], parts[3], true
}

func authorizeLocaleRoute(w http.ResponseWriter, r *http.Request, sec *SecurityHandler) bool {
	const bearerPrefix = "Bearer "

	authorization := r.Header.Get("Authorization")
	if !strings.HasPrefix(authorization, bearerPrefix) {
		message := "The access token you sent could not be found or is invalid."
		_ = WriteContentfulManagementErrorResponse(w, http.StatusUnauthorized, "AccessTokenInvalid", &message, nil)
		return false
	}

	accessToken := strings.TrimPrefix(authorization, bearerPrefix)
	operationName := localeOperationNameForMethod(r.Method)
	_, err := sec.HandleAccessToken(r.Context(), operationName, cm.AccessToken{Token: accessToken})
	if errors.Is(err, ErrAccessTokenInvalid) {
		message := "The access token you sent could not be found or is invalid."
		_ = WriteContentfulManagementErrorResponse(w, http.StatusUnauthorized, "AccessTokenInvalid", &message, nil)
		return false
	}

	if err != nil {
		message := "Failed to authorize access token."
		_ = WriteContentfulManagementErrorResponse(w, http.StatusUnauthorized, "AccessTokenInvalid", &message, nil)
		return false
	}

	return true
}

func localeOperationNameForMethod(method string) cm.OperationName {
	switch method {
	case http.MethodGet:
		return cm.GetLocaleOperation
	case http.MethodPut:
		return cm.PutLocaleOperation
	case http.MethodDelete:
		return cm.DeleteLocaleOperation
	default:
		return cm.OperationName(method)
	}
}

func (h *Handler) handleGetLocale(w http.ResponseWriter, spaceID, localeID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	locale := h.locales.Get(spaceID, localeID)
	if locale == nil {
		_ = WriteContentfulManagementErrorResponse(w, http.StatusNotFound, cm.ErrorSysIDNotFound, new("Locale not found"), nil)
		return
	}

	_ = WriteContentfulManagementResponse(w, http.StatusOK, locale)
}

func (h *Handler) handlePutLocale(w http.ResponseWriter, r *http.Request, spaceID, localeID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	request := cm.LocaleRequest{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		_ = WriteContentfulManagementErrorResponse(w, http.StatusBadRequest, "BadRequest", new("Invalid locale request body"), nil)
		return
	}

	existing := h.locales.Get(spaceID, localeID)
	if existing != nil {
		version, versionSet := parseContentfulVersionHeader(r.Header.Get("X-Contentful-Version"))
		if !versionSet || version != existing.Sys.Version {
			_ = WriteContentfulManagementResponse(w, http.StatusConflict, NewContentfulManagementErrorStatusCodeVersionMismatch(nil, nil).Response)
			return
		}

		UpdateLocaleFromRequest(existing, &request)

		_ = WriteContentfulManagementResponse(w, http.StatusOK, existing)
		return
	}

	locale := NewLocaleFromRequest(spaceID, localeID, &request)
	h.locales.Set(spaceID, localeID, &locale)

	_ = WriteContentfulManagementResponse(w, http.StatusCreated, &locale)
}

func (h *Handler) handleDeleteLocale(w http.ResponseWriter, r *http.Request, spaceID, localeID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	locale := h.locales.Get(spaceID, localeID)
	if locale == nil {
		_ = WriteContentfulManagementErrorResponse(w, http.StatusNotFound, cm.ErrorSysIDNotFound, new("Locale not found"), nil)
		return
	}

	version, versionSet := parseContentfulVersionHeader(r.Header.Get("X-Contentful-Version"))
	if versionSet && version != locale.Sys.Version {
		_ = WriteContentfulManagementResponse(w, http.StatusConflict, NewContentfulManagementErrorStatusCodeVersionMismatch(nil, nil).Response)
		return
	}

	h.locales.Delete(spaceID, localeID)

	w.WriteHeader(http.StatusNoContent)
}

func parseContentfulVersionHeader(raw string) (version int, set bool) {
	if raw == "" {
		return 0, false
	}

	parsed, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false
	}

	return parsed, true
}
