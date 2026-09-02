package provider_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

var errWebhookBasicPasswordLifecycle = errors.New("webhook Basic password lifecycle mismatch")

//nolint:gosec // Raw response fixture deliberately contains no password.
const rawWebhookWithoutBasicPassword = `{
  "sys": {
    "space": {"sys": {"type": "Link", "linkType": "Space", "id": "space"}},
    "type": "WebhookDefinition",
    "id": "raw-webhook",
    "version": 1
  },
  "name": "Basic webhook",
  "url": "https://example.com/webhook",
  "topics": [],
  "filters": null,
  "httpBasicUsername": "user",
  "headers": [],
  "transformation": null,
  "active": true
}`

func TestAccWebhookBasicPasswordOmittedFromRawCMAResponse(t *testing.T) {
	t.Parallel()

	config := webhookBasicPasswordConfig("Basic webhook", `"basic-password-test-sentinel"`, "")
	ContentfulProviderMockedResourceTest(t, rawWebhookPasswordOmissionHandler{}, resource.TestCase{Steps: []resource.TestStep{
		{
			Config: config,
			Check: resource.TestCheckResourceAttr(
				"contentful_webhook.test", "http_basic_password", "basic-password-test-sentinel",
			),
		},
		{
			Config: config,
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionNoop),
			}},
		},
	}})
}

type rawWebhookPasswordOmissionHandler struct{}

func (rawWebhookPasswordOmissionHandler) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	responseWriter.Header().Set("Content-Type", "application/json")

	switch {
	case request.Method == http.MethodPost && request.URL.Path == "/spaces/space/webhook_definitions":
		responseWriter.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(responseWriter, rawWebhookWithoutBasicPassword)
	case request.Method == http.MethodGet && request.URL.Path == "/spaces/space/webhook_definitions/raw-webhook":
		_, _ = io.WriteString(responseWriter, rawWebhookWithoutBasicPassword)
	case request.Method == http.MethodDelete && request.URL.Path == "/spaces/space/webhook_definitions/raw-webhook":
		responseWriter.WriteHeader(http.StatusNoContent)
	default:
		http.NotFound(responseWriter, request)
	}
}

func TestAccWebhookBasicPasswordLifecycle(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")

	passwordPath := tfjsonpath.New("http_basic_password")
	tests := []struct {
		webhookName string
		password    string
		action      plancheck.ResourceActionType
		stateValue  knownvalue.Check
		storedValue cm.OptNilString
	}{
		{"Created webhook", `"password-one"`, plancheck.ResourceActionCreate, knownvalue.StringExact("password-one"), cm.NewOptNilString("password-one")},
		{"Created webhook", `"password-one"`, plancheck.ResourceActionNoop, knownvalue.StringExact("password-one"), cm.NewOptNilString("password-one")},
		{"Updated webhook", `"password-one"`, plancheck.ResourceActionUpdate, knownvalue.StringExact("password-one"), cm.NewOptNilString("password-one")},
		{"Updated webhook", `"password-one"`, plancheck.ResourceActionNoop, knownvalue.StringExact("password-one"), cm.NewOptNilString("password-one")},
		{"Updated webhook", `"password-two"`, plancheck.ResourceActionUpdate, knownvalue.StringExact("password-two"), cm.NewOptNilString("password-two")},
		{"Updated webhook", `"password-two"`, plancheck.ResourceActionNoop, knownvalue.StringExact("password-two"), cm.NewOptNilString("password-two")},
		{"Updated webhook", "null", plancheck.ResourceActionUpdate, knownvalue.Null(), cm.NewOptNilStringNull()},
		{"Updated webhook", "null", plancheck.ResourceActionNoop, knownvalue.Null(), cm.NewOptNilStringNull()},
	}

	steps := make([]resource.TestStep, 0, len(tests))
	for _, test := range tests {
		steps = append(steps, resource.TestStep{
			Config: webhookBasicPasswordConfig(test.webhookName, test.password, ""),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_webhook.test", test.action),
			}},
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("contentful_webhook.test", passwordPath, test.stateValue),
			},
			Check: checkStoredWebhookPassword(server, test.storedValue),
		})
	}

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{Steps: steps})
}

func TestAccWebhookBasicPasswordImportRequiresUpdate(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")
	server.SetWebhookDefinition("space", "imported-webhook", cm.WebhookDefinitionData{
		Name:              "Imported webhook",
		URL:               "https://example.com/webhook",
		Topics:            []string{},
		HttpBasicUsername: cm.NewOptNilString("user"),
		HttpBasicPassword: cm.NewOptNilString("remote-password"),
	})

	config := webhookBasicPasswordConfig("Imported webhook", `"configured-after-import"`, "")
	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{Steps: []resource.TestStep{
		{
			Config:             config,
			ResourceName:       "contentful_webhook.test",
			ImportState:        true,
			ImportStateId:      "space/imported-webhook",
			ImportStatePersist: true,
			ImportStateCheck: func(states []*terraform.InstanceState) error {
				if len(states) != 1 {
					return fmt.Errorf("%w: imported %d webhook states, want 1", errWebhookBasicPasswordLifecycle, len(states))
				}

				if _, exists := states[0].Attributes["http_basic_password"]; exists {
					return fmt.Errorf("%w: import manufactured http_basic_password in state", errWebhookBasicPasswordLifecycle)
				}

				return nil
			},
		},
		{
			Config: config,
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionUpdate),
			}},
			Check: checkStoredWebhookPassword(server, cm.NewOptNilString("configured-after-import")),
		},
		{
			Config: config,
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionNoop),
			}},
		},
	}})
}

func TestAccWebhookBasicPasswordIgnoreChangesUsesEffectivePlan(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{Steps: []resource.TestStep{
		{Config: webhookBasicPasswordConfig("Initial webhook", `"effective-plan-password"`, "")},
		{
			Config: webhookBasicPasswordConfig("Updated webhook", `"changed-config-password"`, "lifecycle { ignore_changes = [http_basic_password] }"),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionUpdate),
				plancheck.ExpectKnownValue("contentful_webhook.test", tfjsonpath.New("http_basic_password"), knownvalue.StringExact("effective-plan-password")),
			}},
			Check: checkStoredWebhookPassword(server, cm.NewOptNilString("effective-plan-password")),
		},
		{
			Config: webhookBasicPasswordConfig("Updated webhook", `"changed-config-password"`, "lifecycle { ignore_changes = [http_basic_password] }"),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionNoop),
			}},
		},
	}})
}

func webhookBasicPasswordConfig(name, password, lifecycle string) string {
	username := `"user"`
	if password == "null" {
		username = "null"
	}

	return fmt.Sprintf(`
resource "contentful_webhook" "test" {
  space_id            = "space"
  name                = %q
  url                 = "https://example.com/webhook"
  topics              = []
  http_basic_username = %s
  http_basic_password = %s
  %s
}
`, name, username, password, lifecycle)
}

func checkStoredWebhookPassword(server *cmt.Server, expected cm.OptNilString) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState := state.RootModule().Resources["contentful_webhook.test"]
		stored, ok := server.StoredWebhookDefinition("space", resourceState.Primary.Attributes["webhook_id"])

		if !ok {
			return fmt.Errorf("%w: stored webhook not found", errWebhookBasicPasswordLifecycle)
		}

		if stored.HttpBasicPassword != expected {
			return fmt.Errorf("%w: stored password %#v, want %#v", errWebhookBasicPasswordLifecycle, stored.HttpBasicPassword, expected)
		}

		return nil
	}
}
