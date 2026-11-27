package provider_test

import (
	"net/http"
	"strings"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"
)

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
	testObservedWebhookResource(t, fixture, &observations, []resource.TestStep{initial, delegate, resume, rotate})
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
