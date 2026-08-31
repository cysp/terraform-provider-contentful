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

func TestAccWebhookMutationResponseContradictionRetainsTruthfulState(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")

	adapter := &mutationJSONResponseAdapter{delegate: server}
	config := func(name, filterType string) string {
		return fmt.Sprintf(`
resource "contentful_webhook" "test" {
  space_id = "space"
  name     = %q
  url      = "https://example.com/webhook"
  topics   = []
  filters = [{
    equals = {
      doc   = "sys.type"
      value = %q
    }
  }]
}
`, name, filterType)
	}

	ContentfulProviderMockedResourceTest(t, adapter, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
		Steps: []resource.TestStep{
			{
				Config: config("Initial webhook", "Asset"),
			},
			{
				PreConfig: func() {
					adapter.arm(http.MethodPut, replaceWebhookResponseFilter("Asset"))
				},
				Config:      config("Updated webhook", "Entry"),
				ExpectError: regexp.MustCompile(`filters response differed meaningfully from the Terraform plan`),
			},
			{
				Config: config("Updated webhook", "Entry"),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionUpdate),
					expectMutationPlanTransition{
						address:   "contentful_webhook.test",
						valuePath: tfjsonpath.New("filters").AtSliceIndex(0).AtMapKey("equals").AtMapKey("value"),
						before:    "Asset",
						after:     "Entry",
					},
				}},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"contentful_webhook.test",
						tfjsonpath.New("filters").AtSliceIndex(0).AtMapKey("equals").AtMapKey("value"),
						knownvalue.StringExact("Entry"),
					),
				},
			},
		},
	})
}

func TestAccWebhookCreateMutationResponseContradictionRetainsTaintedState(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")

	adapter := &mutationJSONResponseAdapter{delegate: server}
	config := `
resource "contentful_webhook" "test" {
  space_id = "space"
  name     = "Created webhook"
  url      = "https://example.com/webhook"
  topics   = []
  filters = [{
    equals = {
      doc   = "sys.type"
      value = "Entry"
    }
  }]
}
`

	ContentfulProviderMockedResourceTest(t, adapter, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					adapter.arm(http.MethodPost, replaceWebhookResponseFilter("Asset"))
				},
				Config:      config,
				ExpectError: regexp.MustCompile(`filters response differed meaningfully from the Terraform plan`),
			},
			{
				Config: config,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionReplace),
					expectMutationPlanTransition{
						address:   "contentful_webhook.test",
						valuePath: tfjsonpath.New("filters").AtSliceIndex(0).AtMapKey("equals").AtMapKey("value"),
						before:    "Asset",
						after:     "Entry",
					},
				}},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"contentful_webhook.test",
						tfjsonpath.New("filters").AtSliceIndex(0).AtMapKey("equals").AtMapKey("value"),
						knownvalue.StringExact("Entry"),
					),
				},
			},
		},
	})
}

func TestAccWebhookMutationReconciliationUsesEffectivePlanWithIgnoreChanges(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")

	config := func(name, filterType string, ignoreFilters bool) string {
		lifecycle := ""
		if ignoreFilters {
			lifecycle = "lifecycle { ignore_changes = [filters] }"
		}

		return fmt.Sprintf(`
resource "contentful_webhook" "test" {
  space_id = "space"
  name     = %q
  url      = "https://example.com/webhook"
  topics   = []
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

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{Steps: []resource.TestStep{
		{
			Config: config("Initial webhook", "Asset", false),
		},
		{
			Config: config("Updated webhook", "Entry", true),
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
			Config: config("Updated webhook", "Entry", true),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionNoop),
			}},
		},
	}})
}

func replaceWebhookResponseFilter(value string) func(map[string]any) {
	return func(response map[string]any) {
		response["filters"] = []any{map[string]any{
			"equals": []any{map[string]any{"doc": "sys.type"}, value},
		}}
	}
}
