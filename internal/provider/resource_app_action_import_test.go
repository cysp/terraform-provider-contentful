package provider_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccAppActionResourceIdentityImport(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.SetAppDefinition("org", "app", cm.AppDefinitionData{Name: "App"})
	result, err := server.Handler().CreateAppAction(t.Context(), &cm.AppActionCreateData{Name: "Action", Category: "Custom", Type: "endpoint", URL: cm.NewOptString("https://example.invalid/action"), Parameters: []byte(testJSON([]any{}))}, cm.CreateAppActionParams{OrganizationID: "org", AppDefinitionID: "app"})
	require.NoError(t, err)

	action, ok := result.(*cm.AppAction)
	require.True(t, ok)

	imported := actionLegacyConfig + fmt.Sprintf(`
 import {
  to = contentful_app_action.test
  identity = {
   organization_id = "org"
   app_definition_id = "app"
   app_action_id = %q
  }
 }
 `, action.Sys.ID)

	var (
		requestMutex sync.Mutex
		requests     []string
	)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			requestMutex.Lock()

			requests = append(requests, r.Method+" "+r.URL.Path)
			requestMutex.Unlock()
		}

		server.ServeHTTP(w, r)
	})
	checks := []statecheck.StateCheck{
		statecheck.ExpectKnownValue(actionAddress, tfjsonpath.New("id"), knownvalue.StringExact("org/app/"+action.Sys.ID)),
		statecheck.ExpectKnownValue(actionAddress, tfjsonpath.New("name"), knownvalue.StringExact("Action")),
		statecheck.ExpectKnownValue(actionAddress, tfjsonpath.New("parameters"), knownvalue.StringExact(testJSON([]any{}))),
		statecheck.ExpectIdentity(actionAddress, map[string]knownvalue.Check{"organization_id": knownvalue.StringExact("org"), "app_definition_id": knownvalue.StringExact("app"), "app_action_id": knownvalue.StringExact(action.Sys.ID)}),
	}
	testAccMockedResource(t, handler, resource.TestCase{Steps: []resource.TestStep{
		{Config: imported, ConfigStateChecks: checks},
		{Config: actionLegacyConfig, ConfigStateChecks: checks, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}},
	}})
	requestMutex.Lock()
	defer requestMutex.Unlock()

	assert.Equal(t, []string{"DELETE " + actionCollectionPath + "/" + action.Sys.ID}, requests)
	readResult, err := server.Handler().GetAppAction(t.Context(), cm.GetAppActionParams{OrganizationID: "org", AppDefinitionID: "app", AppActionID: action.Sys.ID})
	require.NoError(t, err)

	status, ok := readResult.(cm.StatusCodeResponse)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, status.GetStatusCode())
}

//nolint:dupl,maintidx // Keep lifecycle cases and independent payload expectations together.
func TestAccAppActionResourceExternalDefinitions(t *testing.T) {
	t.Parallel()

	var (
		nested = testJSON(map[string]any{
			"type": "object",
			"properties": map[string]any{
				"items": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"enabled": map[string]any{"type": "boolean", "default": false},
							"count":   map[string]any{"type": "number", "minimum": 0},
						},
						"required":             []any{"count"},
						"additionalProperties": false,
					},
				},
			},
			"required":             []any{"items"},
			"additionalProperties": false,
		})
		result = testJSON(map[string]any{
			"type": "object",
			"properties": map[string]any{
				"count":    map[string]any{"type": "integer"},
				"warnings": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			},
			"required": []any{"count"},
		})
		legacy = testJSON([]any{
			map[string]any{"id": "text", "name": "Text", "type": "Symbol", "description": "", "default": ""},
			map[string]any{"id": "number", "name": "Number", "type": "Number", "required": false, "default": 0},
			map[string]any{"id": "flag", "name": "Flag", "type": "Boolean", "default": false},
			map[string]any{"id": "choice", "name": "Choice", "type": "Enum", "options": []any{"a", "b"}},
		})
		builtin = testJSON([]any{
			map[string]any{
				"id":          "entryIds",
				"name":        "Entry Ids",
				"description": "Ids of the entries you want to trigger the action for",
				"type":        "Symbol",
				"required":    true,
			},
		})
	)
	for _, test := range []struct{ name, category, parameters, schema, request string }{
		{"legacy", "Custom", legacy, "", testJSON(map[string]any{
			"name":        "Renamed",
			"category":    "Custom",
			"type":        "endpoint",
			"url":         "https://example.invalid/external",
			"description": "",
			"parameters": []any{
				map[string]any{"id": "text", "name": "Text", "type": "Symbol", "description": "", "default": ""},
				map[string]any{"id": "number", "name": "Number", "type": "Number", "required": false, "default": 0},
				map[string]any{"id": "flag", "name": "Flag", "type": "Boolean", "default": false},
				map[string]any{"id": "choice", "name": "Choice", "type": "Enum", "options": []any{"a", "b"}},
			},
			"resultSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"count":    map[string]any{"type": "integer"},
					"warnings": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
				},
				"required": []any{"count"},
			},
		})},
		{"schema", "Custom", "", nested, testJSON(map[string]any{
			"name":        "Renamed",
			"category":    "Custom",
			"type":        "endpoint",
			"url":         "https://example.invalid/external",
			"description": "",
			"parametersSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"items": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"enabled": map[string]any{"type": "boolean", "default": false},
								"count":   map[string]any{"type": "number", "minimum": 0},
							},
							"required":             []any{"count"},
							"additionalProperties": false,
						},
					},
				},
				"required":             []any{"items"},
				"additionalProperties": false,
			},
			"resultSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"count":    map[string]any{"type": "integer"},
					"warnings": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
				},
				"required": []any{"count"},
			},
		})},
		{"builtin", "Entries.v1.0", builtin, nested, testJSON(map[string]any{
			"name":        "Renamed",
			"category":    "Entries.v1.0",
			"type":        "endpoint",
			"url":         "https://example.invalid/external",
			"description": "",
			"parametersSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"items": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"enabled": map[string]any{"type": "boolean", "default": false},
								"count":   map[string]any{"type": "number", "minimum": 0},
							},
							"required":             []any{"count"},
							"additionalProperties": false,
						},
					},
				},
				"required":             []any{"items"},
				"additionalProperties": false,
			},
			"resultSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"count":    map[string]any{"type": "integer"},
					"warnings": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
				},
				"required": []any{"count"},
			},
		})},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var requestMutex sync.Mutex

			renamed, deleted := false, false

			var mutations []string

			fields := fmt.Sprintf(" category=%q\n type=\"endpoint\"\n url=\"https://example.invalid/external\"\n description=\"\"\n", test.category)
			if test.parameters != "" && test.category == "Custom" {
				fields += fmt.Sprintf(" parameters=%q\n", test.parameters)
			}

			if test.schema != "" {
				fields += fmt.Sprintf(" parameters_schema=%q\n", test.schema)
			}

			fields += fmt.Sprintf(" result_schema=%q\n", result)

			response := cm.AppAction{
				Sys: cm.AppActionSys{
					ID: "external", Type: cm.AppActionSysTypeAppAction,
					Organization: cm.NewOrganizationLink("org"), AppDefinition: cm.NewAppDefinitionLink("app"),
				},
				Name: "Action", Category: test.category, Type: "endpoint",
				URL: cm.NewOptString("https://example.invalid/external"), Description: cm.NewOptString(""),
				Parameters: []byte(test.parameters), ParametersSchema: []byte(test.schema), ResultSchema: []byte(result),
			}
			notFound := cmt.NewContentfulManagementError(cm.ErrorSysIDNotFound, nil, nil)

			configuration := actionConfigPrefix + fields + "}"
			imported := configuration + `
import {
 to=contentful_app_action.test
 id="org/app/external"
}`
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestMutex.Lock()
				defer requestMutex.Unlock()

				assert.Equal(t, actionCollectionPath+"/external", r.URL.Path)
				assert.Empty(t, r.Header.Get("X-Contentful-Version"))
				w.Header().Set("Content-Type", "application/json")

				if r.Method != http.MethodGet {
					mutations = append(mutations, r.Method)
				}

				switch r.Method {
				case http.MethodPut:
					body, err := io.ReadAll(r.Body)
					assert.NoError(t, err)
					assert.JSONEq(t, test.request, string(body))

					renamed = true
				case http.MethodDelete:
					deleted = true

					w.WriteHeader(http.StatusNoContent)

					return
				case http.MethodGet:
				default:
					t.Errorf("unexpected method %s", r.Method)
				}

				if deleted {
					w.WriteHeader(http.StatusNotFound)
					assert.NoError(t, json.NewEncoder(w).Encode(&notFound))

					return
				}

				if renamed {
					response.Name = "Renamed"
				}

				assert.NoError(t, json.NewEncoder(w).Encode(&response))
			})
			changed := strings.Replace(configuration, `name = "Action"`, `name = "Renamed"`, 1)

			checks := []statecheck.StateCheck{statecheck.ExpectKnownValue(actionAddress, tfjsonpath.New("app_action_id"), knownvalue.StringExact("external")), statecheck.ExpectKnownValue(actionAddress, tfjsonpath.New("description"), knownvalue.StringExact(""))}
			if test.category != "Custom" {
				checks = append(checks, statecheck.ExpectKnownValue(actionAddress, tfjsonpath.New("parameters"), knownvalue.Null()))
			}

			testAccMockedResource(t, handler, resource.TestCase{Steps: []resource.TestStep{
				{Config: imported, ConfigStateChecks: checks},
				{Config: configuration, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}},
				{Config: changed, ConfigStateChecks: checks, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(actionAddress, plancheck.ResourceActionUpdate)}}},
			}})
			requestMutex.Lock()
			defer requestMutex.Unlock()

			assert.Equal(t, []string{http.MethodPut, http.MethodDelete}, mutations)
			assert.True(t, deleted)
		})
	}
}
