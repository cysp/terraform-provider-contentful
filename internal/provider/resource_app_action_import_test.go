package provider_test

import (
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
	result, err := server.Handler().CreateAppAction(t.Context(), &cm.AppActionData{Name: "Action", Category: "Custom", Type: "endpoint", URL: cm.NewOptString("https://example.invalid/action"), Parameters: []byte(`[]`)}, cm.CreateAppActionParams{OrganizationID: "org", AppDefinitionID: "app"})
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
		statecheck.ExpectKnownValue(actionAddress, tfjsonpath.New("parameters"), knownvalue.StringExact("[]")),
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

func TestAccAppActionResourceExternalDefinitions(t *testing.T) {
	t.Parallel()

	const (
		nested  = `{"type":"object","properties":{"items":{"type":"array","items":{"type":"object","properties":{"enabled":{"type":"boolean","default":false},"count":{"type":"number","minimum":0}},"required":["count"],"additionalProperties":false}}},"required":["items"],"additionalProperties":false}`
		result  = `{"type":"object","properties":{"count":{"type":"integer"},"warnings":{"type":"array","items":{"type":"string"}}},"required":["count"]}`
		legacy  = `[{"id":"text","name":"Text","type":"Symbol","description":"","default":""},{"id":"number","name":"Number","type":"Number","required":false,"default":0},{"id":"flag","name":"Flag","type":"Boolean","default":false},{"id":"choice","name":"Choice","type":"Enum","options":["a","b"]}]`
		builtin = `[{"id":"entryIds","name":"Entry Ids","description":"Ids of the entries you want to trigger the action for","type":"Symbol","required":true}]`
	)
	for _, test := range []struct{ name, category, parameters, schema string }{
		{"legacy", "Custom", legacy, ""},
		{"schema", "Custom", "", nested},
		{"builtin", "Entries.v1.0", builtin, nested},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var requestMutex sync.Mutex

			renamed, deleted := false, false

			var mutations []string

			writeFields := fmt.Sprintf(`,"category":%q,"type":"endpoint","url":"https://example.invalid/external","description":""`, test.category)

			fields := fmt.Sprintf(" category=%q\n type=\"endpoint\"\n url=\"https://example.invalid/external\"\n description=\"\"\n", test.category)
			if test.parameters != "" && test.category == "Custom" {
				writeFields += `,"parameters":` + test.parameters
				fields += fmt.Sprintf(" parameters=%q\n", test.parameters)
			}

			if test.schema != "" {
				writeFields += `,"parametersSchema":` + test.schema
				fields += fmt.Sprintf(" parameters_schema=%q\n", test.schema)
			}

			writeFields += `,"resultSchema":` + result
			fields += fmt.Sprintf(" result_schema=%q\n", result)

			responseFields := writeFields
			if test.category != "Custom" {
				responseFields += `,"parameters":` + test.parameters
			}

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
					assert.JSONEq(t, `{"name":"Renamed"`+writeFields+`}`, string(body))

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
					_, _ = io.WriteString(w, `{"sys":{"type":"Error","id":"NotFound"}}`)

					return
				}

				name := "Action"
				if renamed {
					name = "Renamed"
				}

				_, _ = fmt.Fprintf(w, `{"sys":{"id":"external","type":"AppAction","organization":{"sys":{"type":"Link","linkType":"Organization","id":"org"}},"appDefinition":{"sys":{"type":"Link","linkType":"AppDefinition","id":"app"}}},"name":%q%s}`, name, responseFields)
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
