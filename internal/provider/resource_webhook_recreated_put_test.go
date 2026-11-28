package provider_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccWebhookResourceGuardsSubmittedCredentialsAfterRecreation(t *testing.T) {
	t.Parallel()

	const address = "contentful_webhook.test"

	server := newTestCMAServer(t)
	server.RegisterSpaceEnvironment("space", "master")
	server.SetWebhookDefinition("space", "imported", cm.WebhookDefinitionData{
		Name: "Imported", URL: "https://example.invalid/webhook", Topics: []string{"Entry.publish"}, Active: cm.NewOptBool(true),
	})
	fixture := &webhookSecretRecoveryFixture{server: server}
	initial := webhookRecoveryHeaderConfig(`"X-Sibling" = { secret=true, value_wo="unchanged" }`, "", "") + `
import {
  to = contentful_webhook.test
  id = "space/imported"
}
`
	rotated := strings.Replace(initial, `http_basic_username    = "user"`, `http_basic_username    = "changed-user"`, 1)
	first := config.Variables{"password": config.StringVariable(webhookPriorPassword), "header": config.StringVariable(webhookPriorHeader)}
	next := config.Variables{"password": config.StringVariable(webhookRotatedPassword), "header": config.StringVariable(webhookRotatedHeader)}

	testAccMockedResource(t, fixture, resource.TestCase{Steps: []resource.TestStep{
		{Config: initial, ConfigVariables: first},
		{
			Config: rotated, ConfigVariables: next,
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{webhookPlanCheck(func(*tfjson.Plan) {
				recreateWebhookAtVersion(t, fixture, 1)
				fixture.onNextResponse(http.MethodPut, func(w http.ResponseWriter, rec *httptest.ResponseRecorder) {
					_, _ = io.WriteString(w, mutateTestJSON(rec.Body.String(), func(doc map[string]any) {
						delete(doc, "httpBasicUsername")
						delete(doc, "httpBasicPassword")
						delete(doc, "headers")
					}))
				})
			})}},
			ExpectError: regexp.MustCompile("Webhook was recreated"),
		},
		{
			Config: rotated, ConfigVariables: next,
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction(address, plancheck.ResourceActionUpdate),
				webhookPlanCheck(func(plan *tfjson.Plan) {
					// The next GET restores the full list. Its drift record exposes
					// exactly what Terraform checkpointed on the preceding failed PUT.
					require.Len(t, plan.ResourceDrift, 1)
					drift := plan.ResourceDrift[0]
					require.Equal(t, address, drift.Address)
					prior, ok := drift.Change.Before.(map[string]any)
					require.True(t, ok)
					assert.Equal(t, "changed-user", prior["http_basic_username"])
					assert.Nil(t, prior["http_basic_password"])
					assert.Equal(t, map[string]any{"X-Managed": map[string]any{"secret": true, "value": nil, "value_wo": nil}}, prior["headers"])
				}),
			}},
			PostApplyFunc: func() {
				writes := fixture.mutations()
				require.Len(t, writes, 3)
				require.Equal(t, "1", writes[1].version)
				require.Equal(t, "2", writes[2].version)

				headers := []any{map[string]any{"key": "X-Managed", "secret": true, "value": webhookRotatedHeader}, map[string]any{"key": "X-Sibling", "secret": true}}
				require.Equal(t, headers, testJSONObject(writes[1].body)["headers"])
				headers[1] = map[string]any{"key": "X-Sibling", "secret": true, "value": "unchanged"}
				require.Equal(t, headers, testJSONObject(writes[2].body)["headers"])
			},
		},
	}})
}

func TestAccWebhookResourceGuardsDelegatedHeaderAfterRecreation(t *testing.T) {
	t.Parallel()
	fixture := newWebhookDelegationFixture(t)

	var observations webhookProtocolObservations

	const key = "X-Review"

	initial := webhookHeadersConfig("Original", `headers = { X-Review = { secret = true, value = "prior" } }`)
	delegated := webhookHeadersConfig("Renamed", `headers = { X-Review = { secret = true } }`)
	replacement := cm.WebhookDefinitionData{Name: "Replacement", URL: "https://example.com/webhook", Topics: []string{"Entry.publish"}, Active: cm.NewOptBool(true), Headers: cm.WebhookDefinitionRequestHeaders{{Key: key, Value: cm.NewOptNilString("public-on-replacement")}}}

	var webhookID string

	put := func(version cm.OptInt) {
		t.Helper()
		reply, err := fixture.server.Handler().UpdateWebhookDefinition(t.Context(), &replacement, cm.UpdateWebhookDefinitionParams{SpaceID: "space", WebhookDefinitionID: webhookID, XContentfulVersion: version})
		require.NoError(t, err)
		require.IsType(t, &cm.WebhookDefinitionStatusCode{}, reply)
	}

	testObservedWebhookResource(t, fixture, &observations, []resource.TestStep{
		{Config: initial},
		{Config: delegated, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{afterPlan(func() {
			webhookID = observations.lastApply().model(t).WebhookID.ValueString()
			removed, err := fixture.server.Handler().DeleteWebhookDefinition(t.Context(), cm.DeleteWebhookDefinitionParams{SpaceID: "space", WebhookDefinitionID: webhookID})
			require.NoError(t, err)
			require.IsType(t, &cm.NoContent{}, removed)
			put(cm.OptInt{})
		})}}, ExpectError: regexp.MustCompile("Webhook was recreated")},
		{RefreshState: true, PreConfig: func() {
			reply := observations.lastApply()
			header := reply.model(t).Headers.Elements()[key].Value()
			require.True(t, header.Value.IsNull(), "discard the readable replacement value during the failed PUT")
			require.True(t, header.Secret.ValueBool())
			require.Empty(t, reply.verifiers(t))

			writes := fixture.mutations()
			require.Len(t, writes, 2)
			require.Equal(t, "0", writes[1].version)
			require.Equal(t, []any{map[string]any{"key": key, "secret": true}}, testJSONObject(writes[1].body)["headers"])
		}, ExpectError: regexp.MustCompile("Unexpected webhook secret header in response")},
		{Config: delegated, PreConfig: func() {
			replacement.Name = "Renamed"
			replacement.Headers = nil

			put(cm.NewOptInt(1))

			replacement.Headers = cm.WebhookDefinitionRequestHeaders{{Key: key, Secret: cm.NewOptBool(true), Value: cm.NewOptNilString("externally-managed")}}

			put(cm.NewOptInt(2))
		}, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionNoop)}}},
	})
}
