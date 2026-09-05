package provider_test

import (
	"fmt"
	"net/http"
	"regexp"
	"testing"

	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

func webhookMutationConfig(name, filterType string, ignoreFilters bool) string {
	lifecycle := ""
	if ignoreFilters {
		lifecycle = "lifecycle { ignore_changes = [filters] }"
	}

	return fmt.Sprintf(`
resource "contentful_webhook" "test" {
  space_id = "space"
  name     = %q
  url      = "https://example.com/webhook"
  topics   = ["Entry.publish"]
  filters = [{
    equals = {
      doc   = "sys.type"
      value = %q
    }
  }]
  %s
}
`, name, filterType, lifecycle)
}

func TestAccWebhookResourceConsistencyErrorsRetainResponseState(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")

	adapter := &mutationJSONResponseAdapter{delegate: server}
	filterValuePath := tfjsonpath.New("filters").AtSliceIndex(0).AtMapKey("equals").AtMapKey("value")

	ContentfulProviderMockedResourceTest(t, adapter, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					adapter.mutateNext(http.MethodPost, replaceWebhookResponseFilter("Asset"))
				},
				Config:      webhookMutationConfig("Created webhook", "Entry", false),
				ExpectError: regexp.MustCompile(`Contentful returned different webhook filters`),
			},
			{
				Config: webhookMutationConfig("Created webhook", "Entry", false),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionReplace),
					expectResponseStateInPlan{
						address:   "contentful_webhook.test",
						valuePath: filterValuePath,
						value:     "Asset",
					},
				}},
			},
			{
				PreConfig: func() {
					adapter.mutateNext(http.MethodPut, replaceWebhookResponseFilter("Entry"))
				},
				Config:      webhookMutationConfig("Updated webhook", "Asset", false),
				ExpectError: regexp.MustCompile(`Contentful returned different webhook filters`),
			},
			{
				Config: webhookMutationConfig("Updated webhook", "Asset", false),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionUpdate),
					expectResponseStateInPlan{
						address:   "contentful_webhook.test",
						valuePath: filterValuePath,
						value:     "Entry",
					},
				}},
			},
		},
	})
}

func TestAccWebhookResourceUnsupportedResponsePropertiesPreventFalseConvergence(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")

	adapter := &mutationJSONResponseAdapter{delegate: server}
	filterValuePath := tfjsonpath.New("filters").AtSliceIndex(0).AtMapKey("equals").AtMapKey("value")
	webhookIDPath := tfjsonpath.New("webhook_id")

	ContentfulProviderMockedResourceTest(t, adapter, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					adapter.mutateNext(http.MethodPost, func(response map[string]any) {
						response["filters"] = []any{map[string]any{
							"equals":         []any{map[string]any{"doc": "sys.type"}, "Entry"},
							"futureOperator": []any{map[string]any{"doc": "sys.id"}, "entry"},
						}}
					})
				},
				Config:      webhookMutationConfig("Created webhook", "Entry", false),
				ExpectError: regexp.MustCompile(`Provider cannot fully represent webhook filters`),
			},
			{
				Config:             webhookMutationConfig("Created webhook", "Entry", false),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{PostApplyPreRefresh: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionReplace),
					expectResponseStateInPlan{
						address:            "contentful_webhook.test",
						valuePath:          filterValuePath,
						value:              "Entry",
						nonEmptyBeforePath: &webhookIDPath,
					},
				}},
			},
		},
	})
}

func TestAccWebhookResourceMutationReconciliationUsesEffectivePlanWithIgnoreChanges(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{Steps: []resource.TestStep{
		{
			Config: webhookMutationConfig("Initial webhook", "Asset", false),
		},
		{
			Config: webhookMutationConfig("Updated webhook", "Entry", true),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionUpdate),
			}},
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(
					"contentful_webhook.test",
					tfjsonpath.New("filters").AtSliceIndex(0).AtMapKey("equals").AtMapKey("value"),
					knownvalue.StringExact("Asset"),
				),
			},
		},
		{
			Config: webhookMutationConfig("Updated webhook", "Entry", true),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionNoop),
			}},
		},
	}})
}

func TestAccWebhookResourceKnownDefaultContradictionRetainsResponseState(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")

	adapter := &mutationJSONResponseAdapter{delegate: server}
	ContentfulProviderMockedResourceTest(t, adapter, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					adapter.mutateNext(http.MethodPost, func(response map[string]any) {
						response["active"] = false
					})
				},
				Config: webhookMutationConfig("Defaulted active webhook", "Entry", false),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectKnownValue("contentful_webhook.test", tfjsonpath.New("active"), knownvalue.Bool(true)),
				}},
				ExpectError: regexp.MustCompile(`Contentful returned a different webhook active value`),
			},
			{
				Config: webhookMutationConfig("Defaulted active webhook", "Entry", false),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionReplace),
					expectResponseStateInPlan{
						address:   "contentful_webhook.test",
						valuePath: tfjsonpath.New("active"),
						value:     false,
					},
				}},
			},
		},
	})
}

func replaceWebhookResponseFilter(value string) func(map[string]any) {
	return func(response map[string]any) {
		response["filters"] = []any{map[string]any{
			"equals": []any{map[string]any{"doc": "sys.type"}, value},
		}}
	}
}
