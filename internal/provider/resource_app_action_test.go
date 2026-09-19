package provider_test

import (
	"encoding/json"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"regexp"
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

const (
	actionAddress        = "contentful_app_action.test"
	actionCollectionPath = "/organizations/org/app_definitions/app/actions"
	actionPath           = actionCollectionPath + "/action"
	actionConfigPrefix   = `resource "contentful_app_action" "test" {
 organization_id = "org"
 app_definition_id = "app"
 name = "Action"
`
)

const actionLegacyConfig = actionConfigPrefix + `category = "Custom"
 type = "endpoint"
 url = "https://example.invalid/action"
 parameters = jsonencode([])
}`

const actionSchemaConfig = actionConfigPrefix + `category = "Custom"
 type = "endpoint"
 url = "https://example.invalid/action"
 parameters_schema = jsonencode({type="object", properties={message={type="string"}}})
}`

const actionBuiltinConfig = actionConfigPrefix + `category = "Entries.v1.0"
 type = "endpoint"
 url = "https://example.invalid/action"
}`

const (
	actionLegacyBody      = `{"name":"Action","category":"Custom","type":"endpoint","url":"https://example.invalid/action","parameters":[]}`
	actionSchemaBody      = `{"name":"Action","category":"Custom","type":"endpoint","url":"https://example.invalid/action","parametersSchema":{"type":"object","properties":{"message":{"type":"string"}}}}`
	actionBuiltinBody     = `{"name":"Action","category":"Entries.v1.0","type":"endpoint","url":"https://example.invalid/action"}`
	actionBuiltinResponse = `{"name":"Action","category":"Entries.v1.0","type":"endpoint","url":"https://example.invalid/action","parameters":[{"id":"entryIds","name":"Entry Ids","description":"Ids of the entries you want to trigger the action for","type":"Symbol","required":true}]}`
	actionResponseSys     = `"sys":{"id":"action","type":"AppAction","organization":{"sys":{"type":"Link","linkType":"Organization","id":"org"}},"appDefinition":{"sys":{"type":"Link","linkType":"AppDefinition","id":"app"}}}`
)

type actionMutationFixture struct{ config, request, response string }

// The service fixture is independent of provider request conversion and the
// generated server: each transition has an authored wire request and response.
func testActionLifecycle(t *testing.T, cases []actionMutationFixture) {
	t.Helper()

	var requestMutex sync.Mutex

	next := 0
	current := ""
	deleted := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestMutex.Lock()
		defer requestMutex.Unlock()

		assert.Empty(t, r.Header.Get("X-Contentful-Version"))
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodPost, http.MethodPut:
			if next >= len(cases) {
				t.Errorf("unexpected mutation %s", r.Method)
				w.WriteHeader(http.StatusInternalServerError)

				return
			}

			expectedMethod := http.MethodPut
			expectedPath := actionPath

			if next == 0 {
				expectedMethod = http.MethodPost
				expectedPath = actionCollectionPath
			}

			assert.Equal(t, expectedMethod, r.Method)
			assert.Equal(t, expectedPath, r.URL.Path)
			raw, err := io.ReadAll(r.Body)
			assert.NoError(t, err)
			assert.JSONEq(t, cases[next].request, string(raw))
			assert.Equal(t, "application/vnd.contentful.management.v1+json", r.Header.Get("Content-Type"))

			current = cases[next].response
			next++

			if r.Method == http.MethodPost {
				w.WriteHeader(http.StatusCreated)
			}

			_, err = w.Write([]byte("{" + actionResponseSys + "," + strings.TrimPrefix(current, "{")))
			assert.NoError(t, err)
		case http.MethodGet:
			assert.Equal(t, actionPath, r.URL.Path)

			if deleted {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"sys":{"type":"Error","id":"NotFound"}}`))

				return
			}

			_, err := w.Write([]byte("{" + actionResponseSys + "," + strings.TrimPrefix(current, "{")))
			assert.NoError(t, err)
		case http.MethodDelete:
			assert.Equal(t, actionPath, r.URL.Path)

			deleted = true

			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected method %s", r.Method)
		}
	})

	var steps []resource.TestStep
	for index, test := range cases {
		steps = append(steps, resource.TestStep{Config: test.config, ConfigPlanChecks: resource.ConfigPlanChecks{PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}, ConfigStateChecks: []statecheck.StateCheck{statecheck.ExpectKnownValue(actionAddress, tfjsonpath.New("id"), knownvalue.StringExact("org/app/action"))}})
		if index == 0 {
			steps = append(steps, resource.TestStep{ResourceName: actionAddress, ImportState: true, ImportStateId: "org/app/action", ImportStateVerify: true})
			steps = append(steps, resource.TestStep{Config: test.config, ResourceName: actionAddress, ImportState: true, ImportStateKind: resource.ImportBlockWithResourceIdentity})
		}
	}

	testAccMockedResource(t, handler, resource.TestCase{Steps: steps})
	requestMutex.Lock()
	defer requestMutex.Unlock()

	assert.Equal(t, len(cases), next)
	assert.True(t, deleted)
}

func TestAccAppActionResourceTransitions(t *testing.T) {
	t.Parallel()

	fullConfig := strings.Replace(actionSchemaConfig, "\n}", "\n description = \"Description\"\n result_schema = jsonencode({type=\"object\"})\n}", 1)
	fullBody := strings.TrimSuffix(actionSchemaBody, "}") + `,"description":"Description","resultSchema":{"type":"object"}}`
	functionConfig := actionConfigPrefix + `category="Custom"
 type="function-invocation"
 function_id="function"
 parameters=jsonencode([{id="enabled",name="Enabled",type="Boolean",required=false,default=false}])
}`
	functionBody := `{"name":"Action","category":"Custom","type":"function-invocation","function":{"sys":{"type":"Link","linkType":"Function","id":"function"}},"parameters":[{"id":"enabled","name":"Enabled","type":"Boolean","required":false,"default":false}]}`
	testActionLifecycle(t, []actionMutationFixture{
		{actionLegacyConfig, actionLegacyBody, actionLegacyBody},
		{fullConfig, fullBody, fullBody},
		{actionSchemaConfig, actionSchemaBody, actionSchemaBody},
		{functionConfig, functionBody, functionBody},
		{actionBuiltinConfig, actionBuiltinBody, actionBuiltinResponse},
		{actionSchemaConfig, actionSchemaBody, actionSchemaBody},
		{actionLegacyConfig, actionLegacyBody, actionLegacyBody},
	})
}

func TestAccAppActionResourceBuiltinSchema(t *testing.T) {
	t.Parallel()

	config := strings.Replace(actionBuiltinConfig, "\n}", "\n parameters_schema=jsonencode({type=\"object\"})\n}", 1)
	body := strings.TrimSuffix(actionBuiltinBody, "}") + `,"parametersSchema":{"type":"object"}}`
	response := strings.TrimSuffix(actionBuiltinResponse, "}") + `,"parametersSchema":{"type":"object"}}`
	updatedConfig := strings.Replace(config, `name = "Action"`, `name = "Updated"`, 1)
	testActionLifecycle(t, []actionMutationFixture{{config, body, response}, {updatedConfig, strings.Replace(body, `"Action"`, `"Updated"`, 1), strings.Replace(response, `"Action"`, `"Updated"`, 1)}})
}

func TestAccAppActionResourceIgnoreChanges(t *testing.T) {
	t.Parallel()

	config := strings.Replace(actionLegacyConfig, "\n}", "\n lifecycle { ignore_changes = [parameters] }\n}", 1)
	update := strings.Replace(strings.Replace(config, `name = "Action"`, `name = "Updated"`, 1), `jsonencode([])`, `jsonencode([{id="x", name="X", type="Symbol"}])`, 1)
	body := strings.Replace(actionLegacyBody, `"Action"`, `"Updated"`, 1)
	testActionLifecycle(t, []actionMutationFixture{{config, actionLegacyBody, actionLegacyBody}, {update, body, body}})
}

func TestAccAppActionResourceInvalidConfig(t *testing.T) {
	t.Parallel()

	for _, test := range []struct{ name, config, pattern string }{
		{"missing input", strings.Replace(actionLegacyConfig, " parameters = jsonencode([])\n", "", 1), "Invalid Custom App Action input"},
		{"both inputs", strings.Replace(actionLegacyConfig, "\n}", "\n parameters_schema=jsonencode({})\n}", 1), "Invalid Custom App Action input"},
		{"builtin parameters", strings.Replace(actionLegacyConfig, `"Custom"`, `"Entries.v1.0"`, 1), "Built-in App Action parameters are read-only"},
		{"both executors", strings.Replace(actionLegacyConfig, "\n}", "\n function_id=\"fn\"\n}", 1), "Invalid App Action configuration"},
		{"HTTP URL", strings.Replace(actionLegacyConfig, "https://", "http://", 1), "Invalid App Action URL"},
		{"null schema", strings.Replace(actionLegacyConfig, "\n}", "\n result_schema=\"null\"\n}", 1), "Invalid App Action JSON"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			testAccMockedResource(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				t.Error("unexpected request")
				w.WriteHeader(http.StatusInternalServerError)
			}), resource.TestCase{Steps: []resource.TestStep{{Config: test.config, PlanOnly: true, ExpectError: regexp.MustCompile(test.pattern)}}})
		})
	}
}

func TestAccAppActionDataSources(t *testing.T) {
	t.Parallel()

	const (
		functionBody = `{"name":"Function Action","category":"Future.v2.0","type":"function-invocation","description":"A full definition","function":{"sys":{"type":"Link","linkType":"Function","id":"function"}},"parameters":[],"parametersSchema":{"type":"object"},"resultSchema":{"type":"string"}}`
		functionID   = "z-function"
	)

	functionResponse := strings.Replace("{"+actionResponseSys+","+strings.TrimPrefix(functionBody, "{"), `"id":"action"`, `"id":"z-function"`, 1)
	endpointResponse := "{" + actionResponseSys + "," + strings.TrimPrefix(actionBuiltinResponse, "{")

	var requestMutex sync.Mutex

	skips := map[string]bool{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		assert.Equal(t, http.MethodGet, r.Method)

		if r.URL.Path == actionCollectionPath+"/"+functionID {
			_, _ = w.Write([]byte(functionResponse))

			return
		}

		assert.Equal(t, actionCollectionPath, r.URL.Path)
		assert.Equal(t, "100", r.URL.Query().Get("limit"))
		skip := r.URL.Query().Get("skip")

		requestMutex.Lock()
		skips[skip] = true
		requestMutex.Unlock()

		action := json.RawMessage(functionResponse)

		page := map[string]any{"sys": map[string]any{"type": "Array"}, "total": 2, "skip": 0, "limit": 1, "items": []json.RawMessage{action}}
		if skip == "1" {
			page["skip"] = 1
			page["items"] = []json.RawMessage{json.RawMessage(endpointResponse)}
		} else {
			assert.Equal(t, "0", skip)
		}

		assert.NoError(t, json.NewEncoder(w).Encode(page))
	})
	functionFields := map[string]knownvalue.Check{
		"app_action_id":     knownvalue.StringExact("z-function"),
		"name":              knownvalue.StringExact("Function Action"),
		"category":          knownvalue.StringExact("Future.v2.0"),
		"type":              knownvalue.StringExact("function-invocation"),
		"description":       knownvalue.StringExact("A full definition"),
		"url":               knownvalue.Null(),
		"function_id":       knownvalue.StringExact("function"),
		"parameters":        knownvalue.StringExact("[]"),
		"parameters_schema": knownvalue.StringExact(`{"type":"object"}`),
		"result_schema":     knownvalue.StringExact(`{"type":"string"}`),
	}

	checks := make([]statecheck.StateCheck, 0, 3+len(functionFields))

	checks = append(checks,
		statecheck.ExpectKnownValue("data.contentful_app_action.one", tfjsonpath.New("id"), knownvalue.StringExact("org/app/z-function")),
		statecheck.ExpectKnownValue("data.contentful_app_actions.all", tfjsonpath.New("id"), knownvalue.StringExact("org/app")),
		statecheck.ExpectKnownValue("data.contentful_app_actions.all", tfjsonpath.New("app_actions"), knownvalue.ListExact([]knownvalue.Check{
			knownvalue.ObjectExact(map[string]knownvalue.Check{
				"app_action_id": knownvalue.StringExact("action"), "name": knownvalue.StringExact("Action"), "category": knownvalue.StringExact("Entries.v1.0"), "type": knownvalue.StringExact("endpoint"),
				"description": knownvalue.Null(), "url": knownvalue.StringExact("https://example.invalid/action"), "function_id": knownvalue.Null(),
				"parameters":        knownvalue.StringExact(`[{"description":"Ids of the entries you want to trigger the action for","id":"entryIds","name":"Entry Ids","required":true,"type":"Symbol"}]`),
				"parameters_schema": knownvalue.Null(), "result_schema": knownvalue.Null(),
			}),
			knownvalue.ObjectExact(functionFields),
		})),
	)
	for name, expected := range functionFields {
		checks = append(checks, statecheck.ExpectKnownValue("data.contentful_app_action.one", tfjsonpath.New(name), expected))
	}

	testAccMockedResource(t, handler, resource.TestCase{Steps: []resource.TestStep{{Config: `data "contentful_app_action" "one" {
 organization_id="org"
 app_definition_id="app"
 app_action_id="z-function"
}
data "contentful_app_actions" "all" {
 organization_id="org"
 app_definition_id="app"
}`, ConfigStateChecks: checks}}})
	requestMutex.Lock()
	defer requestMutex.Unlock()

	require.True(t, skips["0"])
	require.True(t, skips["1"])
}

func TestAccAppActionResourceIgnoreCategory(t *testing.T) {
	t.Parallel()

	config := strings.Replace(actionBuiltinConfig, "\n}", "\n lifecycle { ignore_changes = [category] }\n}", 1)
	update := strings.Replace(strings.Replace(config, `"Entries.v1.0"`, `"Custom"`, 1), `name = "Action"`, `name = "Updated"`, 1)
	update = strings.Replace(update, "\n}", "\n parameters_schema=jsonencode({type=\"object\"})\n}", 1)
	body := strings.TrimSuffix(strings.Replace(actionBuiltinBody, `"Action"`, `"Updated"`, 1), "}") + `,"parametersSchema":{"type":"object"}}`
	response := strings.TrimSuffix(strings.Replace(actionBuiltinResponse, `"Action"`, `"Updated"`, 1), "}") + `,"parametersSchema":{"type":"object"}}`
	testActionLifecycle(t, []actionMutationFixture{{config, actionBuiltinBody, actionBuiltinResponse}, {update, body, response}})
}

func TestAccAppActionResourceDisappears(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.SetAppDefinition("org", "app", cm.AppDefinitionData{Name: "App"})
	testAccMockedResource(t, server, resource.TestCase{Steps: []resource.TestStep{
		{Config: actionLegacyConfig},
		{PreConfig: func() {
			response, err := server.Handler().GetAppActions(t.Context(), cm.GetAppActionsParams{OrganizationID: "org", AppDefinitionID: "app"})
			require.NoError(t, err)

			actions, ok := response.(*cm.AppActionCollection)
			require.True(t, ok)
			require.Len(t, actions.Items, 1)
			deleted, err := server.Handler().DeleteAppAction(t.Context(), cm.DeleteAppActionParams{OrganizationID: "org", AppDefinitionID: "app", AppActionID: actions.Items[0].Sys.ID})
			require.NoError(t, err)
			require.IsType(t, &cm.NoContent{}, deleted)
		}, Config: actionLegacyConfig, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(actionAddress, plancheck.ResourceActionCreate)}}},
	}})
}

func TestAccAppActionResourceIgnoreBuiltinParameters(t *testing.T) {
	t.Parallel()

	config := strings.Replace(actionBuiltinConfig, "\n}", "\n lifecycle { ignore_changes = [category, parameters] }\n}", 1)
	update := strings.Replace(strings.Replace(config, `"Entries.v1.0"`, `"Custom"`, 1), `name = "Action"`, `name = "Updated"`, 1)
	update = strings.Replace(update, "\n}", "\n parameters=jsonencode([])\n}", 1)
	body := strings.Replace(actionBuiltinBody, `"Action"`, `"Updated"`, 1)
	response := strings.Replace(actionBuiltinResponse, `"Action"`, `"Updated"`, 1)
	testActionLifecycle(t, []actionMutationFixture{{config, actionBuiltinBody, actionBuiltinResponse}, {update, body, response}})
}

func TestAccAppActionResourceTimeoutOnly(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, config string
		drift        bool
	}{
		{"custom", actionSchemaConfig, false},
		{"builtin", actionBuiltinConfig, false},
		{"ignored drift", strings.Replace(actionSchemaConfig, "\n}", "\n lifecycle { ignore_changes = [url] }\n}", 1), true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
			require.NoError(t, err)
			server.SetAppDefinition("org", "app", cm.AppDefinitionData{Name: "App"})

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.NotEqual(t, http.MethodPut, r.Method, "timeout-only updates must not write to Contentful")

				if test.drift && r.Method == http.MethodGet {
					response := httptest.NewRecorder()
					server.ServeHTTP(response, r)
					maps.Copy(w.Header(), response.Header())
					w.WriteHeader(response.Code)
					_, _ = w.Write([]byte(strings.ReplaceAll(response.Body.String(), "https://", "http://")))

					return
				}

				server.ServeHTTP(w, r)
			})
			update := strings.Replace(test.config, "\n}", "\n timeouts = { update=\"3m\" }\n}", 1)
			testAccMockedResource(t, handler, resource.TestCase{Steps: []resource.TestStep{
				{Config: test.config},
				{Config: update, ConfigPlanChecks: resource.ConfigPlanChecks{PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}, ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(actionAddress, tfjsonpath.New("timeouts").AtMapKey("update"), knownvalue.StringExact("3m")),
				}},
			}})
		})
	}
}

func TestAccAppActionResourceScopeReplacement(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)

	scopes := []struct{ organization, definition string }{
		{"org", "app"}, {"org", "other-app"}, {"other-org", "third-app"},
	}

	steps := make([]resource.TestStep, 0, len(scopes))
	for i, scope := range scopes {
		server.SetAppDefinition(scope.organization, scope.definition, cm.AppDefinitionData{Name: "App"})
		config := strings.Replace(strings.Replace(actionLegacyConfig, `organization_id = "org"`, `organization_id = "`+scope.organization+`"`, 1), `app_definition_id = "app"`, `app_definition_id = "`+scope.definition+`"`, 1)

		step := resource.TestStep{Config: config, ConfigStateChecks: []statecheck.StateCheck{
			statecheck.ExpectKnownValue(actionAddress, tfjsonpath.New("id"), knownvalue.StringRegexp(regexp.MustCompile("^"+scope.organization+"/"+scope.definition+"/[^/]+$"))),
		}}
		if i > 0 {
			step.ConfigPlanChecks.PreApply = []plancheck.PlanCheck{plancheck.ExpectResourceAction(actionAddress, plancheck.ResourceActionDestroyBeforeCreate)}
		}

		steps = append(steps, step)
		if i == 0 {
			// An organization-only change must plan replacement even though an
			// App Definition cannot belong to both organizations during apply.
			steps = append(steps, resource.TestStep{
				Config:   strings.Replace(config, `organization_id = "org"`, `organization_id = "other-org"`, 1),
				PlanOnly: true, ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectResourceAction(actionAddress, plancheck.ResourceActionDestroyBeforeCreate)}},
			})
		}
	}

	testAccMockedResource(t, server, resource.TestCase{Steps: steps})

	for _, scope := range scopes {
		response, err := server.Handler().GetAppActions(t.Context(), cm.GetAppActionsParams{OrganizationID: scope.organization, AppDefinitionID: scope.definition})
		require.NoError(t, err)

		actions, ok := response.(*cm.AppActionCollection)
		require.True(t, ok)
		assert.Empty(t, actions.Items, "replacement and final destroy must delete every action")
	}
}
