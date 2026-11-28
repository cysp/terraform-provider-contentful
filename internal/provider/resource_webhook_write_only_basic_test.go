package provider_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

func TestAccWebhookResourceWriteOnlyBasicPassword(t *testing.T) {
	const (
		first     = "original-synthetic-basic"
		planned   = "planned-synthetic-basic"
		applied   = "applied-synthetic-basic"
		variable  = "TF_VAR_webhook_basic_rotation"
		address   = "contentful_webhook.test"
		writeOnly = "http_basic_password_wo = var.webhook_basic_rotation"
	)
	t.Setenv(variable, first)

	server := newTestCMAServer(t)
	server.RegisterSpaceEnvironment("space", "master")
	recorder := &taxonomyRequestBodyRecorder{next: server}
	adapter := &mutationJSONResponseAdapter{delegate: recorder}

	var observations webhookProtocolObservations

	config := func(name, password string) string {
		return fmt.Sprintf(`
variable "webhook_basic_rotation" {
  type      = string
  sensitive = true
  ephemeral = true
}
resource "contentful_webhook" "test" {
  space_id            = "space"
  name                = %q
  url                 = "https://example.invalid/webhook"
  topics              = ["Entry.publish"]
  http_basic_username = "user"
  %s
  headers = { X-Key = { secret = true, value = "ordinary-header" } }
}
`, name, password)
	}
	body := func(name, password string) string {
		return testJSON(map[string]any{
			"name": name, "url": "https://example.invalid/webhook", "active": true, "topics": []string{"Entry.publish"},
			"httpBasicUsername": "user", "httpBasicPassword": password, "filters": nil, "transformation": nil,
			"headers": []map[string]any{{"key": "X-Key", "secret": true, "value": "ordinary-header"}},
		})
	}

	var (
		webhookID          string
		original, baseline webhookPrivateVerifiers
	)

	requirePut := func(number int, version, name, password string) {
		t.Helper()

		puts := recorder.matchingRequests(http.MethodPut, "")
		require.Len(t, puts, number)
		put := puts[number-1]
		require.Equal(t, "/spaces/space/webhook_definitions/"+webhookID, put.path)
		require.Equal(t, version, put.version)
		require.JSONEq(t, body(name, password), string(put.body))
	}
	managed := config("Managed", writeOnly)
	changed := config("Changed", writeOnly)
	check := func(key string, value knownvalue.Check) statecheck.StateCheck {
		return statecheck.ExpectKnownValue(address, tfjsonpath.New(key), value)
	}

	var rejected bool

	resource.Test(t, validateAcceptanceTestCase(t, resource.TestCase{
		ProtoV6ProviderFactories: observations.providerFactories(t, adapter),
		ErrorCheck: func(err error) error {
			message := err.Error()
			if strings.Contains(message, first) {
				t.Fatal("Terraform diagnostic exposed the write-only password")
			}

			if rejected || !strings.Contains(message, "Error running apply:") || !strings.Contains(message, "Contentful returned a different webhook name") {
				return err
			}

			rejected = true

			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: managed,
				Check: func(state *terraform.State) error {
					webhookID = state.RootModule().Resources[address].Primary.Attributes["webhook_id"]
					require.NotEmpty(t, webhookID)

					created := recorder.matchingRequests(http.MethodPost, "")
					require.Len(t, created, 1)
					require.Equal(t, "/spaces/space/webhook_definitions", created[0].path)
					require.Empty(t, created[0].version)
					require.JSONEq(t, body("Managed", first), string(created[0].body))

					reply := observations.lastApply()
					original = reply.verifiers(t)
					require.NotNil(t, original.Basic)
					require.Equal(t, "user", original.Basic.Username)
					require.NotEmpty(t, original.Basic.Hash)
					require.NotContains(t, string(reply.privateFields(t)["write_only_secret_hashes"]), first)

					return nil
				},
				ConfigStateChecks: []statecheck.StateCheck{check("http_basic_password", knownvalue.Null()), check("http_basic_password_wo", knownvalue.Null())},
			},
			{
				// Applying the saved no-op cannot rotate a password, even if the ephemeral input changes.
				Config: managed, ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply:             []plancheck.PlanCheck{plancheck.ExpectEmptyPlan(), afterPlan(func() { t.Setenv(variable, planned) })},
					PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectResourceAction(address, plancheck.ResourceActionUpdate)},
				},
				PostApplyFunc: func() { require.Empty(t, recorder.matchingRequests(http.MethodPut, "")) },
			},
			{
				Config: managed,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(address, plancheck.ResourceActionUpdate), afterPlan(func() { t.Setenv(variable, applied) }),
				}},
				PostApplyFunc: func() {
					requirePut(1, "0", "Managed", applied)

					rotated := observations.lastApply().verifiers(t)
					require.NotNil(t, rotated.Basic)
					require.NotEqual(t, original.Basic.Hash, rotated.Basic.Hash)
				},
			},
			{
				Config:            config("Managed", fmt.Sprintf("http_basic_password = %q", applied)),
				ConfigStateChecks: []statecheck.StateCheck{check("http_basic_password", knownvalue.StringExact(applied))},
				PostApplyFunc:     func() { require.Nil(t, observations.lastApply().verifiers(t).Basic) },
			},
			{Config: managed, PostApplyFunc: func() {
				requirePut(1, "0", "Managed", applied)

				baseline = observations.lastApply().verifiers(t)
				require.NotNil(t, baseline.Basic)
			}},
			{Config: changed, PreConfig: func() {
				t.Setenv(variable, first)
				adapter.mutateNext(http.MethodPut, func(doc map[string]any) { doc["name"] = "Contradictory" })
			}},
			{
				// A plan without refresh exposes exactly what Core persisted after the failed PUT.
				Config: changed, PlanOnly: true, ExpectNonEmptyPlan: true,
				PreConfig: func() {
					require.True(t, rejected, "contradictory PUT was accepted")
					requirePut(2, "1", "Changed", first)
					require.Equal(t, baseline, observations.lastApply().verifiers(t))
					t.Setenv(variable, planned)
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{PostApplyPreRefresh: []plancheck.PlanCheck{
					testAccPriorState{check: check("name", knownvalue.StringExact("Contradictory"))},
					expectResponseStateInPlan{address: address, valuePath: tfjsonpath.New("version"), value: json.Number("2")},
					testAccPriorState{check: check("http_basic_password", knownvalue.Null())},
					testAccPriorState{check: check("http_basic_password_wo", knownvalue.Null())},
					testAccPriorState{check: check("headers", knownvalue.MapExact(map[string]knownvalue.Check{
						"X-Key": knownvalue.ObjectExact(map[string]knownvalue.Check{"secret": knownvalue.Bool(true), "value": knownvalue.StringExact("ordinary-header"), "value_wo": knownvalue.Null()}),
					}))},
				}},
			},
			{
				// Refresh accepts the stored acknowledgement. A differing input deliberately sends a new PUT.
				Config: changed, PreConfig: func() { t.Setenv(variable, applied) },
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}},
			},
			{Config: changed, PreConfig: func() { t.Setenv(variable, planned) }, PostApplyFunc: func() { requirePut(3, "2", "Changed", planned) }},
		},
	}))
}
