package cmtesting_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocaleMutationHTTPRejectsInvalidFallbacksAndCodes(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		method, localeID, code, fallback string
		disableDelivery                  bool
		status                           int
		errorID                          string
	}{
		"create duplicate code":   {http.MethodPost, "", "de-DE", "null", false, http.StatusUnprocessableEntity, "ValidationFailed"},
		"create missing fallback": {http.MethodPost, "", "de-CH", `"missing"`, false, http.StatusUnprocessableEntity, "ValidationFailed"},
		"create empty fallback":   {http.MethodPost, "", "de-CH", `""`, false, http.StatusUnprocessableEntity, "ValidationFailed"},
		"create self fallback":    {http.MethodPost, "", "de-CH", `"de-CH"`, false, http.StatusUnprocessableEntity, "ValidationFailed"},
		"update duplicate code":   {http.MethodPut, "locale-c", "fr-FR", "null", false, http.StatusInternalServerError, "ServerError"},
		"update missing fallback": {http.MethodPut, "locale-c", "it-IT", `"missing"`, false, http.StatusUnprocessableEntity, "ValidationFailed"},
		"update empty fallback":   {http.MethodPut, "locale-c", "it-IT", `""`, false, http.StatusUnprocessableEntity, "ValidationFailed"},
		"update self fallback":    {http.MethodPut, "locale-c", "it-IT", `"it-IT"`, false, http.StatusUnprocessableEntity, "ValidationFailed"},
		"update fallback cycle":   {http.MethodPut, "locale-a", "de-DE", `"it-IT"`, false, http.StatusUnprocessableEntity, "ValidationFailed"},
		"default fallback":        {http.MethodPut, "locale-default", "en-US", `"de-DE"`, false, http.StatusUnprocessableEntity, "ValidationFailed"},
		"rename fallback target":  {http.MethodPut, "locale-a", "de-AT", "null", false, http.StatusUnprocessableEntity, "ValidationFailed"},
		"disable fallback target": {http.MethodPut, "locale-a", "de-DE", "null", true, http.StatusUnprocessableEntity, "ValidationFailed"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server := newLocaleValidationServer(t)
			before := localeCollectionJSON(t, server)

			body := `{"name":"Changed","code":"` + test.code + `","fallbackCode":` + test.fallback + `,"contentDeliveryApi":true,"contentManagementApi":true,"optional":false}`
			if test.disableDelivery {
				body = strings.Replace(body, `"contentDeliveryApi":true`, `"contentDeliveryApi":false`, 1)
			}

			response := localeHTTPMutation(t, server, test.method, test.localeID, "7", body)

			assert.Equal(t, test.status, response.Code, response.Body.String())
			assert.Contains(t, response.Body.String(), `"id":"`+test.errorID+`"`)
			assert.JSONEq(t, before, localeCollectionJSON(t, server), "rejected mutation must not change values or versions")
		})
	}
}

func TestDeleteLocaleRequiresParentEnvironment(t *testing.T) {
	t.Parallel()

	for _, environmentID := range []string{"environment", "alias"} {
		t.Run(environmentID, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
			require.NoError(t, err)

			locale := cm.Locale{
				Sys:  cm.LocaleSys{ID: "locale", Type: cm.LocaleSysTypeLocale, Space: cm.NewSpaceLink("space"), Environment: cm.NewEnvironmentLink("environment"), Version: cm.NewOptInt(7)},
				Name: "English", Code: "en-US", ContentManagementApi: true, ContentDeliveryApi: true,
			}
			server.SetLocale(locale)
			server.SetEnvironmentAlias(cmt.NewEnvironmentAliasFromEnvironmentAliasData("space", "alias", cm.EnvironmentAliasData{Environment: cm.NewEnvironmentLink("environment")}))

			response, err := server.Handler().DeleteLocale(t.Context(), cm.DeleteLocaleParams{SpaceID: "space", EnvironmentID: environmentID, LocaleID: "locale"})
			require.NoError(t, err)

			failure, ok := response.(*cm.ErrorStatusCode)
			require.True(t, ok)
			assert.Equal(t, http.StatusNotFound, failure.StatusCode)

			server.RegisterSpaceEnvironment("space", "environment")
			retained, err := server.Handler().GetLocale(t.Context(), cm.GetLocaleParams{SpaceID: "space", EnvironmentID: "environment", LocaleID: "locale"})
			require.NoError(t, err)
			assert.Equal(t, &locale, retained)
		})
	}
}

func TestDeleteLocaleHTTPRejectsDefaultAndFallbackTargets(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		localeID, errorID string
		referenceDefault  bool
		status            int
	}{
		"unreferenced default": {"locale-default", "ServerError", false, http.StatusInternalServerError},
		"referenced default":   {"locale-default", "ServerError", true, http.StatusInternalServerError},
		"nondefault fallback":  {"locale-a", "ValidationFailed", false, http.StatusUnprocessableEntity},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server := newLocaleValidationServer(t)
			if test.referenceDefault {
				dependent := localeHTTPMutation(t, server, http.MethodPut, "locale-a", "7", `{"name":"German","code":"de-DE","fallbackCode":"en-US","contentDeliveryApi":true,"contentManagementApi":true,"optional":false}`)
				require.Equal(t, http.StatusOK, dependent.Code, dependent.Body.String())
			}

			before := localeCollectionJSON(t, server)
			response := localeHTTPMutation(t, server, http.MethodDelete, test.localeID, "", "")
			assert.Equal(t, test.status, response.Code, response.Body.String())
			assert.Contains(t, response.Body.String(), `"id":"`+test.errorID+`"`)
			assert.JSONEq(t, before, localeCollectionJSON(t, server))
		})
	}
}

func TestLocaleHTTPAcceptsFallbackChainsAndRemoval(t *testing.T) {
	t.Parallel()

	server := newLocaleValidationServer(t)
	created := localeHTTPMutation(t, server, http.MethodPost, "", "", `{"name":"Swiss German","code":"de-CH","fallbackCode":"it-IT","contentDeliveryApi":true,"contentManagementApi":true,"optional":true}`)
	require.Equal(t, http.StatusCreated, created.Code, created.Body.String())

	var locale cm.Locale
	require.NoError(t, json.Unmarshal(created.Body.Bytes(), &locale))
	assert.Equal(t, "de-CH", locale.Code)
	assert.NotEmpty(t, locale.Sys.ID)
	assert.NotEqual(t, locale.Code, locale.Sys.ID)
	assert.Equal(t, cm.NewOptNilString("it-IT"), locale.FallbackCode)
	assert.Equal(t, cm.NewOptInt(1), locale.Sys.Version)
	assert.False(t, locale.Default)

	// Removing the intermediate fallback breaks the dependency on locale-b.
	cleared := localeHTTPMutation(t, server, http.MethodPut, "locale-c", "7", `{"name":"Italian","code":"it-IT","fallbackCode":null,"contentDeliveryApi":true,"contentManagementApi":true,"optional":false}`)
	require.Equal(t, http.StatusOK, cleared.Code, cleared.Body.String())
	require.NoError(t, json.Unmarshal(cleared.Body.Bytes(), &locale))
	assert.Equal(t, cm.NewOptNilStringNull(), locale.FallbackCode)
	assert.Equal(t, cm.NewOptInt(8), locale.Sys.Version)

	renamed := localeHTTPMutation(t, server, http.MethodPut, "locale-b", "7", `{"name":"French (Canada)","code":"fr-CA","fallbackCode":"de-DE","contentDeliveryApi":true,"contentManagementApi":true,"optional":false}`)
	require.Equal(t, http.StatusOK, renamed.Code, renamed.Body.String())
	require.NoError(t, json.Unmarshal(renamed.Body.Bytes(), &locale))
	assert.Equal(t, "locale-b", locale.Sys.ID)
	assert.Equal(t, "fr-CA", locale.Code)
	assert.Equal(t, cm.NewOptNilString("de-DE"), locale.FallbackCode)
	assert.Equal(t, cm.NewOptInt(8), locale.Sys.Version)

	// A locale that uses a fallback can be deleted when nothing refers to it.
	deleted := localeHTTPMutation(t, server, http.MethodDelete, "locale-b", "", "")
	require.Equal(t, http.StatusNoContent, deleted.Code, deleted.Body.String())
	deleted = localeHTTPMutation(t, server, http.MethodDelete, "locale-a", "", "")
	require.Equal(t, http.StatusNoContent, deleted.Code, deleted.Body.String())
	missing := localeHTTPMutation(t, server, http.MethodDelete, "locale-a", "", "")
	assert.Equal(t, http.StatusNotFound, missing.Code, missing.Body.String())
}

func TestLocaleHTTPDefaultFallbackLifecycle(t *testing.T) {
	t.Parallel()

	server := newLocaleValidationServer(t)
	renamed := localeHTTPMutation(t, server, http.MethodPut, "locale-default", "7", `{"name":"Default","code":"default","fallbackCode":null,"contentDeliveryApi":true,"contentManagementApi":true,"optional":false}`)
	require.Equal(t, http.StatusOK, renamed.Code, renamed.Body.String())

	var locale cm.Locale
	require.NoError(t, json.Unmarshal(renamed.Body.Bytes(), &locale))
	assert.Equal(t, "locale-default", locale.Sys.ID)
	assert.Equal(t, "default", locale.Code)
	assert.True(t, locale.Default)
	assert.Equal(t, cm.NewOptInt(8), locale.Sys.Version)

	dependent := localeHTTPMutation(t, server, http.MethodPut, "locale-a", "7", `{"name":"German","code":"de-DE","fallbackCode":"default","contentDeliveryApi":true,"contentManagementApi":true,"optional":false}`)
	require.Equal(t, http.StatusOK, dependent.Code, dependent.Body.String())
	response := localeHTTPMutation(t, server, http.MethodPut, "locale-default", "8", `{"name":"Default","code":"default","fallbackCode":null,"contentDeliveryApi":false,"contentManagementApi":true,"optional":false}`)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &locale))
	assert.Equal(t, "locale-default", locale.Sys.ID)
	assert.True(t, locale.Default)
	assert.False(t, locale.ContentDeliveryApi)
	assert.Equal(t, cm.NewOptInt(9), locale.Sys.Version)

	created := localeHTTPMutation(t, server, http.MethodPost, "", "", `{"name":"Spanish","code":"es-ES","fallbackCode":"default","contentDeliveryApi":true,"contentManagementApi":true,"optional":false}`)
	require.Equal(t, http.StatusCreated, created.Code, created.Body.String())
	disabled := localeHTTPMutation(t, server, http.MethodPut, "locale-c", "7", `{"name":"Italian","code":"it-IT","fallbackCode":"fr-FR","contentDeliveryApi":false,"contentManagementApi":true,"optional":false}`)
	require.Equal(t, http.StatusOK, disabled.Code, disabled.Body.String())
	before := localeCollectionJSON(t, server)
	rejected := localeHTTPMutation(t, server, http.MethodPost, "", "", `{"name":"Portuguese","code":"pt-PT","fallbackCode":"it-IT","contentDeliveryApi":true,"contentManagementApi":true,"optional":false}`)
	assert.Equal(t, http.StatusUnprocessableEntity, rejected.Code, rejected.Body.String())
	assert.JSONEq(t, before, localeCollectionJSON(t, server))
}

func newLocaleValidationServer(t *testing.T) *cmt.Server {
	t.Helper()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	for _, fixture := range []struct {
		id, code, fallback string
		defaultLocale      bool
	}{
		{"locale-default", "en-US", "", true},
		{"locale-a", "de-DE", "", false},
		{"locale-b", "fr-FR", "de-DE", false},
		{"locale-c", "it-IT", "fr-FR", false},
	} {
		fallback := cm.NewOptNilStringNull()
		if fixture.fallback != "" {
			fallback = cm.NewOptNilString(fixture.fallback)
		}

		// Keep fixture values independent of the mutation helper being tested.
		server.SetLocale(cm.Locale{
			Sys: cm.LocaleSys{
				ID:          fixture.id,
				Type:        cm.LocaleSysTypeLocale,
				Space:       cm.NewSpaceLink("space"),
				Environment: cm.NewEnvironmentLink("environment"),
				Version:     cm.NewOptInt(7),
			},
			Name:                 "Original " + fixture.code,
			Code:                 fixture.code,
			FallbackCode:         fallback,
			ContentDeliveryApi:   true,
			ContentManagementApi: true,
			Optional:             false,
			Default:              fixture.defaultLocale,
		})
	}

	return server
}

func localeHTTPMutation(t *testing.T, server *cmt.Server, method, localeID, version, body string) *httptest.ResponseRecorder {
	t.Helper()

	url := "/spaces/space/environments/environment/locales"
	if localeID != "" {
		url += "/" + localeID
	}

	request := httptest.NewRequestWithContext(t.Context(), method, url, strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+cmt.ValidAccessToken)
	request.Header.Set("Content-Type", "application/vnd.contentful.management.v1+json")

	if version != "" {
		request.Header.Set("X-Contentful-Version", version)
	}

	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)

	return response
}

func localeCollectionJSON(t *testing.T, server *cmt.Server) string {
	t.Helper()

	response, err := server.Handler().GetLocales(t.Context(), cm.GetLocalesParams{SpaceID: "space", EnvironmentID: "environment"})
	require.NoError(t, err)

	collection, ok := response.(*cm.LocaleCollection)
	require.True(t, ok)

	encoded, err := json.Marshal(collection)
	require.NoError(t, err)

	return string(encoded)
}

func TestLocaleHTTPNoOpAndVersionLocking(t *testing.T) {
	t.Parallel()
	server := newLocaleValidationServer(t)
	body := `{"name":"Original it-IT","code":"it-IT","fallbackCode":"fr-FR","contentDeliveryApi":true,"contentManagementApi":true,"optional":false}`

	before := localeCollectionJSON(t, server)
	for _, version := range []string{"6", "7", "107"} {
		response := localeHTTPMutation(t, server, http.MethodPut, "locale-c", version, body)
		require.Equal(t, http.StatusOK, response.Code, response.Body.String())
		assert.JSONEq(t, before, localeCollectionJSON(t, server))
	}

	changed := strings.Replace(body, "Original it-IT", "", 1)
	stale := localeHTTPMutation(t, server, http.MethodPut, "locale-c", "6", changed)
	require.Equal(t, http.StatusConflict, stale.Code)
	assert.JSONEq(t, before, localeCollectionJSON(t, server))
	accepted := localeHTTPMutation(t, server, http.MethodPut, "locale-c", "7", changed)
	require.Equal(t, http.StatusOK, accepted.Code, accepted.Body.String())

	var locale cm.Locale
	require.NoError(t, json.Unmarshal(accepted.Body.Bytes(), &locale))
	assert.Empty(t, locale.Name, "PUT preserves an empty name")
	assert.Equal(t, cm.NewOptInt(8), locale.Sys.Version)
	repeated := localeHTTPMutation(t, server, http.MethodPut, "locale-c", "7", changed)
	require.Equal(t, http.StatusOK, repeated.Code)
	assert.JSONEq(t, accepted.Body.String(), repeated.Body.String())
}

func TestLocaleHTTPCodeSyntaxPrecedesVersionCheck(t *testing.T) {
	t.Parallel()

	for _, code := range []string{"", "x", "en_CA", "abcdefghijkl"} {
		t.Run(code, func(t *testing.T) {
			t.Parallel()
			server := newLocaleValidationServer(t)
			before := localeCollectionJSON(t, server)
			response := localeHTTPMutation(t, server, http.MethodPut, "locale-c", "6", `{"name":"Changed","code":"`+code+`","fallbackCode":null,"contentDeliveryApi":true,"contentManagementApi":true,"optional":false}`)
			assert.Equal(t, http.StatusUnprocessableEntity, response.Code, response.Body.String())
			assert.JSONEq(t, before, localeCollectionJSON(t, server))
		})
	}
}

func TestLocaleMutationsThroughEnvironmentAliasAndParentDeletion(t *testing.T) {
	t.Parallel()
	server := newLocaleValidationServer(t)
	server.SetEnvironmentAlias(cmt.NewEnvironmentAliasFromEnvironmentAliasData("space", "alias", cm.EnvironmentAliasData{Environment: cm.NewEnvironmentLink("environment")}))

	payload := &cm.LocaleData{Name: "Spanish", Code: "es-ES", FallbackCode: cm.NewNilStringNull(), ContentDeliveryApi: true, ContentManagementApi: true}
	created, err := server.Handler().CreateLocale(t.Context(), payload, cm.CreateLocaleParams{SpaceID: "space", EnvironmentID: "alias"})
	require.NoError(t, err)

	result, ok := created.(*cm.LocaleStatusCode)
	require.True(t, ok)
	require.Equal(t, http.StatusCreated, result.StatusCode)
	assert.Equal(t, "alias", result.Response.Sys.Environment.Sys.ID)
	localeID := result.Response.Sys.ID
	direct, err := server.Handler().GetLocale(t.Context(), cm.GetLocaleParams{SpaceID: "space", EnvironmentID: "environment", LocaleID: localeID})
	require.NoError(t, err)

	locale, ok := direct.(*cm.Locale)
	require.True(t, ok)
	assert.Equal(t, "environment", locale.Sys.Environment.Sys.ID)

	payload.Name = "Spanish (updated)"
	updated, err := server.Handler().PutLocale(t.Context(), payload, cm.PutLocaleParams{SpaceID: "space", EnvironmentID: "alias", LocaleID: localeID, XContentfulVersion: 1})
	require.NoError(t, err)

	result, ok = updated.(*cm.LocaleStatusCode)
	require.True(t, ok)
	assert.Equal(t, cm.NewOptInt(2), result.Response.Sys.Version)
	assert.Equal(t, "alias", result.Response.Sys.Environment.Sys.ID)
	deleted, err := server.Handler().DeleteLocale(t.Context(), cm.DeleteLocaleParams{SpaceID: "space", EnvironmentID: "alias", LocaleID: localeID})
	require.NoError(t, err)
	assert.IsType(t, &cm.NoContent{}, deleted)
	_, err = server.Handler().DeleteEnvironment(t.Context(), cm.DeleteEnvironmentParams{SpaceID: "space", EnvironmentID: "environment"})
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")
	missing, err := server.Handler().GetLocale(t.Context(), cm.GetLocaleParams{SpaceID: "space", EnvironmentID: "alias", LocaleID: "locale-default"})
	require.NoError(t, err)

	failure, ok := missing.(*cm.ErrorStatusCode)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, failure.StatusCode)
}

func TestLocaleHTTPCreateRejectsBlankNames(t *testing.T) {
	t.Parallel()

	for name, value := range map[string]string{"empty": "", "spaces": "  "} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server := newLocaleValidationServer(t)
			before := localeCollectionJSON(t, server)
			response := localeHTTPMutation(t, server, http.MethodPost, "", "", `{"name":"`+value+`","code":"es-ES","fallbackCode":null,"contentDeliveryApi":true,"contentManagementApi":true,"optional":false}`)
			require.Equal(t, http.StatusUnprocessableEntity, response.Code, response.Body.String())
			assert.JSONEq(t, before, localeCollectionJSON(t, server))
		})
	}
}
