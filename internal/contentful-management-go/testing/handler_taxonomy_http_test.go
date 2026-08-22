package cmtesting_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const taxonomyConceptRequestPath = "/organizations/organization/taxonomy/concepts/concept"

func TestContentfulManagementServerMatchesObservedTaxonomyDeleteVersionBehavior(t *testing.T) {
	t.Parallel()

	for resourceType, collectionPath := range map[string]string{
		"concept": "/organizations/organization/taxonomy/concepts",
		"scheme":  "/organizations/organization/taxonomy/concept-schemes",
	} {
		t.Run(resourceType, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(100))
			require.NoError(t, err)

			testServer := httptest.NewServer(server)
			t.Cleanup(testServer.Close)

			correctPath := collectionPath + "/correct"
			status, body := taxonomyHTTPRequestAt(t, testServer, correctPath, http.MethodPut, 0, `{"prefLabel":{"en-US":"Correct"}}`)
			require.Equal(t, http.StatusCreated, status)
			assert.Equal(t, 1, decodeTaxonomyConceptResponse(t, body).Sys.Version)

			status, body = taxonomyHTTPRequestAtWithVersionHeader(t, testServer, correctPath, http.MethodDelete, "1", "")
			assert.Equal(t, http.StatusNoContent, status)
			assert.Empty(t, body)

			status, _ = taxonomyHTTPRequestAt(t, testServer, correctPath, http.MethodGet, 0, "")
			assert.Equal(t, http.StatusNotFound, status)

			validationPath := collectionPath + "/validation"
			status, body = taxonomyHTTPRequestAt(t, testServer, validationPath, http.MethodPut, 0, `{"prefLabel":{"en-US":"Validation"}}`)
			require.Equal(t, http.StatusCreated, status)
			assert.Equal(t, 1, decodeTaxonomyConceptResponse(t, body).Sys.Version)

			status, body = taxonomyHTTPRequestAtWithVersionHeader(t, testServer, validationPath, http.MethodDelete, "", "")
			assert.Equal(t, http.StatusUnprocessableEntity, status)
			assert.JSONEq(t, `{
				"sys":{"type":"Error","id":"ValidationFailed"},
				"message":"Validation error",
				"details":{
					"flatten":{"formErrors":[],"fieldErrors":{"x-contentful-version":["Invalid input: expected number, received NaN"]}},
					"errors":[{"name":"invalid_type","path":["x-contentful-version"],"details":"Invalid input: expected number, received NaN"}]
				}
			}`, string(body))
			assertTaxonomyResourceVersion(t, testServer, validationPath, 1)

			status, body = taxonomyHTTPRequestAtWithVersionHeader(t, testServer, validationPath, http.MethodDelete, "0", "")
			assert.Equal(t, http.StatusUnprocessableEntity, status)
			assert.JSONEq(t, `{
				"sys":{"type":"Error","id":"ValidationFailed"},
				"message":"Validation error",
				"details":{
					"flatten":{"formErrors":[],"fieldErrors":{"x-contentful-version":["Too small: expected number to be >0"]}},
					"errors":[{"name":"too_small","path":["x-contentful-version"],"details":"Too small: expected number to be >0"}]
				}
			}`, string(body))
			assertTaxonomyResourceVersion(t, testServer, validationPath, 1)

			status, body = taxonomyHTTPRequestAtWithVersionHeader(t, testServer, validationPath, http.MethodPatch, "1", `[{"op":"replace","path":"/prefLabel","value":{"en-US":"Version 2"}}]`)
			require.Equal(t, http.StatusOK, status)
			assert.Equal(t, 2, decodeTaxonomyConceptResponse(t, body).Sys.Version)

			status, body = taxonomyHTTPRequestAtWithVersionHeader(t, testServer, validationPath, http.MethodDelete, "1", "")
			assert.Equal(t, http.StatusConflict, status)
			assert.JSONEq(t, `{
				"sys":{"type":"Error","id":"VersionMismatch"},
				"message":"Version mismatch",
				"details":"Version mismatch, expected 2, got 1."
			}`, string(body))
			assertTaxonomyResourceVersion(t, testServer, validationPath, 2)

			status, body = taxonomyHTTPRequestAtWithVersionHeader(t, testServer, validationPath, http.MethodDelete, "2", "")
			assert.Equal(t, http.StatusNoContent, status)
			assert.Empty(t, body)

			status, _ = taxonomyHTTPRequestAt(t, testServer, validationPath, http.MethodGet, 0, "")
			assert.Equal(t, http.StatusNotFound, status)
		})
	}
}

func assertTaxonomyResourceVersion(t *testing.T, server *httptest.Server, path string, expected int) {
	t.Helper()

	status, body := taxonomyHTTPRequestAt(t, server, path, http.MethodGet, 0, "")
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, expected, decodeTaxonomyConceptResponse(t, body).Sys.Version)
}

func TestContentfulManagementServerRejectsIncompleteConceptLabelMaps(t *testing.T) {
	t.Parallel()

	for shape, labelMap := range map[string]string{"empty": `{}`, "nonpreferred_only": `{"fr-FR":[]}`} {
		for _, field := range []string{"altLabels", "hiddenLabels"} {
			t.Run(shape+"_"+field, func(t *testing.T) {
				t.Parallel()

				server, err := cmt.NewContentfulManagementServer()
				require.NoError(t, err)

				testServer := httptest.NewServer(server)
				t.Cleanup(testServer.Close)

				status, responseBody := taxonomyHTTPRequest(t, testServer, http.MethodPut, 0, `{
				"prefLabel": {"en-US": "Concept"},
				"`+field+`": `+labelMap+`
			}`)

				assert.Equal(t, http.StatusUnprocessableEntity, status)
				assertTaxonomyLabelValidationError(t, responseBody, field, "en-US")

				status, _ = taxonomyHTTPRequest(t, testServer, http.MethodGet, 0, "")
				assert.Equal(t, http.StatusNotFound, status, "a rejected PUT must not create the concept")
			})
		}
	}
}

func TestContentfulManagementServerRejectsIncompletePatchedConceptLabelMaps(t *testing.T) {
	t.Parallel()

	for shape, labelMap := range map[string]string{"empty": `{}`, "nonpreferred_only": `{"fr-FR":[]}`} {
		for _, operation := range []string{"add", "replace"} {
			for _, field := range []string{"altLabels", "hiddenLabels"} {
				t.Run(shape+"_"+operation+"_"+field, func(t *testing.T) {
					t.Parallel()

					server, err := cmt.NewContentfulManagementServer()
					require.NoError(t, err)

					testServer := httptest.NewServer(server)
					t.Cleanup(testServer.Close)

					status, _ := taxonomyHTTPRequest(t, testServer, http.MethodPut, 0, `{
					"prefLabel": {"en-US": "Concept"}
				}`)
					require.Equal(t, http.StatusCreated, status)

					status, responseBody := taxonomyHTTPRequest(t, testServer, http.MethodPatch, 1, `[
					{"op": "`+operation+`", "path": "/`+field+`", "value": `+labelMap+`}
				]`)

					assert.Equal(t, http.StatusUnprocessableEntity, status)
					assertTaxonomyLabelValidationError(t, responseBody, field, "en-US")

					if shape == "empty" && operation == "add" && field == "altLabels" {
						assert.JSONEq(t, `{
					"sys": {"type": "Error", "id": "ValidationFailed"},
					"message": "Validation error",
					"details": {
						"flatten": {
							"formErrors": [],
							"fieldErrors": {
								"`+field+`": ["Invalid input: expected array, received undefined"]
							}
						},
						"errors": [{
							"name": "invalid_type",
							"path": ["`+field+`", "en-US"],
							"details": "Invalid input: expected array, received undefined"
						}]
					}
					}`, string(responseBody))
					}

					status, responseBody = taxonomyHTTPRequest(t, testServer, http.MethodGet, 0, "")
					require.Equal(t, http.StatusOK, status)

					current := decodeTaxonomyConceptResponse(t, responseBody)
					assert.Equal(t, 1, current.Sys.Version, "a rejected PATCH must not increment the version")
					assert.Equal(t, map[string]string{"en-US": "Concept"}, current.PrefLabel)
					assert.Equal(t, map[string][]string{"en-US": {}}, current.AltLabels)
					assert.Equal(t, map[string][]string{"en-US": {}}, current.HiddenLabels)
				})
			}
		}
	}
}

func TestContentfulManagementServerAcceptsCompletePatchedConceptLabelMaps(t *testing.T) {
	t.Parallel()

	for _, operation := range []string{"add", "replace"} {
		for _, field := range []string{"altLabels", "hiddenLabels"} {
			t.Run(operation+"_"+field, func(t *testing.T) {
				t.Parallel()

				server, err := cmt.NewContentfulManagementServer()
				require.NoError(t, err)

				testServer := httptest.NewServer(server)
				t.Cleanup(testServer.Close)

				status, _ := taxonomyHTTPRequest(t, testServer, http.MethodPut, 0, `{
					"prefLabel": {"en-US": "Concept"}
				}`)
				require.Equal(t, http.StatusCreated, status)

				status, responseBody := taxonomyHTTPRequest(t, testServer, http.MethodPatch, 1, `[
					{"op": "`+operation+`", "path": "/`+field+`", "value": {"en-US": ["Term"]}}
				]`)
				require.Equal(t, http.StatusOK, status)

				current := decodeTaxonomyConceptResponse(t, responseBody)
				assert.Equal(t, 2, current.Sys.Version)

				if field == "altLabels" {
					assert.Equal(t, map[string][]string{"en-US": {"Term"}}, current.AltLabels)
					assert.Equal(t, map[string][]string{"en-US": {}}, current.HiddenLabels)
				} else {
					assert.Equal(t, map[string][]string{"en-US": {}}, current.AltLabels)
					assert.Equal(t, map[string][]string{"en-US": {"Term"}}, current.HiddenLabels)
				}
			})
		}
	}
}

func TestContentfulManagementServerCanonicalizesOmittedAndAcceptsCompleteConceptLabelMaps(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		request          string
		wantAltLabels    map[string][]string
		wantHiddenLabels map[string][]string
	}{
		"omitted": {
			request:          `{"prefLabel":{"en-US":"Concept"}}`,
			wantAltLabels:    map[string][]string{"en-US": {}},
			wantHiddenLabels: map[string][]string{"en-US": {}},
		},
		"complete": {
			request:          `{"prefLabel":{"en-US":"Concept"},"altLabels":{"en-US":["Term"]},"hiddenLabels":{"en-US":[]}}`,
			wantAltLabels:    map[string][]string{"en-US": {"Term"}},
			wantHiddenLabels: map[string][]string{"en-US": {}},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer()
			require.NoError(t, err)

			testServer := httptest.NewServer(server)
			t.Cleanup(testServer.Close)

			status, responseBody := taxonomyHTTPRequest(t, testServer, http.MethodPut, 0, test.request)
			require.Equal(t, http.StatusCreated, status)

			current := decodeTaxonomyConceptResponse(t, responseBody)
			assert.Equal(t, 1, current.Sys.Version)
			assert.Equal(t, map[string]string{"en-US": "Concept"}, current.PrefLabel)
			assert.Equal(t, test.wantAltLabels, current.AltLabels)
			assert.Equal(t, test.wantHiddenLabels, current.HiddenLabels)
		})
	}
}

func TestContentfulManagementServerRejectsPatchedConceptPreferredLabelWithoutEnUS(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)

	testServer := httptest.NewServer(server)
	t.Cleanup(testServer.Close)

	status, _ := taxonomyHTTPRequest(t, testServer, http.MethodPut, 0, `{
		"prefLabel": {"en-US": "Concept"}
	}`)
	require.Equal(t, http.StatusCreated, status)

	status, responseBody := taxonomyHTTPRequest(t, testServer, http.MethodPatch, 1, `[
		{"op": "replace", "path": "/prefLabel", "value": {"fr-FR": "Concept"}}
	]`)

	assert.Equal(t, http.StatusUnprocessableEntity, status)
	assertTaxonomyPrefLabelValidationError(t, responseBody)

	status, responseBody = taxonomyHTTPRequest(t, testServer, http.MethodGet, 0, "")
	require.Equal(t, http.StatusOK, status)
	current := decodeTaxonomyConceptResponse(t, responseBody)
	assert.Equal(t, 1, current.Sys.Version)
	assert.Equal(t, map[string]string{"en-US": "Concept"}, current.PrefLabel)
}

func TestContentfulManagementServerRejectsTaxonomyPreferredLabelWithoutEnUS(t *testing.T) {
	t.Parallel()

	for name, requestPath := range map[string]string{
		"concept":        taxonomyConceptRequestPath,
		"concept scheme": "/organizations/organization/taxonomy/concept-schemes/scheme",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer()
			require.NoError(t, err)

			testServer := httptest.NewServer(server)
			t.Cleanup(testServer.Close)

			status, responseBody := taxonomyHTTPRequestAt(t, testServer, requestPath, http.MethodPut, 0, `{"prefLabel":{"fr-FR":"Concept"}}`)
			assert.Equal(t, http.StatusUnprocessableEntity, status)
			assertTaxonomyPrefLabelValidationError(t, responseBody)

			status, _ = taxonomyHTTPRequestAt(t, testServer, requestPath, http.MethodGet, 0, "")
			assert.Equal(t, http.StatusNotFound, status)

			status, _ = taxonomyHTTPRequestAt(t, testServer, requestPath, http.MethodPut, 0, `{"prefLabel":{"en-US":"Concept"}}`)
			require.Equal(t, http.StatusCreated, status)
			status, responseBody = taxonomyHTTPRequestAt(t, testServer, requestPath, http.MethodPatch, 1, `[{"op":"replace","path":"/prefLabel","value":{"fr-FR":"Concept"}}]`)
			assert.Equal(t, http.StatusUnprocessableEntity, status)
			assertTaxonomyPrefLabelValidationError(t, responseBody)

			status, responseBody = taxonomyHTTPRequestAt(t, testServer, requestPath, http.MethodGet, 0, "")
			require.Equal(t, http.StatusOK, status)
			current := decodeTaxonomyConceptResponse(t, responseBody)
			assert.Equal(t, 1, current.Sys.Version, "a rejected PATCH must not increment the version")
			assert.Equal(t, map[string]string{"en-US": "Concept"}, current.PrefLabel)
		})
	}
}

func TestContentfulManagementServerFiltersConceptSchemePreferredLabelLocales(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)

	testServer := httptest.NewServer(server)
	t.Cleanup(testServer.Close)

	path := "/organizations/organization/taxonomy/concept-schemes/scheme"

	status, body := taxonomyHTTPRequestAt(t, testServer, path, http.MethodPut, 0, `{"prefLabel":{"en-US":"Scheme","fr-FR":"Schema"},"definition":{"fr-FR":"Definition"}}`)
	require.Equal(t, http.StatusCreated, status)
	current := decodeTaxonomyConceptResponse(t, body)
	assert.Equal(t, 1, current.Sys.Version)
	assert.Equal(t, map[string]string{"en-US": "Scheme"}, current.PrefLabel)

	status, body = taxonomyHTTPRequestAt(t, testServer, path, http.MethodPatch, 1, `[{"op":"replace","path":"/prefLabel","value":{"en-US":"Updated","fr-FR":"Mis a jour"}}]`)
	require.Equal(t, http.StatusOK, status)
	current = decodeTaxonomyConceptResponse(t, body)
	assert.Equal(t, 2, current.Sys.Version)
	assert.Equal(t, map[string]string{"en-US": "Updated"}, current.PrefLabel)
}

func TestContentfulManagementServerFiltersAllObservedTaxonomyLocalizedFields(t *testing.T) {
	t.Parallel()

	conceptFields := []string{"note", "changeNote", "definition", "editorialNote", "example", "historyNote", "scopeNote"}
	fields := append([]string{"prefLabel"}, conceptFields...)

	for name, requestPath := range map[string]string{
		"concept": "/organizations/organization/taxonomy/concepts/concept",
		"scheme":  "/organizations/organization/taxonomy/concept-schemes/scheme",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer()
			require.NoError(t, err)

			testServer := httptest.NewServer(server)
			t.Cleanup(testServer.Close)

			var builder strings.Builder
			builder.WriteString(`{"prefLabel":{"en-US":"US","fr-FR":"FR"}`)

			for _, field := range conceptFields {
				if name == "scheme" && field != "definition" {
					continue
				}

				builder.WriteString(`,"` + field + `":{"en-US":"US","fr-FR":"FR"}`)
			}

			builder.WriteString(`}`)
			body := builder.String()
			status, response := taxonomyHTTPRequestAt(t, testServer, requestPath, http.MethodPut, 0, body)
			require.Equal(t, http.StatusCreated, status)
			assertTaxonomyResponseOnlyEnUS(t, response, fieldsForTaxonomyResource(name, fields))

			patch := `[{"op":"replace","path":"/prefLabel","value":{"en-US":"US2","fr-FR":"FR2"}}]`
			status, response = taxonomyHTTPRequestAt(t, testServer, requestPath, http.MethodPatch, 1, patch)
			require.Equal(t, http.StatusOK, status)
			assertTaxonomyResponseOnlyEnUS(t, response, fieldsForTaxonomyResource(name, fields))
		})
	}
}

func fieldsForTaxonomyResource(name string, fields []string) []string {
	if name == "concept" {
		return fields
	}

	return []string{"prefLabel", "definition"}
}

func assertTaxonomyResponseOnlyEnUS(t *testing.T, body []byte, fields []string) {
	t.Helper()

	document := map[string]json.RawMessage{}
	require.NoError(t, json.Unmarshal(body, &document))

	for _, field := range fields {
		labels := map[string]string{}
		require.NoError(t, json.Unmarshal(document[field], &labels), field)
		assert.Equal(t, map[string]string{"en-US": labels["en-US"]}, labels, field)
	}
}

func TestContentfulManagementServerCanonicalizesEmptyLocalizedFieldsToNull(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		requestPath string
		fields      []string
	}{
		"concept": {
			requestPath: "/organizations/organization/taxonomy/concepts/concept-empty",
			fields:      []string{"note", "changeNote", "definition", "editorialNote", "example", "historyNote", "scopeNote"},
		},
		"scheme": {
			requestPath: "/organizations/organization/taxonomy/concept-schemes/scheme-empty",
			fields:      []string{"definition"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer()
			require.NoError(t, err)

			testServer := httptest.NewServer(server)
			t.Cleanup(testServer.Close)

			var request strings.Builder
			request.WriteString(`{"prefLabel":{"en-US":"US"}`)

			for _, field := range test.fields {
				request.WriteString(`,"` + field + `":{"en-US":"value"}`)
			}

			request.WriteString(`}`)
			status, _ := taxonomyHTTPRequestAt(t, testServer, test.requestPath, http.MethodPut, 0, request.String())
			require.Equal(t, http.StatusCreated, status)

			operations := make([]string, 0, len(test.fields))
			for _, field := range test.fields {
				operations = append(operations, `{"op":"replace","path":"/`+field+`","value":{}}`)
			}

			status, body := taxonomyHTTPRequestAt(t, testServer, test.requestPath, http.MethodPatch, 1, `[`+strings.Join(operations, ",")+`]`)
			require.Equal(t, http.StatusOK, status)

			for _, field := range test.fields {
				assertTaxonomyResponseFieldNull(t, body, field)
			}

			status, body = taxonomyHTTPRequestAt(t, testServer, test.requestPath, http.MethodGet, 0, "")
			require.Equal(t, http.StatusOK, status)

			for _, field := range test.fields {
				assertTaxonomyResponseFieldNull(t, body, field)
			}
		})
	}
}

func assertTaxonomyResponseFieldNull(t *testing.T, body []byte, field string) {
	t.Helper()

	document := map[string]json.RawMessage{}
	require.NoError(t, json.Unmarshal(body, &document))
	assert.JSONEq(t, "null", string(document[field]))
}

func TestContentfulManagementServerFiltersTaxonomyLocalesAndRejectsLabelRemoval(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)

	testServer := httptest.NewServer(server)
	t.Cleanup(testServer.Close)

	status, responseBody := taxonomyHTTPRequest(t, testServer, http.MethodPut, 0, `{"prefLabel":{"en-US":"Concept","fr-FR":"Concept"},"altLabels":{"en-US":["Term"],"fr-FR":["Terme"]}}`)
	require.Equal(t, http.StatusCreated, status)
	current := decodeTaxonomyConceptResponse(t, responseBody)
	assert.Equal(t, map[string]string{"en-US": "Concept"}, current.PrefLabel)
	assert.Equal(t, map[string][]string{"en-US": {"Term"}}, current.AltLabels)

	for _, field := range []string{"altLabels", "hiddenLabels"} {
		status, responseBody = taxonomyHTTPRequest(t, testServer, http.MethodPatch, 1, `[{"op":"remove","path":"/`+field+`"}]`)
		assert.Equal(t, http.StatusUnprocessableEntity, status, field)
		assertTaxonomyLabelRemovalValidationError(t, responseBody, field)
		status, responseBody = taxonomyHTTPRequest(t, testServer, http.MethodGet, 0, "")
		require.Equal(t, http.StatusOK, status)
		current := decodeTaxonomyConceptResponse(t, responseBody)
		assert.Equal(t, 1, current.Sys.Version)
		assert.Equal(t, map[string][]string{"en-US": {"Term"}}, current.AltLabels)
		assert.Equal(t, map[string][]string{"en-US": {}}, current.HiddenLabels)
	}
}

func TestContentfulManagementServerAllowsTaxonomyLabelRemoveThenAdd(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)

	testServer := httptest.NewServer(server)
	t.Cleanup(testServer.Close)

	status, responseBody := taxonomyHTTPRequest(t, testServer, http.MethodPut, 0, `{"prefLabel":{"en-US":"Concept"},"altLabels":{"en-US":["Old alt"]},"hiddenLabels":{"en-US":["Old hidden"]}}`)
	require.Equal(t, http.StatusCreated, status)
	assert.Equal(t, 1, decodeTaxonomyConceptResponse(t, responseBody).Sys.Version)

	status, responseBody = taxonomyHTTPRequest(t, testServer, http.MethodPatch, 1, `[{"op":"remove","path":"/altLabels"},{"op":"add","path":"/altLabels","value":{"en-US":["New alt"]}}]`)
	require.Equal(t, http.StatusOK, status)
	current := decodeTaxonomyConceptResponse(t, responseBody)
	assert.Equal(t, 2, current.Sys.Version)
	assert.Equal(t, map[string][]string{"en-US": {"New alt"}}, current.AltLabels)
	assert.Equal(t, map[string][]string{"en-US": {"Old hidden"}}, current.HiddenLabels)

	status, responseBody = taxonomyHTTPRequest(t, testServer, http.MethodPatch, 2, `[{"op":"remove","path":"/hiddenLabels"},{"op":"add","path":"/hiddenLabels","value":{"en-US":["New hidden"]}}]`)
	require.Equal(t, http.StatusOK, status)
	current = decodeTaxonomyConceptResponse(t, responseBody)
	assert.Equal(t, 3, current.Sys.Version)
	assert.Equal(t, map[string][]string{"en-US": {"New alt"}}, current.AltLabels)
	assert.Equal(t, map[string][]string{"en-US": {"New hidden"}}, current.HiddenLabels)
}

type taxonomyConceptWireResponse struct {
	Sys struct {
		Version int `json:"version"`
	} `json:"sys"`
	PrefLabel    map[string]string   `json:"prefLabel"`
	AltLabels    map[string][]string `json:"altLabels"`
	HiddenLabels map[string][]string `json:"hiddenLabels"`
}

type taxonomyLabelValidationErrorWireResponse struct {
	Sys struct {
		ID string `json:"id"`
	} `json:"sys"`
	Message string `json:"message"`
	Details struct {
		Flatten struct {
			FormErrors  []string            `json:"formErrors"`
			FieldErrors map[string][]string `json:"fieldErrors"`
		} `json:"flatten"`
		Errors []struct {
			Name    string   `json:"name"`
			Path    []string `json:"path"`
			Details string   `json:"details"`
		} `json:"errors"`
	} `json:"details"`
}

func decodeTaxonomyConceptResponse(t *testing.T, body []byte) taxonomyConceptWireResponse {
	t.Helper()

	var response taxonomyConceptWireResponse
	require.NoError(t, json.Unmarshal(body, &response))

	return response
}

func assertTaxonomyLabelValidationError(t *testing.T, body []byte, field, locale string) {
	t.Helper()

	const detail = "Invalid input: expected array, received undefined"

	var response taxonomyLabelValidationErrorWireResponse
	require.NoError(t, json.Unmarshal(body, &response))
	assert.Equal(t, "ValidationFailed", response.Sys.ID)
	assert.Equal(t, "Validation error", response.Message)
	assert.Empty(t, response.Details.Flatten.FormErrors)
	assert.Equal(t, map[string][]string{field: {detail}}, response.Details.Flatten.FieldErrors)
	require.Len(t, response.Details.Errors, 1)
	assert.Equal(t, "invalid_type", response.Details.Errors[0].Name)
	assert.Equal(t, []string{field, locale}, response.Details.Errors[0].Path)
	assert.Equal(t, detail, response.Details.Errors[0].Details)
}

func assertTaxonomyPrefLabelValidationError(t *testing.T, body []byte) {
	t.Helper()

	var response taxonomyLabelValidationErrorWireResponse
	require.NoError(t, json.Unmarshal(body, &response))
	assert.Equal(t, map[string][]string{"prefLabel": {"Invalid input: expected string, received undefined"}}, response.Details.Flatten.FieldErrors)
	require.Len(t, response.Details.Errors, 1)
	assert.Equal(t, []string{"prefLabel", "en-US"}, response.Details.Errors[0].Path)
}

func assertTaxonomyLabelRemovalValidationError(t *testing.T, body []byte, field string) {
	t.Helper()

	var response taxonomyLabelValidationErrorWireResponse
	require.NoError(t, json.Unmarshal(body, &response))
	assert.Equal(t, map[string][]string{field: {"Invalid input: expected object, received null"}}, response.Details.Flatten.FieldErrors)
	require.Len(t, response.Details.Errors, 1)
	assert.Equal(t, []string{field}, response.Details.Errors[0].Path)
}

func taxonomyHTTPRequest(t *testing.T, server *httptest.Server, method string, version int, body string) (int, []byte) {
	t.Helper()

	return taxonomyHTTPRequestAt(t, server, taxonomyConceptRequestPath, method, version, body)
}

func taxonomyHTTPRequestAt(t *testing.T, server *httptest.Server, requestPath, method string, version int, body string) (int, []byte) {
	t.Helper()

	versionHeader := ""
	if version != 0 {
		versionHeader = strconv.Itoa(version)
	}

	return taxonomyHTTPRequestAtWithVersionHeader(t, server, requestPath, method, versionHeader, body)
}

func taxonomyHTTPRequestAtWithVersionHeader(t *testing.T, server *httptest.Server, requestPath, method, versionHeader, body string) (int, []byte) {
	t.Helper()

	request, err := http.NewRequestWithContext(t.Context(), method, server.URL+requestPath, bytes.NewBufferString(body))
	require.NoError(t, err)
	request.Header.Set("Authorization", "Bearer "+cmt.ValidAccessToken)

	if method == http.MethodPatch {
		request.Header.Set("Content-Type", "application/json-patch+json")
	} else if body != "" {
		request.Header.Set("Content-Type", "application/vnd.contentful.management.v1+json")
	}

	if versionHeader != "" {
		request.Header.Set("X-Contentful-Version", versionHeader)
	}

	response, err := server.Client().Do(request)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, response.Body.Close())
	}()

	responseBody, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	return response.StatusCode, responseBody
}
