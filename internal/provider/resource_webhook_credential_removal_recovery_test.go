package provider_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/stretchr/testify/require"
)

func TestAccWebhookResourceReassertsHeaderMissingFromPUTResponse(t *testing.T) {
	t.Parallel()

	server := newTestCMAServer(t)
	server.RegisterSpaceEnvironment("space", "master")
	server.SetWebhookDefinition("space", "imported", cm.WebhookDefinitionData{
		Name: "Imported", URL: "https://example.invalid/webhook", Topics: []string{"Entry.publish"}, Active: cm.NewOptBool(true),
	})
	fixture := &webhookSecretRecoveryFixture{server: server}
	config := func(value string) string {
		return fmt.Sprintf(`
import {
  to = contentful_webhook.test
  id = "space/imported"
}
resource "contentful_webhook" "test" {
  space_id = "space"
  name = "Imported"
  url = "https://example.invalid/webhook"
  topics = ["Entry.publish"]
  http_basic_username = "user"
  http_basic_password_wo = %q
  headers = {
    X-Managed = { secret = true, value_wo = %q }
    X-Sibling = { secret = true, value_wo = %q }
  }
}
`, webhookPriorPassword, value, webhookSiblingHeader)
	}

	testAccMockedResource(t, fixture, resource.TestCase{Steps: []resource.TestStep{
		{Config: config(webhookPriorHeader)},
		{
			Config: config(webhookRotatedHeader),
			PreConfig: func() {
				fixture.onNextResponse(http.MethodPut, func(w http.ResponseWriter, rec *httptest.ResponseRecorder) {
					_, _ = io.WriteString(w, mutateTestJSON(rec.Body.String(), func(doc map[string]any) {
						doc["headers"] = []any{map[string]any{"key": "X-Sibling", "secret": true}}
					}))
				})
			},
			ExpectError: regexp.MustCompile("Contentful returned different webhook headers"),
		},
		{
			Config: config(webhookPriorHeader),
			PreConfig: func() {
				sibling := cm.WebhookDefinitionRequestHeader{Key: "X-Sibling", Secret: cm.NewOptBool(true)}
				fixture.updateRemoteHeaders(t, 2, cm.WebhookDefinitionRequestHeaders{sibling})
				fixture.updateRemoteHeaders(t, 3, cm.WebhookDefinitionRequestHeaders{sibling, {Key: "X-Managed", Secret: cm.NewOptBool(true), Value: cm.NewOptNilString("externally-recreated")}})
			},
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionUpdate),
			}},
			PostApplyFunc: func() {
				writes := fixture.mutations()
				require.Len(t, writes, 3)
				require.Equal(t, webhookExpectedRecoveryPut("4", webhookPriorPassword, webhookPriorHeader, true), writes[2])
				fixture.requireStored(t, webhookPriorPassword, webhookPriorHeader, true)
			},
		},
	}})
}

type basicRecoveryCase struct {
	name       string
	remove     bool
	edit       func(map[string]any)
	diagnostic string
	wantUser   string
	keepHash   bool
	recreate   bool
	protected  bool
}

func TestAccWebhookResourceRecoversBasicCredentialDescriptors(t *testing.T) {
	t.Parallel()

	const (
		unexpectedName = "Contentful returned a different webhook name"
		unexpectedUser = "Contentful returned a different webhook Basic username"
	)
	for _, test := range []basicRecoveryCase{
		{name: "changed username", edit: func(doc map[string]any) { doc["name"] = "Unexpected name" }, diagnostic: unexpectedName, wantUser: "new-user"},
		{name: "omitted username", edit: func(doc map[string]any) { delete(doc, "httpBasicUsername") }, diagnostic: unexpectedUser, wantUser: "user", keepHash: true},
		{name: "null username omitted password", edit: func(doc map[string]any) { doc["httpBasicUsername"] = nil }, diagnostic: unexpectedUser, recreate: true},
		{name: "null username null password", edit: func(doc map[string]any) { doc["httpBasicUsername"] = nil; doc["httpBasicPassword"] = nil }, diagnostic: unexpectedUser, recreate: true},
		{name: "null username readable password", edit: func(doc map[string]any) {
			doc["httpBasicUsername"] = nil
			doc["httpBasicPassword"] = webhookRotatedPassword
		}, diagnostic: unexpectedUser, wantUser: "user", keepHash: true, protected: true},
		{name: "confirmed removal", remove: true, edit: func(doc map[string]any) { doc["name"] = "Unexpected name" }, diagnostic: unexpectedName, recreate: true},
		{name: "successful removal with omitted username", remove: true, edit: func(doc map[string]any) { delete(doc, "httpBasicUsername") }, recreate: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			fixture := newSeededWebhookFixture(t)

			var observations webhookProtocolObservations

			config := webhookRecoveryConfig(true, "")
			initial := importedWebhookRecoveryStep(t, fixture)
			changed := strings.Replace(config, `http_basic_username    = "user"`, `http_basic_username    = "new-user"`, 1)
			targetUser := "new-user"
			attempt := webhookPUTWithFields("1", webhookRotatedPassword, "", map[string]any{"httpBasicUsername": targetUser})

			if test.remove {
				changed = strings.Replace(config, "  http_basic_username    = \"user\"\n  http_basic_password_wo = var.password\n", "", 1)
				attempt = webhookPUTWithFields("1", "", "", map[string]any{"httpBasicUsername": nil, "httpBasicPassword": nil})
			}

			update := webhookRecoveryStep(changed, webhookRotatedPassword, webhookPriorHeader)

			var prior webhookPrivateVerifiers

			update.PreConfig = func() {
				prior = observations.lastApply().verifiers(t)

				webhookReplyFields(fixture, http.MethodPut, test.edit)
			}
			if test.diagnostic != "" {
				update.ExpectError = regexp.MustCompile(test.diagnostic)
			}

			if test.recreate {
				targetUser = "user"
			}

			retry := webhookRecoveryStep(strings.Replace(config, `http_basic_username    = "user"`, fmt.Sprintf("http_basic_username    = %q", targetUser), 1), webhookPriorPassword, webhookPriorHeader)
			retry.PreConfig = func() {
				response := observations.lastApply()
				test.checkReply(t, response, prior)

				fixture.requirePuts(t, webhookExpectedRecoveryPut("0", webhookPriorPassword, webhookPriorHeader, true), attempt)

				if test.remove {
					stored, ok := fixture.server.StoredWebhookDefinition("space", "imported")
					require.True(t, ok)
					require.Nil(t, stored.HttpBasicUsername.ValueStringPointer())
					require.Nil(t, stored.HttpBasicPassword.ValueStringPointer())
				}

				if test.recreate {
					setRemoteWebhookBasic(t, fixture, 2, targetUser, webhookRotatedPassword)
				}
			}
			retry.PostApplyFunc = func() {
				version := "2"
				if test.recreate {
					version = "3"
				}

				fixture.requirePuts(t, webhookExpectedRecoveryPut("0", webhookPriorPassword, webhookPriorHeader, true), attempt, webhookPUTWithFields(version, webhookPriorPassword, "", map[string]any{"httpBasicUsername": targetUser}))
				fixture.requireStored(t, webhookPriorPassword, webhookPriorHeader, true)
			}
			testObservedWebhookResource(t, fixture, &observations, []resource.TestStep{initial, update, retry})
		})
	}
}

func (test basicRecoveryCase) checkReply(t *testing.T, response webhookProtocolResult, prior webhookPrivateVerifiers) {
	t.Helper()
	model := response.model(t)

	expected := webhookPrivateVerifiers{Headers: prior.Headers}
	if test.keepHash {
		expected.Basic = prior.Basic
	}

	require.Equal(t, expected, response.verifiers(t))
	require.Equal(t, int64(2), model.Version.ValueInt64())
	require.True(t, model.HTTPBasicPassword.IsNull())
	require.True(t, model.HTTPBasicPasswordWO.IsNull())
	require.Equal(t, test.wantUser, model.HTTPBasicUsername.ValueString())

	if test.wantUser == "" {
		require.True(t, model.HTTPBasicUsername.IsNull())
	}

	if test.protected {
		require.Contains(t, testJSON(response.diagnostics), "Unexpected webhook Basic password in response")
		requireNoWebhookRecoverySecrets(t, testJSON(response.diagnostics))
	}
}

func TestAccWebhookResourceRecoveryAfterHeaderRemoval(t *testing.T) {
	t.Parallel()

	for _, secret := range []bool{true, false} {
		t.Run(fmt.Sprintf("recreated secret=%t", secret), func(t *testing.T) {
			t.Parallel()
			fixture := newSeededWebhookFixture(t)

			var observations webhookProtocolObservations

			config := webhookRecoveryConfig(true, "")
			initial := importedWebhookRecoveryStep(t, fixture)
			removed := webhookPUTWithFields("1", webhookPriorPassword, "", map[string]any{"headers": []map[string]any{{"key": "X-Sibling", "secret": true}}})
			update := webhookRecoveryStep(strings.Replace(config, `"X-Managed" = { secret=true, value_wo=var.header }`, "", 1), webhookPriorPassword, webhookPriorHeader)
			update.PreConfig = func() { contradictNextWebhookName(fixture) }
			update.ExpectError = regexp.MustCompile("Contentful returned a different webhook name")
			value := webhookRotatedHeader

			if !secret {
				config = webhookRecoveryConfig(true, "X-Managed")
				value = "public-replacement"
			}

			reconcile := webhookRecoveryStep(config, webhookPriorPassword, webhookPriorHeader)
			reconcile.PreConfig = func() {
				response := observations.lastApply()
				require.Equal(t, webhookPrivateVerifiers{Basic: observations.firstApply().verifiers(t).Basic}, response.verifiers(t))
				headers := response.model(t).Headers.Elements()
				require.Len(t, headers, 1)
				require.True(t, headers["X-Sibling"].Value().Secret.ValueBool())
				require.True(t, headers["X-Sibling"].Value().Value.IsNull())
				require.True(t, headers["X-Sibling"].Value().ValueWO.IsNull())
				fixture.requirePuts(t, webhookExpectedRecoveryPut("0", webhookPriorPassword, webhookPriorHeader, true), removed)
				stored, ok := fixture.server.StoredWebhookDefinition("space", "imported")
				require.True(t, ok)
				require.Len(t, stored.Headers, 1)
				require.Equal(t, webhookSiblingHeader, stored.Headers[0].Value.Or(""))
				fixture.updateRemoteHeaders(t, 2, cm.WebhookDefinitionRequestHeaders{
					{Key: "X-Managed", Secret: cm.NewOptBool(secret), Value: cm.NewOptNilString(value)},
					{Key: "X-Sibling", Secret: cm.NewOptBool(true)},
				})
			}
			reconcile.PostApplyFunc = func() {
				expected := []webhookMutation{webhookExpectedRecoveryPut("0", webhookPriorPassword, webhookPriorHeader, true), removed}
				if secret {
					expected = append(expected, webhookExpectedRecoveryPut("3", webhookPriorPassword, webhookPriorHeader, true))
					fixture.requireStored(t, webhookPriorPassword, webhookPriorHeader, true)
				}

				fixture.requirePuts(t, expected...)
			}
			testObservedWebhookResource(t, fixture, &observations, []resource.TestStep{initial, update, reconcile})
		})
	}
}

func TestAccWebhookResourceRecoveryMissingCredentialDescriptors(t *testing.T) {
	t.Parallel()
	fixture := newSeededWebhookFixture(t)

	var observations webhookProtocolObservations

	config := webhookRecoveryConfig(true, "")
	initial := importedWebhookRecoveryStep(t, fixture)
	remove := strings.Replace(config, "  http_basic_username    = \"user\"\n  http_basic_password_wo = var.password\n", "", 1)
	remove = strings.Replace(remove, `"X-Managed" = { secret=true, value_wo=var.header }`, "", 1)
	update := webhookRecoveryStep(remove, webhookPriorPassword, webhookPriorHeader)
	update.PreConfig = func() {
		webhookReplyFields(fixture, http.MethodPut, func(doc map[string]any) {
			doc["name"] = "Unexpected name"
			delete(doc, "headers")
			delete(doc, "httpBasicUsername")
		})
	}
	update.ExpectError = regexp.MustCompile("Contentful returned a different webhook name")
	reconcile := webhookRecoveryStep(remove, webhookPriorPassword, webhookPriorHeader)
	reconcile.PreConfig = func() {
		result := observations.lastApply()
		model := result.model(t)
		require.Equal(t, "user", model.HTTPBasicUsername.ValueString())
		headers := model.Headers.Elements()
		require.Len(t, headers, 2)

		for _, name := range []string{"X-Managed", "X-Sibling"} {
			h := headers[name].Value()
			require.True(t, h.Secret.ValueBool())
			require.True(t, h.Value.IsNull() && h.ValueWO.IsNull())
		}

		require.Equal(t, observations.firstApply().verifiers(t), result.verifiers(t))
		fixture.requirePuts(t, webhookExpectedRecoveryPut("0", webhookPriorPassword, webhookPriorHeader, true), webhookPUTWithFields("1", "", "", map[string]any{"httpBasicUsername": nil, "httpBasicPassword": nil, "headers": []map[string]any{{"key": "X-Sibling", "secret": true}}}))
	}
	reconcile.PostApplyFunc = func() { require.Equal(t, webhookPrivateVerifiers{}, observations.lastRead().verifiers(t)) }
	testObservedWebhookResource(t, fixture, &observations, []resource.TestStep{initial, update, reconcile})
}
