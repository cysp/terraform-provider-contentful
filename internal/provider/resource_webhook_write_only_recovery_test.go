package provider_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"
)

func TestAccWebhookResourceWriteOnlyLostUpdate(t *testing.T) {
	t.Parallel()
	fixture := newSeededWebhookFixture(t)

	var observed webhookProtocolObservations

	hcl := webhookRecoveryConfig(true, "")
	first := webhookExpectedRecoveryPut("0", webhookPriorPassword, webhookPriorHeader, true)
	attempted := webhookExpectedRecoveryPut("1", webhookRotatedPassword, webhookRotatedHeader, true)

	var prior webhookPrivateVerifiers

	rotate := webhookRecoveryStep(hcl, webhookRotatedPassword, webhookRotatedHeader)
	rotate.PreConfig = func() {
		prior = observed.lastApply().verifiers(t)
		require.NotNil(t, prior.Basic)
		require.NotEmpty(t, prior.Headers)
		// The fixture commits the request before returning an ambiguous response.
		fixture.onNextResponse(http.MethodPut, func(w http.ResponseWriter, _ *httptest.ResponseRecorder) {
			http.Error(w, "fixture discarded acknowledged response", http.StatusServiceUnavailable)
		})
	}
	rotate.ExpectError = regexp.MustCompile("Failed to update webhook")

	reverted := webhookRecoveryStep(hcl, webhookPriorPassword, webhookPriorHeader)
	reverted.PlanOnly = true
	reverted.PreConfig = func() {
		require.Equal(t, prior, observed.lastApply().verifiers(t))
		fixture.requirePuts(t, first, attempted)
		fixture.requireStored(t, webhookRotatedPassword, webhookRotatedHeader, true)
	}

	// Reconcile the header while the password remains externally owned.
	delegated := webhookExpectedRecoveryPut("2", "", webhookRotatedHeader, true)
	delegated.body = mutateTestJSON(delegated.body, func(doc map[string]any) {
		delete(doc, "httpBasicUsername")
		delete(doc, "httpBasicPassword")
	})
	header := webhookRecoveryStep(strings.Replace(hcl, "http_basic_password_wo = var.password", "", 1), webhookPriorPassword, webhookRotatedHeader)
	header.PostApplyFunc = func() {
		fixture.requirePuts(t, first, attempted, delegated)
		require.Equal(t, prior.Basic, observed.lastApply().verifiers(t).Basic)
		fixture.requireStored(t, webhookRotatedPassword, webhookRotatedHeader, true)
	}

	priorPassword := webhookRecoveryStep(hcl, webhookPriorPassword, webhookRotatedHeader)
	priorPassword.PlanOnly = true
	acknowledged := webhookExpectedRecoveryPut("2", webhookRotatedPassword, "", true)
	confirm := webhookRecoveryStep(hcl, webhookRotatedPassword, webhookRotatedHeader)
	confirm.PostApplyFunc = func() {
		fixture.requirePuts(t, first, attempted, delegated, acknowledged)
		require.NotEqual(t, prior, observed.lastApply().verifiers(t))
		fixture.requireStored(t, webhookRotatedPassword, webhookRotatedHeader, true)
	}

	removeSibling := webhookRecoveryStep(webhookRecoveryConfig(false, ""), webhookRotatedPassword, webhookRotatedHeader)
	removeSibling.PostApplyFunc = func() {
		fixture.requirePuts(t, first, attempted, delegated, acknowledged, webhookExpectedRecoveryPut("2", webhookRotatedPassword, "", false))
		fixture.requireStored(t, webhookRotatedPassword, webhookRotatedHeader, false)
	}
	testObservedWebhookResource(t, fixture, &observed, []resource.TestStep{
		importedWebhookRecoveryStep(t, fixture), rotate, reverted, header, priorPassword, confirm, removeSibling,
	})
}

func TestAccWebhookResourceWriteOnlyCreateRecovery(t *testing.T) {
	t.Parallel()

	runtime, fixture := newWebhookRecoveryFixture(t)
	runtime.writeConfig(t, webhookRecoveryConfig(false, ""))
	fixture.onNextResponse(http.MethodPost, func(w http.ResponseWriter, rec *httptest.ResponseRecorder) {
		body := mutateTestJSON(rec.Body.String(), func(document map[string]any) {
			document["name"] = "Returned name"
		})
		w.WriteHeader(rec.Code)
		_, _ = io.WriteString(w, body)
	})

	output, err := runWebhookRecoveryCLI(t, runtime, webhookPriorPassword, webhookPriorHeader, "apply", "-auto-approve")
	require.Error(t, err, output)
	require.Contains(t, output, "Contentful returned a different webhook name")

	created := webhookExpectedRecoveryPut("", webhookPriorPassword, webhookPriorHeader, false)
	created.method = http.MethodPost
	created.path = "/spaces/space/webhook_definitions"
	require.Equal(t, []webhookMutation{created}, fixture.mutations())

	instance := readWebhookRecoveryInstance(t, runtime)
	require.Equal(t, "tainted", instance.Status)
	require.Equal(t, "Returned name", instance.Attributes["name"])
	require.Empty(t, instance.verifiers)
	webhookID, ok := instance.Attributes["webhook_id"].(string)
	require.True(t, ok)
	require.Equal(t, "space/"+webhookID, instance.Attributes["id"])
	stored, exists := fixture.server.StoredWebhookDefinition("space", webhookID)
	require.True(t, exists)
	require.Equal(t, cm.NewOptNilString(webhookPriorPassword), stored.HttpBasicPassword)
}

// The effective ordinary value authorizes publication of those bytes. A different
// readable response at the former write-only path could contain the old secret.
func TestAccWebhookResourceWriteOnlyHeaderBecomesOrdinary(t *testing.T) {
	t.Parallel()
	fixture := newSeededWebhookFixture(t)

	var (
		observations webhookProtocolObservations
		prior        webhookPrivateVerifiers
	)

	config := webhookRecoveryConfig(true, "")
	// Only the Terraform storage mode changes. CMA's secret flag stays true.
	ordinary := webhookRecoveryStep(strings.Replace(config, "value_wo=var.header", `value="public-replacement"`, 1), webhookPriorPassword, webhookPriorHeader)
	ordinary.PreConfig = func() {
		prior = observations.lastApply().verifiers(t)
		require.Contains(t, prior.Headers, "X-Managed")
		webhookReplyFields(fixture, http.MethodPut, func(doc map[string]any) {
			doc["headers"] = []map[string]any{
				{"key": "X-Managed", "secret": false, "value": webhookPriorHeader},
				{"key": "X-Sibling", "secret": true},
			}
		})
	}
	ordinary.ExpectError = regexp.MustCompile("(?s)Error running apply:.*Unexpected webhook secret header in response")
	retry := webhookRecoveryStep(ordinary.Config, webhookPriorPassword, webhookPriorHeader)
	retry.PreConfig = func() {
		reply := observations.lastApply()
		require.Equal(t, prior, reply.verifiers(t))
		require.True(t, reply.model(t).Headers.Elements()["X-Managed"].Value().Value.IsNull())

		writes := fixture.mutations()
		require.Len(t, writes, 2)
		require.Equal(t, "1", writes[1].version)
		require.Equal(t, []any{
			map[string]any{"key": "X-Managed", "secret": true, "value": "public-replacement"},
			map[string]any{"key": "X-Sibling", "secret": true},
		}, testJSONObject(writes[1].body)["headers"])
	}
	testObservedWebhookResource(t, fixture, &observations, []resource.TestStep{importedWebhookRecoveryStep(t, fixture), ordinary, retry})
}

// A case-only rename has a new Terraform key. Publishing an explicitly
// configured ordinary value at that spelling is safe when the response agrees.
func TestAccWebhookResourceWriteOnlyCaseRenameToOrdinary(t *testing.T) {
	t.Parallel()

	runtime, fixture, first := importWebhookRecoveryFixture(t)
	runtime.writeConfig(t, webhookRecoveryConfig(true, "x-managed"))

	output, err := runWebhookRecoveryCLI(t, runtime, webhookPriorPassword, webhookPriorHeader, "apply", "-auto-approve")
	require.NoError(t, err, output)

	update := webhookExpectedRecoveryPut("1", webhookPriorPassword, "", true)
	update.body = mutateTestJSON(update.body, func(document map[string]any) {
		document["headers"] = []map[string]any{
			{"key": "X-Sibling", "secret": true},
			{"key": "x-managed", "secret": false, "value": "public-replacement"},
		}
	})
	fixture.requirePuts(t, first, update)

	instance := readWebhookRecoveryInstance(t, runtime)
	require.Equal(t, map[string]any{
		"X-Sibling": map[string]any{"secret": true, "value": nil, "value_wo": nil},
		"x-managed": map[string]any{"secret": false, "value": "public-replacement", "value_wo": nil},
	}, instance.Attributes["headers"])
}

func TestAccWebhookResourceWriteOnlyHeaderRecreatedAfterRefresh(t *testing.T) {
	t.Parallel()

	runtime, fixture, first := importWebhookRecoveryFixture(t)

	sibling := cm.WebhookDefinitionRequestHeader{Key: "X-Sibling", Secret: cm.NewOptBool(true), Value: cm.NewOptNilString(webhookSiblingHeader)}
	fixture.updateRemoteHeaders(t, 1, cm.WebhookDefinitionRequestHeaders{sibling})

	output, err := runWebhookRecoveryCLI(t, runtime, webhookPriorPassword, webhookPriorHeader, "apply", "-refresh-only", "-auto-approve")
	require.NoError(t, err, output)
	require.Equal(t, map[string]any{"X-Sibling": map[string]any{"secret": true, "value": nil, "value_wo": nil}}, readWebhookRecoveryInstance(t, runtime).Attributes["headers"])
	fixture.requirePuts(t, first)

	fixture.updateRemoteHeaders(t, 2, cm.WebhookDefinitionRequestHeaders{{Key: "X-Managed", Secret: cm.NewOptBool(true), Value: cm.NewOptNilString("externally-recreated-secret")}, sibling})

	output, err = runWebhookRecoveryCLI(t, runtime, webhookPriorPassword, webhookPriorHeader, "apply", "-auto-approve")
	require.NoError(t, err, output)
	fixture.requirePuts(t, first, webhookExpectedRecoveryPut("3", webhookPriorPassword, webhookPriorHeader, true))
	fixture.requireStored(t, webhookPriorPassword, webhookPriorHeader, true)
}

func (fixture *webhookSecretRecoveryFixture) updateRemoteHeaders(t *testing.T, version int, headers cm.WebhookDefinitionRequestHeaders) {
	t.Helper()
	response, err := fixture.server.Handler().UpdateWebhookDefinition(t.Context(), &cm.WebhookDefinitionData{
		Name: "Imported", URL: "https://example.invalid/webhook", Topics: []string{"Entry.publish"}, Active: cm.NewOptBool(true),
		HttpBasicUsername: cm.NewOptNilString("user"), HttpBasicPassword: cm.NewOptNilString(webhookPriorPassword),
		Headers: headers,
	}, cm.UpdateWebhookDefinitionParams{SpaceID: "space", WebhookDefinitionID: "imported", XContentfulVersion: cm.NewOptInt(version)})
	require.NoError(t, err)
	require.IsType(t, &cm.WebhookDefinitionStatusCode{}, response)
}

func TestAccWebhookResourceDelegatesHeaderByName(t *testing.T) {
	t.Parallel()
	runtime, fixture, first := importWebhookRecoveryFixture(t)

	const outside = "synthetic-externally-rotated-sibling"
	fixture.updateRemoteHeaders(t, 1, cm.WebhookDefinitionRequestHeaders{
		{Key: "X-Managed", Secret: cm.NewOptBool(true)},
		{Key: "X-Sibling", Secret: cm.NewOptBool(true), Value: cm.NewOptNilString(outside)},
	})

	runtime.writeConfig(t, webhookRecoveryHeaderConfig("", "", `lifecycle { ignore_changes = [headers["X-Sibling"]] }`))
	output, err := runWebhookRecoveryCLI(t, runtime, webhookPriorPassword, webhookRotatedHeader, "apply", "-auto-approve")
	require.NoError(t, err, output)
	fixture.requirePuts(t, first, webhookExpectedRecoveryPut("2", webhookPriorPassword, webhookRotatedHeader, true))
	stored, exists := fixture.server.StoredWebhookDefinition("space", "imported")
	require.True(t, exists)
	require.Equal(t, cm.WebhookDefinitionHeaders{
		{Key: "X-Managed", Secret: cm.NewOptBool(true), Value: cm.NewOptString(webhookRotatedHeader)},
		{Key: "X-Sibling", Secret: cm.NewOptBool(true), Value: cm.NewOptString(outside)},
	}, stored.Headers)
	output, err = runWebhookRecoveryCLI(t, runtime, webhookPriorPassword, webhookRotatedHeader, "plan", "-detailed-exitcode")
	require.NoError(t, err, output)

	// Ignoring the write-only leaf suppresses a rotation-only plan. An update
	// scheduled by a different argument still receives apply-time Config.
	runtime.writeConfig(t, webhookRecoveryHeaderConfig("", "", `lifecycle { ignore_changes = [headers["X-Sibling"], headers["X-Managed"].value_wo] }`))
	output, err = runWebhookRecoveryCLI(t, runtime, webhookPriorPassword, webhookPriorHeader, "plan", "-detailed-exitcode")
	require.NoError(t, err, output)
	output, err = runWebhookRecoveryCLI(t, runtime, webhookRotatedPassword, webhookPriorHeader, "apply", "-auto-approve")
	require.NoError(t, err, output)
	fixture.requirePuts(t, first, webhookExpectedRecoveryPut("2", webhookPriorPassword, webhookRotatedHeader, true), webhookExpectedRecoveryPut("3", webhookRotatedPassword, webhookPriorHeader, true))
}

func TestAccWebhookResourceWriteOnlyStalePlan(t *testing.T) {
	t.Parallel()
	runtime, fixture, first := importWebhookRecoveryFixture(t)
	before := readWebhookRecoveryInstance(t, runtime)
	out, err := runWebhookRecoveryCLI(t, runtime, webhookRotatedPassword, webhookRotatedHeader, "plan", "-out=rotate.tfplan")
	require.NoError(t, err, out)
	fixture.updateRemoteHeaders(t, 1, cm.WebhookDefinitionRequestHeaders{
		{Key: "X-Managed", Secret: cm.NewOptBool(true)},
		{Key: "X-Sibling", Secret: cm.NewOptBool(true), Value: cm.NewOptNilString("externally-rotated")},
	})
	out, err = runWebhookRecoveryCLI(t, runtime, webhookRotatedPassword, webhookRotatedHeader, "apply", "rotate.tfplan")
	require.Error(t, err, out)
	require.Contains(t, out, "HTTP 409")
	fixture.requirePuts(t, first, webhookExpectedRecoveryPut("1", webhookRotatedPassword, webhookRotatedHeader, true))
	after := readWebhookRecoveryInstance(t, runtime)
	require.Equal(t, before.Attributes, after.Attributes)
	require.Equal(t, before.Private, after.Private)

	fields, found := fixture.server.StoredWebhookDefinition("space", "imported")
	require.True(t, found)
	require.Equal(t, 2, fields.Sys.Version)
	require.Equal(t, cm.NewOptNilString(webhookPriorPassword), fields.HttpBasicPassword)
	require.Equal(t, cm.WebhookDefinitionHeaders{
		{Key: "X-Managed", Secret: cm.NewOptBool(true), Value: cm.NewOptString(webhookPriorHeader)},
		{Key: "X-Sibling", Secret: cm.NewOptBool(true), Value: cm.NewOptString("externally-rotated")},
	}, fields.Headers)
}

func TestAccWebhookResourceWriteOnlyReadableResponse(t *testing.T) {
	t.Parallel()

	for _, method := range []string{http.MethodPut, http.MethodGet} {
		for _, name := range []string{"", "X-Managed", "x-managed"} {
			t.Run(method+"/"+name, func(t *testing.T) {
				t.Parallel()
				fixture := newSeededWebhookFixture(t)

				var observations webhookProtocolObservations

				initial := importedWebhookRecoveryStep(t, fixture)
				config := webhookRecoveryConfig(true, "")
				echo := func(doc map[string]any) {
					if name == "" {
						doc["httpBasicPassword"] = webhookRotatedPassword
						delete(doc, "httpBasicUsername")

						return
					}

					doc["headers"] = []map[string]any{{"key": name, "secret": false, "value": webhookRotatedHeader}, {"key": "X-Sibling", "secret": true}}
				}
				diagnostic := regexp.MustCompile("Unexpected webhook (Basic password|secret header) in response")
				fault := webhookRecoveryStep(config, webhookRotatedPassword, webhookRotatedHeader)
				fault.PreConfig = func() { webhookReplyFields(fixture, method, echo) }
				fault.ExpectError = diagnostic
				steps := []resource.TestStep{initial, fault}
				checkSafe := func(response webhookProtocolResult, operation string) {
					require.Contains(t, response.logs, operation)
					requireNoWebhookRecoverySecrets(t, response.logs)
					requireNoWebhookRecoverySecrets(t, testJSON(response.diagnostics))

					for _, field := range response.privateFields(t) {
						requireNoWebhookRecoverySecrets(t, string(field))
					}

					model := response.model(t)
					require.True(t, model.HTTPBasicPassword.IsNull() && model.HTTPBasicPasswordWO.IsNull())
					requireNoWebhookRecoverySecrets(t, model.Headers.String())
				}

				if method == http.MethodPut {
					repeat := webhookRecoveryStep(config, webhookRotatedPassword, webhookRotatedHeader)
					repeat.ExpectError = diagnostic
					repeat.PreConfig = func() {
						response := observations.lastApply()
						checkSafe(response, "webhook.update")

						expected := observations.firstApply().verifiers(t)
						if name == "x-managed" {
							expected.Headers = nil
						}

						require.Equal(t, expected, response.verifiers(t))
						fixture.requireStored(t, webhookRotatedPassword, webhookRotatedHeader, true)
						webhookReplyFields(fixture, http.MethodGet, echo)
					}
					steps = append(steps, repeat)
				}
				// A clean refresh after the rejected echo can reassert the configured bytes.
				repair := webhookRecoveryStep(config, webhookRotatedPassword, webhookRotatedHeader)

				var priorPrivate []byte

				repair.PreConfig = func() {
					failedRead := observations.lastRead()
					checkSafe(failedRead, "webhook.read")
					priorPrivate = failedRead.priorPrivate
					require.NotEmpty(t, priorPrivate)

					expected := []webhookMutation{webhookExpectedRecoveryPut("0", webhookPriorPassword, webhookPriorHeader, true)}
					if method == http.MethodPut {
						expected = append(expected, webhookExpectedRecoveryPut("1", webhookRotatedPassword, webhookRotatedHeader, true))
					}

					fixture.requirePuts(t, expected...)
				}
				repair.Check = func(*terraform.State) error {
					require.Equal(t, priorPrivate, observations.lastRead().priorPrivate, "Core must retry Read with the checkpoint preceding its rejected echo")

					return nil
				}
				testObservedWebhookResource(t, fixture, &observations, append(steps, repair))
			})
		}
	}
}

func TestAccWebhookResourceOrdinaryHeaderEchoOnUnrelatedContradiction(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"X-Managed", "x-managed"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			fixture := newSeededWebhookFixture(t)

			var observations webhookProtocolObservations

			initial := importedWebhookRecoveryStep(t, fixture)
			exact := name == "X-Managed"

			config := webhookRecoveryConfig(true, name)
			if exact {
				config = strings.Replace(config, "secret=false", "secret=true", 1)
			}

			const value = "public-replacement"

			headers := []map[string]any{{"key": name, "secret": exact, "value": value}, {"key": "X-Sibling", "secret": true}}
			if !exact {
				headers[0], headers[1] = headers[1], headers[0]
			}

			update := webhookRecoveryStep(config, webhookPriorPassword, webhookPriorHeader)
			update.PreConfig = func() {
				webhookReplyFields(fixture, http.MethodPut, func(doc map[string]any) {
					doc["name"] = "Unexpected name"
					if exact {
						doc["headers"] = headers
					}
				})
			}
			update.ExpectError = regexp.MustCompile("Contentful returned a different webhook name")
			refresh := webhookRecoveryStep(config, webhookPriorPassword, webhookPriorHeader)
			refresh.PreConfig = func() {
				fixture.requirePuts(t, webhookExpectedRecoveryPut("0", webhookPriorPassword, webhookPriorHeader, true), webhookPUTWithFields("1", webhookPriorPassword, "", map[string]any{"headers": headers}))

				result := observations.lastApply()
				header := result.model(t).Headers.Elements()[name].Value()
				require.Equal(t, exact, header.Secret.ValueBool())
				require.Equal(t, value, header.Value.ValueString())
				require.True(t, header.ValueWO.IsNull())
				require.Equal(t, webhookPrivateVerifiers{Basic: observations.firstApply().verifiers(t).Basic}, result.verifiers(t))

				if exact {
					webhookReplyFields(fixture, http.MethodGet, func(doc map[string]any) { doc["headers"] = headers })
				}
			}
			refresh.ConfigPlanChecks = resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionNoop)}}
			testObservedWebhookResource(t, fixture, &observations, []resource.TestStep{initial, update, refresh})
		})
	}
}
