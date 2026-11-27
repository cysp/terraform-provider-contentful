package provider_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"

	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/stretchr/testify/require"
)

func TestAccWebhookResourceRefreshLowerVersionDropsCredentialHistory(t *testing.T) {
	t.Parallel()

	for _, ordinaryBasic := range []bool{true, false} {
		t.Run(fmt.Sprintf("ordinary_basic=%t", ordinaryBasic), func(t *testing.T) {
			t.Parallel()
			fixture := newSeededWebhookFixture(t)

			var observed webhookProtocolObservations

			hcl := webhookRecoveryConfig(true, "")
			if ordinaryBasic {
				hcl = strings.Replace(hcl, "http_basic_password_wo = var.password", "http_basic_password = "+strconv.Quote(webhookPriorPassword), 1)
			} else {
				hcl = strings.Replace(hcl, "value_wo=var.header", "value=sensitive("+strconv.Quote(webhookPriorHeader)+")", 1)
			}

			mixed := webhookRecoveryStep(hcl, webhookPriorPassword, webhookPriorHeader)
			first := webhookExpectedRecoveryPut("0", webhookPriorPassword, webhookPriorHeader, true)

			var priorPrivate []byte

			mixed.PostApplyFunc = func() {
				reply := observed.lastRead()
				require.EqualValues(t, 1, reply.model(t).Version.ValueInt64())
				require.NotEmpty(t, reply.verifiers(t))
				priorPrivate = reply.private

				fixture.requirePuts(t, first)
			}

			// Only the version is comparable. The new instance may omit createdAt.
			rejected := webhookRecoveryStep(hcl, webhookPriorPassword, webhookPriorHeader)
			rejected.PreConfig = func() {
				recreateWebhookAtVersion(t, fixture, 0)
				webhookReplyFields(fixture, http.MethodGet, func(doc map[string]any) {
					omitWebhookCreationTime(doc)

					if ordinaryBasic {
						doc["headers"] = []map[string]any{{"key": "X-Managed", "secret": false, "value": webhookRotatedHeader}}
					} else {
						doc["httpBasicPassword"] = webhookRotatedPassword
					}
				})
			}
			rejected.ExpectError = regexp.MustCompile("Unexpected webhook")
			retry := webhookRecoveryStep(hcl, webhookPriorPassword, webhookPriorHeader)
			retry.PreConfig = func() { webhookReplyFields(fixture, http.MethodGet, omitWebhookCreationTime) }
			retry.ConfigPlanChecks = resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionUpdate),
				webhookPlanCheck(func(plan *tfjson.Plan) {
					reply := observed.lastRead()
					require.Equal(t, priorPrivate, reply.priorPrivate, "Terraform retained private state after the rejected refresh")
					require.Len(t, plan.ResourceDrift, 1)
					change := plan.ResourceDrift[0].Change
					before, ok := change.Before.(map[string]any)
					require.True(t, ok)
					require.Equal(t, "1", testJSON(before["version"]))

					state, ok := change.After.(map[string]any)
					require.True(t, ok)
					require.Equal(t, "0", testJSON(state["version"]))
					require.Equal(t, "user", state["http_basic_username"])
					require.Nil(t, state["http_basic_password"])

					hidden := map[string]any{"secret": true, "value": nil, "value_wo": nil}
					require.Equal(t, map[string]any{"X-Managed": hidden, "X-Sibling": hidden}, state["headers"])
					private := reply.privateFields(t)
					require.Equal(t, "0", string(private["version"]))
					require.NotContains(t, private, "write_only_secret_hashes")
					require.NotContains(t, private, "created_at")
					fixture.requirePuts(t, first)
				}),
			}}
			retry.PostApplyFunc = func() {
				fixture.requirePuts(t, first, webhookExpectedRecoveryPut("0", webhookPriorPassword, webhookPriorHeader, true))
				fixture.requireStored(t, webhookPriorPassword, webhookPriorHeader, true)
			}
			testObservedWebhookResource(t, fixture, &observed, []resource.TestStep{importedWebhookRecoveryStep(t, fixture), mixed, rejected, retry})
		})
	}
}

func TestAccWebhookResourceRefreshChangedCreationTime(t *testing.T) {
	t.Parallel()

	for _, version := range []int{1, 2} {
		t.Run(strconv.Itoa(version), func(t *testing.T) {
			t.Parallel()
			fixture := newSeededWebhookFixture(t)

			var observed webhookProtocolObservations

			retry := webhookRecoveryStep(webhookRecoveryConfig(true, ""), webhookPriorPassword, webhookPriorHeader)
			retry.PreConfig = func() { recreateWebhookAtVersion(t, fixture, version) }
			retry.ConfigPlanChecks = resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionUpdate),
				webhookPlanCheck(func(*tfjson.Plan) { require.Empty(t, observed.lastRead().verifiers(t)) }),
			}}
			retry.PostApplyFunc = func() {
				fixture.requirePuts(t, webhookExpectedRecoveryPut("0", webhookPriorPassword, webhookPriorHeader, true), webhookExpectedRecoveryPut(strconv.Itoa(version), webhookPriorPassword, webhookPriorHeader, true))
				fixture.requireStored(t, webhookPriorPassword, webhookPriorHeader, true)
			}
			testObservedWebhookResource(t, fixture, &observed, []resource.TestStep{importedWebhookRecoveryStep(t, fixture), retry})
		})
	}
}

func TestAccWebhookResourceSavedPlanCrossesCreationTime(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name string
		echo bool
	}{
		{name: "redacted reply"}, {name: "unsolicited password", echo: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			fixture := newSeededWebhookFixture(t)

			var observed webhookProtocolObservations

			hcl := webhookRecoveryConfig(true, "")

			const rotation = "synthetic-password-after-recreation"

			initial := importedWebhookRecoveryStep(t, fixture)
			update := webhookRecoveryStep(hcl, rotation, webhookPriorHeader)
			update.ConfigPlanChecks = resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionUpdate),
				webhookPlanCheck(func(*tfjson.Plan) {
					recreateWebhookAtVersion(t, fixture, 1)

					if test.echo {
						webhookReplyFields(fixture, http.MethodPut, func(doc map[string]any) {
							delete(doc, "httpBasicUsername")
							doc["httpBasicPassword"] = rotation
						})
					}
				}),
			}}
			update.ExpectError = regexp.MustCompile("Webhook was recreated")
			first := webhookExpectedRecoveryPut("0", webhookPriorPassword, webhookPriorHeader, true)
			attempt := webhookExpectedRecoveryPut("1", rotation, "", true)
			checkRecovery := func() {
				reply := observed.lastApply()
				state := reply.model(t)
				require.Equal(t, "user", state.HTTPBasicUsername.ValueString())
				require.True(t, state.HTTPBasicPassword.IsNull())
				require.EqualValues(t, 2, state.Version.ValueInt64())
				headers := state.Headers.Elements()
				require.Len(t, headers, 2)

				for _, name := range []string{"X-Managed", "X-Sibling"} {
					entry, exists := headers[name]
					require.True(t, exists)
					require.True(t, entry.Value().Secret.ValueBool())
					require.True(t, entry.Value().Value.IsNull())
				}

				private := reply.privateFields(t)
				require.NotContains(t, private, "write_only_secret_hashes")
				require.Equal(t, "2", string(private["version"]))

				stored, exists := fixture.server.StoredWebhookDefinition("space", "imported")
				require.True(t, exists)

				created, present := stored.Sys.CreatedAt.Get()
				require.True(t, present)
				require.JSONEq(t, testJSON(created), string(private["created_at"]))
				fixture.requirePuts(t, first, attempt)
				fixture.requireStored(t, rotation, webhookRotatedHeader, true)
			}
			steps := []resource.TestStep{initial, update}
			password := rotation
			retry := webhookRecoveryStep(hcl, rotation, webhookPriorHeader)
			retry.PreConfig = checkRecovery

			if test.echo {
				// The failed PUT omitted the username. The next refresh must
				// still guard a credential received under that descriptor.
				read := webhookRecoveryStep(hcl, webhookPriorPassword, webhookPriorHeader)
				read.PreConfig = func() {
					checkRecovery()
					requireWebhookEchoDiagnostic(t, observed.lastApply(), rotation)
					webhookReplyFields(fixture, http.MethodGet, func(doc map[string]any) { doc["httpBasicPassword"] = rotation })
				}
				read.ExpectError = regexp.MustCompile("Unexpected webhook Basic password in response")
				steps = append(steps, read)
				password = webhookPriorPassword
				retry = webhookRecoveryStep(hcl, password, webhookPriorHeader)
				retry.PreConfig = func() {
					requireWebhookEchoDiagnostic(t, observed.lastRead(), rotation)
					webhookReplyFields(fixture, http.MethodGet, omitWebhookCreationTime)
				}
			}

			retry.ConfigPlanChecks = resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionUpdate)}}
			retry.PostApplyFunc = func() {
				fixture.requirePuts(t, first, attempt, webhookExpectedRecoveryPut("2", password, webhookPriorHeader, true))
				fixture.requireStored(t, password, webhookPriorHeader, true)
			}
			testObservedWebhookResource(t, fixture, &observed, append(steps, retry))
		})
	}
}

func requireWebhookEchoDiagnostic(t *testing.T, reply webhookProtocolResult, password string) {
	t.Helper()

	found := false
	for _, d := range reply.diagnostics {
		found = found || d.Summary == "Unexpected webhook Basic password in response"
		require.NotContains(t, d.Summary+" "+d.Detail, password)
	}

	require.True(t, found, "expected a Basic password redaction diagnostic")
	require.NotContains(t, reply.logs, password)
}

// Install different hidden bytes at the saved plan's address and version.
// Direct fixture traffic is independent of the provider request recorder.
func recreateWebhookAtVersion(t *testing.T, fixture *webhookSecretRecoveryFixture, version int) {
	t.Helper()

	const path = "/spaces/space/webhook_definitions/imported"

	request := func(method, lock string, body map[string]any, status int) {
		t.Helper()
		req := httptest.NewRequestWithContext(t.Context(), method, path, strings.NewReader(testJSON(body)))
		req.Header.Set("Authorization", "Bearer "+cmt.ValidAccessToken)
		req.Header.Set("Content-Type", "application/vnd.contentful.management.v1+json")

		if lock != "" {
			req.Header.Set("X-Contentful-Version", lock)
		}

		rec := httptest.NewRecorder()
		fixture.server.ServeHTTP(rec, req)
		require.Equal(t, status, rec.Code)
	}
	request(http.MethodDelete, "", nil, http.StatusNoContent)

	body := map[string]any{
		"name": "New instance", "url": "https://example.invalid/webhook", "topics": []string{"Entry.publish"}, "active": true,
		"httpBasicUsername": "user", "httpBasicPassword": webhookRotatedPassword,
		"headers": []map[string]any{{"key": "X-Managed", "secret": true, "value": webhookRotatedHeader}, {"key": "X-Sibling", "secret": true, "value": webhookSiblingHeader}},
	}
	request(http.MethodPut, "", body, http.StatusOK)

	for i := range version {
		body["name"] = fmt.Sprintf("External update %d", i)
		if i == version-1 {
			body["name"] = "Imported"
		}

		request(http.MethodPut, strconv.Itoa(i), body, http.StatusOK)
	}
}

func omitWebhookCreationTime(doc map[string]any) {
	sys, _ := doc["sys"].(map[string]any)
	delete(sys, "createdAt")
}
