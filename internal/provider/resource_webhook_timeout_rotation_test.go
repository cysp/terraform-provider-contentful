package provider_test

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

type webhookAfterPlan func()

func (f webhookAfterPlan) CheckPlan(context.Context, plancheck.CheckPlanRequest, *plancheck.CheckPlanResponse) {
	f()
}

func TestAccWebhookResourceIgnoredWriteOnlyBasicPasswordOnTimeoutUpdate(t *testing.T) {
	const (
		original = "synthetic-installed-password"
		rotated  = "synthetic-timeout-rotation"
		variable = "TF_VAR_webhook_timeout_password"
		address  = "contentful_webhook.test"
	)
	// The framework inherits TF_VAR_* separately for planning and saved-plan apply.
	t.Setenv(variable, original)

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")
	recorder := &taxonomyRequestBodyRecorder{next: server}
	adapter := &mutationJSONResponseAdapter{delegate: recorder}
	seed := cm.WebhookDefinitionData{
		Name: "Imported", URL: "https://example.invalid/webhook", Topics: []string{"Entry.publish"}, Active: cm.NewOptBool(true),
		HttpBasicUsername: cm.NewOptNilString("user"), HttpBasicPassword: cm.NewOptNilString(original),
		Headers: cm.WebhookDefinitionRequestHeaders{
			{Key: "X-Imported", Secret: cm.NewOptBool(true), Value: cm.NewOptNilString("externally-managed-header")},
			{Key: "x-imported", Secret: cm.NewOptBool(true), Value: cm.NewOptNilString("another-header")},
		},
	}
	seed.Filters.SetToNull()
	seed.Transformation.SetToNull()
	server.SetWebhookDefinition("space", "imported", seed)

	configure := func(lifecycle string) string {
		return fmt.Sprintf(`
variable "webhook_timeout_password" {
 type = string
 ephemeral = true
 sensitive = true
}
resource "contentful_webhook" "test" {
 space_id = "space"
 name = "Imported"
 url = "https://example.invalid/webhook"
 topics = ["Entry.publish"]
 http_basic_username = "user"
 http_basic_password_wo = var.webhook_timeout_password
 %s
}`, lifecycle)
	}
	initial := configure("lifecycle { ignore_changes = [headers] }")
	timeout := configure(`timeouts = { update = "4m" }
 lifecycle { ignore_changes = [headers] }`)
	omitHeaderValue := func() {
		adapter.mutateNext(http.MethodGet, func(doc map[string]any) {
			doc["headers"] = []map[string]any{{"key": "X-External", "secret": false}}
		})
	}
	testAccMockedResource(t, adapter, resource.TestCase{Steps: []resource.TestStep{
		{
			Config: initial, ResourceName: address, ImportState: true, ImportStateId: "space/imported", ImportStatePersist: true,
			ImportStateCheck: testAccImportAttributes(map[string]string{"headers.%": "2", "headers.X-Imported.secret": "true", "headers.x-imported.secret": "true"},
				"headers.X-Imported.value", "headers.X-Imported.value_wo", "headers.x-imported.value", "headers.x-imported.value_wo"),
		},
		{Config: initial},
		{
			Config: configure(`timeouts = { update = "3m" }
 lifecycle { ignore_changes = [http_basic_password_wo, headers] }`),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction(address, plancheck.ResourceActionUpdate),
				webhookAfterPlan(func() { t.Setenv(variable, rotated) }),
			}},
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(address, tfjsonpath.New("http_basic_password"), knownvalue.Null()),
			},
			PostApplyFunc: func() {
				puts := recorder.matchingRequests(http.MethodPut, "")
				require.Len(t, puts, 2)

				for i, password := range []string{original, rotated} {
					require.Equal(t, "/spaces/space/webhook_definitions/imported", puts[i].path)
					// An initial PUT of the same remote bytes leaves version zero.
					require.Equal(t, "0", puts[i].version)
					require.JSONEq(t, testJSON(map[string]any{
						"name": "Imported", "url": "https://example.invalid/webhook", "active": true, "topics": []string{"Entry.publish"},
						"httpBasicUsername": "user", "httpBasicPassword": password, "filters": nil, "transformation": nil,
						"headers": []map[string]any{{"key": "X-Imported", "secret": true}, {"key": "x-imported", "secret": true}},
					}), string(puts[i].body))
				}

				stored, ok := server.StoredWebhookDefinition("space", "imported")
				require.True(t, ok)
				require.Equal(t, rotated, stored.HttpBasicPassword.Or(""))
				require.Equal(t, "externally-managed-header", stored.Headers[0].Value.Or(""))
				require.Equal(t, "another-header", stored.Headers[1].Value.Or(""))
				require.Equal(t, 1, stored.Sys.Version)
			},
		},
		{
			// A timeout-only change can checkpoint an incomplete GET without sending a PUT.
			Config: timeout, PreConfig: omitHeaderValue,
			ConfigStateChecks: []statecheck.StateCheck{statecheck.ExpectKnownValue(address, tfjsonpath.New("headers"), knownvalue.MapExact(map[string]knownvalue.Check{
				"X-External": knownvalue.ObjectExact(map[string]knownvalue.Check{"value": knownvalue.Null(), "value_wo": knownvalue.Null(), "secret": knownvalue.Bool(false)}),
			}))},
		},
		{
			Config: timeout, PreConfig: func() { t.Setenv(variable, original); omitHeaderValue() },
			ExpectError: regexp.MustCompile("Missing webhook header value"),
		},
	}})
	require.Len(t, recorder.matchingRequests(http.MethodPut, ""), 2)
}
