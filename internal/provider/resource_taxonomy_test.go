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
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	errUnexpectedTaxonomyResponse    = errors.New("unexpected taxonomy response")
	errUnexpectedTaxonomyImportState = errors.New("unexpected imported taxonomy concept state")
)

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
			Config:            taxonomyLifecycleConfig(ids, taxonomyLifecycleClear),
			ResourceName:      "contentful_taxonomy_concept.test",
			ImportState:       true,
			ImportStateId:     taxonomyAcceptanceOrganizationID + "/" + ids.concept,
			ImportStateVerify: true,
			// Import has no prior locale representation to preserve. The ownership
			// and imported-label scenarios below check CMA-added empty locales separately.
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

	var broader, related, schemeConcepts, topConcepts []string

	var notations knownvalue.Check

	switch stage {
	case taxonomyLifecycleCreate:
		conceptURI = knownvalue.Null()
		conceptDefinition = knownvalue.Null()
		schemeURI = knownvalue.StringExact("https://example.com/schemes/initial")
		schemeDefinition = localizedKnownValue("Initial scheme definition")
		conceptLabel, altLabel = "Initial concept", ""
		schemeLabel = "Lifecycle scheme"
		notations = stringListKnownValue(nil)
		broader, related = []string{ids.parent1}, []string{ids.related}
		schemeConcepts, topConcepts = []string{ids.concept, ids.parent1, ids.related}, []string{ids.parent1}
	case taxonomyLifecycleUpdate:
		conceptURI = knownvalue.StringExact("https://example.com/concepts/updated")
		conceptDefinition = localizedKnownValue("Updated definition")
		schemeURI = knownvalue.StringExact("https://example.com/schemes/updated")
		schemeDefinition = localizedKnownValue("Updated scheme definition")
		conceptLabel, altLabel = "Updated concept", "Updated alternative"
		schemeLabel = "Lifecycle scheme"
		notations = stringListKnownValue([]string{"UPDATED"})
		broader, related = []string{ids.parent2}, []string{ids.parent1}
		schemeConcepts, topConcepts = []string{ids.concept, ids.parent2}, []string{ids.parent2}
	case taxonomyLifecyclePreserveComputed:
		conceptURI = knownvalue.Null()
		conceptDefinition = knownvalue.Null()
		schemeURI = knownvalue.Null()
		schemeDefinition = knownvalue.Null()
		conceptLabel, altLabel = "Preserved concept", "Updated alternative"
		schemeLabel = "Preserved scheme"
		notations = stringListKnownValue([]string{"UPDATED"})
		broader, related = []string{ids.parent2}, []string{ids.parent1}
		schemeConcepts, topConcepts = []string{ids.concept, ids.parent2}, []string{ids.parent2}
	case taxonomyLifecycleClear:
		conceptURI, conceptDefinition = knownvalue.Null(), knownvalue.Null()
		schemeURI, schemeDefinition = knownvalue.Null(), knownvalue.Null()
		conceptLabel, altLabel = "Cleared concept", ""
		schemeLabel = "Cleared scheme"
		notations = stringListKnownValue(nil)
		broader, related = []string{}, []string{}
		schemeConcepts, topConcepts = []string{}, []string{}
	}

	checks = append(
		checks,
		statecheck.ExpectKnownValue(conceptAddress, tfjsonpath.New("uri"), conceptURI),
		statecheck.ExpectKnownValue(conceptAddress, tfjsonpath.New("pref_label"), localizedKnownValue(conceptLabel)),
		statecheck.ExpectKnownValue(conceptAddress, tfjsonpath.New("definition"), conceptDefinition),
		statecheck.ExpectKnownValue(conceptAddress, tfjsonpath.New("notations"), notations),
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
				resource.TestCheckResourceAttr("contentful_taxonomy_concept.test", "alt_labels.en-US.0", "Furniture"),
				resource.TestCheckResourceAttr("contentful_taxonomy_concept.test", "hidden_labels.%", "1"),
				resource.TestCheckResourceAttr("contentful_taxonomy_concept.test", "hidden_labels.en-US.0", "Furnishings"),
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
func TestAccTaxonomyApplySurvivesResponseOwnedCascade(t *testing.T) {
	parallelWhenMocked(t)

	tests := map[string]struct {
		path          string
		initialConfig string
		changedConfig string
		seed          func(*cmt.Server) error
		attach        func(*cmt.Server) error
		delete        func(*cmt.Server) error
		refreshed     resource.TestCheckFunc
		checks        resource.TestCheckFunc
	}{
		"concept": {
			path: "/taxonomy/concepts/furniture", initialConfig: taxonomyConceptConfig("Furniture"), changedConfig: taxonomyConceptConfig("Home furniture"),
			seed: func(server *cmt.Server) error {
				response, err := server.Handler().PutTaxonomyConcept(context.Background(), &cm.TaxonomyConceptRequest{PrefLabel: cm.LocalizedString{"en-US": "Remote"}}, cm.PutTaxonomyConceptParams{OrganizationID: "organization-id", TaxonomyConceptID: "remote"})
				if err != nil {
					return fmt.Errorf("create remote taxonomy concept: %w", err)
				}

				if _, ok := response.(*cm.TaxonomyConcept); !ok {
					return fmt.Errorf("%w: remote taxonomy concept: %T", errUnexpectedTaxonomyResponse, response)
				}

				return nil
			},
			attach: attachRemoteConceptToConcept, delete: func(server *cmt.Server) error {
				return deleteTaxonomyConceptRemote(server, "organization-id", "remote")
			},
			refreshed: resource.TestCheckResourceAttr("contentful_taxonomy_concept.test", "broader_concept_ids.0", "remote"),
			checks:    resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttr("contentful_taxonomy_concept.test", "pref_label.en-US", "Home furniture"), resource.TestCheckResourceAttr("contentful_taxonomy_concept.test", "broader_concept_ids.#", "0"), resource.TestCheckResourceAttr("contentful_taxonomy_concept.test", "related_concept_ids.#", "0")),
		},
		"scheme": {
			path: "/taxonomy/concept-schemes/products", initialConfig: taxonomyConceptSchemeOmittedCollectionsConfig("Products"), changedConfig: taxonomyConceptSchemeOmittedCollectionsConfig("Home products"),
			seed: func(server *cmt.Server) error {
				response, err := server.Handler().PutTaxonomyConcept(context.Background(), &cm.TaxonomyConceptRequest{PrefLabel: cm.LocalizedString{"en-US": "Remote"}}, cm.PutTaxonomyConceptParams{OrganizationID: "organization-id", TaxonomyConceptID: "remote"})
				if err != nil {
					return fmt.Errorf("create remote taxonomy concept: %w", err)
				}

				if _, ok := response.(*cm.TaxonomyConcept); !ok {
					return fmt.Errorf("%w: remote taxonomy concept: %T", errUnexpectedTaxonomyResponse, response)
				}

				return nil
			},
			attach: attachRemoteConceptToScheme, delete: func(server *cmt.Server) error {
				return deleteTaxonomyConceptRemote(server, "organization-id", "remote")
			},
			refreshed: resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttr("contentful_taxonomy_concept_scheme.test", "concept_ids.0", "remote"), resource.TestCheckResourceAttr("contentful_taxonomy_concept_scheme.test", "top_concept_ids.0", "remote")),
			checks:    resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttr("contentful_taxonomy_concept_scheme.test", "pref_label.en-US", "Home products"), resource.TestCheckResourceAttr("contentful_taxonomy_concept_scheme.test", "concept_ids.#", "0"), resource.TestCheckResourceAttr("contentful_taxonomy_concept_scheme.test", "top_concept_ids.#", "0")),
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			parallelWhenMocked(t)

			server, err := cmt.NewContentfulManagementServer()
			require.NoError(t, err)

			hook := &taxonomyRequestHook{next: server}
			recorder := &taxonomyRequestBodyRecorder{next: hook}
			ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{Steps: []resource.TestStep{
				{Config: test.initialConfig},
				{PreConfig: func() { require.NoError(t, test.seed(server)); require.NoError(t, test.attach(server)) }, Config: test.initialConfig, Check: test.refreshed},
				{PreConfig: func() {
					recorder.reset()
					hook.runOnce(http.MethodPatch, test.path, func() error { return test.delete(server) })
				}, Config: test.changedConfig, Check: test.checks},
			}})
			require.True(t, hook.wasCalled())

			patches := recorder.matchingRequests(http.MethodPatch, test.path)
			require.NotEmpty(t, patches)

			for _, request := range patches {
				assert.Equal(t, "2", request.version)
				assert.Equal(t, patches[0].body, request.body)
			}

			var patch []struct {
				Path string `json:"path"`
			}
			require.NoError(t, json.Unmarshal(patches[0].body, &patch))

			paths := make([]string, 0, len(patch))
			for _, operation := range patch {
				paths = append(paths, operation.Path)
			}

			require.Equal(t, []string{"/prefLabel"}, paths)
			assert.GreaterOrEqual(t, len(recorder.matchingRequests(http.MethodGet, test.path)), 2)
		})
	}
}

//nolint:paralleltest
func TestAccTaxonomyResourcesRejectResponseIdentityRetargeting(t *testing.T) {
	parallelWhenMocked(t)

	tests := map[string]struct {
		method        string
		path          string
		resourceName  string
		initialConfig string
		changedConfig string
		update        bool
	}{
		"concept create": {
			method: http.MethodPut, path: "/taxonomy/concepts/furniture", resourceName: "taxonomy concept",
			initialConfig: taxonomyConceptConfig("Furniture"), changedConfig: taxonomyConceptConfig("Furniture"),
		},
		"concept update": {
			method: http.MethodPatch, path: "/taxonomy/concepts/furniture", resourceName: "taxonomy concept",
			initialConfig: taxonomyConceptConfig("Furniture"), changedConfig: taxonomyConceptConfig("Home furniture"), update: true,
		},
		"scheme create": {
			method: http.MethodPut, path: "/taxonomy/concept-schemes/products", resourceName: "taxonomy concept scheme",
			initialConfig: taxonomyConceptSchemeConfig("Products"), changedConfig: taxonomyConceptSchemeConfig("Products"),
		},
		"scheme update": {
			method: http.MethodPatch, path: "/taxonomy/concept-schemes/products", resourceName: "taxonomy concept scheme",
			initialConfig: taxonomyConceptSchemeConfig("Products"), changedConfig: taxonomyConceptSchemeConfig("Home products"), update: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			parallelWhenMocked(t)

			server, err := cmt.NewContentfulManagementServer()
			require.NoError(t, err)

			mutator := &taxonomyResponseMutator{next: server}
			recorder := &taxonomyRequestBodyRecorder{next: mutator}
			steps := []resource.TestStep{}

			if test.update {
				steps = append(steps, resource.TestStep{Config: test.initialConfig})
			}

			steps = append(steps, resource.TestStep{
				PreConfig: func() {
					mutator.replaceIdentityOnce(test.method, test.path, "other-organization", "other-resource")
				},
				Config:      test.changedConfig,
				ExpectError: regexp.MustCompile("Unexpected Contentful " + test.resourceName + " response"),
			})

			if test.update {
				steps = append(steps, resource.TestStep{
					Config: test.changedConfig,
					ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					}},
				})
			} else {
				steps = append(steps, resource.TestStep{Config: test.changedConfig})
			}

			ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{
				AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
				Steps:                steps,
			})

			if test.update {
				requests := recorder.matchingRequests(http.MethodPatch, test.path)
				require.NotEmpty(t, requests)

				for _, request := range requests {
					assert.Equal(t, "1", request.version)
					assert.Equal(t, requests[0].body, request.body)
				}
			} else {
				deletes := recorder.matchingRequests(http.MethodDelete, test.path)
				require.NotEmpty(t, deletes)

				assert.Equal(t, "1", deletes[0].version)
			}
		})
	}
}

//nolint:paralleltest
func TestAccTaxonomyConceptResourcePreservesExplicitEmptyLabelMapsAgainstCanonicalResponse(t *testing.T) {
	parallelWhenMocked(t)

	tests := map[string]struct {
		update bool
	}{
		"create": {},
		"update": {update: true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			parallelWhenMocked(t)

			server, err := cmt.NewContentfulManagementServer()
			if err != nil {
				t.Fatal(err)
			}

			recorder := &taxonomyRequestBodyRecorder{next: server}
			base := taxonomyConceptEmptyLabelMapsConfig("Chair")
			steps := []resource.TestStep{}

			if test.update {
				steps = append(steps, resource.TestStep{Config: base})
			}

			updated := taxonomyConceptEmptyLabelMapsConfig("Chaise")
			steps = append(
				steps,
				resource.TestStep{
					Config: updated,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("contentful_taxonomy_concept.test", "alt_labels.%", "0"),
						resource.TestCheckResourceAttr("contentful_taxonomy_concept.test", "hidden_labels.%", "0"),
					),
				},
				resource.TestStep{
					Config: updated,
					ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					}},
				},
			)

			ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{Steps: steps})

			if !test.update {
				return
			}

			requests := recorder.matchingRequests(http.MethodPatch, "/taxonomy/concepts/furniture")
			require.Len(t, requests, 1)

			var patch []struct {
				Path string `json:"path"`
			}

			require.NoError(t, json.Unmarshal(requests[0].body, &patch))
			require.Len(t, patch, 1)
			require.Equal(t, []string{"/prefLabel"}, []string{patch[0].Path})
		})
	}
}

//nolint:paralleltest
func TestAccTaxonomyConceptResourceProjectsOutOfBandLabelLocalesByOwnership(t *testing.T) {
	parallelWhenMocked(t)

	tests := map[string]struct {
		config         string
		expectedAction plancheck.ResourceActionType
		check          resource.TestCheckFunc
	}{
		"configured maps remove drift": {
			config:         taxonomyConceptConfiguredLabelMapsConfig("Furniture"),
			expectedAction: plancheck.ResourceActionUpdate,
			check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("contentful_taxonomy_concept.test", "alt_labels.%", "1"),
				resource.TestCheckResourceAttr("contentful_taxonomy_concept.test", "alt_labels.en-US.0", "Furniture"),
				resource.TestCheckResourceAttr("contentful_taxonomy_concept.test", "hidden_labels.%", "1"),
				resource.TestCheckResourceAttr("contentful_taxonomy_concept.test", "hidden_labels.en-US.0", "Furnishings"),
			),
		},
		"explicit empty maps remove drift": {
			config:         taxonomyConceptEmptyLabelMapsConfig("Furniture"),
			expectedAction: plancheck.ResourceActionUpdate,
			check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("contentful_taxonomy_concept.test", "alt_labels.%", "0"),
				resource.TestCheckResourceAttr("contentful_taxonomy_concept.test", "hidden_labels.%", "0"),
			),
		},
		"omitted maps retain drift": {
			config:         taxonomyConceptConfig("Furniture"),
			expectedAction: plancheck.ResourceActionNoop,
			check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("contentful_taxonomy_concept.test", "alt_labels.fr-FR.0", "Fauteuil"),
				resource.TestCheckResourceAttr("contentful_taxonomy_concept.test", "hidden_labels.fr-FR.0", "Siege"),
			),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			parallelWhenMocked(t)

			server, err := cmt.NewContentfulManagementServer()
			if err != nil {
				t.Fatal(err)
			}

			ContentfulProviderMockedResourceTest(t, server, resource.TestCase{Steps: []resource.TestStep{
				{Config: test.config},
				{
					PreConfig: func() {
						injectTaxonomyConceptLabelLocaleIntoStoredResponse(t, server, "organization-id", "furniture", "fr-FR", "Fauteuil", "Siege")
					},
					Config: test.config,
					ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_taxonomy_concept.test", test.expectedAction),
					}},
					Check: test.check,
				},
			}})
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
			method: http.MethodPatch, path: "/taxonomy/concepts/furniture", expectedError: `version precondition was not\s+satisfied`,
			bumpVersion: bumpTaxonomyConceptVersion,
		},
		"concept delete": {
			initialConfig: taxonomyConceptConfig("Furniture"), changedConfig: "# intentionally empty\n",
			method: http.MethodDelete, path: "/taxonomy/concepts/furniture", expectedError: `version precondition was not\s+satisfied`,
			bumpVersion: bumpTaxonomyConceptVersion,
		},
		"scheme update": {
			initialConfig: taxonomyConceptSchemeConfig("Products"), changedConfig: taxonomyConceptSchemeConfig("Home products"),
			method: http.MethodPatch, path: "/taxonomy/concept-schemes/products", expectedError: `version precondition was not\s+satisfied`,
			bumpVersion: bumpTaxonomyConceptSchemeVersion,
		},
		"scheme delete": {
			initialConfig: taxonomyConceptSchemeConfig("Products"), changedConfig: "# intentionally empty\n",
			method: http.MethodDelete, path: "/taxonomy/concept-schemes/products", expectedError: `version precondition was not\s+satisfied`,
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
func TestAccTaxonomyResourcesRecoverRemoteStateAfterMutationMismatch(t *testing.T) {
	parallelWhenMocked(t)

	tests := map[string]struct {
		path             string
		initialConfig    string
		mismatchedConfig string
		recoveredConfig  string
		prepareRemote    func(*cmt.Server) error
		mutateRemote     func(*cmt.Server) error
	}{
		"concept owned label map": {
			path:             "/taxonomy/concepts/furniture",
			initialConfig:    taxonomyConceptRecoveryLabelMapsConfig("Furniture", "Furniture"),
			mismatchedConfig: taxonomyConceptRecoveryLabelMapsConfig("Home furniture", "Furniture"),
			recoveredConfig:  taxonomyConceptRecoveredLabelMapConfig("Final furniture"),
			mutateRemote: func(server *cmt.Server) error {
				return mutateTaxonomyConceptAltLabels(server, "en-US", []string{"Remote furniture"})
			},
		},
		"scheme owned concepts": {
			path:             "/taxonomy/concept-schemes/products",
			initialConfig:    taxonomyConceptSchemeConfig("Products"),
			mismatchedConfig: taxonomyConceptSchemeConfig("Home products"),
			recoveredConfig:  taxonomyConceptSchemeRecoveredConceptsConfig("Final products"),
			prepareRemote: func(server *cmt.Server) error {
				request := cm.TaxonomyConceptRequest{PrefLabel: cm.LocalizedString{"en-US": "Furniture"}}

				_, err := server.Handler().PutTaxonomyConcept(context.Background(), &request, cm.PutTaxonomyConceptParams{
					OrganizationID: "organization-id", TaxonomyConceptID: "furniture",
				})
				if err != nil {
					return fmt.Errorf("create taxonomy concept: %w", err)
				}

				return nil
			},
			mutateRemote: func(server *cmt.Server) error {
				response, err := server.Handler().GetTaxonomyConceptScheme(context.Background(), cm.GetTaxonomyConceptSchemeParams{
					OrganizationID: "organization-id", TaxonomyConceptSchemeID: "products",
				})
				if err != nil {
					return fmt.Errorf("get taxonomy concept scheme: %w", err)
				}

				scheme, ok := response.(*cm.TaxonomyConceptScheme)
				if !ok {
					return fmt.Errorf("%w: taxonomy concept scheme: %T", errUnexpectedTaxonomyResponse, response)
				}

				scheme.Concepts = []cm.TaxonomyConceptLink{cm.NewTaxonomyConceptLink("furniture")}

				return nil
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			parallelWhenMocked(t)

			server, err := cmt.NewContentfulManagementServer()
			require.NoError(t, err)

			if test.prepareRemote != nil {
				require.NoError(t, test.prepareRemote(server))
			}

			recorder := &taxonomyRequestBodyRecorder{next: server}

			ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{
				AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
				Steps: []resource.TestStep{
					{Config: test.initialConfig},
					{
						PreConfig: func() {
							require.NoError(t, test.mutateRemote(server))
						},
						Config:      test.mismatchedConfig,
						ExpectError: regexp.MustCompile("Unexpected Contentful taxonomy"),
					},
					{Config: test.recoveredConfig},
				},
			})

			requests := recorder.matchingRequests(http.MethodPatch, test.path)
			require.Len(t, requests, 2)
			assert.Equal(t, "2", requests[1].version, "recovered remote version must be used")

			var patch []struct {
				Path string `json:"path"`
			}
			require.NoError(t, json.Unmarshal(requests[1].body, &patch))
			require.Len(t, patch, 1)
			assert.Equal(t, "/prefLabel", patch[0].Path)
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
func TestAccTaxonomyResourcesDoNotGETInsideUpdate(t *testing.T) {
	parallelWhenMocked(t)

	tests := map[string]struct {
		initialConfig string
		changedConfig string
		path          string
	}{
		"concept": {
			initialConfig: taxonomyConceptConfig("Furniture"), changedConfig: taxonomyConceptConfig("Home furniture"),
			path: "/taxonomy/concepts/furniture",
		},
		"scheme": {
			initialConfig: taxonomyConceptSchemeConfig("Products"), changedConfig: taxonomyConceptSchemeConfig("Home products"),
			path: "/taxonomy/concept-schemes/products",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			parallelWhenMocked(t)

			server, err := cmt.NewContentfulManagementServer()
			if err != nil {
				t.Fatal(err)
			}

			recorder := &taxonomyRequestRecorder{next: server}
			ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{Steps: []resource.TestStep{
				{Config: test.initialConfig},
				{
					PreConfig: recorder.reset,
					Config:    test.changedConfig,
				},
			}})

			if count := recorder.count(http.MethodGet, test.path); count != 2 {
				t.Errorf("plan and post-apply refresh GET count = %d, want 2", count)
			}

			if count := recorder.count(http.MethodPatch, test.path); count != 1 {
				t.Errorf("update PATCH count = %d, want 1", count)
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
						ExpectError: regexp.MustCompile(`version precondition was not\s+satisfied`),
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
func TestAccTaxonomyCreateRequestPreservesCollectionOwnership(t *testing.T) {
	parallelWhenMocked(t)

	tests := []struct {
		name           string
		config         string
		path           string
		absent         []string
		expectedFields map[string]string
	}{
		{
			name:   "concept omitted collections are omitted",
			config: taxonomyConceptConfig("Furniture"),
			path:   "/taxonomy/concepts/furniture",
			absent: []string{"altLabels", "hiddenLabels", "notations", "broader", "related"},
		},
		{
			name:   "concept explicit empty collections are sent",
			config: taxonomyConceptAllCollectionsConfig("Furniture"),
			path:   "/taxonomy/concepts/furniture",
			expectedFields: map[string]string{
				"altLabels":    `{"en-US":[]}`,
				"hiddenLabels": `{"en-US":[]}`,
				"notations":    `[]`,
				"broader":      `[]`,
				"related":      `[]`,
			},
		},
		{
			name:   "scheme omitted response-owned arrays are omitted",
			config: taxonomyConceptSchemeOmittedCollectionsConfig("Products"),
			path:   "/taxonomy/concept-schemes/products",
			absent: []string{"topConcepts", "concepts"},
		},
		{
			name:           "scheme explicit empty arrays remain empty",
			config:         taxonomyConceptSchemeConfig("Products"),
			path:           "/taxonomy/concept-schemes/products",
			expectedFields: map[string]string{"topConcepts": `[]`, "concepts": `[]`},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parallelWhenMocked(t)

			server, err := cmt.NewContentfulManagementServer()
			require.NoError(t, err)

			recorder := &taxonomyRequestBodyRecorder{next: server}
			ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{Steps: []resource.TestStep{{Config: test.config}}})

			fields, ok := recorder.request(http.MethodPut, test.path)
			require.True(t, ok, "missing PUT request for %s", test.path)

			for _, field := range test.absent {
				assert.NotContains(t, fields, field)
			}

			for field, want := range test.expectedFields {
				value, ok := fields[field]
				require.True(t, ok, "missing request field %s", field)
				assert.JSONEq(t, want, string(value))
			}
		})
	}
}

//nolint:paralleltest
func TestAccTaxonomyConceptResourceImportProjectsNonemptyLabelMaps(t *testing.T) {
	parallelWhenMocked(t)

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)

	request := cm.TaxonomyConceptRequest{
		PrefLabel:    cm.LocalizedString{"en-US": "Furniture"},
		AltLabels:    cm.NewOptLocalizedStringList(cm.LocalizedStringList{"en-US": {"Furnishings"}}),
		HiddenLabels: cm.NewOptLocalizedStringList(cm.LocalizedStringList{"en-US": {"Household goods"}}),
	}
	_, err = server.Handler().PutTaxonomyConcept(t.Context(), &request, cm.PutTaxonomyConceptParams{OrganizationID: "organization-id", TaxonomyConceptID: "furniture"})
	require.NoError(t, err)

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{Steps: []resource.TestStep{{
		Config: taxonomyConceptConfig("Furniture"), ResourceName: "contentful_taxonomy_concept.test", ImportState: true, ImportStateId: "organization-id/furniture", ImportStateCheck: taxonomyConceptImportLabelMapsCheck(),
	}}})
}

func taxonomyConceptConfig(label string) string {
	return fmt.Sprintf(`
resource "contentful_taxonomy_concept" "test" {
  organization_id = "organization-id"
  concept_id      = "furniture"
  pref_label      = { "en-US" = %q }
}
`, label)
}

func taxonomyConceptAllCollectionsConfig(label string) string {
	return fmt.Sprintf(`
resource "contentful_taxonomy_concept" "test" {
  organization_id = "organization-id"
  concept_id      = "furniture"
  pref_label      = { "en-US" = %q }
  alt_labels      = {}
  hidden_labels   = {}
  notations       = []
  broader_concept_ids = []
  related_concept_ids = []
}
`, label)
}

func taxonomyConceptConfiguredLabelMapsConfig(label string) string {
	return fmt.Sprintf(`
resource "contentful_taxonomy_concept" "test" {
  organization_id = "organization-id"
  concept_id      = "furniture"
  pref_label      = { "en-US" = %q }
  alt_labels      = { "en-US" = ["Furniture"] }
  hidden_labels   = { "en-US" = ["Furnishings"] }
}
`, label)
}

func taxonomyConceptImportLabelMapsCheck() resource.ImportStateCheckFunc {
	return func(states []*terraform.InstanceState) error {
		if len(states) != 1 {
			return fmt.Errorf("%w: expected one resource, got %d", errUnexpectedTaxonomyImportState, len(states))
		}

		expected := map[string]string{
			"alt_labels.%":          "1",
			"alt_labels.en-US.#":    "1",
			"alt_labels.en-US.0":    "Furnishings",
			"hidden_labels.%":       "1",
			"hidden_labels.en-US.#": "1",
			"hidden_labels.en-US.0": "Household goods",
		}
		for name, want := range expected {
			if got := states[0].Attributes[name]; got != want {
				return fmt.Errorf("%w: attribute %q got %q, want %q", errUnexpectedTaxonomyImportState, name, got, want)
			}
		}

		return nil
	}
}

func taxonomyConceptRecoveredLabelMapConfig(label string) string {
	return taxonomyConceptRecoveryLabelMapsConfig(label, "Remote furniture")
}

func taxonomyConceptRecoveryLabelMapsConfig(label, altLabel string) string {
	return fmt.Sprintf(`
resource "contentful_taxonomy_concept" "test" {
  organization_id = "organization-id"
  concept_id      = "furniture"
  pref_label      = { "en-US" = %q }
  alt_labels      = { "en-US" = [%q] }
  hidden_labels   = { "en-US" = ["Furnishings"] }
  notations       = []
  broader_concept_ids = []
  related_concept_ids = []
}
`, label, altLabel)
}

func taxonomyConceptEmptyLabelMapsConfig(label string) string {
	return fmt.Sprintf(`
resource "contentful_taxonomy_concept" "test" {
  organization_id = "organization-id"
  concept_id      = "furniture"
  pref_label      = { "en-US" = %q }
  alt_labels      = {}
  hidden_labels   = {}
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

func taxonomyConceptSchemeRecoveredConceptsConfig(label string) string {
	return fmt.Sprintf(`
resource "contentful_taxonomy_concept_scheme" "test" {
  organization_id   = "organization-id"
  concept_scheme_id = "products"
  pref_label        = { "en-US" = %q }
  top_concept_ids   = []
  concept_ids       = ["furniture"]
}
`, label)
}

func taxonomyConceptSchemeOmittedCollectionsConfig(label string) string {
	return fmt.Sprintf(`
resource "contentful_taxonomy_concept_scheme" "test" {
  organization_id   = "organization-id"
  concept_scheme_id = "products"
  pref_label        = { "en-US" = %q }
}
`, label)
}

func attachRemoteConceptToConcept(server *cmt.Server) error {
	response, err := server.Handler().GetTaxonomyConcept(context.Background(), cm.GetTaxonomyConceptParams{OrganizationID: "organization-id", TaxonomyConceptID: "furniture"})
	if err != nil {
		return fmt.Errorf("get taxonomy concept: %w", err)
	}

	concept, ok := response.(*cm.TaxonomyConcept)
	if !ok {
		return fmt.Errorf("%w: taxonomy concept: %T", errUnexpectedTaxonomyResponse, response)
	}

	patch := cm.TaxonomyPatch{{Op: cm.TaxonomyPatchItemOpAdd, Path: "/broader", Value: jx.Raw(`[{"sys":{"id":"remote","linkType":"TaxonomyConcept","type":"Link"}}]`)}}

	_, err = server.Handler().PatchTaxonomyConcept(context.Background(), patch, cm.PatchTaxonomyConceptParams{OrganizationID: "organization-id", TaxonomyConceptID: "furniture", XContentfulVersion: concept.Sys.Version})
	if err != nil {
		return fmt.Errorf("patch taxonomy concept: %w", err)
	}

	return nil
}

func attachRemoteConceptToScheme(server *cmt.Server) error {
	response, err := server.Handler().GetTaxonomyConceptScheme(context.Background(), cm.GetTaxonomyConceptSchemeParams{OrganizationID: "organization-id", TaxonomyConceptSchemeID: "products"})
	if err != nil {
		return fmt.Errorf("get taxonomy concept scheme: %w", err)
	}

	scheme, ok := response.(*cm.TaxonomyConceptScheme)
	if !ok {
		return fmt.Errorf("%w: taxonomy concept scheme: %T", errUnexpectedTaxonomyResponse, response)
	}

	patch := cm.TaxonomyPatch{{Op: cm.TaxonomyPatchItemOpAdd, Path: "/concepts", Value: jx.Raw(`[{"sys":{"id":"remote","linkType":"TaxonomyConcept","type":"Link"}}]`)}, {Op: cm.TaxonomyPatchItemOpAdd, Path: "/topConcepts", Value: jx.Raw(`[{"sys":{"id":"remote","linkType":"TaxonomyConcept","type":"Link"}}]`)}}

	_, err = server.Handler().PatchTaxonomyConceptScheme(context.Background(), patch, cm.PatchTaxonomyConceptSchemeParams{OrganizationID: "organization-id", TaxonomyConceptSchemeID: "products", XContentfulVersion: scheme.Sys.Version})
	if err != nil {
		return fmt.Errorf("patch taxonomy concept scheme: %w", err)
	}

	return nil
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

func mutateTaxonomyConceptAltLabels(server *cmt.Server, locale string, labels []string) error {
	response, err := server.Handler().GetTaxonomyConcept(context.Background(), cm.GetTaxonomyConceptParams{
		OrganizationID: "organization-id", TaxonomyConceptID: "furniture",
	})
	if err != nil {
		return fmt.Errorf("get taxonomy concept: %w", err)
	}

	concept, ok := response.(*cm.TaxonomyConcept)
	if !ok {
		return fmt.Errorf("%w: taxonomy concept: %T", errUnexpectedTaxonomyResponse, response)
	}

	concept.AltLabels = cm.NewOptLocalizedStringList(cm.LocalizedStringList{locale: labels})

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

// injectTaxonomyConceptLabelLocaleIntoStoredResponse creates an adversarial
// response-only locale that CMA normalization would not currently return. It
// deliberately bypasses request handling to exercise Read ownership policy.
func injectTaxonomyConceptLabelLocaleIntoStoredResponse(t *testing.T, server *cmt.Server, organizationID, conceptID, locale, altLabel, hiddenLabel string) {
	t.Helper()

	ctx := context.Background()

	response, err := server.Handler().GetTaxonomyConcept(ctx, cm.GetTaxonomyConceptParams{
		OrganizationID: organizationID, TaxonomyConceptID: conceptID,
	})
	if err != nil {
		t.Fatal(err)
	}

	concept, ok := response.(*cm.TaxonomyConcept)
	if !ok {
		t.Fatalf("get taxonomy concept returned %T", response)
	}

	altLabels, _ := concept.AltLabels.Get()
	hiddenLabels, _ := concept.HiddenLabels.Get()
	altLabels = maps.Clone(altLabels)
	hiddenLabels = maps.Clone(hiddenLabels)
	altLabels[locale] = []string{altLabel}
	hiddenLabels[locale] = []string{hiddenLabel}

	concept.AltLabels = cm.NewOptLocalizedStringList(altLabels)
	concept.HiddenLabels = cm.NewOptLocalizedStringList(hiddenLabels)
	concept.Sys.Version++
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

//nolint:paralleltest
func TestAccTaxonomyReadIdentityFailuresRetainPriorState(t *testing.T) {
	parallelWhenMocked(t)

	tests := map[string]struct {
		path          string
		initialConfig string
		mutate        func(*taxonomyResponseMutator, string)
		expectError   *regexp.Regexp
		resourceName  string
	}{
		"concept changed": {
			path: "/taxonomy/concepts/furniture", initialConfig: taxonomyConceptConfig("Furniture"), resourceName: "contentful_taxonomy_concept.test", expectError: regexp.MustCompile("Unexpected Identity Change"),
			mutate: func(mutator *taxonomyResponseMutator, path string) {
				mutator.replaceIdentityOnce(http.MethodGet, path, "other-organization", "other-resource")
			},
		},
		"concept missing": {
			path: "/taxonomy/concepts/furniture", initialConfig: taxonomyConceptConfig("Furniture"), resourceName: "contentful_taxonomy_concept.test", expectError: regexp.MustCompile("Failed to read taxonomy concept"),
			mutate: func(mutator *taxonomyResponseMutator, path string) { mutator.removeIdentityOnce(http.MethodGet, path) },
		},
		"scheme changed": {
			path: "/taxonomy/concept-schemes/products", initialConfig: taxonomyConceptSchemeConfig("Products"), resourceName: "contentful_taxonomy_concept_scheme.test", expectError: regexp.MustCompile("Unexpected Identity Change"),
			mutate: func(mutator *taxonomyResponseMutator, path string) {
				mutator.replaceIdentityOnce(http.MethodGet, path, "other-organization", "other-resource")
			},
		},
		"scheme missing": {
			path: "/taxonomy/concept-schemes/products", initialConfig: taxonomyConceptSchemeConfig("Products"), resourceName: "contentful_taxonomy_concept_scheme.test", expectError: regexp.MustCompile("Failed to read taxonomy concept scheme"),
			mutate: func(mutator *taxonomyResponseMutator, path string) { mutator.removeIdentityOnce(http.MethodGet, path) },
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			parallelWhenMocked(t)

			server, err := cmt.NewContentfulManagementServer()
			require.NoError(t, err)

			mutator := &taxonomyResponseMutator{next: server}

			ContentfulProviderMockedResourceTest(t, mutator, resource.TestCase{Steps: []resource.TestStep{
				{Config: test.initialConfig},
				{PreConfig: func() { test.mutate(mutator, test.path) }, Config: test.initialConfig, ExpectError: test.expectError},
				{Config: test.initialConfig, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(test.resourceName, plancheck.ResourceActionNoop)}}},
			}})
		})
	}
}

//nolint:paralleltest
func TestAccTaxonomyExplicitEmptyLocalizedMapsRemainStable(t *testing.T) {
	parallelWhenMocked(t)

	tests := map[string]struct {
		resourceName string
		path         string
		config       string
		seedDrift    func(*cmt.Server) error
		fields       []string
	}{
		"concept": {
			resourceName: "contentful_taxonomy_concept.test", path: "/taxonomy/concepts/furniture", config: taxonomyConceptEmptyLocalizedStringsConfig("Furniture"),
			seedDrift: func(server *cmt.Server) error {
				response, err := server.Handler().PatchTaxonomyConcept(t.Context(), cm.TaxonomyPatch{{Op: cm.TaxonomyPatchItemOpAdd, Path: "/note", Value: jx.Raw(`{"en-US":"remote"}`)}}, cm.PatchTaxonomyConceptParams{OrganizationID: "organization-id", TaxonomyConceptID: "furniture", XContentfulVersion: 1})
				if err != nil {
					return fmt.Errorf("patch taxonomy concept: %w", err)
				}

				if _, ok := response.(*cm.TaxonomyConcept); !ok {
					return fmt.Errorf("%w: %T", errUnexpectedTaxonomyResponse, response)
				}

				return nil
			},
			fields: []string{"note", "changeNote", "definition", "editorialNote", "example", "historyNote", "scopeNote"},
		},
		"scheme": {
			resourceName: "contentful_taxonomy_concept_scheme.test", path: "/taxonomy/concept-schemes/products", config: taxonomyConceptSchemeEmptyDefinitionConfig("Products"),
			seedDrift: func(server *cmt.Server) error {
				response, err := server.Handler().PatchTaxonomyConceptScheme(t.Context(), cm.TaxonomyPatch{{Op: cm.TaxonomyPatchItemOpAdd, Path: "/definition", Value: jx.Raw(`{"en-US":"remote"}`)}}, cm.PatchTaxonomyConceptSchemeParams{OrganizationID: "organization-id", TaxonomyConceptSchemeID: "products", XContentfulVersion: 1})
				if err != nil {
					return fmt.Errorf("patch taxonomy concept scheme: %w", err)
				}

				if _, ok := response.(*cm.TaxonomyConceptScheme); !ok {
					return fmt.Errorf("%w: %T", errUnexpectedTaxonomyResponse, response)
				}

				return nil
			},
			fields: []string{"definition"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			parallelWhenMocked(t)

			server, err := cmt.NewContentfulManagementServer()
			require.NoError(t, err)

			recorder := &taxonomyRequestBodyRecorder{next: server}

			ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{Steps: []resource.TestStep{
				{Config: test.config, ConfigPlanChecks: resource.ConfigPlanChecks{PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}},
				{Config: test.config, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(test.resourceName, plancheck.ResourceActionNoop)}}},
				{PreConfig: func() { require.NoError(t, test.seedDrift(server)) }, Config: test.config, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(test.resourceName, plancheck.ResourceActionUpdate)}, PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}},
			}})

			put, ok := recorder.request(http.MethodPut, test.path)
			require.True(t, ok)

			for _, field := range test.fields {
				require.Contains(t, put, field)
				assert.JSONEq(t, `{}`, string(put[field]))
			}

			patches := recorder.matchingRequests(http.MethodPatch, test.path)
			require.NotEmpty(t, patches)
			assert.Equal(t, "2", patches[len(patches)-1].version)

			if name == "concept" {
				assertTaxonomyEmptyPatch(t, patches[len(patches)-1].body, []string{"/note"})
			} else {
				assertTaxonomyEmptyPatch(t, patches[len(patches)-1].body, []string{"/definition"})
			}
		})
	}
}

//nolint:paralleltest
func TestAccTaxonomyImportRemoteNullThenConfiguresEmptyLocalizedMap(t *testing.T) {
	parallelWhenMocked(t)

	tests := map[string]struct {
		resourceName string
		importID     string
		path         string
		config       string
		seed         func(*cmt.Server) error
	}{
		"concept": {
			resourceName: "contentful_taxonomy_concept.test", importID: "organization-id/furniture", path: "/taxonomy/concepts/furniture", config: taxonomyConceptEmptyLocalizedStringsConfig("Furniture"),
			seed: func(server *cmt.Server) error {
				_, err := server.Handler().PutTaxonomyConcept(t.Context(), &cm.TaxonomyConceptRequest{PrefLabel: cm.LocalizedString{"en-US": "Furniture"}}, cm.PutTaxonomyConceptParams{OrganizationID: "organization-id", TaxonomyConceptID: "furniture"})
				if err != nil {
					return fmt.Errorf("put taxonomy concept: %w", err)
				}

				return nil
			},
		},
		"scheme": {
			resourceName: "contentful_taxonomy_concept_scheme.test", importID: "organization-id/products", path: "/taxonomy/concept-schemes/products", config: taxonomyConceptSchemeEmptyDefinitionConfig("Products"),
			seed: func(server *cmt.Server) error {
				_, err := server.Handler().PutTaxonomyConceptScheme(t.Context(), &cm.TaxonomyConceptSchemeRequest{PrefLabel: cm.LocalizedString{"en-US": "Products"}}, cm.PutTaxonomyConceptSchemeParams{OrganizationID: "organization-id", TaxonomyConceptSchemeID: "products"})
				if err != nil {
					return fmt.Errorf("put taxonomy concept scheme: %w", err)
				}

				return nil
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			parallelWhenMocked(t)

			server, err := cmt.NewContentfulManagementServer()
			require.NoError(t, err)
			require.NoError(t, test.seed(server))
			recorder := &taxonomyRequestBodyRecorder{next: server}

			ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{Steps: []resource.TestStep{
				{Config: test.config, ResourceName: test.resourceName, ImportState: true, ImportStateId: test.importID, ImportStatePersist: true},
				{Config: test.config, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(test.resourceName, plancheck.ResourceActionUpdate)}, PostApplyPostRefresh: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}},
				{Config: test.config, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(test.resourceName, plancheck.ResourceActionNoop)}}},
			}})

			patches := recorder.matchingRequests(http.MethodPatch, test.path)
			require.Len(t, patches, 1)
			assert.Equal(t, "1", patches[0].version)

			if name == "concept" {
				assertTaxonomyEmptyPatch(t, patches[0].body, []string{"/changeNote", "/definition", "/editorialNote", "/example", "/historyNote", "/note", "/scopeNote"})
			} else {
				assertTaxonomyEmptyPatch(t, patches[0].body, []string{"/definition"})
			}
		})
	}
}

func taxonomyConceptEmptyLocalizedStringsConfig(label string) string {
	return fmt.Sprintf(`
resource "contentful_taxonomy_concept" "test" {
  organization_id = "organization-id"
  concept_id      = "furniture"
  pref_label      = { "en-US" = %q }
  note = {}
  change_note = {}
  definition = {}
  editorial_note = {}
  example = {}
  history_note = {}
  scope_note = {}
}
`, label)
}

func taxonomyConceptSchemeEmptyDefinitionConfig(label string) string {
	return fmt.Sprintf(`
resource "contentful_taxonomy_concept_scheme" "test" {
  organization_id   = "organization-id"
  concept_scheme_id = "products"
  pref_label        = { "en-US" = %q }
  definition        = {}
}
`, label)
}
