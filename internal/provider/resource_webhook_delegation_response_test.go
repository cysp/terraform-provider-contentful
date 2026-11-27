package provider_test

import (
	"context"
	"net/http"
	"regexp"
	"strings"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"
)

type webhookPlanCheck func(*tfjson.Plan)

func newWebhookDelegationFixture(t *testing.T) *webhookSecretRecoveryFixture {
	t.Helper()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")

	return &webhookSecretRecoveryFixture{server: server}
}

func TestAccWebhookResourceUnconfirmedWriteOnlyPassword(t *testing.T) {
	t.Parallel()
	fixture := newWebhookDelegationFixture(t)

	var (
		observed webhookProtocolObservations
		prior    webhookPrivateVerifiers
	)

	rotated := webhookBasicAuthConfig("Webhook", `"user"`, "null", `http_basic_password_wo = "changed-value"`)
	testObservedWebhookResource(t, fixture, &observed, []resource.TestStep{
		{Config: webhookBasicAuthConfig("Webhook", `"user"`, "null", `http_basic_password_wo = "initial-value"`), PostApplyFunc: func() {
			prior = observed.lastApply().verifiers(t)
			require.NotNil(t, prior.Basic)
		}},
		{Config: rotated, PreConfig: func() {
			webhookReplyFields(fixture, http.MethodPut, func(doc map[string]any) { doc["httpBasicPassword"] = nil })
		}, ExpectError: regexp.MustCompile("Unconfirmed webhook Basic password")},
		{Config: rotated, PreConfig: func() { require.Equal(t, prior, observed.lastApply().verifiers(t)) }, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionUpdate)}}, PostApplyFunc: func() {
			writes := fixture.mutations()
			require.Len(t, writes, 3)

			for i, version := range []string{"0", "1"} {
				require.Equal(t, http.MethodPut, writes[i+1].method)
				require.Equal(t, version, writes[i+1].version)
				require.Equal(t, "changed-value", testJSONObject(writes[i+1].body)["httpBasicPassword"])
			}
		}},
	})
}

func TestAccWebhookResourceDelegatedPasswordAfterDeletion(t *testing.T) {
	t.Parallel()
	fixture := newSeededWebhookFixture(t)

	var observed webhookProtocolObservations

	update := webhookRecoveryStep(strings.Replace(webhookRecoveryConfig(false, ""), "http_basic_password_wo = var.password", "", 1), webhookPriorPassword, webhookRotatedHeader)
	update.ConfigPlanChecks = resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
		plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionUpdate),
		webhookPlanCheck(func(*tfjson.Plan) {
			deletion, err := fixture.server.Handler().DeleteWebhookDefinition(t.Context(), cm.DeleteWebhookDefinitionParams{SpaceID: "space", WebhookDefinitionID: "imported"})
			require.NoError(t, err)
			require.IsType(t, &cm.NoContent{}, deletion)
		}),
	}}
	update.ExpectError = regexp.MustCompile("Contentful returned a different webhook Basic username")
	cleanup := webhookRecoveryStep(update.Config, webhookPriorPassword, webhookRotatedHeader)
	cleanup.Destroy = true
	cleanup.PreConfig = func() {
		attempt := webhookExpectedRecoveryPut("1", "", webhookRotatedHeader, false)
		attempt.body = mutateTestJSON(attempt.body, func(doc map[string]any) {
			delete(doc, "httpBasicUsername")
			delete(doc, "httpBasicPassword")
		})
		fixture.requirePuts(t, webhookExpectedRecoveryPut("0", webhookPriorPassword, webhookPriorHeader, true), attempt)
		stored, exists := fixture.server.StoredWebhookDefinition("space", "imported")
		require.True(t, exists)
		require.True(t, stored.Active.Or(false))
		require.True(t, stored.HttpBasicUsername.IsNull())
		require.True(t, stored.HttpBasicPassword.IsNull())

		reply := observed.lastApply()
		state := reply.model(t)
		require.Equal(t, types.Int64Value(0), state.Version)
		require.True(t, state.HTTPBasicUsername.IsNull())
		require.Empty(t, reply.verifiers(t))
	}
	testObservedWebhookResource(t, fixture, &observed, []resource.TestStep{importedWebhookRecoveryStep(t, fixture), update, cleanup})
}

func TestAccWebhookResourceDelegatedOrdinaryPasswordEcho(t *testing.T) {
	t.Parallel()
	fixture := newWebhookDelegationFixture(t)

	var observed webhookProtocolObservations

	delegated := webhookBasicAuthConfig("Renamed", `"user"`, "null", "")
	testObservedWebhookResource(t, fixture, &observed, []resource.TestStep{
		{Config: webhookBasicAuthConfig("Original", `"user"`, `"initial-value"`, "")},
		{RefreshState: true, PreConfig: func() {
			webhookReplyFields(fixture, http.MethodGet, func(doc map[string]any) {
				delete(doc, "httpBasicUsername")
				delete(doc, "httpBasicPassword")
			})
		}, Check: func(*terraform.State) error {
			state := observed.lastRead().model(t)
			require.Equal(t, "user", state.HTTPBasicUsername.ValueString())
			require.Equal(t, "initial-value", state.HTTPBasicPassword.ValueString())

			return nil
		}},
		{Config: delegated, PreConfig: func() {
			webhookReplyFields(fixture, http.MethodPut, func(doc map[string]any) { doc["httpBasicPassword"] = webhookRotatedPassword })
		}, ExpectError: regexp.MustCompile("Unexpected webhook Basic password in response")},
		{Config: delegated, PreConfig: func() {
			reply := observed.lastApply()
			require.True(t, reply.model(t).HTTPBasicPassword.IsNull())
			require.Empty(t, reply.verifiers(t))

			for _, d := range reply.diagnostics {
				requireNoWebhookRecoverySecrets(t, d.Summary+" "+d.Detail)
			}

			requireNoWebhookRecoverySecrets(t, reply.logs)
		}},
	})
}

func TestAccWebhookResourceDelegatedNullPasswordResponse(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, ordinary, extra string
		wantVerifier          bool
	}{
		{name: "ordinary", ordinary: `"initial-value"`},
		{name: "write-only", ordinary: "null", extra: `http_basic_password_wo = "initial-value"`, wantVerifier: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			fixture := newWebhookDelegationFixture(t)

			var (
				observed webhookProtocolObservations
				prior    webhookPrivateVerifiers
			)

			nullPassword := func(doc map[string]any) { doc["httpBasicPassword"] = nil }

			testObservedWebhookResource(t, fixture, &observed, []resource.TestStep{
				{Config: webhookBasicAuthConfig("Original", `"user"`, test.ordinary, test.extra), PostApplyFunc: func() {
					prior = observed.lastApply().verifiers(t)
					require.Equal(t, test.wantVerifier, prior.Basic != nil)
				}},
				{Config: webhookBasicAuthConfig("Renamed", `"user"`, "null", ""), PreConfig: func() {
					webhookReplyFields(fixture, http.MethodPut, nullPassword)
				}, PostApplyFunc: func() {
					writes := fixture.mutations()
					require.Len(t, writes, 2)
					require.Equal(t, http.MethodPut, writes[1].method)
					require.Equal(t, "0", writes[1].version)
					body := testJSONObject(writes[1].body)
					require.NotContains(t, body, "httpBasicUsername")
					require.NotContains(t, body, "httpBasicPassword")

					reply := observed.lastApply()
					state := reply.model(t)
					require.Equal(t, "user", state.HTTPBasicUsername.ValueString())
					require.True(t, state.HTTPBasicPassword.IsNull())
					require.Equal(t, prior, reply.verifiers(t))
				}},
				{RefreshState: true, PreConfig: func() { webhookReplyFields(fixture, http.MethodGet, nullPassword) }, Check: func(*terraform.State) error {
					require.Empty(t, observed.lastRead().verifiers(t))

					return nil
				}},
			})
		})
	}
}

func (check webhookPlanCheck) CheckPlan(_ context.Context, req plancheck.CheckPlanRequest, _ *plancheck.CheckPlanResponse) {
	check(req.Plan)
}

func TestAccWebhookResourceDelegatedOrdinaryHeaderEcho(t *testing.T) {
	t.Parallel()

	for _, ignored := range []bool{false, true} {
		name := "omitted value"
		if ignored {
			name = "omitted key retained by lifecycle"
		}

		t.Run(name, func(t *testing.T) {
			t.Parallel()
			fixture := newWebhookDelegationFixture(t)

			var observed webhookProtocolObservations

			const sibling = `"X-Sibling" = {secret=true,value_wo="unchanged"}`

			original := webhookHeadersConfig("Original", `headers = { "X-Managed" = {secret=true,value="prior"}, `+sibling+` }`)

			headers := `headers = { "X-Managed" = {secret=true}, ` + sibling + ` }`
			if ignored {
				headers = `headers = { ` + sibling + ` }
 lifecycle { ignore_changes = [headers["X-Managed"]] }`
			}

			delegated := webhookHeadersConfig("Renamed", headers)
			testObservedWebhookResource(t, fixture, &observed, []resource.TestStep{
				{Config: original},
				{Config: delegated, PreConfig: func() {
					webhookReplyFields(fixture, http.MethodPut, func(doc map[string]any) {
						headers, ok := doc["headers"].([]any)
						require.True(t, ok)

						headers[0] = map[string]any{"key": "X-Managed", "secret": false, "value": webhookRotatedHeader}
					})
				}, ExpectError: regexp.MustCompile("Unexpected webhook secret header in response")},
				{Config: delegated, PreConfig: func() {
					writes := fixture.mutations()
					require.Len(t, writes, 2)
					require.Equal(t, "0", writes[1].version)
					require.Equal(t, []any{map[string]any{"key": "X-Managed", "secret": true}, map[string]any{"key": "X-Sibling", "secret": true}}, testJSONObject(writes[1].body)["headers"])

					reply := observed.lastApply()
					require.True(t, reply.model(t).Headers.Elements()["X-Managed"].Value().Value.IsNull())
					require.Equal(t, observed.firstApply().verifiers(t), reply.verifiers(t))
					requireNoWebhookRecoverySecrets(t, reply.logs)
				}, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionNoop),
				}}},
			})
		})
	}
}
