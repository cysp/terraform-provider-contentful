package cmtesting_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/require"
)

func TestLivePreviewVariablesHTTPFailures(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		method, environment, version, body string
		status                             int
		response                           string
	}{
		"missing version":       {method: "PUT", environment: "environment", body: `{"variables":{}}`, status: 400, response: `{"sys":{"type":"Error","id":"BadRequest"},"message":"The 'x-contentful-version' header is missing or invalid."}`},
		"invalid version":       {method: "PUT", environment: "environment", version: "abc", body: `{"variables":{}}`, status: 400, response: `{"sys":{"type":"Error","id":"BadRequest"},"message":"The 'x-contentful-version' header is missing or invalid."}`},
		"negative version":      {method: "PUT", environment: "environment", version: "-1", body: `{"variables":{}}`, status: 400, response: `{"sys":{"type":"Error","id":"BadRequest"},"message":"The 'x-contentful-version' header is missing or invalid."}`},
		"missing variables":     {method: "PUT", environment: "environment", version: "0", body: `{}`, status: 422, response: `{"sys":{"type":"Error","id":"ValidationFailed"},"message":"Validation error","details":{"errors":[{"name":"type","type":"Object","details":"The type of \"value\" is incorrect, expected type: Object"}]}}`},
		"malformed JSON":        {method: "PUT", environment: "environment", version: "0", body: `{`, status: 400, response: `{"statusCode":400,"error":"Bad Request","message":"Invalid request payload JSON format"}`},
		"missing parent read":   {method: "GET", environment: "missing", status: 404, response: `{"sys":{"type":"Error","id":"NotFound"},"message":"The resource could not be found.","details":{"type":"Environment","id":"missing"}}`},
		"missing parent update": {method: "PUT", environment: "missing", version: "0", body: `{"variables":{}}`, status: 404, response: `{"sys":{"type":"Error","id":"NotFound"},"message":"The resource could not be found.","details":{"type":"Environment","id":"missing"}}`},
		"missing parent delete": {method: "DELETE", environment: "missing", status: 404, response: `{"sys":{"type":"Error","id":"NotFound"},"message":"The resource could not be found.","details":{"type":"Environment","id":"missing"}}`},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer()
			require.NoError(t, err)
			server.RegisterSpaceEnvironment("space", "environment")

			request := httptest.NewRequestWithContext(t.Context(), test.method, "/spaces/space/environments/"+test.environment+"/live_preview/variables", strings.NewReader(test.body))
			request.Header.Set("Authorization", "Bearer "+cmt.ValidAccessToken)
			request.Header.Set("Content-Type", "application/vnd.contentful.management.v1+json")

			if test.version != "" {
				request.Header.Set("X-Contentful-Version", test.version)
			}

			response := httptest.NewRecorder()
			server.ServeHTTP(response, request)
			require.Equal(t, test.status, response.Code)
			require.JSONEq(t, test.response, response.Body.String())
			// A rejected write must not create a variables document.
			read := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/spaces/space/environments/environment/live_preview/variables", nil)
			read.Header.Set("Authorization", "Bearer "+cmt.ValidAccessToken)

			result := httptest.NewRecorder()
			server.ServeHTTP(result, read)
			require.Equal(t, http.StatusNotFound, result.Code)
		})
	}
}

func TestLivePreviewVariablesHTTPValidationAndReplacement(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		variables string
		status    int
		stored    string
		failure   string
	}{
		"null root":            {variables: `null`, status: 422},
		"string root":          {variables: `"wrong"`, status: 422},
		"number root":          {variables: `123`, status: 422},
		"boolean root":         {variables: `true`, status: 422},
		"global number":        {variables: `{"probe":123}`, status: 422, failure: `{"sys":{"type":"Error","id":"ValidationFailed"},"message":"Validation error","details":{"errors":[{"name":"type","type":"Text","value":123,"details":"The type of \"value\" is incorrect, expected type: Text","path":["probe"],"i18nContext":{"code":"CmaError.Field.Validation.IncorrectType","parameters":{"schemaType":{"type":"string","value":"Text"}}}}]}}`},
		"global boolean":       {variables: `{"probe":true}`, status: 422},
		"localized number":     {variables: `{"probe":{"en-US":123}}`, status: 422},
		"localized boolean":    {variables: `{"probe":{"en-US":true}}`, status: 422},
		"localized array":      {variables: `{"probe":{"en-US":["first"]}}`, status: 422},
		"localized object":     {variables: `{"probe":{"en-US":{}}}`, status: 422},
		"unknown locale":       {variables: `{"probe":{"zz-ZZ":"value"}}`, status: 422, failure: `{"sys":{"type":"Error","id":"ValidationFailed"},"message":"Validation error","details":{"errors":[{"name":"unknown","value":"value","details":"The property \"zz-ZZ\" is not allowed here.","path":["probe","zz-ZZ"],"i18nContext":{"code":"CmaError.Field.Validation.UnknownProperty","parameters":{"propertyName":{"type":"string","value":"zz-ZZ"}}}}]}}`},
		"wrong case locale":    {variables: `{"probe":{"en-us":"value"}}`, status: 422},
		"arbitrary locale":     {variables: `{"probe":{"arbitrary":null}}`, status: 422},
		"global string array":  {variables: `{"probe":["first","second"]}`, status: 422},
		"proto key":            {variables: `{"__proto__":"value"}`, status: 400, failure: `{"statusCode":400,"error":"Bad Request","message":"Invalid request payload JSON format"}`},
		"root empty array":     {variables: `[]`, status: 200, stored: `{}`},
		"root string array":    {variables: `["first","second"]`, status: 200, stored: `{"0":"first","1":"second"}`},
		"variable empty array": {variables: `{"probe":[]}`, status: 200, stored: `{"probe":{}}`},
		"mixed empty values":   {variables: `{"empty":"","global":null,"localized":{"en-US":null},"map":{}}`, status: 200, stored: `{"empty":"","global":null,"localized":{"en-US":null},"map":{}}`},
		"unrestricted names":   {variables: `{"":"empty","a.b[c]/d-e_f":"punctuation","café":"unicode","constructor":"accepted","prototype":"accepted"}`, status: 200, stored: `{"":"empty","a.b[c]/d-e_f":"punctuation","café":"unicode","constructor":"accepted","prototype":"accepted"}`},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
			require.NoError(t, err)
			server.RegisterSpaceEnvironment("space", "environment")

			const initial = `{"keep":"original","localized":{"en-US":"existing"}}`

			created := requestLivePreviewVariables(t, server, http.MethodPut, "0", `{"variables":`+initial+`}`)
			require.Equal(t, 200, created.Code)
			before := requestLivePreviewVariables(t, server, http.MethodGet, "", "")
			response := requestLivePreviewVariables(t, server, http.MethodPut, "1", `{"variables":`+test.variables+`}`)
			require.Equal(t, test.status, response.Code, response.Body.String())

			if test.failure != "" {
				require.JSONEq(t, test.failure, response.Body.String())
			}

			after := requestLivePreviewVariables(t, server, http.MethodGet, "", "")
			if test.status != 200 {
				require.JSONEq(t, before.Body.String(), after.Body.String())

				return
			}

			var stored struct {
				Sys struct {
					Version int `json:"version"`
				} `json:"sys"`
				Variables json.RawMessage `json:"variables"`
			}
			require.NoError(t, json.Unmarshal(after.Body.Bytes(), &stored))
			require.Equal(t, 2, stored.Sys.Version)
			require.JSONEq(t, test.stored, string(stored.Variables))
		})
	}
}

func TestLivePreviewVariablesHTTPMultipleValidationErrors(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")
	created := requestLivePreviewVariables(t, server, http.MethodPut, "0", `{"variables":{"keep":"original"}}`)
	require.Equal(t, http.StatusOK, created.Code)
	before := requestLivePreviewVariables(t, server, http.MethodGet, "", "")
	response := requestLivePreviewVariables(t, server, http.MethodPut, "1", `{"variables":{"global":123,"localized":{"en-US":true,"zz-ZZ":"rejected"}}}`)
	require.Equal(t, http.StatusUnprocessableEntity, response.Code)

	var failure struct {
		Sys struct {
			Type string `json:"type"`
			ID   string `json:"id"`
		} `json:"sys"`
		Message string `json:"message"`
		Details struct {
			Errors []map[string]any `json:"errors"`
		} `json:"details"`
	}
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &failure))
	require.Equal(t, "Error", failure.Sys.Type)
	require.Equal(t, "ValidationFailed", failure.Sys.ID)
	require.Equal(t, "Validation error", failure.Message)

	var expectedErrors []map[string]any
	require.NoError(t, json.Unmarshal([]byte(`[
		{"name":"type","type":"Text","value":123,"details":"The type of \"value\" is incorrect, expected type: Text","path":["global"],"i18nContext":{"code":"CmaError.Field.Validation.IncorrectType","parameters":{"schemaType":{"type":"string","value":"Text"}}}},
		{"name":"type","type":"Text","value":true,"details":"The type of \"value\" is incorrect, expected type: Text","path":["localized","en-US"],"i18nContext":{"code":"CmaError.Field.Validation.IncorrectType","parameters":{"schemaType":{"type":"string","value":"Text"}}}},
		{"name":"unknown","value":"rejected","details":"The property \"zz-ZZ\" is not allowed here.","path":["localized","zz-ZZ"],"i18nContext":{"code":"CmaError.Field.Validation.UnknownProperty","parameters":{"propertyName":{"type":"string","value":"zz-ZZ"}}}}
	]`), &expectedErrors))
	require.ElementsMatch(t, expectedErrors, failure.Details.Errors)

	after := requestLivePreviewVariables(t, server, http.MethodGet, "", "")
	require.Equal(t, http.StatusOK, after.Code)
	require.JSONEq(t, before.Body.String(), after.Body.String())

	var stored struct {
		Sys struct {
			Version int `json:"version"`
		} `json:"sys"`
		Variables json.RawMessage `json:"variables"`
	}
	require.NoError(t, json.Unmarshal(after.Body.Bytes(), &stored))
	require.Equal(t, 1, stored.Sys.Version)
	require.JSONEq(t, `{"keep":"original"}`, string(stored.Variables))
}

func TestLivePreviewVariablesHTTPTextLength(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		text      string
		localized bool
		status    int
	}{
		"ASCII boundary":     {text: strings.Repeat("a", 50000), status: 200},
		"ASCII too long":     {text: strings.Repeat("a", 50001), status: 422},
		"localized too long": {text: strings.Repeat("a", 50001), localized: true, status: 422},
		"BMP":                {text: strings.Repeat("é", 50000), status: 200},
		"astral":             {text: strings.Repeat("😀", 25001), status: 200},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
			require.NoError(t, err)
			server.RegisterSpaceEnvironment("space", "environment")
			created := requestLivePreviewVariables(t, server, http.MethodPut, "0", `{"variables":{"keep":"original"}}`)
			require.Equal(t, 200, created.Code)
			before := requestLivePreviewVariables(t, server, http.MethodGet, "", "")
			value, err := json.Marshal(test.text)
			require.NoError(t, err)

			fragment := string(value)
			expectedPath := []string{"probe"}

			if test.localized {
				fragment = `{"en-US":` + fragment + `}`

				expectedPath = append(expectedPath, "en-US")
			}

			response := requestLivePreviewVariables(t, server, http.MethodPut, "1", `{"variables":{"probe":`+fragment+`}}`)
			require.Equal(t, test.status, response.Code)

			if test.status == 200 {
				after := requestLivePreviewVariables(t, server, http.MethodGet, "", "")

				var stored struct {
					Variables json.RawMessage `json:"variables"`
				}
				require.NoError(t, json.Unmarshal(after.Body.Bytes(), &stored))
				require.JSONEq(t, `{"probe":`+fragment+`}`, string(stored.Variables))

				return
			}

			var failure struct {
				Details struct {
					Errors []struct {
						Value   string          `json:"value"`
						Path    []string        `json:"path"`
						Details string          `json:"details"`
						Context json.RawMessage `json:"i18nContext"`
					} `json:"errors"`
				} `json:"details"`
			}
			require.NoError(t, json.Unmarshal(response.Body.Bytes(), &failure))
			require.Len(t, failure.Details.Errors, 1)
			item := failure.Details.Errors[0]
			require.Equal(t, test.text, item.Value)
			require.Equal(t, expectedPath, item.Path)
			require.Equal(t, "Maximum Text length is 50000 characters", item.Details)
			require.JSONEq(t, `{"code":"CmaError.Field.Validation.InvalidTextLength","parameters":{"maxLength":{"type":"number","value":50000}}}`, string(item.Context))
			after := requestLivePreviewVariables(t, server, http.MethodGet, "", "")
			require.JSONEq(t, before.Body.String(), after.Body.String())
		})
	}
}

func requestLivePreviewVariables(t *testing.T, server http.Handler, method, version, body string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequestWithContext(t.Context(), method, "/spaces/space/environments/environment/live_preview/variables", strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+cmt.ValidAccessToken)
	request.Header.Set("Content-Type", "application/vnd.contentful.management.v1+json")

	if version != "" {
		request.Header.Set("X-Contentful-Version", version)
	}

	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)

	return response
}

func TestLivePreviewVariablesHTTPConflictPreservesDocument(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")
	first := requestLivePreviewVariables(t, server, http.MethodPut, "0", `{"variables":{"keep":"original"}}`)
	require.Equal(t, 200, first.Code)
	second := requestLivePreviewVariables(t, server, http.MethodPut, "1", `{"variables":{"keep":"updated"}}`)
	require.Equal(t, 200, second.Code)

	before := requestLivePreviewVariables(t, server, http.MethodGet, "", "")
	for _, version := range []string{"0", "1", "999999"} {
		conflict := requestLivePreviewVariables(t, server, http.MethodPut, version, `{"variables":{}}`)
		require.Equal(t, 409, conflict.Code)
		require.JSONEq(t, `{"sys":{"type":"Error","id":"VersionMismatch"},"message":"The given version value is not the current one"}`, conflict.Body.String())
		after := requestLivePreviewVariables(t, server, http.MethodGet, "", "")
		require.JSONEq(t, before.Body.String(), after.Body.String())
	}
}
