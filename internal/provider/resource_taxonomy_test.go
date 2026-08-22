package provider_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"regexp"
	"strings"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errUnexpectedTaxonomyResponse = errors.New("unexpected taxonomy response")

const taxonomyAcceptanceOrganizationID = "2zuSjSO4A0e6GKBrhJRe2m"

//nolint:paralleltest
func TestAccTaxonomyResourcesLifecycle(t *testing.T) {
	parallelWhenMocked(t)

	server, err := cmt.NewContentfulManagementServer()
	if err != nil {
		t.Fatal(err)
	}

	suffix := acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")
	ids := taxonomyLifecycleIDs{
		concept: "acctest_concept_" + suffix,
		parent1: "acctest_parent1_" + suffix,
		parent2: "acctest_parent2_" + suffix,
		related: "acctest_related_" + suffix,
		scheme:  "acctest_scheme_" + suffix,
	}

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{Steps: []resource.TestStep{
		{
			Config: taxonomyLifecycleConfig(ids, taxonomyLifecycleCreate),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: taxonomyLifecycleActions(plancheck.ResourceActionCreate),
				PostApplyPostRefresh: []plancheck.PlanCheck{
					plancheck.ExpectEmptyPlan(),
				},
			},
			ConfigStateChecks: taxonomyLifecycleStateChecks(ids, taxonomyLifecycleCreate),
		},
		{
			Config: taxonomyLifecycleConfig(ids, taxonomyLifecycleCreate),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_taxonomy_concept.test", plancheck.ResourceActionNoop),
				plancheck.ExpectResourceAction("contentful_taxonomy_concept_scheme.test", plancheck.ResourceActionNoop),
			}},
			ConfigStateChecks: append(taxonomyLifecycleStateChecks(ids, taxonomyLifecycleCreate), taxonomyConceptSchemeMembershipCheck(ids, true)),
		},
		{
			Config: taxonomyLifecycleConfig(ids, taxonomyLifecycleUpdate),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_taxonomy_concept.test", plancheck.ResourceActionUpdate),
					plancheck.ExpectResourceAction("contentful_taxonomy_concept_scheme.test", plancheck.ResourceActionUpdate),
				},
				PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
			},
			ConfigStateChecks: append(taxonomyLifecycleStateChecks(ids, taxonomyLifecycleUpdate), taxonomyConceptSchemeMembershipCheck(ids, true)),
		},
		{
			Config: taxonomyLifecycleConfig(ids, taxonomyLifecyclePreserveComputed),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_taxonomy_concept.test", plancheck.ResourceActionUpdate),
					plancheck.ExpectResourceAction("contentful_taxonomy_concept_scheme.test", plancheck.ResourceActionUpdate),
				},
				PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
			},
			ConfigStateChecks: append(taxonomyLifecycleStateChecks(ids, taxonomyLifecyclePreserveComputed), taxonomyConceptSchemeMembershipCheck(ids, true)),
		},
		{
			Config: taxonomyLifecycleConfig(ids, taxonomyLifecycleClear),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_taxonomy_concept.test", plancheck.ResourceActionUpdate),
					plancheck.ExpectResourceAction("contentful_taxonomy_concept_scheme.test", plancheck.ResourceActionUpdate),
				},
				PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
			},
			ConfigStateChecks: taxonomyLifecycleStateChecks(ids, taxonomyLifecycleClear),
		},
		{
			Config: taxonomyLifecycleConfig(ids, taxonomyLifecycleClear),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_taxonomy_concept.test", plancheck.ResourceActionNoop),
				plancheck.ExpectResourceAction("contentful_taxonomy_concept_scheme.test", plancheck.ResourceActionNoop),
			}},
			ConfigStateChecks: append(taxonomyLifecycleStateChecks(ids, taxonomyLifecycleClear), taxonomyConceptSchemeMembershipCheck(ids, false)),
		},
		{
			Config:                  taxonomyLifecycleConfig(ids, taxonomyLifecycleClear),
			ResourceName:            "contentful_taxonomy_concept.test",
			ImportState:             true,
			ImportStateId:           taxonomyAcceptanceOrganizationID + "/" + ids.concept,
			ImportStateVerify:       true,
			ImportStateVerifyIgnore: []string{"alt_labels", "hidden_labels"},
		},
		{
			Config:            taxonomyLifecycleConfig(ids, taxonomyLifecycleClear),
			ResourceName:      "contentful_taxonomy_concept_scheme.test",
			ImportState:       true,
			ImportStateId:     taxonomyAcceptanceOrganizationID + "/" + ids.scheme,
			ImportStateVerify: true,
		},
	}})
}

type taxonomyLifecycleStage int

const (
	taxonomyLifecycleCreate taxonomyLifecycleStage = iota
	taxonomyLifecycleUpdate
	taxonomyLifecyclePreserveComputed
	taxonomyLifecycleClear
)

type taxonomyLifecycleIDs struct {
	concept string
	parent1 string
	parent2 string
	related string
	scheme  string
}

func taxonomyLifecycleActions(action plancheck.ResourceActionType) []plancheck.PlanCheck {
	return []plancheck.PlanCheck{
		plancheck.ExpectResourceAction("contentful_taxonomy_concept.parent1", action),
		plancheck.ExpectResourceAction("contentful_taxonomy_concept.parent2", action),
		plancheck.ExpectResourceAction("contentful_taxonomy_concept.related", action),
		plancheck.ExpectResourceAction("contentful_taxonomy_concept.test", action),
		plancheck.ExpectResourceAction("contentful_taxonomy_concept_scheme.test", action),
	}
}

func taxonomyLifecycleStateChecks(ids taxonomyLifecycleIDs, stage taxonomyLifecycleStage) []statecheck.StateCheck {
	conceptAddress := "contentful_taxonomy_concept.test"
	schemeAddress := "contentful_taxonomy_concept_scheme.test"
	checks := []statecheck.StateCheck{
		statecheck.ExpectIdentity(conceptAddress, map[string]knownvalue.Check{
			"organization_id": knownvalue.StringExact(taxonomyAcceptanceOrganizationID),
			"concept_id":      knownvalue.StringExact(ids.concept),
		}),
		statecheck.ExpectIdentity(schemeAddress, map[string]knownvalue.Check{
			"organization_id":   knownvalue.StringExact(taxonomyAcceptanceOrganizationID),
			"concept_scheme_id": knownvalue.StringExact(ids.scheme),
		}),
	}

	var conceptURI, conceptDefinition, schemeURI, schemeDefinition knownvalue.Check

	var conceptLabel, altLabel string

	var schemeLabel string

	var notations, broader, related, schemeConcepts, topConcepts []string

	switch stage {
	case taxonomyLifecycleCreate:
		conceptURI = knownvalue.Null()
		conceptDefinition = knownvalue.Null()
		schemeURI = knownvalue.StringExact("https://example.com/schemes/initial")
		schemeDefinition = localizedKnownValue("Initial scheme definition")
		conceptLabel, altLabel = "Initial concept", ""
		schemeLabel = "Lifecycle scheme"
		notations, broader, related = []string{}, []string{ids.parent1}, []string{ids.related}
		schemeConcepts, topConcepts = []string{ids.concept, ids.parent1, ids.related}, []string{ids.parent1}
	case taxonomyLifecycleUpdate:
		conceptURI = knownvalue.StringExact("https://example.com/concepts/updated")
		conceptDefinition = localizedKnownValue("Updated definition")
		schemeURI = knownvalue.StringExact("https://example.com/schemes/updated")
		schemeDefinition = localizedKnownValue("Updated scheme definition")
		conceptLabel, altLabel = "Updated concept", "Updated alternative"
		schemeLabel = "Lifecycle scheme"
		notations, broader, related = []string{"UPDATED"}, []string{ids.parent2}, []string{ids.parent1}
		schemeConcepts, topConcepts = []string{ids.concept, ids.parent2}, []string{ids.parent2}
	case taxonomyLifecyclePreserveComputed:
		conceptURI = knownvalue.Null()
		conceptDefinition = knownvalue.Null()
		schemeURI = knownvalue.Null()
		schemeDefinition = knownvalue.Null()
		conceptLabel, altLabel = "Preserved concept", "Updated alternative"
		schemeLabel = "Preserved scheme"
		notations, broader, related = []string{"UPDATED"}, []string{ids.parent2}, []string{ids.parent1}
		schemeConcepts, topConcepts = []string{ids.concept, ids.parent2}, []string{ids.parent2}
	case taxonomyLifecycleClear:
		conceptURI, conceptDefinition = knownvalue.Null(), knownvalue.Null()
		schemeURI, schemeDefinition = knownvalue.Null(), knownvalue.Null()
		conceptLabel, altLabel = "Cleared concept", ""
		schemeLabel = "Cleared scheme"
		notations, broader, related = []string{}, []string{}, []string{}
		schemeConcepts, topConcepts = []string{}, []string{}
	}

	checks = append(
		checks,
		statecheck.ExpectKnownValue(conceptAddress, tfjsonpath.New("uri"), conceptURI),
		statecheck.ExpectKnownValue(conceptAddress, tfjsonpath.New("pref_label"), localizedKnownValue(conceptLabel)),
		statecheck.ExpectKnownValue(conceptAddress, tfjsonpath.New("definition"), conceptDefinition),
		statecheck.ExpectKnownValue(conceptAddress, tfjsonpath.New("notations"), stringListKnownValue(notations)),
		statecheck.ExpectKnownValue(conceptAddress, tfjsonpath.New("broader_concept_ids"), stringListKnownValue(broader)),
		statecheck.ExpectKnownValue(conceptAddress, tfjsonpath.New("related_concept_ids"), stringListKnownValue(related)),
		statecheck.ExpectKnownValue(schemeAddress, tfjsonpath.New("uri"), schemeURI),
		statecheck.ExpectKnownValue(schemeAddress, tfjsonpath.New("pref_label"), localizedKnownValue(schemeLabel)),
		statecheck.ExpectKnownValue(schemeAddress, tfjsonpath.New("definition"), schemeDefinition),
		statecheck.ExpectKnownValue(schemeAddress, tfjsonpath.New("concept_ids"), stringListKnownValue(schemeConcepts)),
		statecheck.ExpectKnownValue(schemeAddress, tfjsonpath.New("top_concept_ids"), stringListKnownValue(topConcepts)),
		statecheck.ExpectKnownValue(schemeAddress, tfjsonpath.New("total_concepts"), knownvalue.Int64Exact(int64(len(schemeConcepts)))),
	)

	switch stage {
	case taxonomyLifecycleClear:
		checks = append(checks, statecheck.ExpectKnownValue(conceptAddress, tfjsonpath.New("alt_labels"), knownvalue.MapExact(map[string]knownvalue.Check{})))
	case taxonomyLifecycleCreate:
		checks = append(checks, statecheck.ExpectKnownValue(conceptAddress, tfjsonpath.New("alt_labels"), knownvalue.MapExact(map[string]knownvalue.Check{
			"en-US": knownvalue.ListExact([]knownvalue.Check{}),
		})))
	case taxonomyLifecycleUpdate, taxonomyLifecyclePreserveComputed:
		checks = append(checks, statecheck.ExpectKnownValue(conceptAddress, tfjsonpath.New("alt_labels"), knownvalue.MapExact(map[string]knownvalue.Check{
			"en-US": knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact(altLabel)}),
		})))
	}

	return checks
}

//nolint:ireturn
func taxonomyConceptSchemeMembershipCheck(ids taxonomyLifecycleIDs, member bool) statecheck.StateCheck {
	values := []knownvalue.Check{}
	if member {
		values = append(values, knownvalue.StringExact(ids.scheme))
	}

	return statecheck.ExpectKnownValue(
		"contentful_taxonomy_concept.test",
		tfjsonpath.New("concept_scheme_ids"),
		knownvalue.SetExact(values),
	)
}

//nolint:ireturn
func localizedKnownValue(value string) knownvalue.Check {
	return knownvalue.MapExact(map[string]knownvalue.Check{"en-US": knownvalue.StringExact(value)})
}

//nolint:ireturn
func stringListKnownValue(values []string) knownvalue.Check {
	checks := make([]knownvalue.Check, 0, len(values))
	for _, value := range values {
		checks = append(checks, knownvalue.StringExact(value))
	}

	return knownvalue.ListExact(checks)
}

func taxonomyLifecycleConfig(ids taxonomyLifecycleIDs, stage taxonomyLifecycleStage) string {
	conceptAttributes := `
  broader_concept_ids = [contentful_taxonomy_concept.parent1.concept_id]
  related_concept_ids = [contentful_taxonomy_concept.related.concept_id]`
	conceptLabel := "Initial concept"
	schemeAttributes := `
  uri         = "https://example.com/schemes/initial"
  definition  = { "en-US" = "Initial scheme definition" }
  concept_ids = [contentful_taxonomy_concept.test.concept_id, contentful_taxonomy_concept.parent1.concept_id, contentful_taxonomy_concept.related.concept_id]
  top_concept_ids = [contentful_taxonomy_concept.parent1.concept_id]`
	schemeLabel := "Lifecycle scheme"

	switch stage {
	case taxonomyLifecycleCreate:
	case taxonomyLifecycleUpdate:
		conceptLabel = "Updated concept"
		conceptAttributes = `
  uri         = "https://example.com/concepts/updated"
  alt_labels  = { "en-US" = ["Updated alternative"] }
  definition  = { "en-US" = "Updated definition" }
  notations   = ["UPDATED"]
  broader_concept_ids = [contentful_taxonomy_concept.parent2.concept_id]
  related_concept_ids = [contentful_taxonomy_concept.parent1.concept_id]`
		schemeAttributes = `
  uri         = "https://example.com/schemes/updated"
  definition  = { "en-US" = "Updated scheme definition" }
  concept_ids = [contentful_taxonomy_concept.test.concept_id, contentful_taxonomy_concept.parent2.concept_id]
  top_concept_ids = [contentful_taxonomy_concept.parent2.concept_id]`
	case taxonomyLifecyclePreserveComputed:
		conceptLabel = "Preserved concept"
		conceptAttributes = ""
		schemeLabel = "Preserved scheme"
		schemeAttributes = ""
	case taxonomyLifecycleClear:
		conceptLabel = "Cleared concept"
		conceptAttributes = `
  alt_labels = {}
  notations = []
  broader_concept_ids = []
  related_concept_ids = []`
		schemeLabel = "Cleared scheme"
		schemeAttributes = `
  concept_ids = []
  top_concept_ids = []`
	}

	return fmt.Sprintf(`
resource "contentful_taxonomy_concept" "parent1" {
  organization_id = %[1]q
  concept_id = %[2]q
  pref_label = { "en-US" = "Parent one" }
}

resource "contentful_taxonomy_concept" "parent2" {
  organization_id = %[1]q
  concept_id = %[3]q
  pref_label = { "en-US" = "Parent two" }
}

resource "contentful_taxonomy_concept" "related" {
  organization_id = %[1]q
  concept_id = %[4]q
  pref_label = { "en-US" = "Related" }
}

resource "contentful_taxonomy_concept" "test" {
  organization_id = %[1]q
  concept_id = %[5]q
  pref_label = { "en-US" = %[6]q }
%[7]s
}

resource "contentful_taxonomy_concept_scheme" "test" {
  organization_id = %[1]q
  concept_scheme_id = %[8]q
  pref_label = { "en-US" = %[9]q }
%[10]s
}
`, taxonomyAcceptanceOrganizationID, ids.parent1, ids.parent2, ids.related, ids.concept, conceptLabel, conceptAttributes, ids.scheme, schemeLabel, schemeAttributes)
}

//nolint:paralleltest
func TestAccTaxonomyResourcesRecoverFromDeletion(t *testing.T) {
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
			ConfigDirectory: config.StaticDirectory("testdata/TestAccTaxonomyResourcesCreateUpdate"), ConfigVariables: base,
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
			ConfigDirectory: config.StaticDirectory("testdata/TestAccTaxonomyResourcesCreateUpdate"), ConfigVariables: updated,
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_taxonomy_concept.test", plancheck.ResourceActionUpdate),
				plancheck.ExpectResourceAction("contentful_taxonomy_concept_scheme.test", plancheck.ResourceActionUpdate),
			}},
		},
		{
			PreConfig: func() {
				deleteTaxonomyConceptSchemeOutOfBand(t, server, "organization-id", "products")
			},
			ConfigDirectory: config.StaticDirectory("testdata/TestAccTaxonomyResourcesCreateUpdate"), ConfigVariables: updated,
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_taxonomy_concept.test", plancheck.ResourceActionNoop),
				plancheck.ExpectResourceAction("contentful_taxonomy_concept_scheme.test", plancheck.ResourceActionCreate),
			}},
		},
		{
			PreConfig: func() {
				deleteTaxonomyConceptOutOfBand(t, server, "organization-id", "furniture")
			},
			ConfigDirectory: config.StaticDirectory("testdata/TestAccTaxonomyResourcesCreateUpdate"), ConfigVariables: updated,
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_taxonomy_concept.test", plancheck.ResourceActionCreate),
				plancheck.ExpectResourceAction("contentful_taxonomy_concept_scheme.test", plancheck.ResourceActionUpdate),
			}},
		},
	}})
}

//nolint:paralleltest
func TestAccTaxonomyResourcesRecoverFromUnexpectedResponses(t *testing.T) {
	parallelWhenMocked(t)

	tests := map[string]struct {
		method       string
		path         string
		resourceName string
		update       bool
	}{
		"concept create": {method: http.MethodPut, path: "/taxonomy/concepts/furniture", resourceName: "taxonomy concept"},
		"concept update": {method: http.MethodPatch, path: "/taxonomy/concepts/furniture", resourceName: "taxonomy concept", update: true},
		"scheme create":  {method: http.MethodPut, path: "/taxonomy/concept-schemes/products", resourceName: "taxonomy concept scheme"},
		"scheme update":  {method: http.MethodPatch, path: "/taxonomy/concept-schemes/products", resourceName: "taxonomy concept scheme", update: true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			parallelWhenMocked(t)

			server, err := cmt.NewContentfulManagementServer()
			if err != nil {
				t.Fatal(err)
			}

			mutator := &taxonomyResponseMutator{next: server}
			recorder := &taxonomyRequestBodyRecorder{next: mutator}
			base := taxonomyConfigVariables("Furniture", "Products")
			steps := []resource.TestStep{}

			if test.update {
				steps = append(steps, resource.TestStep{
					ConfigDirectory: config.StaticDirectory("testdata/TestAccTaxonomyResourcesCreateUpdate"),
					ConfigVariables: base,
				})
			}

			updated := maps.Clone(base)
			if strings.Contains(test.path, "concept-schemes") {
				updated["scheme_label"] = config.StringVariable("Home products")
			} else {
				updated["concept_label"] = config.StringVariable("Home furniture")
			}

			steps = append(steps, resource.TestStep{
				PreConfig: func() {
					mutator.dropPreferredLabelOnce(test.method, test.path, "en-US")
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccTaxonomyResourcesCreateUpdate"),
				ConfigVariables: updated,
				ExpectError:     regexp.MustCompile("Unexpected Contentful " + test.resourceName + " response"),
			})
			steps = append(steps, resource.TestStep{
				ConfigDirectory: config.StaticDirectory("testdata/TestAccTaxonomyResourcesCreateUpdate"),
				ConfigVariables: updated,
			})

			ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{
				AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
				Steps:                steps,
			})

			requests := recorder.matchingRequests(http.MethodPatch, test.path)
			if test.update {
				require.Len(t, requests, 2)
				assert.Equal(t, "2", requests[len(requests)-1].version)

				var patch []struct {
					Path string `json:"path"`
				}
				require.NoError(t, json.Unmarshal(requests[len(requests)-1].body, &patch))
				require.Len(t, patch, 1)
				assert.Equal(t, "/prefLabel", patch[0].Path)
			} else {
				require.Empty(t, requests)
				require.Len(t, recorder.matchingRequests(http.MethodGet, test.path), 1, "errored Create recovery should fetch its missing private version")

				deletes := recorder.matchingRequests(http.MethodDelete, test.path)
				require.NotEmpty(t, deletes)
				assert.Equal(t, "1", deletes[0].version, "tainted Create recovery must use its fetched response version")
			}
		})
	}
}

//nolint:paralleltest
func TestAccTaxonomyInvalidCreateResponseVersionCheckpointsState(t *testing.T) {
	parallelWhenMocked(t)

	tests := map[string]struct {
		path         string
		config       string
		resourceName string
	}{
		"concept": {
			path: "/taxonomy/concepts/furniture", config: taxonomyConceptConfig("Furniture"), resourceName: "contentful_taxonomy_concept.test",
		},
		"scheme": {
			path: "/taxonomy/concept-schemes/products", config: taxonomyConceptSchemeConfig("Products"), resourceName: "contentful_taxonomy_concept_scheme.test",
		},
	}

	for name, test := range tests {
		for _, responseVersion := range []int{0, -1} {
			t.Run(fmt.Sprintf("%s/version_%d", name, responseVersion), func(t *testing.T) {
				parallelWhenMocked(t)

				server, err := cmt.NewContentfulManagementServer()
				require.NoError(t, err)

				mutator := &taxonomyResponseMutator{next: server}
				recorder := &taxonomyRequestBodyRecorder{next: mutator}
				ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{
					AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
					Steps: []resource.TestStep{
						{
							PreConfig:   func() { mutator.versionOnce(http.MethodPut, test.path, responseVersion) },
							Config:      test.config,
							ExpectError: regexp.MustCompile(`Invalid taxonomy resource version`),
						},
						{
							Config: test.config,
							ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
								plancheck.ExpectResourceAction(test.resourceName, plancheck.ResourceActionReplace),
							}},
						},
					},
				})

				require.Len(t, recorder.matchingRequests(http.MethodPut, test.path), 2)
				require.Len(t, recorder.matchingRequests(http.MethodGet, test.path), 1, "tainted replacement must fetch the version missing from private state")

				deletes := recorder.matchingRequests(http.MethodDelete, test.path)
				require.GreaterOrEqual(t, len(deletes), 2, "tainted replacement and test cleanup should each delete")

				for _, request := range deletes {
					assert.Equal(t, "1", request.version)
				}
			})
		}
	}
}

//nolint:paralleltest
func TestAccTaxonomyInvalidUpdateResponseVersionCheckpointsState(t *testing.T) {
	parallelWhenMocked(t)

	tests := map[string]struct {
		path          string
		resourceName  string
		initialConfig string
		changedConfig string
		finalConfig   string
		changedLabel  string
		finalLabel    string
	}{
		"concept": {
			path: "/taxonomy/concepts/furniture", resourceName: "contentful_taxonomy_concept.test",
			initialConfig: taxonomyConceptConfig("Furniture"), changedConfig: taxonomyConceptConfig("Home furniture"), finalConfig: taxonomyConceptConfig("Domestic furniture"),
			changedLabel: "Home furniture", finalLabel: "Domestic furniture",
		},
		"scheme": {
			path: "/taxonomy/concept-schemes/products", resourceName: "contentful_taxonomy_concept_scheme.test",
			initialConfig: taxonomyConceptSchemeConfig("Products"), changedConfig: taxonomyConceptSchemeConfig("Home products"), finalConfig: taxonomyConceptSchemeConfig("Domestic products"),
			changedLabel: "Home products", finalLabel: "Domestic products",
		},
	}

	for name, test := range tests {
		for _, responseVersion := range []int{0, -1} {
			t.Run(fmt.Sprintf("%s/version_%d", name, responseVersion), func(t *testing.T) {
				parallelWhenMocked(t)

				server, err := cmt.NewContentfulManagementServer()
				require.NoError(t, err)

				mutator := &taxonomyResponseMutator{next: server}
				recorder := &taxonomyRequestBodyRecorder{next: mutator}
				ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{
					AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
					Steps: []resource.TestStep{
						{Config: test.initialConfig},
						{
							PreConfig:   func() { mutator.versionOnce(http.MethodPatch, test.path, responseVersion) },
							Config:      test.changedConfig,
							ExpectError: regexp.MustCompile(`Invalid taxonomy resource version`),
						},
						{
							Config: test.changedConfig,
							ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
								plancheck.ExpectResourceAction(test.resourceName, plancheck.ResourceActionNoop),
							}},
							Check: resource.TestCheckResourceAttr(test.resourceName, "pref_label.en-US", test.changedLabel),
						},
						{
							Config:      test.finalConfig,
							ExpectError: regexp.MustCompile(`changed after\s+this Terraform plan was created`),
						},
						{
							RefreshState:       true,
							ExpectNonEmptyPlan: true,
							Check:              resource.TestCheckResourceAttr(test.resourceName, "pref_label.en-US", test.changedLabel),
						},
						{
							Config: test.finalConfig,
							Check:  resource.TestCheckResourceAttr(test.resourceName, "pref_label.en-US", test.finalLabel),
						},
					},
				})

				patches := recorder.matchingRequests(http.MethodPatch, test.path)
				require.Len(t, patches, 3)
				assert.Equal(t, "1", patches[0].version, "the initial Update must use the prior private version")
				assert.Equal(t, "1", patches[1].version, "an update before refresh must retain the stale optimistic-lock token")
				assert.Equal(t, "2", patches[2].version, "refresh must acquire the remote response version")

				assertTaxonomyPreferredLabelPatch(t, patches[0].body, test.changedLabel)
				assertTaxonomyPreferredLabelPatch(t, patches[1].body, test.finalLabel)

				deletes := recorder.matchingRequests(http.MethodDelete, test.path)
				require.NotEmpty(t, deletes)

				for _, request := range deletes {
					assert.Equal(t, "3", request.version, "successful Update must store its positive response version")
				}
			})
		}
	}
}

func assertTaxonomyPreferredLabelPatch(t *testing.T, body []byte, label string) {
	t.Helper()

	var patch []struct {
		Path  string          `json:"path"`
		Value json.RawMessage `json:"value"`
	}
	require.NoError(t, json.Unmarshal(body, &patch))
	require.Len(t, patch, 1)
	assert.Equal(t, "/prefLabel", patch[0].Path)
	assert.JSONEq(t, fmt.Sprintf(`{"en-US":%q}`, label), string(patch[0].Value))
}

//nolint:paralleltest
func TestAccTaxonomyResourcesRejectNormalizedLabels(t *testing.T) {
	parallelWhenMocked(t)

	tests := map[string]struct {
		method       string
		path         string
		resourceName string
		update       bool
	}{
		"concept create": {method: http.MethodPut, path: "/taxonomy/concepts/", resourceName: "taxonomy concept"},
		"concept update": {method: http.MethodPatch, path: "/taxonomy/concepts/", resourceName: "taxonomy concept", update: true},
		"scheme create":  {method: http.MethodPut, path: "/taxonomy/concept-schemes/", resourceName: "taxonomy concept scheme"},
		"scheme update":  {method: http.MethodPatch, path: "/taxonomy/concept-schemes/", resourceName: "taxonomy concept scheme", update: true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			parallelWhenMocked(t)

			server, err := cmt.NewContentfulManagementServer()
			if err != nil {
				t.Fatal(err)
			}

			mutator := &taxonomyResponseMutator{next: server}
			base := taxonomyConfigVariables("Furniture", "Products")
			steps := []resource.TestStep{}

			if test.update {
				steps = append(steps, resource.TestStep{
					ConfigDirectory: config.StaticDirectory("testdata/TestAccTaxonomyResourcesCreateUpdate"),
					ConfigVariables: base,
				})
			}

			updated := maps.Clone(base)
			if strings.Contains(test.path, "concept-schemes") {
				updated["scheme_label"] = config.StringVariable("Home products")
			} else {
				updated["concept_label"] = config.StringVariable("Home furniture")
			}

			steps = append(steps, resource.TestStep{
				PreConfig: func() {
					mutator.dropPreferredLabelOnce(test.method, test.path, "en-US")
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccTaxonomyResourcesCreateUpdate"),
				ConfigVariables: updated,
				ExpectError:     regexp.MustCompile("Contentful normalized " + test.resourceName + " configuration"),
			})

			ContentfulProviderMockedResourceTest(t, mutator, resource.TestCase{Steps: steps})
		})
	}
}

//nolint:paralleltest
func TestAccTaxonomyConceptPreservesConfiguredLabelMaps(t *testing.T) {
	parallelWhenMocked(t)

	tests := map[string]struct {
		method string
		update bool
	}{
		"create": {method: http.MethodPut},
		"update": {method: http.MethodPatch, update: true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			parallelWhenMocked(t)

			server, err := cmt.NewContentfulManagementServer()
			if err != nil {
				t.Fatal(err)
			}

			mutator := &taxonomyResponseMutator{next: server}
			base := taxonomyConfigVariables("Furniture", "Products")
			steps := []resource.TestStep{}

			if test.update {
				steps = append(steps, resource.TestStep{
					ConfigDirectory: config.StaticDirectory("testdata/TestAccTaxonomyResourcesCreateUpdate"),
					ConfigVariables: base,
				})
			}

			updated := maps.Clone(base)
			updated["concept_label"] = config.StringVariable("Home furniture")
			steps = append(
				steps,
				resource.TestStep{
					PreConfig: func() {
						mutator.addEmptyLabelLocaleOnce(test.method, "/taxonomy/concepts/", "en-US")
					},
					ConfigDirectory: config.StaticDirectory("testdata/TestAccTaxonomyResourcesCreateUpdate"),
					ConfigVariables: updated,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("contentful_taxonomy_concept.test", "alt_labels.%", "1"),
						resource.TestCheckResourceAttr("contentful_taxonomy_concept.test", "hidden_labels.%", "1"),
						resource.TestCheckNoResourceAttr("contentful_taxonomy_concept.test", "alt_labels.en-US"),
						resource.TestCheckNoResourceAttr("contentful_taxonomy_concept.test", "hidden_labels.en-US"),
					),
				},
				resource.TestStep{
					ConfigDirectory: config.StaticDirectory("testdata/TestAccTaxonomyResourcesCreateUpdate"),
					ConfigVariables: updated,
					PlanOnly:        true,
				},
			)

			ContentfulProviderMockedResourceTest(t, mutator, resource.TestCase{Steps: steps})
		})
	}
}

//nolint:paralleltest
func TestAccTaxonomyResourcesSurfaceVersionConflicts(t *testing.T) {
	parallelWhenMocked(t)

	tests := map[string]struct {
		initialConfig string
		changedConfig string
		method        string
		path          string
		expectedError string
		bumpVersion   func(*cmt.Server) error
	}{
		"concept update": {
			initialConfig: taxonomyConceptConfig("Furniture"), changedConfig: taxonomyConceptConfig("Home furniture"),
			method: http.MethodPatch, path: "/taxonomy/concepts/furniture", expectedError: `changed after\s+this Terraform plan was created`,
			bumpVersion: bumpTaxonomyConceptVersion,
		},
		"concept delete": {
			initialConfig: taxonomyConceptConfig("Furniture"), changedConfig: "# intentionally empty\n",
			method: http.MethodDelete, path: "/taxonomy/concepts/furniture", expectedError: `changed after\s+this Terraform plan was created`,
			bumpVersion: bumpTaxonomyConceptVersion,
		},
		"scheme update": {
			initialConfig: taxonomyConceptSchemeConfig("Products"), changedConfig: taxonomyConceptSchemeConfig("Home products"),
			method: http.MethodPatch, path: "/taxonomy/concept-schemes/products", expectedError: `changed after\s+this Terraform plan was created`,
			bumpVersion: bumpTaxonomyConceptSchemeVersion,
		},
		"scheme delete": {
			initialConfig: taxonomyConceptSchemeConfig("Products"), changedConfig: "# intentionally empty\n",
			method: http.MethodDelete, path: "/taxonomy/concept-schemes/products", expectedError: `changed after\s+this Terraform plan was created`,
			bumpVersion: bumpTaxonomyConceptSchemeVersion,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			parallelWhenMocked(t)

			server, err := cmt.NewContentfulManagementServer()
			if err != nil {
				t.Fatal(err)
			}

			hook := &taxonomyRequestHook{next: server}
			ContentfulProviderMockedResourceTest(t, hook, resource.TestCase{Steps: []resource.TestStep{
				{Config: test.initialConfig},
				{
					PreConfig: func() {
						hook.runOnce(test.method, test.path, func() error { return test.bumpVersion(server) })
					},
					Config: test.changedConfig, ExpectError: regexp.MustCompile(test.expectedError),
				},
				{
					Config: test.initialConfig,
					ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					}},
				},
			}})

			if !hook.wasCalled() {
				t.Fatal("version-conflict hook was not called")
			}
		})
	}
}

//nolint:paralleltest
func TestAccTaxonomyResourcesAllowConcurrentDeletion(t *testing.T) {
	parallelWhenMocked(t)

	tests := map[string]struct {
		initialConfig string
		path          string
		deleteRemote  func(*cmt.Server) error
	}{
		"concept": {
			initialConfig: taxonomyConceptConfig("Furniture"), path: "/taxonomy/concepts/furniture",
			deleteRemote: func(server *cmt.Server) error {
				return deleteTaxonomyConceptRemote(server, "organization-id", "furniture")
			},
		},
		"scheme": {
			initialConfig: taxonomyConceptSchemeConfig("Products"), path: "/taxonomy/concept-schemes/products",
			deleteRemote: func(server *cmt.Server) error {
				return deleteTaxonomyConceptSchemeRemote(server, "organization-id", "products")
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			parallelWhenMocked(t)

			server, err := cmt.NewContentfulManagementServer()
			if err != nil {
				t.Fatal(err)
			}

			hook := &taxonomyRequestHook{next: server}
			ContentfulProviderMockedResourceTest(t, hook, resource.TestCase{Steps: []resource.TestStep{
				{Config: test.initialConfig},
				{
					PreConfig: func() {
						hook.runOnce(http.MethodDelete, test.path, func() error { return test.deleteRemote(server) })
					},
					Config: "# intentionally empty\n",
				},
			}})

			if !hook.wasCalled() {
				t.Fatal("concurrent-deletion hook was not called")
			}
		})
	}
}

//nolint:paralleltest
func TestAccTaxonomyResourcesSurfaceReadFailures(t *testing.T) {
	parallelWhenMocked(t)

	tests := map[string]struct {
		config        string
		path          string
		expectedError string
		resource      string
	}{
		"concept": {
			config: taxonomyConceptConfig("Furniture"), path: "/taxonomy/concepts/furniture",
			expectedError: "Failed to read taxonomy concept", resource: "contentful_taxonomy_concept.test",
		},
		"scheme": {
			config: taxonomyConceptSchemeConfig("Products"), path: "/taxonomy/concept-schemes/products",
			expectedError: "Failed to read taxonomy concept scheme", resource: "contentful_taxonomy_concept_scheme.test",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			parallelWhenMocked(t)

			server, err := cmt.NewContentfulManagementServer()
			if err != nil {
				t.Fatal(err)
			}

			failer := &taxonomyResponseFailure{next: server}
			ContentfulProviderMockedResourceTest(t, failer, resource.TestCase{Steps: []resource.TestStep{
				{Config: test.config},
				{
					PreConfig:   func() { failer.failOnce(http.MethodGet, test.path, 1) },
					Config:      test.config,
					ExpectError: regexp.MustCompile(test.expectedError),
				},
				{
					Config: test.config,
					ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(test.resource, plancheck.ResourceActionNoop),
					}},
				},
			}})

			if !failer.wasCalled() {
				t.Fatal("read-failure handler was not called")
			}
		})
	}
}

//nolint:paralleltest
func TestAccTaxonomyResourcesSurfacePreUpdateFailures(t *testing.T) {
	parallelWhenMocked(t)

	tests := map[string]struct {
		initialConfig string
		changedConfig string
		path          string
		expectedError string
		resource      string
	}{
		"concept": {
			initialConfig: taxonomyConceptConfig("Furniture"), changedConfig: taxonomyConceptConfig("Home furniture"),
			path: "/taxonomy/concepts/furniture", expectedError: "Failed to refresh taxonomy concept before update", resource: "contentful_taxonomy_concept.test",
		},
		"scheme": {
			initialConfig: taxonomyConceptSchemeConfig("Products"), changedConfig: taxonomyConceptSchemeConfig("Home products"),
			path: "/taxonomy/concept-schemes/products", expectedError: "Failed to refresh taxonomy concept scheme before update", resource: "contentful_taxonomy_concept_scheme.test",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			parallelWhenMocked(t)

			server, err := cmt.NewContentfulManagementServer()
			if err != nil {
				t.Fatal(err)
			}

			failer := &taxonomyResponseFailure{next: server}
			ContentfulProviderMockedResourceTest(t, failer, resource.TestCase{Steps: []resource.TestStep{
				{Config: test.initialConfig},
				{
					PreConfig:   func() { failer.failOnce(http.MethodGet, test.path, 2) },
					Config:      test.changedConfig,
					ExpectError: regexp.MustCompile(test.expectedError),
				},
				{
					Config: test.initialConfig,
					ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(test.resource, plancheck.ResourceActionNoop),
					}},
				},
			}})

			if !failer.wasCalled() {
				t.Fatal("pre-update failure handler was not called")
			}
		})
	}
}

//nolint:paralleltest
func TestAccTaxonomyTaintedReplacementFetchesMissingDeleteVersion(t *testing.T) {
	parallelWhenMocked(t)

	tests := map[string]struct {
		resourceName string
		config       string
		path         string
		bumpVersion  func(*cmt.Server) error
	}{
		"concept": {
			resourceName: "contentful_taxonomy_concept.test", config: taxonomyConceptConfig("Furniture"),
			path: "/taxonomy/concepts/furniture", bumpVersion: bumpTaxonomyConceptVersion,
		},
		"scheme": {
			resourceName: "contentful_taxonomy_concept_scheme.test", config: taxonomyConceptSchemeConfig("Products"),
			path: "/taxonomy/concept-schemes/products", bumpVersion: bumpTaxonomyConceptSchemeVersion,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			parallelWhenMocked(t)

			server, err := cmt.NewContentfulManagementServer()
			require.NoError(t, err)

			recorder := &taxonomyRequestBodyRecorder{next: server}
			ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{
				AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
				Steps: []resource.TestStep{
					{Config: test.config},
					{
						PreConfig: func() { require.NoError(t, test.bumpVersion(server)) },
						Config:    test.config, Taint: []string{test.resourceName},
					},
				},
			})

			gets := recorder.matchingRequests(http.MethodGet, test.path)
			require.Len(t, gets, 1, "only deletion without private version should fetch the resource")

			deletes := recorder.matchingRequests(http.MethodDelete, test.path)
			require.GreaterOrEqual(t, len(deletes), 2, "tainted replacement and test cleanup should each delete")
			assert.Equal(t, "2", deletes[0].version, "tainted replacement must delete the updated resource version")

			for _, request := range deletes[1:] {
				assert.Equal(t, "1", request.version, "cleanup must use the replacement resource's stored version")
			}
		})
	}
}

//nolint:paralleltest
func TestAccTaxonomyTaintedReplacementPreservesDeleteVersionLock(t *testing.T) {
	parallelWhenMocked(t)

	tests := map[string]struct {
		resourceName string
		config       string
		path         string
		bumpVersion  func(*cmt.Server) error
	}{
		"concept": {
			resourceName: "contentful_taxonomy_concept.test", config: taxonomyConceptConfig("Furniture"),
			path: "/taxonomy/concepts/furniture", bumpVersion: bumpTaxonomyConceptVersion,
		},
		"scheme": {
			resourceName: "contentful_taxonomy_concept_scheme.test", config: taxonomyConceptSchemeConfig("Products"),
			path: "/taxonomy/concept-schemes/products", bumpVersion: bumpTaxonomyConceptSchemeVersion,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			parallelWhenMocked(t)

			server, err := cmt.NewContentfulManagementServer()
			require.NoError(t, err)

			hook := &taxonomyRequestHook{next: server}
			recorder := &taxonomyRequestBodyRecorder{next: hook}
			ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{
				AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
				Steps: []resource.TestStep{
					{Config: test.config},
					{
						PreConfig: func() {
							require.NoError(t, test.bumpVersion(server))
							hook.runOnce(http.MethodDelete, test.path, func() error { return test.bumpVersion(server) })
						},
						Config: test.config, Taint: []string{test.resourceName},
						ExpectError: regexp.MustCompile(`changed after\s+this Terraform plan was created`),
					},
					{Config: test.config},
				},
			})

			require.True(t, hook.wasCalled(), "concurrent change hook was not called")

			gets := recorder.matchingRequests(http.MethodGet, test.path)
			require.Len(t, gets, 2, "each deletion attempt without private version should fetch once")

			deletes := recorder.matchingRequests(http.MethodDelete, test.path)
			require.GreaterOrEqual(t, len(deletes), 3)
			assert.Equal(t, "2", deletes[0].version, "the first attempt must retain the fetched lock token")
			assert.Equal(t, "3", deletes[1].version, "the retry must fetch the version advanced by the concurrent change")

			for _, request := range deletes[2:] {
				assert.Equal(t, "1", request.version, "cleanup must use the replacement resource's stored version")
			}
		})
	}
}

//nolint:paralleltest
func TestAccTaxonomyTaintedReplacementHandlesMissingResource(t *testing.T) {
	parallelWhenMocked(t)

	tests := map[string]struct {
		resourceName string
		config       string
		path         string
		deleteRemote func(*cmt.Server) error
	}{
		"concept": {
			resourceName: "contentful_taxonomy_concept.test", config: taxonomyConceptConfig("Furniture"), path: "/taxonomy/concepts/furniture",
			deleteRemote: func(server *cmt.Server) error {
				return deleteTaxonomyConceptRemote(server, "organization-id", "furniture")
			},
		},
		"scheme": {
			resourceName: "contentful_taxonomy_concept_scheme.test", config: taxonomyConceptSchemeConfig("Products"), path: "/taxonomy/concept-schemes/products",
			deleteRemote: func(server *cmt.Server) error {
				return deleteTaxonomyConceptSchemeRemote(server, "organization-id", "products")
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			parallelWhenMocked(t)

			server, err := cmt.NewContentfulManagementServer()
			require.NoError(t, err)

			recorder := &taxonomyRequestBodyRecorder{next: server}
			ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{
				AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
				Steps: []resource.TestStep{
					{Config: test.config},
					{
						PreConfig: func() { require.NoError(t, test.deleteRemote(server)) },
						Config:    test.config, Taint: []string{test.resourceName},
					},
				},
			})

			require.Len(t, recorder.matchingRequests(http.MethodGet, test.path), 1, "missing private version should trigger one GET")

			deletes := recorder.matchingRequests(http.MethodDelete, test.path)
			require.Len(t, deletes, 1, "the 404 fallback must not send DELETE for the already absent resource")
			assert.Equal(t, "1", deletes[0].version, "test cleanup must use the replacement resource's stored version")
		})
	}
}

//nolint:paralleltest
func TestAccTaxonomyResourcesUseImportedVersion(t *testing.T) {
	parallelWhenMocked(t)

	tests := map[string]struct {
		resourceName       string
		importID           string
		path               string
		initialConfig      string
		changedConfig      string
		importVerifyIgnore []string
	}{
		"concept": {
			resourceName: "contentful_taxonomy_concept.test", importID: "organization-id/furniture", path: "/taxonomy/concepts/furniture",
			initialConfig: taxonomyConceptConfig("Furniture"), changedConfig: taxonomyConceptConfig("Home furniture"),
			importVerifyIgnore: []string{"alt_labels", "hidden_labels"},
		},
		"scheme": {
			resourceName: "contentful_taxonomy_concept_scheme.test", importID: "organization-id/products", path: "/taxonomy/concept-schemes/products",
			initialConfig: taxonomyConceptSchemeConfig("Products"), changedConfig: taxonomyConceptSchemeConfig("Home products"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			parallelWhenMocked(t)

			server, err := cmt.NewContentfulManagementServer()
			require.NoError(t, err)

			recorder := &taxonomyRequestBodyRecorder{next: server}
			ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{
				AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
				Steps: []resource.TestStep{
					{Config: test.initialConfig},
					{
						Config:                  test.initialConfig,
						ResourceName:            test.resourceName,
						ImportState:             true,
						ImportStateId:           test.importID,
						ImportStateVerify:       true,
						ImportStateVerifyIgnore: test.importVerifyIgnore,
					},
					{Config: test.changedConfig},
				},
			})

			requests := recorder.matchingRequests(http.MethodPatch, test.path)
			require.Len(t, requests, 1)
			assert.Equal(t, "1", requests[0].version)

			var patch []struct {
				Path string `json:"path"`
			}
			require.NoError(t, json.Unmarshal(requests[0].body, &patch))
			require.Len(t, patch, 1)
			assert.Equal(t, "/prefLabel", patch[0].Path)
		})
	}
}

//nolint:paralleltest
func taxonomyConceptConfig(label string) string {
	return fmt.Sprintf(`
resource "contentful_taxonomy_concept" "test" {
  organization_id = "organization-id"
  concept_id      = "furniture"
  pref_label      = { "en-US" = %q }
}
`, label)
}

func taxonomyConceptSchemeConfig(label string) string {
	return fmt.Sprintf(`
resource "contentful_taxonomy_concept_scheme" "test" {
  organization_id  = "organization-id"
  concept_scheme_id = "products"
  pref_label        = { "en-US" = %q }
  top_concept_ids   = []
  concept_ids       = []
}
`, label)
}

func bumpTaxonomyConceptVersion(server *cmt.Server) error {
	ctx := context.Background()

	response, err := server.Handler().GetTaxonomyConcept(ctx, cm.GetTaxonomyConceptParams{
		OrganizationID: "organization-id", TaxonomyConceptID: "furniture",
	})
	if err != nil {
		return fmt.Errorf("get taxonomy concept: %w", err)
	}

	concept, ok := response.(*cm.TaxonomyConcept)
	if !ok {
		return fmt.Errorf("%w: taxonomy concept: %T", errUnexpectedTaxonomyResponse, response)
	}

	patchResponse, err := server.Handler().PatchTaxonomyConcept(ctx, cm.TaxonomyPatch{}, cm.PatchTaxonomyConceptParams{
		OrganizationID: "organization-id", TaxonomyConceptID: "furniture", XContentfulVersion: concept.Sys.Version,
	})
	if err != nil {
		return fmt.Errorf("patch taxonomy concept: %w", err)
	}

	if _, ok := patchResponse.(*cm.TaxonomyConcept); !ok {
		return fmt.Errorf("%w: taxonomy concept patch: %T", errUnexpectedTaxonomyResponse, patchResponse)
	}

	return nil
}

func bumpTaxonomyConceptSchemeVersion(server *cmt.Server) error {
	ctx := context.Background()

	response, err := server.Handler().GetTaxonomyConceptScheme(ctx, cm.GetTaxonomyConceptSchemeParams{
		OrganizationID: "organization-id", TaxonomyConceptSchemeID: "products",
	})
	if err != nil {
		return fmt.Errorf("get taxonomy concept scheme: %w", err)
	}

	scheme, ok := response.(*cm.TaxonomyConceptScheme)
	if !ok {
		return fmt.Errorf("%w: taxonomy concept scheme: %T", errUnexpectedTaxonomyResponse, response)
	}

	patchResponse, err := server.Handler().PatchTaxonomyConceptScheme(ctx, cm.TaxonomyPatch{}, cm.PatchTaxonomyConceptSchemeParams{
		OrganizationID: "organization-id", TaxonomyConceptSchemeID: "products", XContentfulVersion: scheme.Sys.Version,
	})
	if err != nil {
		return fmt.Errorf("patch taxonomy concept scheme: %w", err)
	}

	if _, ok := patchResponse.(*cm.TaxonomyConceptScheme); !ok {
		return fmt.Errorf("%w: taxonomy concept scheme patch: %T", errUnexpectedTaxonomyResponse, patchResponse)
	}

	return nil
}

func taxonomyConfigVariables(conceptLabel, schemeLabel string) config.Variables {
	return config.Variables{
		"organization_id":   config.StringVariable("organization-id"),
		"concept_id":        config.StringVariable("furniture"),
		"concept_scheme_id": config.StringVariable("products"),
		"concept_label":     config.StringVariable(conceptLabel),
		"scheme_label":      config.StringVariable(schemeLabel),
	}
}
