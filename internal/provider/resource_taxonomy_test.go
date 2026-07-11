package provider_test

import (
	"maps"
	"testing"

	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

//nolint:paralleltest
func TestAccTaxonomyResourcesCreateUpdate(t *testing.T) {
	parallelWhenMocked(t)

	server, err := cmt.NewContentfulManagementServer()
	if err != nil {
		t.Fatal(err)
	}

	base := config.Variables{
		"organization_id":   config.StringVariable("organization-id"),
		"concept_id":        config.StringVariable("furniture"),
		"concept_scheme_id": config.StringVariable("products"),
		"concept_label":     config.StringVariable("Furniture"),
		"scheme_label":      config.StringVariable("Products"),
	}
	updated := maps.Clone(base)
	updated["concept_label"] = config.StringVariable("Home furniture")
	updated["scheme_label"] = config.StringVariable("Home products")

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{Steps: []resource.TestStep{
		{
			ConfigDirectory: config.TestNameDirectory(), ConfigVariables: base,
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_taxonomy_concept.test", plancheck.ResourceActionCreate),
				plancheck.ExpectResourceAction("contentful_taxonomy_concept_scheme.test", plancheck.ResourceActionCreate),
			}},
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("contentful_taxonomy_concept.test", "id", "organization-id/furniture"),
				resource.TestCheckResourceAttr("contentful_taxonomy_concept.test", "pref_label.en-US", "Furniture"),
				resource.TestCheckResourceAttr("contentful_taxonomy_concept.test", "alt_labels.%", "1"),
				resource.TestCheckResourceAttr("contentful_taxonomy_concept.test", "alt_labels.en-GB.0", "Furniture"),
				resource.TestCheckResourceAttr("contentful_taxonomy_concept.test", "hidden_labels.%", "1"),
				resource.TestCheckResourceAttr("contentful_taxonomy_concept.test", "hidden_labels.en-GB.0", "Furnishings"),
				resource.TestCheckResourceAttr("contentful_taxonomy_concept_scheme.test", "total_concepts", "1"),
			),
		},
		{
			ConfigDirectory: config.TestNameDirectory(), ConfigVariables: updated,
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_taxonomy_concept.test", plancheck.ResourceActionUpdate),
				plancheck.ExpectResourceAction("contentful_taxonomy_concept_scheme.test", plancheck.ResourceActionUpdate),
			}},
		},
		{
			ConfigDirectory: config.TestNameDirectory(), ConfigVariables: updated,
			ResourceName: "contentful_taxonomy_concept.test", ImportState: true, ImportStateId: "organization-id/furniture", ImportStateVerify: true,
			ImportStateVerifyIgnore: []string{"alt_labels", "hidden_labels"},
		},
		{
			ConfigDirectory: config.TestNameDirectory(), ConfigVariables: updated,
			ResourceName: "contentful_taxonomy_concept_scheme.test", ImportState: true, ImportStateId: "organization-id/products", ImportStateVerify: true,
		},
	}})
}
