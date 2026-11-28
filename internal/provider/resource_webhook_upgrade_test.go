package provider_test

import (
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

func TestAccWebhookResourceRefreshStateWithoutVersion(t *testing.T) {
	const address = "contentful_webhook.test"

	// The fixture was imported with v0.1.0. Attach the provider under test at the same source address.
	t.Setenv(resource.EnvTfAccProviderNamespace, "cysp")

	server := newTestCMAServer(t)
	server.RegisterSpaceEnvironment("space", "master")

	fields := cm.WebhookDefinitionData{
		Name: "Imported", URL: "https://example.com/webhook", Topics: []string{"Entry.publish"},
		Filters: cm.NewOptNilWebhookDefinitionFilterArray([]cm.WebhookDefinitionFilter{cmWebhookEqualsFilter("Entry")}),
		Active:  cm.NewOptBool(true),
	}
	server.SetWebhookDefinition("space", "webhook", fields)

	fields.Name = "Edited externally"
	updated, err := server.Handler().UpdateWebhookDefinition(t.Context(), &fields, cm.UpdateWebhookDefinitionParams{
		SpaceID: "space", WebhookDefinitionID: "webhook", XContentfulVersion: cm.NewOptInt(0),
	})
	require.NoError(t, err)
	require.IsType(t, &cm.WebhookDefinitionStatusCode{}, updated)

	recorder := &webhookWriteRecorder{next: server}
	version := tfjsonpath.New("version")
	testAccMockedResource(t, recorder, resource.TestCase{Steps: []resource.TestStep{
		{
			ConfigDirectory: config.TestNameDirectory(),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction(address, plancheck.ResourceActionNoop),
				plancheck.ExpectKnownValue(address, version, knownvalue.Int64Exact(1)),
			}},
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(address, version, knownvalue.Int64Exact(1)),
			},
			PostApplyFunc: func() {
				recorder.mu.Lock()
				defer recorder.mu.Unlock()

				require.Empty(t, recorder.writes)
			},
		},
	}})
}
