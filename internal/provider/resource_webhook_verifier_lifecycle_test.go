package provider_test

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/stretchr/testify/require"
)

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
