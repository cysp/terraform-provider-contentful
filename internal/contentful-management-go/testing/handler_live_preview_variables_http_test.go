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
		"missing version": {method: http.MethodPut, environment: "environment", body: testJSON(map[string]any{"variables": map[string]any{}}), status: http.StatusBadRequest, response: testJSON(map[string]any{
			"sys":     map[string]any{"type": "Error", "id": "BadRequest"},
			"message": "The 'x-contentful-version' header is missing or invalid.",
		})},
		"invalid version": {method: http.MethodPut, environment: "environment", version: "abc", body: testJSON(map[string]any{"variables": map[string]any{}}), status: http.StatusBadRequest, response: testJSON(map[string]any{
			"sys":     map[string]any{"type": "Error", "id": "BadRequest"},
			"message": "The 'x-contentful-version' header is missing or invalid.",
		})},
		"negative version": {method: http.MethodPut, environment: "environment", version: "-1", body: testJSON(map[string]any{"variables": map[string]any{}}), status: http.StatusBadRequest, response: testJSON(map[string]any{
			"sys":     map[string]any{"type": "Error", "id": "BadRequest"},
			"message": "The 'x-contentful-version' header is missing or invalid.",
		})},
		"missing variables": {method: http.MethodPut, environment: "environment", version: "0", body: testJSON(map[string]any{}), status: http.StatusUnprocessableEntity, response: testJSON(map[string]any{
			"sys":     map[string]any{"type": "Error", "id": "ValidationFailed"},
			"message": "Validation error",
			"details": map[string]any{
				"errors": []any{
					map[string]any{
						"name":    "type",
						"type":    "Object",
						"details": "The type of \"value\" is incorrect, expected type: Object",
					},
				},
			},
		})},
		"malformed JSON": {method: http.MethodPut, environment: "environment", version: "0", body: `{`, status: http.StatusBadRequest, response: testJSON(map[string]any{"statusCode": 400, "error": "Bad Request", "message": "Invalid request payload JSON format"})},
		"missing parent read": {method: http.MethodGet, environment: "missing", status: http.StatusNotFound, response: testJSON(map[string]any{
			"sys":     map[string]any{"type": "Error", "id": "NotFound"},
			"message": "The resource could not be found.",
			"details": map[string]any{"type": "Environment", "id": "missing"},
		})},
		"missing parent update": {method: http.MethodPut, environment: "missing", version: "0", body: testJSON(map[string]any{"variables": map[string]any{}}), status: http.StatusNotFound, response: testJSON(map[string]any{
			"sys":     map[string]any{"type": "Error", "id": "NotFound"},
			"message": "The resource could not be found.",
			"details": map[string]any{"type": "Environment", "id": "missing"},
		})},
		"missing parent delete": {method: http.MethodDelete, environment: "missing", status: http.StatusNotFound, response: testJSON(map[string]any{
			"sys":     map[string]any{"type": "Error", "id": "NotFound"},
			"message": "The resource could not be found.",
			"details": map[string]any{"type": "Environment", "id": "missing"},
		})},
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
		"null root":    {variables: testJSON(nil), status: http.StatusUnprocessableEntity},
		"string root":  {variables: testJSON("wrong"), status: http.StatusUnprocessableEntity},
		"number root":  {variables: testJSON(123), status: http.StatusUnprocessableEntity},
		"boolean root": {variables: testJSON(true), status: http.StatusUnprocessableEntity},
		"global number": {variables: testJSON(map[string]any{"probe": 123}), status: http.StatusUnprocessableEntity, failure: testJSON(map[string]any{
			"sys":     map[string]any{"type": "Error", "id": "ValidationFailed"},
			"message": "Validation error",
			"details": map[string]any{
				"errors": []any{
					map[string]any{
						"name":    "type",
						"type":    "Text",
						"value":   123,
						"details": "The type of \"value\" is incorrect, expected type: Text",
						"path":    []any{"probe"},
						"i18nContext": map[string]any{
							"code":       "CmaError.Field.Validation.IncorrectType",
							"parameters": map[string]any{"schemaType": map[string]any{"type": "string", "value": "Text"}},
						},
					},
				},
			},
		})},
		"global boolean":    {variables: testJSON(map[string]any{"probe": true}), status: http.StatusUnprocessableEntity},
		"localized number":  {variables: testJSON(map[string]any{"probe": map[string]any{"en-US": 123}}), status: http.StatusUnprocessableEntity},
		"localized boolean": {variables: testJSON(map[string]any{"probe": map[string]any{"en-US": true}}), status: http.StatusUnprocessableEntity},
		"localized array":   {variables: testJSON(map[string]any{"probe": map[string]any{"en-US": []any{"first"}}}), status: http.StatusUnprocessableEntity},
		"localized object":  {variables: testJSON(map[string]any{"probe": map[string]any{"en-US": map[string]any{}}}), status: http.StatusUnprocessableEntity},
		"unknown locale": {variables: testJSON(map[string]any{"probe": map[string]any{"zz-ZZ": "value"}}), status: http.StatusUnprocessableEntity, failure: testJSON(map[string]any{
			"sys":     map[string]any{"type": "Error", "id": "ValidationFailed"},
			"message": "Validation error",
			"details": map[string]any{
				"errors": []any{
					map[string]any{
						"name":    "unknown",
						"value":   "value",
						"details": "The property \"zz-ZZ\" is not allowed here.",
						"path":    []any{"probe", "zz-ZZ"},
						"i18nContext": map[string]any{
							"code":       "CmaError.Field.Validation.UnknownProperty",
							"parameters": map[string]any{"propertyName": map[string]any{"type": "string", "value": "zz-ZZ"}},
						},
					},
				},
			},
		})},
		"wrong case locale":    {variables: testJSON(map[string]any{"probe": map[string]any{"en-us": "value"}}), status: http.StatusUnprocessableEntity},
		"arbitrary locale":     {variables: testJSON(map[string]any{"probe": map[string]any{"arbitrary": nil}}), status: http.StatusUnprocessableEntity},
		"global string array":  {variables: testJSON(map[string]any{"probe": []any{"first", "second"}}), status: http.StatusUnprocessableEntity},
		"proto key":            {variables: testJSON(map[string]any{"__proto__": "value"}), status: http.StatusBadRequest, failure: testJSON(map[string]any{"statusCode": 400, "error": "Bad Request", "message": "Invalid request payload JSON format"})},
		"root empty array":     {variables: testJSON([]any{}), status: http.StatusOK, stored: testJSON(map[string]any{})},
		"root string array":    {variables: testJSON([]any{"first", "second"}), status: http.StatusOK, stored: testJSON(map[string]any{"0": "first", "1": "second"})},
		"variable empty array": {variables: testJSON(map[string]any{"probe": []any{}}), status: http.StatusOK, stored: testJSON(map[string]any{"probe": map[string]any{}})},
		"mixed empty values":   {variables: testJSON(map[string]any{"empty": "", "global": nil, "localized": map[string]any{"en-US": nil}, "map": map[string]any{}}), status: http.StatusOK, stored: testJSON(map[string]any{"empty": "", "global": nil, "localized": map[string]any{"en-US": nil}, "map": map[string]any{}})},
		"unrestricted names": {variables: testJSON(map[string]any{
			"":             "empty",
			"a.b[c]/d-e_f": "punctuation",
			"café":         "unicode",
			"constructor":  "accepted",
			"prototype":    "accepted",
		}), status: http.StatusOK, stored: testJSON(map[string]any{
			"":             "empty",
			"a.b[c]/d-e_f": "punctuation",
			"café":         "unicode",
			"constructor":  "accepted",
			"prototype":    "accepted",
		})},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
			require.NoError(t, err)
			server.RegisterSpaceEnvironment("space", "environment")

			initial := testJSON(map[string]any{"keep": "original", "localized": map[string]any{"en-US": "existing"}})

			created := requestLivePreviewVariables(t, server, http.MethodPut, "0", testJSON(map[string]any{"variables": json.RawMessage(initial)}))
			require.Equal(t, http.StatusOK, created.Code)
			before := requestLivePreviewVariables(t, server, http.MethodGet, "", "")
			response := requestLivePreviewVariables(t, server, http.MethodPut, "1", testJSON(map[string]any{"variables": json.RawMessage(test.variables)}))
			require.Equal(t, test.status, response.Code, response.Body.String())

			if test.failure != "" {
				require.JSONEq(t, test.failure, response.Body.String())
			}

			after := requestLivePreviewVariables(t, server, http.MethodGet, "", "")
			if test.status != http.StatusOK {
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
	created := requestLivePreviewVariables(t, server, http.MethodPut, "0", testJSON(map[string]any{"variables": map[string]any{"keep": "original"}}))
	require.Equal(t, http.StatusOK, created.Code)
	before := requestLivePreviewVariables(t, server, http.MethodGet, "", "")
	response := requestLivePreviewVariables(t, server, http.MethodPut, "1", testJSON(map[string]any{
		"variables": map[string]any{"global": 123, "localized": map[string]any{"en-US": true, "zz-ZZ": "rejected"}},
	}))
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
	require.NoError(t, json.Unmarshal([]byte(testJSON([]any{
		map[string]any{
			"name":    "type",
			"type":    "Text",
			"value":   123,
			"details": "The type of \"value\" is incorrect, expected type: Text",
			"path":    []any{"global"},
			"i18nContext": map[string]any{
				"code":       "CmaError.Field.Validation.IncorrectType",
				"parameters": map[string]any{"schemaType": map[string]any{"type": "string", "value": "Text"}},
			},
		},
		map[string]any{
			"name":    "type",
			"type":    "Text",
			"value":   true,
			"details": "The type of \"value\" is incorrect, expected type: Text",
			"path":    []any{"localized", "en-US"},
			"i18nContext": map[string]any{
				"code":       "CmaError.Field.Validation.IncorrectType",
				"parameters": map[string]any{"schemaType": map[string]any{"type": "string", "value": "Text"}},
			},
		},
		map[string]any{
			"name":    "unknown",
			"value":   "rejected",
			"details": "The property \"zz-ZZ\" is not allowed here.",
			"path":    []any{"localized", "zz-ZZ"},
			"i18nContext": map[string]any{
				"code":       "CmaError.Field.Validation.UnknownProperty",
				"parameters": map[string]any{"propertyName": map[string]any{"type": "string", "value": "zz-ZZ"}},
			},
		},
	})), &expectedErrors))
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
	require.JSONEq(t, testJSON(map[string]any{"keep": "original"}), string(stored.Variables))
}

func TestLivePreviewVariablesHTTPTextLength(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		text      string
		localized bool
		status    int
	}{
		"ASCII boundary":     {text: strings.Repeat("a", 50000), status: http.StatusOK},
		"ASCII too long":     {text: strings.Repeat("a", 50001), status: http.StatusUnprocessableEntity},
		"localized too long": {text: strings.Repeat("a", 50001), localized: true, status: http.StatusUnprocessableEntity},
		"BMP":                {text: strings.Repeat("é", 50000), status: http.StatusOK},
		"astral":             {text: strings.Repeat("😀", 25001), status: http.StatusOK},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
			require.NoError(t, err)
			server.RegisterSpaceEnvironment("space", "environment")
			created := requestLivePreviewVariables(t, server, http.MethodPut, "0", testJSON(map[string]any{"variables": map[string]any{"keep": "original"}}))
			require.Equal(t, http.StatusOK, created.Code)
			before := requestLivePreviewVariables(t, server, http.MethodGet, "", "")
			value, err := json.Marshal(test.text)
			require.NoError(t, err)

			fragment := string(value)
			expectedPath := []string{"probe"}

			if test.localized {
				fragment = testJSON(map[string]any{"en-US": json.RawMessage(fragment)})

				expectedPath = append(expectedPath, "en-US")
			}

			response := requestLivePreviewVariables(t, server, http.MethodPut, "1", testJSON(map[string]any{"variables": map[string]any{"probe": json.RawMessage(fragment)}}))
			require.Equal(t, test.status, response.Code)

			if test.status == http.StatusOK {
				after := requestLivePreviewVariables(t, server, http.MethodGet, "", "")

				var stored struct {
					Variables json.RawMessage `json:"variables"`
				}
				require.NoError(t, json.Unmarshal(after.Body.Bytes(), &stored))
				require.JSONEq(t, testJSON(map[string]any{"probe": json.RawMessage(fragment)}), string(stored.Variables))

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
			require.JSONEq(t, testJSON(map[string]any{
				"code":       "CmaError.Field.Validation.InvalidTextLength",
				"parameters": map[string]any{"maxLength": map[string]any{"type": "number", "value": 50000}},
			}), string(item.Context))
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
	first := requestLivePreviewVariables(t, server, http.MethodPut, "0", testJSON(map[string]any{"variables": map[string]any{"keep": "original"}}))
	require.Equal(t, http.StatusOK, first.Code)
	second := requestLivePreviewVariables(t, server, http.MethodPut, "1", testJSON(map[string]any{"variables": map[string]any{"keep": "updated"}}))
	require.Equal(t, http.StatusOK, second.Code)

	before := requestLivePreviewVariables(t, server, http.MethodGet, "", "")
	for _, version := range []string{"0", "1", "999999"} {
		conflict := requestLivePreviewVariables(t, server, http.MethodPut, version, testJSON(map[string]any{"variables": map[string]any{}}))
		require.Equal(t, http.StatusConflict, conflict.Code)
		require.JSONEq(t, testJSON(map[string]any{
			"sys":     map[string]any{"type": "Error", "id": "VersionMismatch"},
			"message": "The given version value is not the current one",
		}), conflict.Body.String())
		after := requestLivePreviewVariables(t, server, http.MethodGet, "", "")
		require.JSONEq(t, before.Body.String(), after.Body.String())
	}
}
