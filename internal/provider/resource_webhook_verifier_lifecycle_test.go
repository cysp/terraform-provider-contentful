package provider_test

import (
	"net/http"
	"regexp"
	"strings"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"
)

func contradictNextWebhookName(fixture *webhookSecretRecoveryFixture) {
	webhookReplyFields(fixture, http.MethodPut, func(doc map[string]any) { doc["name"] = "Unexpected name" })
}

func TestAccWebhookResourceRefreshBasicCredentialPresence(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name     string
		edit     func(map[string]any)
		reassert bool
	}{
		{"explicit null password", func(doc map[string]any) { doc["httpBasicPassword"] = nil }, true},
		{"omitted credentials", func(doc map[string]any) { delete(doc, "httpBasicUsername"); delete(doc, "httpBasicPassword") }, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			fixture := newSeededWebhookFixture(t)

			var observations webhookProtocolObservations

			first := importedWebhookRecoveryStep(t, fixture)
			next := webhookRecoveryStep(webhookRecoveryConfig(true, ""), webhookPriorPassword, webhookPriorHeader)
			next.PreConfig = func() { webhookReplyFields(fixture, http.MethodGet, test.edit) }

			action := plancheck.ResourceActionNoop
			if test.reassert {
				action = plancheck.ResourceActionUpdate
			}

			next.ConfigPlanChecks = resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("contentful_webhook.test", action)}}
			next.Check = func(*terraform.State) error {
				// Core may read again for its final plan. Inspect the response to the injected GET.
				prior := observations.firstApply().verifiers(t)
				require.NotNil(t, prior.Basic)
				require.NotEmpty(t, prior.Headers["X-Managed"])
				expected := prior
				writes := []webhookMutation{webhookExpectedRecoveryPut("0", webhookPriorPassword, webhookPriorHeader, true)}

				if test.reassert {
					expected.Basic = nil

					writes = append(writes, webhookExpectedRecoveryPut("1", webhookPriorPassword, "", true))
				}

				read := observations.lastRead()
				require.Equal(t, "user", read.model(t).HTTPBasicUsername.ValueString())
				require.Equal(t, expected, read.verifiers(t))
				fixture.requirePuts(t, writes...)

				return nil
			}
			testObservedWebhookResource(t, fixture, &observations, []resource.TestStep{first, next})
		})
	}
}

func TestAccWebhookResourceRetainsDelegatedHeaderVerifier(t *testing.T) {
	t.Parallel()
	fixture := newSeededWebhookFixture(t)

	var observations webhookProtocolObservations

	config := webhookRecoveryConfig(true, "")
	initial := importedWebhookRecoveryStep(t, fixture)
	delegate := webhookRecoveryStep(strings.Replace(config, ", value_wo=var.header", "", 1), webhookRotatedPassword, webhookPriorHeader)

	var priorHash string

	delegate.PreConfig = func() {
		records := observations.lastApply().verifiers(t)
		require.NotNil(t, records.Basic)
		require.Equal(t, "user", records.Basic.Username)
		require.NotEmpty(t, records.Basic.Hash)
		priorHash = records.Headers["X-Managed"]
		require.NotEmpty(t, priorHash)
	}
	delegate.PostApplyFunc = func() { require.Equal(t, priorHash, observations.lastApply().verifiers(t).Headers["X-Managed"]) }
	resume := webhookRecoveryStep(config, webhookRotatedPassword, webhookPriorHeader)
	resume.ConfigPlanChecks = resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionNoop)}}
	rotate := webhookRecoveryStep(config, webhookRotatedPassword, webhookRotatedHeader)
	rotate.PostApplyFunc = func() {
		fixture.requirePuts(t, webhookExpectedRecoveryPut("0", webhookPriorPassword, webhookPriorHeader, true), webhookExpectedRecoveryPut("1", webhookRotatedPassword, "", true), webhookExpectedRecoveryPut("2", webhookRotatedPassword, webhookRotatedHeader, true))
	}

	const newURL = "https://example.invalid/other-path"

	move := webhookRecoveryStep(strings.Replace(config, "https://example.invalid/webhook", newURL, 1), webhookRotatedPassword, webhookRotatedHeader)
	move.PreConfig = func() {
		fixture.updateRemoteHeaders(t, 3, cm.WebhookDefinitionRequestHeaders{
			{Key: "X-Managed", Secret: cm.NewOptBool(true), Value: cm.NewOptNilString("external-rotation")},
			{Key: "X-Sibling", Secret: cm.NewOptBool(true), Value: cm.NewOptNilString("external-sibling")},
		})
	}
	move.PostApplyFunc = func() {
		writes := fixture.mutations()
		require.Len(t, writes, 4)
		last := writes[3]
		require.Equal(t, "4", last.version)
		body := testJSONObject(last.body)
		require.Equal(t, newURL, body["url"])
		require.Equal(t, []any{map[string]any{"key": "X-Managed", "secret": true, "value": webhookRotatedHeader}, map[string]any{"key": "X-Sibling", "secret": true}}, body["headers"])

		stored, ok := fixture.server.StoredWebhookDefinition("space", "imported")
		require.True(t, ok)
		require.Equal(t, webhookRotatedHeader, stored.Headers[0].Value.Or(""))
		require.Equal(t, "external-sibling", stored.Headers[1].Value.Or(""))
	}
	testObservedWebhookResource(t, fixture, &observations, []resource.TestStep{initial, delegate, resume, rotate, move})
}

func TestAccWebhookResourceRefreshNewlySecretHeader(t *testing.T) {
	t.Parallel()

	const previous = "public-replacement"

	fixture := newSeededWebhookFixture(t)
	sibling := cm.WebhookDefinitionRequestHeader{Key: "X-Sibling", Secret: cm.NewOptBool(true), Value: cm.NewOptNilString(webhookSiblingHeader)}
	fixture.server.SetWebhookDefinition("space", "imported", cm.WebhookDefinitionData{
		Name: "Imported", URL: "https://example.invalid/webhook", Topics: []string{"Entry.publish"}, Active: cm.NewOptBool(true),
		HttpBasicUsername: cm.NewOptNilString("user"), HttpBasicPassword: cm.NewOptNilString(webhookPriorPassword),
		Headers: cm.WebhookDefinitionRequestHeaders{{Key: "X-Managed", Secret: cm.NewOptBool(false), Value: cm.NewOptNilString(previous)}, sibling},
	})
	ordinary := webhookRecoveryStep(webhookRecoveryConfig(true, "X-Managed")+webhookRecoveryImportBlock, webhookPriorPassword, webhookPriorHeader)
	reconcile := webhookRecoveryStep(webhookRecoveryConfig(true, ""), webhookPriorPassword, previous)
	reconcile.PreConfig = func() {
		fixture.updateRemoteHeaders(t, 1, cm.WebhookDefinitionRequestHeaders{sibling})
		fixture.updateRemoteHeaders(t, 2, cm.WebhookDefinitionRequestHeaders{{Key: "X-Managed", Secret: cm.NewOptBool(true), Value: cm.NewOptNilString(webhookRotatedHeader)}, sibling})
	}
	reconcile.PostApplyFunc = func() {
		fixture.requirePuts(t, webhookPUTWithFields("0", webhookPriorPassword, previous, map[string]any{"headers": []map[string]any{
			{"key": "X-Managed", "secret": false, "value": previous}, {"key": "X-Sibling", "secret": true},
		}}), webhookExpectedRecoveryPut("3", webhookPriorPassword, previous, true))
		fixture.requireStored(t, webhookPriorPassword, previous, true)
	}
	testAccMockedResource(t, fixture, resource.TestCase{Steps: []resource.TestStep{ordinary, reconcile}})
}

func TestAccWebhookResourceRefreshOrphanHeaderVerifier(t *testing.T) {
	t.Parallel()
	fixture := newSeededWebhookFixture(t)

	var observations webhookProtocolObservations

	config := webhookRecoveryConfig(true, "")
	initial := importedWebhookRecoveryStep(t, fixture)
	remove := webhookRecoveryStep(strings.Replace(config, `"X-Managed" = { secret=true, value_wo=var.header }`, "", 1), webhookPriorPassword, webhookPriorHeader)

	var original []byte

	sibling := cm.WebhookDefinitionRequestHeader{Key: "X-Sibling", Secret: cm.NewOptBool(true), Value: cm.NewOptNilString(webhookSiblingHeader)}
	remove.PreConfig = func() {
		original = observations.lastApply().privateFields(t)["write_only_secret_hashes"]
		fixture.updateRemoteHeaders(t, 1, cm.WebhookDefinitionRequestHeaders{sibling})
	}
	restore := webhookRecoveryStep(config, webhookPriorPassword, webhookPriorHeader)
	restore.PreConfig = func() {
		observations.injectNextReadVerifiers(original)
		fixture.updateRemoteHeaders(t, 2, cm.WebhookDefinitionRequestHeaders{{Key: "X-Managed", Secret: cm.NewOptBool(true), Value: cm.NewOptNilString("external")}, sibling})
	}
	restore.Check = func(*terraform.State) error {
		require.Equal(t, webhookPrivateVerifiers{Basic: observations.firstApply().verifiers(t).Basic}, observations.lastRead().verifiers(t))
		fixture.requirePuts(t, webhookExpectedRecoveryPut("0", webhookPriorPassword, webhookPriorHeader, true), webhookExpectedRecoveryPut("3", webhookPriorPassword, webhookPriorHeader, true))

		return nil
	}
	testObservedWebhookResource(t, fixture, &observations, []resource.TestStep{initial, remove, restore})
}

func TestAccWebhookResourceRecoverySupersededHeaderVerifier(t *testing.T) {
	t.Parallel()
	fixture := newSeededWebhookFixture(t)

	var observations webhookProtocolObservations

	config := webhookRecoveryConfig(true, "")
	initial := importedWebhookRecoveryStep(t, fixture)

	const ordinary = "synthetic-ordinary-header"

	migrate := webhookRecoveryStep(strings.Replace(config, "value_wo=var.header", `value="`+ordinary+`"`, 1), webhookPriorPassword, webhookPriorHeader)

	var original []byte

	migrate.PreConfig = func() { original = observations.lastApply().privateFields(t)["write_only_secret_hashes"] }
	rotate := webhookRecoveryStep(config, webhookPriorPassword, webhookRotatedHeader)
	rotate.PreConfig = func() { observations.injectNextReadVerifiers(original); contradictNextWebhookName(fixture) }
	rotate.ExpectError = regexp.MustCompile("Contentful returned a different webhook name")
	retry := webhookRecoveryStep(config, webhookPriorPassword, webhookPriorHeader)
	retry.PreConfig = func() {
		prior := observations.firstApply().verifiers(t)
		require.Equal(t, webhookPrivateVerifiers{Basic: prior.Basic}, observations.lastApply().verifiers(t))
	}
	retry.PostApplyFunc = func() {
		fixture.requirePuts(t, webhookExpectedRecoveryPut("0", webhookPriorPassword, webhookPriorHeader, true), webhookExpectedRecoveryPut("1", webhookPriorPassword, ordinary, true), webhookExpectedRecoveryPut("2", webhookPriorPassword, webhookRotatedHeader, true), webhookExpectedRecoveryPut("3", webhookPriorPassword, webhookPriorHeader, true))
	}
	testObservedWebhookResource(t, fixture, &observations, []resource.TestStep{initial, migrate, rotate, retry})
}
