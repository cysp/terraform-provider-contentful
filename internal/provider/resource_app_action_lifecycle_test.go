package provider_test

import (
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

// This mocked/live lifecycle creates only a temporary App Definition and actions.
// Function deployment/invocation and spaces/environments are not touched.
//
//nolint:paralleltest // Live acceptance serializes access to the shared account.
func TestAccAppActionResourceLifecycle(t *testing.T) {
	parallelWhenMocked(t)

	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC must be set for acceptance tests")
	}

	// Use the same organization fixture as the App Definition lifecycle, including CI.
	const organization = "2zuSjSO4A0e6GKBrhJRe2m"

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)

	// The framework owns destruction; an independent client verifies remote absence.
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

	steps := []resource.TestStep{
		{ConfigDirectory: config.TestNameDirectory(), ConfigVariables: initial, ConfigStateChecks: []statecheck.StateCheck{
			statecheck.ExpectKnownValue("data.contentful_app_actions.all", tfjsonpath.New("app_actions"), knownvalue.ListSizeExact(3)),
		}},
		{ConfigDirectory: config.TestNameDirectory(), ConfigVariables: initial, ResourceName: actionAddress, ImportState: true, ImportStateVerify: true},
		{ConfigDirectory: config.TestNameDirectory(), ConfigVariables: initial, ResourceName: "contentful_app_action.builtin", ImportState: true, ImportStateVerify: true},
		{ConfigDirectory: config.TestNameDirectory(), ConfigVariables: initial, ResourceName: "contentful_app_action.function", ImportState: true, ImportStateVerify: true},
		{ConfigDirectory: config.TestNameDirectory(), ConfigVariables: initial, ResourceName: actionAddress, ImportState: true, ImportStateKind: resource.ImportBlockWithResourceIdentity},
		{ConfigDirectory: config.TestNameDirectory(), ConfigVariables: updated, ConfigStateChecks: []statecheck.StateCheck{
			statecheck.ExpectKnownValue(actionAddress, tfjsonpath.New("description"), knownvalue.Null()),
			statecheck.ExpectKnownValue(actionAddress, tfjsonpath.New("result_schema"), knownvalue.Null()),
			statecheck.ExpectKnownValue(actionAddress, tfjsonpath.New("parameters_schema"), knownvalue.Null()),
			statecheck.ExpectKnownValue("contentful_app_action.function", tfjsonpath.New("function_id"), knownvalue.Null()),
		}},
		{ConfigDirectory: config.TestNameDirectory(), ConfigVariables: restored, ConfigStateChecks: []statecheck.StateCheck{
			statecheck.ExpectKnownValue("contentful_app_action.function", tfjsonpath.New("url"), knownvalue.Null()),
		}},
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
