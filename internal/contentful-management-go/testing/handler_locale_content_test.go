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

func localeContentRequest(t *testing.T, server *cmt.Server, method, path, version, body string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequestWithContext(t.Context(), method, "/spaces/space/environments/environment/"+path, strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+cmt.ValidAccessToken)
	request.Header.Set("Content-Type", "application/vnd.contentful.management.v1+json")
	request.Header.Set("X-Contentful-Content-Type", "type")

	if version != "" {
		request.Header.Set("X-Contentful-Version", version)
	}

	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)

	return response
}

func seedLocaleContent(t *testing.T, server *cmt.Server) {
	t.Helper()

	var contentType cm.ContentTypeRequestData
	require.NoError(t, json.Unmarshal([]byte(`{"name":"Type","description":"","displayField":"title","fields":[{"id":"title","name":"Title","type":"Symbol","localized":true,"required":true},{"id":"single","name":"Single","type":"Symbol","localized":false,"required":true},{"id":"defaults","name":"Defaults","type":"Symbol","localized":true,"defaultValue":{"en-US":"A","it-IT":"B"}}]}`), &contentType))
	server.SetContentType("space", "environment", "type", contentType)

	var entry cm.EntryRequest
	require.NoError(t, json.Unmarshal([]byte(`{"fields":{"title":{"en-US":"A","it-IT":"B"},"single":{"en-US":"S"}}}`), &entry))
	server.SetEntry("space", "environment", "type", "entry", entry)
	// Deliberately reuse the seed to prove locale edits stay within an environment.
	server.SetEntry("space", "other", "type", "entry", entry)
	server.SetContentType("space", "other", "type", contentType)
}

func TestLocaleHTTPContentProjectionAndReplacement(t *testing.T) {
	t.Parallel()
	server := newLocaleValidationServer(t)
	seedLocaleContent(t, server)
	beforeEntry := localeContentRequest(t, server, http.MethodGet, "entries/entry", "", "")
	beforeType := localeContentRequest(t, server, http.MethodGet, "content_types/type", "", "")
	disable := `{"name":"Original it-IT","code":"it-IT","fallbackCode":"fr-FR","contentDeliveryApi":true,"contentManagementApi":false,"optional":false}`
	response := localeHTTPMutation(t, server, http.MethodPut, "locale-c", "7", disable)
	require.Equal(t, http.StatusOK, response.Code)
	entry := localeContentRequest(t, server, http.MethodGet, "entries/entry", "", "")
	contentType := localeContentRequest(t, server, http.MethodGet, "content_types/type", "", "")
	assert.Contains(t, entry.Body.String(), `"title":{"en-US":"A"}`)
	assert.Contains(t, contentType.Body.String(), `"defaultValue":{"en-US":"A"}`)
	assert.Contains(t, localeContentRequest(t, server, http.MethodGet, "entries", "", "").Body.String(), `"title":{"en-US":"A"}`)
	assert.Contains(t, localeContentRequest(t, server, http.MethodGet, "content_types", "", "").Body.String(), `"defaultValue":{"en-US":"A"}`)
	assert.Contains(t, localeContentRequest(t, server, http.MethodGet, "locales/locale-c", "", "").Body.String(), `"code":"it-IT"`)

	rejected := localeContentRequest(t, server, http.MethodPut, "entries/entry", "1", `{"fields":{"title":{"en-US":"changed","it-IT":"forbidden"}}}`)
	require.Equal(t, http.StatusUnprocessableEntity, rejected.Code)
	assert.JSONEq(t, entry.Body.String(), localeContentRequest(t, server, http.MethodGet, "entries/entry", "", "").Body.String())
	rejected = localeContentRequest(t, server, http.MethodPut, "content_types/type", "1", `{"name":"Changed","description":"","displayField":"defaults","fields":[{"id":"defaults","name":"Defaults","type":"Symbol","defaultValue":{"it-IT":"forbidden"}}]}`)
	require.Equal(t, http.StatusUnprocessableEntity, rejected.Code)
	assert.JSONEq(t, contentType.Body.String(), localeContentRequest(t, server, http.MethodGet, "content_types/type", "", "").Body.String())

	enable := strings.Replace(disable, `"contentManagementApi":false`, `"contentManagementApi":true`, 1)
	require.Equal(t, http.StatusOK, localeHTTPMutation(t, server, http.MethodPut, "locale-c", "8", enable).Code)
	assert.JSONEq(t, beforeEntry.Body.String(), localeContentRequest(t, server, http.MethodGet, "entries/entry", "", "").Body.String(), "projection must not erase stored values or increment content versions")
	assert.JSONEq(t, beforeType.Body.String(), localeContentRequest(t, server, http.MethodGet, "content_types/type", "", "").Body.String())
	require.Equal(t, http.StatusOK, localeHTTPMutation(t, server, http.MethodPut, "locale-c", "9", disable).Code)
	replaced := localeContentRequest(t, server, http.MethodPut, "entries/entry", "1", `{"fields":{"title":{"en-US":"replacement"},"single":{"en-US":"S"}}}`)
	require.Equal(t, http.StatusOK, replaced.Code)
	require.Equal(t, http.StatusOK, localeHTTPMutation(t, server, http.MethodPut, "locale-c", "10", enable).Code)
	assert.NotContains(t, localeContentRequest(t, server, http.MethodGet, "entries/entry", "", "").Body.String(), "it-IT", "full replacement removes hidden translations")
}

func TestLocaleHTTPCodeChangeAndDeletionTransformContent(t *testing.T) {
	t.Parallel()
	server := newLocaleValidationServer(t)
	seedLocaleContent(t, server)
	renamed := localeHTTPMutation(t, server, http.MethodPut, "locale-c", "7", `{"name":"Renamed","code":"it-CH","fallbackCode":"fr-FR","contentDeliveryApi":true,"contentManagementApi":true,"optional":false}`)
	require.Equal(t, http.StatusOK, renamed.Code)

	for _, path := range []string{"entries/entry", "content_types/type"} {
		response := localeContentRequest(t, server, http.MethodGet, path, "", "")
		assert.Contains(t, response.Body.String(), `"it-CH":"B"`)
		assert.NotContains(t, response.Body.String(), "it-IT")
		assert.Contains(t, response.Body.String(), `"version":1`)
	}

	require.Equal(t, http.StatusNoContent, localeHTTPMutation(t, server, http.MethodDelete, "locale-c", "", "").Code)
	recreated := localeHTTPMutation(t, server, http.MethodPost, "", "", `{"name":"Recreated","code":"it-CH","fallbackCode":null,"contentDeliveryApi":true,"contentManagementApi":true,"optional":false}`)
	require.Equal(t, http.StatusCreated, recreated.Code)

	for _, path := range []string{"entries/entry", "content_types/type"} {
		response := localeContentRequest(t, server, http.MethodGet, path, "", "")
		assert.NotContains(t, response.Body.String(), "it-CH", "recreation must not restore deleted content")
		assert.Contains(t, response.Body.String(), `"version":1`)
	}

	otherResponse, err := server.Handler().GetEntry(t.Context(), cm.GetEntryParams{SpaceID: "space", EnvironmentID: "other", EntryID: "entry"})
	require.NoError(t, err)

	other, ok := otherResponse.(*cm.Entry)
	require.True(t, ok)
	assert.JSONEq(t, `{"en-US":"A","it-IT":"B"}`, string(other.Fields.Value["title"]))
	assert.Equal(t, 1, other.Sys.Version)
}

func TestLocaleHTTPPublicationRequirements(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		optional, editing, delivery, defaultEditing, defaultOptional bool
		fields                                                       string
		status                                                       int
	}{
		"missing required secondary":                   {false, true, true, true, false, `{"title":{"en-US":"A"},"single":{"en-US":"S"}}`, 422},
		"delivery disabled remains required":           {false, true, false, true, false, `{"title":{"en-US":"A"},"single":{"en-US":"S"}}`, 422},
		"editing disabled skips translation":           {false, false, true, true, false, `{"title":{"en-US":"A"},"single":{"en-US":"S"}}`, 200},
		"optional skips translation":                   {true, true, true, true, false, `{"title":{"en-US":"A"},"single":{"en-US":"S"}}`, 200},
		"default optional remains required":            {true, true, true, true, true, `{"title":{"it-IT":"B"},"single":{"en-US":"S"}}`, 422},
		"disabled default skips localized translation": {false, true, true, false, false, `{"title":{"it-IT":"B"},"single":{"en-US":"S"}}`, 200},
		"single field still required":                  {false, true, true, false, false, `{"title":{"it-IT":"B"}}`, 422},
		"absent field with no required translations":   {true, true, true, false, false, `{"single":{"en-US":"S"}}`, 422},
		"null translation":                             {false, true, true, true, false, `{"title":{"en-US":"A","it-IT":null},"single":{"en-US":"S"}}`, 422},
		"empty string translation":                     {false, true, true, true, false, `{"title":{"en-US":"A","it-IT":""},"single":{"en-US":"S"}}`, 200},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			server := newLocaleValidationServer(t)
			seedLocaleContent(t, server)
			// The two irrelevant inventory members must not require translations.
			for _, id := range []string{"locale-a", "locale-b"} {
				response, err := server.Handler().GetLocale(t.Context(), cm.GetLocaleParams{SpaceID: "space", EnvironmentID: "environment", LocaleID: id})
				require.NoError(t, err)

				locale, ok := response.(*cm.Locale)
				require.True(t, ok)

				locale.Optional = true
				server.SetLocale(*locale)
			}

			for _, item := range []struct {
				id                          string
				editing, delivery, optional bool
			}{
				{"locale-default", test.defaultEditing, true, test.defaultOptional},
				{"locale-c", test.editing, test.delivery, test.optional},
			} {
				response, err := server.Handler().GetLocale(t.Context(), cm.GetLocaleParams{SpaceID: "space", EnvironmentID: "environment", LocaleID: item.id})
				require.NoError(t, err)

				locale, ok := response.(*cm.Locale)
				require.True(t, ok)

				locale.ContentManagementApi, locale.ContentDeliveryApi, locale.Optional = item.editing, item.delivery, item.optional
				server.SetLocale(*locale)
			}

			var entry cm.EntryRequest
			require.NoError(t, json.Unmarshal([]byte(`{"fields":`+test.fields+`}`), &entry))
			server.SetEntry("space", "environment", "type", "entry", entry)
			before := localeContentRequest(t, server, http.MethodGet, "entries/entry", "", "")
			response := localeContentRequest(t, server, http.MethodPut, "entries/entry/published", "1", `{}`)
			require.Equal(t, test.status, response.Code, response.Body.String())

			if test.status == 422 {
				assert.Contains(t, response.Body.String(), `"id":"InvalidEntry"`)
				assert.JSONEq(t, before.Body.String(), localeContentRequest(t, server, http.MethodGet, "entries/entry", "", "").Body.String())
			} else {
				assert.Contains(t, response.Body.String(), `"version":2`)
			}
		})
	}
}

func TestLocaleHTTPRequiredFieldFlagsDoNotRelaxPublication(t *testing.T) {
	t.Parallel()

	for _, flag := range []string{"disabled", "omitted"} {
		t.Run(flag, func(t *testing.T) {
			t.Parallel()
			server := newLocaleValidationServer(t)
			seedLocaleContent(t, server)
			updated := localeContentRequest(t, server, http.MethodPut, "content_types/type", "1", `{"name":"Type","description":"","displayField":"title","fields":[{"id":"title","name":"Title","type":"Symbol","localized":true,"required":true,"`+flag+`":true}]}`)
			require.Equal(t, http.StatusOK, updated.Code, updated.Body.String())
			activated := localeContentRequest(t, server, http.MethodPut, "content_types/type/published", "2", `{}`)
			require.Equal(t, http.StatusOK, activated.Code, activated.Body.String())
			updated = localeContentRequest(t, server, http.MethodPut, "entries/entry", "1", `{"fields":{"title":{"en-US":"A"}}}`)
			require.Equal(t, http.StatusOK, updated.Code, updated.Body.String())
			published := localeContentRequest(t, server, http.MethodPut, "entries/entry/published", "2", `{}`)
			require.Equal(t, http.StatusUnprocessableEntity, published.Code, published.Body.String())
			assert.Contains(t, published.Body.String(), `"id":"InvalidEntry"`)
			assert.JSONEq(t, updated.Body.String(), localeContentRequest(t, server, http.MethodGet, "entries/entry", "", "").Body.String())
		})
	}
}
