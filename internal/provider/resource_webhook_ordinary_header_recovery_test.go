package provider_test

import (
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

// An ordinary value remains public even if Contentful's secret flag is true.
// Keep both refresh drift and a contradictory write independently observable.
func TestAccWebhookResourceOrdinaryHeaderRecovery(t *testing.T) {
	t.Parallel()
	fixture := newSeededWebhookFixture(t)

	var observations webhookProtocolObservations

	const config = `resource "contentful_webhook" "test" {
  space_id = "space"
  name = "Ordinary header"
  url = "https://example.invalid/webhook"
  topics = ["Entry.publish"]
  headers = { "X-Ordinary" = { secret = true, value = "configured-public-value" } }
}`

	ordinary := strings.Replace(config, "secret = true", "secret = false", 1)

	testObservedWebhookResource(t, fixture, &observations, []resource.TestStep{
		{Config: config},
		{
			Config: ordinary, PreConfig: func() {
				// An external owner removes the secret header before adding a readable header with the same name.
				webhookID := observations.lastApply().model(t).WebhookID.ValueString()

				updated := cm.WebhookDefinitionData{Name: "Ordinary header", URL: "https://example.invalid/webhook", Topics: []string{"Entry.publish"}, Active: cm.NewOptBool(true)}
				for version, headers := range []cm.WebhookDefinitionRequestHeaders{
					{},
					{{Key: "X-Ordinary", Secret: cm.NewOptBool(false), Value: cm.NewOptNilString("changed-public-value")}},
				} {
					updated.Headers = headers
					response, err := fixture.server.Handler().UpdateWebhookDefinition(t.Context(), &updated, cm.UpdateWebhookDefinitionParams{
						SpaceID: "space", WebhookDefinitionID: webhookID, XContentfulVersion: cm.NewOptInt(version),
					})
					require.NoError(t, err)
					require.IsType(t, &cm.WebhookDefinitionStatusCode{}, response)
				}

				webhookReplyFields(fixture, http.MethodPut, func(doc map[string]any) {
					doc["headers"] = []map[string]any{{"key": "X-Ordinary", "secret": true, "value": "contradictory-public-value"}}
				})
			}, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionUpdate),
				testAccPriorState{check: statecheck.ExpectKnownValue("contentful_webhook.test", tfjsonpath.New("headers"), knownvalue.MapExact(map[string]knownvalue.Check{
					"X-Ordinary": knownvalue.ObjectExact(map[string]knownvalue.Check{
						"secret": knownvalue.Bool(false), "value": knownvalue.StringExact("changed-public-value"), "value_wo": knownvalue.Null(),
					}),
				}))},
			}},
			ExpectError: regexp.MustCompile("(?s)Error running apply:.*Contentful returned different webhook headers"),
		},
		{Config: ordinary, PreConfig: func() {
			headers := observations.lastApply().model(t).Headers.Elements()
			require.Len(t, headers, 1)
			header := headers["X-Ordinary"].Value()
			require.True(t, header.Secret.ValueBool())
			require.Equal(t, "contradictory-public-value", header.Value.ValueString())
			require.True(t, header.ValueWO.IsNull())

			writes := fixture.mutations()
			require.Len(t, writes, 2)
			require.Equal(t, http.MethodPut, writes[1].method)
			require.Equal(t, "2", writes[1].version)
			require.Equal(t, []any{map[string]any{"key": "X-Ordinary", "secret": false, "value": "configured-public-value"}}, testJSONObject(writes[1].body)["headers"])
		}},
	})
}

// A contradictory name does not change the omission fallback for an ordinary
// credential. Exercise the same update with and without a write-only sibling.
func TestAccWebhookResourceOrdinaryCredentialRecovery(t *testing.T) {
	t.Parallel()

	for _, sibling := range []bool{false, true} {
		t.Run(fmt.Sprintf("write_only_sibling=%t", sibling), func(t *testing.T) {
			t.Parallel()
			runtime, fixture := newWebhookRecoveryFixture(t)

			declaration := ""
			if sibling {
				declaration = `"X-Sibling" = { secret = true, value_wo = var.sibling }`
			}

			configure := func(value string) {
				runtime.writeConfig(t, fmt.Sprintf(`
variable "sibling" {
 type = string
 sensitive = true
 ephemeral = true
}
resource "contentful_webhook" "test" {
 space_id = "space"
 name = "Configured"
 url = "https://example.invalid/webhook"
 topics = ["Entry.publish"]
 http_basic_username = "user"
 http_basic_password = %q
 headers = {
   "X-Managed" = {secret=true, value=%q}
   %s
 }
}
`, "password-"+value, "header-"+value, declaration))
			}
			cli := func(command string, options ...string) (string, error) {
				return runtime.run(t.Context(), append([]string{command, "-input=false", "-no-color", "-var=sibling=" + webhookSiblingHeader}, options...)...)
			}

			configure("before")

			out, err := cli("apply", "-auto-approve")
			require.NoError(t, err, out)
			before := readWebhookRecoveryInstance(t, runtime)
			webhookID, ok := before.Attributes["webhook_id"].(string)
			require.True(t, ok)
			configure("after")
			fixture.onNextResponse(http.MethodPut, func(w http.ResponseWriter, rec *httptest.ResponseRecorder) {
				body := mutateTestJSON(rec.Body.String(), func(doc map[string]any) { doc["name"] = "Contradictory" })
				_, _ = io.WriteString(w, body)
			})

			out, err = cli("apply", "-auto-approve")
			require.Error(t, err, out)
			require.Contains(t, out, "Contentful returned a different webhook name")

			writes := fixture.mutations()
			require.Len(t, writes, 2)

			headers := []map[string]any{{"key": "X-Managed", "secret": true, "value": "header-after"}}
			public := map[string]any{"X-Managed": map[string]any{"secret": true, "value": "header-after", "value_wo": nil}}

			if sibling {
				headers = append(headers, map[string]any{"key": "X-Sibling", "secret": true})
				public["X-Sibling"] = map[string]any{"secret": true, "value": nil, "value_wo": nil}

				require.NotEmpty(t, before.verifiers)
			}

			require.Equal(t, webhookMutation{method: http.MethodPut, path: "/spaces/space/webhook_definitions/" + webhookID, version: "0", body: testJSON(map[string]any{
				"active": true, "name": "Configured", "url": "https://example.invalid/webhook", "topics": []string{"Entry.publish"},
				"httpBasicUsername": "user", "httpBasicPassword": "password-after", "headers": headers, "filters": nil, "transformation": nil,
			})}, writes[1])
			after := readWebhookRecoveryInstance(t, runtime)
			require.Equal(t, webhookID, after.Attributes["webhook_id"])
			require.Equal(t, "1", testJSON(after.Attributes["version"]))
			require.Equal(t, "Contradictory", after.Attributes["name"])
			require.Equal(t, before.verifiers, after.verifiers)
			require.Equal(t, "password-after", after.Attributes["http_basic_password"])
			require.Equal(t, public, after.Attributes["headers"])

			out, err = cli("plan", "-detailed-exitcode")
			require.NoError(t, err, out)
			require.Equal(t, writes, fixture.mutations())
		})
	}
}

func TestAccWebhookResourceWrongTargetPUTKeepsPriorState(t *testing.T) {
	t.Parallel()
	fixture := newSeededWebhookFixture(t)
	fixture.server.SetWebhookDefinition("space", "foreign", cm.WebhookDefinitionData{
		Name: "Foreign", URL: "https://example.invalid/foreign", Topics: []string{"Asset.publish"}, Active: cm.NewOptBool(false),
		HttpBasicUsername: cm.NewOptNilString("other-user"),
		Headers:           cm.WebhookDefinitionRequestHeaders{{Key: "X-Managed", Secret: cm.NewOptBool(true), Value: cm.NewOptNilString(rand.Text())}},
	})

	var divert atomic.Bool

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut && divert.CompareAndSwap(true, false) {
			other := r.Clone(r.Context())
			other.Method = http.MethodGet
			other.Body = http.NoBody
			other.URL.Path = "/spaces/space/webhook_definitions/foreign"
			fixture.server.ServeHTTP(w, other)

			return
		}

		fixture.ServeHTTP(w, r)
	})

	var observations webhookProtocolObservations

	priorPassword, priorHeader, nextPassword, nextHeader := rand.Text(), rand.Text(), rand.Text(), rand.Text()
	hcl := func(password, header string) string {
		return webhookBasicAuthConfig("Managed", `"user"`, strconv.Quote(password), fmt.Sprintf(`headers={
"X-Managed"={secret=true,value=sensitive(%q)}
"X-Sibling"={secret=true,value_wo=var.sibling}
}`, header)) + "\nvariable \"sibling\" {\n type=string\n ephemeral=true\n sensitive=true\n}"
	}
	inputs := config.Variables{"sibling": config.StringVariable(webhookSiblingHeader)}
	create := resource.TestStep{Config: hcl(priorPassword, priorHeader), ConfigVariables: inputs}

	var before webhookProtocolResult

	rotate := resource.TestStep{Config: hcl(nextPassword, nextHeader), ConfigVariables: inputs, ExpectError: regexp.MustCompile("Contentful returned a different webhook identity")}
	rotate.PreConfig = func() { before = observations.lastApply(); divert.Store(true) }
	retry := resource.TestStep{Config: rotate.Config, ConfigVariables: inputs, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionUpdate)}}}
	retry.PreConfig = func() {
		response := observations.lastApply()
		prior, after := before.model(t), response.model(t)
		require.True(t, after.Version.IsNull())
		require.NotContains(t, response.privateFields(t), "version")

		prior.Version = after.Version
		require.True(t, prior.Name.Equal(after.Name), "the foreign name must not enter the target checkpoint")
		require.Equal(t, prior, after)
		require.Equal(t, before.verifiers(t), response.verifiers(t))

		stored, ok := fixture.server.StoredWebhookDefinition("space", prior.WebhookID.ValueString())
		require.True(t, ok)
		require.Equal(t, 0, stored.Sys.Version)
		require.Equal(t, stored.HttpBasicPassword.Or(""), priorPassword)
	}
	retry.PostApplyFunc = func() {
		webhookID := before.model(t).WebhookID.ValueString()
		writes := fixture.mutations()
		require.Len(t, writes, 2)
		require.Equal(t, "0", writes[1].version)
		require.Equal(t, "/spaces/space/webhook_definitions/"+webhookID, writes[1].path)

		headers := []map[string]any{{"key": "X-Managed", "secret": true, "value": nextHeader}, {"key": "X-Sibling", "secret": true}}
		require.Equal(t, testJSON(testJSONObject(writes[1].body)["headers"]), testJSON(headers))

		stored, ok := fixture.server.StoredWebhookDefinition("space", webhookID)
		require.True(t, ok)
		require.True(t, stored.HttpBasicPassword.Or("") == nextPassword && stored.Headers[0].Value.Or("") == nextHeader && stored.Headers[1].Value.Or("") == webhookSiblingHeader)
	}
	testObservedWebhookResource(t, handler, &observations, []resource.TestStep{create, rotate, retry})
}

func TestAccWebhookResourceRecoveryForeignBasicUsername(t *testing.T) {
	t.Parallel()
	fixture := newSeededWebhookFixture(t)

	var observations webhookProtocolObservations

	initial := importedWebhookRecoveryStep(t, fixture)
	renamed := strings.Replace(webhookRecoveryConfig(true, ""), `name     = "Imported"`, `name     = "Renamed"`, 1)
	foreign := func(doc map[string]any) {
		sys, _ := doc["sys"].(map[string]any)
		sys["id"] = "foreign"
		sys["createdAt"] = "2000-01-01T00:00:00Z"
		doc["httpBasicUsername"] = "other-user"
	}
	update := webhookRecoveryStep(renamed, webhookPriorPassword, webhookPriorHeader)
	update.PreConfig = func() { webhookReplyFields(fixture, http.MethodPut, foreign) }
	update.ExpectError = regexp.MustCompile("Contentful returned a different webhook identity")

	read := webhookRecoveryStep(renamed, webhookPriorPassword, webhookPriorHeader)
	read.PreConfig = func() {
		prior := observations.firstApply()
		reply := observations.lastApply()
		require.Equal(t, "user", reply.model(t).HTTPBasicUsername.ValueString())
		require.Equal(t, prior.verifiers(t), reply.verifiers(t))
		created := prior.privateFields(t)["created_at"]
		require.NotEmpty(t, created)
		require.Equal(t, created, reply.privateFields(t)["created_at"], "a foreign PUT cannot establish a new creation time")
		fixture.requirePuts(t, webhookExpectedRecoveryPut("0", webhookPriorPassword, webhookPriorHeader, true), webhookPUTWithFields("1", webhookPriorPassword, "", map[string]any{"name": "Renamed"}))
		fixture.requireStored(t, webhookPriorPassword, webhookPriorHeader, true)
		webhookReplyFields(fixture, http.MethodGet, foreign)
	}
	read.ExpectError = regexp.MustCompile("Unexpected Contentful read identity")

	refresh := webhookRecoveryStep(renamed, webhookPriorPassword, webhookPriorHeader)
	refresh.PreConfig = func() {
		reply := observations.lastRead()
		require.Equal(t, observations.lastApply().private, reply.priorPrivate)
		require.Equal(t, reply.priorPrivate, reply.private, "a foreign GET cannot replace the target's private state")
	}
	refresh.ConfigPlanChecks = resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionNoop)}}
	testObservedWebhookResource(t, fixture, &observations, []resource.TestStep{initial, update, read, refresh})
}
