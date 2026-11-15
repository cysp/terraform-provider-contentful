package provider_test

import (
	"fmt"
	"net/http"
	"regexp"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccTaxonomyTaintedReplacementRejectsUnconfirmedDeleteVersion(t *testing.T) {
	t.Parallel()

	for _, kind := range []string{"concept", "concept_scheme"} {
		for _, contradiction := range []string{"organization", "resource"} {
			t.Run(kind+"/"+contradiction, func(t *testing.T) {
				t.Parallel()

				testTaxonomyDeleteIdentityRecovery(t, kind, contradiction)
			})
		}
	}
}

func testTaxonomyDeleteIdentityRecovery(t *testing.T, kind, contradiction string) {
	t.Helper()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)

	mutator := &taxonomyResponseMutator{next: server}
	recorder := &taxonomyRequestBodyRecorder{next: mutator}
	config := fmt.Sprintf(`
resource "contentful_taxonomy_%[1]s" "target" {
  organization_id = "organization-id"
  %[1]s_id = "target"
  pref_label = { "en-US" = "Target" }
}
resource "contentful_taxonomy_%[1]s" "unrelated" {
  organization_id = "organization-id"
  %[1]s_id = "unrelated"
  pref_label = { "en-US" = "Unrelated" }
}
`, kind)
	address := "contentful_taxonomy_" + kind + ".target"

	endpoint := "/taxonomy/concepts/target"
	if kind == "concept_scheme" {
		endpoint = "/taxonomy/concept-schemes/target"
	}

	organizationID, resourceID := "organization-id", "target"
	if contradiction == "organization" {
		organizationID = "other-organization"
	} else {
		resourceID = "unrelated"
	}

	testAccMockedResource(t, recorder, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
		Steps: []resource.TestStep{
			{Config: config},
			{
				PreConfig: func() { mutator.replaceIdentityOnce(http.MethodGet, endpoint, organizationID, resourceID) },
				Config:    config, Taint: []string{address},
				ExpectError: regexp.MustCompile(`Unexpected Contentful taxonomy read identity`),
			},
			{
				PreConfig: func() {
					assert.Empty(t, recorder.matchingRequests(http.MethodDelete, endpoint), "an unconfirmed GET version must not authorize deletion")

					for _, resourceID := range []string{"target", "unrelated"} {
						if kind == "concept" {
							response, readErr := server.Handler().GetTaxonomyConcept(t.Context(), cm.GetTaxonomyConceptParams{OrganizationID: "organization-id", TaxonomyConceptID: resourceID})
							require.NoError(t, readErr)

							concept, ok := response.(*cm.TaxonomyConcept)
							require.True(t, ok, "resource %s must remain present", resourceID)
							assert.Equal(t, 1, concept.Sys.Version)
						} else {
							response, readErr := server.Handler().GetTaxonomyConceptScheme(t.Context(), cm.GetTaxonomyConceptSchemeParams{OrganizationID: "organization-id", TaxonomyConceptSchemeID: resourceID})
							require.NoError(t, readErr)

							scheme, ok := response.(*cm.TaxonomyConceptScheme)
							require.True(t, ok, "resource %s must remain present", resourceID)
							assert.Equal(t, 1, scheme.Sys.Version)
						}
					}
				},
				// With unchanged configuration, replacement proves the failed delete
				// preserved Core's taint; no custom state decoder is needed.
				Config: config,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(address, plancheck.ResourceActionDestroyBeforeCreate),
					testAccPriorState{check: statecheck.ExpectKnownValue(address, tfjsonpath.New("id"), knownvalue.StringExact("organization-id/target"))},
				}},
				Check: func(*terraform.State) error {
					deletes := recorder.matchingRequests(http.MethodDelete, endpoint)
					require.Len(t, deletes, 1, "a subsequent confirmed GET can authorize replacement")
					assert.Equal(t, "1", deletes[0].version)

					return nil
				},
			},
		},
	})
}
