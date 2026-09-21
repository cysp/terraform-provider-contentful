package provider_test

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:paralleltest // Live acceptance serializes access to the shared account.
func TestAccAppActionResourceLifecycle(t *testing.T) {
	parallelWhenMocked(t)

	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC must be set for acceptance tests")
	}

	const organization = "2zuSjSO4A0e6GKBrhJRe2m"

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)

	endpoint, token := cm.DefaultServerURL, os.Getenv("CONTENTFUL_MANAGEMENT_ACCESS_TOKEN")
	if os.Getenv("TF_ACC_MOCKED") != "" {
		inspectionServer := httptest.NewServer(server)
		t.Cleanup(inspectionServer.Close)
		endpoint, token = inspectionServer.URL, cmt.ValidAccessToken
	}

	client, err := cm.NewClient(endpoint, cm.NewAccessTokenSecuritySource(token), cm.WithClient(cm.NewTransportClient(&http.Client{Timeout: time.Minute}, cm.DefaultUserAgent)))
	require.NoError(t, err)

	initial := config.Variables{"organization_id": config.StringVariable(organization)}
	updated := maps.Clone(initial)
	updated["custom_input"] = config.ObjectVariable(map[string]config.Variable{"parameters": config.StringVariable("[]")})
	updated["builtin_category"] = config.StringVariable("Notification.v1.0")
	updated["function_target"] = config.ObjectVariable(map[string]config.Variable{
		"type": config.StringVariable("endpoint"), "url": config.StringVariable("https://example.invalid/function-transition"),
	})
	restored := maps.Clone(updated)
	delete(restored, "function_target")

	custom := map[string]knownvalue.Check{
		"name":        knownvalue.StringExact("Action"),
		"category":    knownvalue.StringExact("Custom"),
		"type":        knownvalue.StringExact("endpoint"),
		"url":         knownvalue.StringExact("https://example.invalid/terraform-action"),
		"function_id": knownvalue.Null(),
		"description": knownvalue.StringExact("Temporary acceptance action"),
		"parameters":  knownvalue.Null(),
		"parameters_schema": knownvalue.StringFunc(func(actual string) error {
			return checkJSONEqual(`{"type":"object","properties":{"message":{"type":"string"}}}`, actual)
		}),
		"result_schema": knownvalue.StringFunc(func(actual string) error { return checkJSONEqual(`{"type":"object"}`, actual) }),
	}
	customUpdated := maps.Clone(custom)
	customUpdated["description"] = knownvalue.Null()
	customUpdated["parameters"] = knownvalue.StringExact("[]")
	customUpdated["parameters_schema"] = knownvalue.Null()
	customUpdated["result_schema"] = knownvalue.Null()

	// Category-owned display text may change independently of the input contract.
	builtinParameters := func(ids ...string) knownvalue.Check {
		checks := make([]knownvalue.Check, 0, len(ids))
		for _, id := range ids {
			checks = append(checks, knownvalue.ObjectPartial(map[string]knownvalue.Check{
				"id": knownvalue.StringExact(id), "type": knownvalue.StringExact("Symbol"), "required": knownvalue.Bool(true),
			}))
		}

		return knownvalue.StringFunc(func(actual string) error {
			var parameters any

			err := json.Unmarshal([]byte(actual), &parameters)
			if err != nil {
				return fmt.Errorf("decode App Action parameters: %w", err)
			}

			return knownvalue.SetExact(checks).CheckValue(parameters)
		})
	}
	builtin := map[string]knownvalue.Check{
		"name":              knownvalue.StringExact("Builtin"),
		"category":          knownvalue.StringExact("Entries.v1.0"),
		"type":              knownvalue.StringExact("endpoint"),
		"url":               knownvalue.StringExact("https://example.invalid/terraform-action"),
		"function_id":       knownvalue.Null(),
		"parameters":        builtinParameters("entryIds"),
		"parameters_schema": knownvalue.Null(),
		"result_schema":     knownvalue.Null(),
	}
	builtinUpdated := maps.Clone(builtin)
	builtinUpdated["category"] = knownvalue.StringExact("Notification.v1.0")
	builtinUpdated["parameters"] = builtinParameters("message", "recipient")

	function := map[string]knownvalue.Check{
		"name":        knownvalue.StringExact("Function action"),
		"category":    knownvalue.StringExact("Custom"),
		"type":        knownvalue.StringExact("function-invocation"),
		"url":         knownvalue.Null(),
		"function_id": knownvalue.StringExact("acceptancefunction"),
		"description": knownvalue.StringExact(""),
		"parameters": knownvalue.StringFunc(func(actual string) error {
			return checkJSONEqual(`[{"id":"text","name":"Text","type":"Symbol"},{"id":"number","name":"Number","type":"Number","required":false,"default":0},{"id":"flag","name":"Flag","type":"Boolean","default":false},{"id":"choice","name":"Choice","type":"Enum","options":["a","b"]}]`, actual)
		}),
		"parameters_schema": knownvalue.Null(),
		"result_schema":     knownvalue.Null(),
	}
	functionUpdated := maps.Clone(function)
	functionUpdated["type"] = knownvalue.StringExact("endpoint")
	functionUpdated["url"] = knownvalue.StringExact("https://example.invalid/function-transition")
	functionUpdated["function_id"] = knownvalue.Null()

	discoveryChecks := func(custom, builtin, function map[string]knownvalue.Check) []statecheck.StateCheck {
		checks := make([]statecheck.StateCheck, 0, 1+len(custom))
		checks = append(checks,
			statecheck.ExpectKnownValue("data.contentful_app_actions.all", tfjsonpath.New("app_actions"), knownvalue.SetExact([]knownvalue.Check{
				knownvalue.ObjectPartial(custom), knownvalue.ObjectPartial(builtin), knownvalue.ObjectPartial(function),
			})),
		)

		for name, check := range custom {
			checks = append(checks, statecheck.ExpectKnownValue("data.contentful_app_action.one", tfjsonpath.New(name), check))
		}

		return checks
	}

	steps := []resource.TestStep{
		{ConfigDirectory: config.TestNameDirectory(), ConfigVariables: initial, ConfigStateChecks: append(discoveryChecks(custom, builtin, function),
			statecheck.ExpectKnownValue("contentful_app_action.builtin", tfjsonpath.New("parameters"), knownvalue.Null()),
		)},
		{ConfigDirectory: config.TestNameDirectory(), ConfigVariables: initial, ResourceName: actionAddress, ImportState: true, ImportStateVerify: true},
		{ConfigDirectory: config.TestNameDirectory(), ConfigVariables: initial, ResourceName: "contentful_app_action.builtin", ImportState: true, ImportStateVerify: true},
		{ConfigDirectory: config.TestNameDirectory(), ConfigVariables: initial, ResourceName: "contentful_app_action.function", ImportState: true, ImportStateVerify: true},
		{ConfigDirectory: config.TestNameDirectory(), ConfigVariables: updated, ConfigStateChecks: append(discoveryChecks(customUpdated, builtinUpdated, functionUpdated),
			statecheck.ExpectKnownValue(actionAddress, tfjsonpath.New("description"), knownvalue.Null()),
			statecheck.ExpectKnownValue(actionAddress, tfjsonpath.New("result_schema"), knownvalue.Null()),
			statecheck.ExpectKnownValue(actionAddress, tfjsonpath.New("parameters_schema"), knownvalue.Null()),
			statecheck.ExpectKnownValue("contentful_app_action.builtin", tfjsonpath.New("parameters"), knownvalue.Null()),
			statecheck.ExpectKnownValue("contentful_app_action.function", tfjsonpath.New("function_id"), knownvalue.Null()),
		)},
		{ConfigDirectory: config.TestNameDirectory(), ConfigVariables: restored, ConfigStateChecks: append(discoveryChecks(customUpdated, builtinUpdated, function),
			statecheck.ExpectKnownValue("contentful_app_action.function", tfjsonpath.New("url"), knownvalue.Null()),
		)},
	}

	for _, name := range []string{"test", "builtin", "function"} {
		address := "contentful_app_action." + name
		identity := statecheck.CompareValue(compare.ValuesSame())

		for i := range steps {
			if !steps[i].ImportState {
				steps[i].ConfigStateChecks = append(steps[i].ConfigStateChecks, identity.AddStateValue(address, tfjsonpath.New("id")))
			}
		}
	}

	testAccMockableResource(t, server, resource.TestCase{
		Steps: steps,
		CheckDestroy: func(state *terraform.State) error {
			for _, address := range []string{actionAddress, "contentful_app_action.builtin", "contentful_app_action.function", "contentful_app_definition.test"} {
				stored := state.RootModule().Resources[address]
				if stored == nil {
					// Failed applies may have created only part of the fixture.
					continue
				}

				attributes := stored.Primary.Attributes

				var (
					result any
					err    error
				)
				if stored.Type == "contentful_app_action" {
					result, err = client.GetAppAction(t.Context(), cm.GetAppActionParams{OrganizationID: organization, AppDefinitionID: attributes["app_definition_id"], AppActionID: attributes["app_action_id"]})
				} else {
					result, err = client.GetAppDefinition(t.Context(), cm.GetAppDefinitionParams{OrganizationID: organization, AppDefinitionID: attributes["app_definition_id"]})
				}

				require.NoError(t, err)

				status, ok := result.(cm.StatusCodeResponse)
				require.True(t, ok, "%s remains after destroy: %T", address, result)
				assert.Equal(t, http.StatusNotFound, status.GetStatusCode(), address)
			}

			return nil
		},
	})
}
