package provider_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
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
	case taxonomyLifecycleUpdate:
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
			method: http.MethodPatch, path: "/taxonomy/concepts/furniture", expectedError: "Failed to update taxonomy concept",
			bumpVersion: bumpTaxonomyConceptVersion,
		},
		"concept delete": {
			initialConfig: taxonomyConceptConfig("Furniture"), changedConfig: "# intentionally empty\n",
			method: http.MethodDelete, path: "/taxonomy/concepts/furniture", expectedError: "Failed to delete taxonomy concept",
			bumpVersion: bumpTaxonomyConceptVersion,
		},
		"scheme update": {
			initialConfig: taxonomyConceptSchemeConfig("Products"), changedConfig: taxonomyConceptSchemeConfig("Home products"),
			method: http.MethodPatch, path: "/taxonomy/concept-schemes/products", expectedError: "Failed to update taxonomy concept scheme",
			bumpVersion: bumpTaxonomyConceptSchemeVersion,
		},
		"scheme delete": {
			initialConfig: taxonomyConceptSchemeConfig("Products"), changedConfig: "# intentionally empty\n",
			method: http.MethodDelete, path: "/taxonomy/concept-schemes/products", expectedError: "Failed to delete taxonomy concept scheme",
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

func deleteTaxonomyConceptRemote(server *cmt.Server, organizationID, conceptID string) error {
	ctx := context.Background()

	response, err := server.Handler().GetTaxonomyConcept(ctx, cm.GetTaxonomyConceptParams{
		OrganizationID: organizationID, TaxonomyConceptID: conceptID,
	})
	if err != nil {
		return fmt.Errorf("get taxonomy concept: %w", err)
	}

	concept, ok := response.(*cm.TaxonomyConcept)
	if !ok {
		return fmt.Errorf("%w: taxonomy concept: %T", errUnexpectedTaxonomyResponse, response)
	}

	deleted, err := server.Handler().DeleteTaxonomyConcept(ctx, cm.DeleteTaxonomyConceptParams{
		OrganizationID: organizationID, TaxonomyConceptID: conceptID, XContentfulVersion: concept.Sys.Version,
	})
	if err != nil {
		return fmt.Errorf("delete taxonomy concept: %w", err)
	}

	if _, ok := deleted.(*cm.NoContent); !ok {
		return fmt.Errorf("%w: taxonomy concept deletion: %T", errUnexpectedTaxonomyResponse, deleted)
	}

	return nil
}

func deleteTaxonomyConceptSchemeRemote(server *cmt.Server, organizationID, schemeID string) error {
	ctx := context.Background()

	response, err := server.Handler().GetTaxonomyConceptScheme(ctx, cm.GetTaxonomyConceptSchemeParams{
		OrganizationID: organizationID, TaxonomyConceptSchemeID: schemeID,
	})
	if err != nil {
		return fmt.Errorf("get taxonomy concept scheme: %w", err)
	}

	scheme, ok := response.(*cm.TaxonomyConceptScheme)
	if !ok {
		return fmt.Errorf("%w: taxonomy concept scheme: %T", errUnexpectedTaxonomyResponse, response)
	}

	deleted, err := server.Handler().DeleteTaxonomyConceptScheme(ctx, cm.DeleteTaxonomyConceptSchemeParams{
		OrganizationID: organizationID, TaxonomyConceptSchemeID: schemeID, XContentfulVersion: scheme.Sys.Version,
	})
	if err != nil {
		return fmt.Errorf("delete taxonomy concept scheme: %w", err)
	}

	if _, ok := deleted.(*cm.NoContent); !ok {
		return fmt.Errorf("%w: taxonomy concept scheme deletion: %T", errUnexpectedTaxonomyResponse, deleted)
	}

	return nil
}

type taxonomyRequestHook struct {
	next http.Handler

	mu     sync.Mutex
	method string
	path   string
	hook   func() error
	called bool
}

type taxonomyResponseFailure struct {
	next http.Handler

	mu               sync.Mutex
	method           string
	path             string
	remainingMatches int
	called           bool
}

func (f *taxonomyResponseFailure) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	if f.consume(request.Method, request.URL.Path) {
		http.Error(responseWriter, "injected taxonomy failure", http.StatusBadRequest)

		return
	}

	f.next.ServeHTTP(responseWriter, request)
}

func (f *taxonomyResponseFailure) failOnce(method, path string, occurrence int) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.method, f.path, f.remainingMatches, f.called = method, path, occurrence, false
}

func (f *taxonomyResponseFailure) consume(method, path string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	if method != f.method || !strings.HasSuffix(path, f.path) || f.called {
		return false
	}

	f.remainingMatches--
	if f.remainingMatches > 0 {
		return false
	}

	f.called = true

	return true
}

func (f *taxonomyResponseFailure) wasCalled() bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.called
}

func (h *taxonomyRequestHook) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	hook := h.consume(request.Method, request.URL.Path)
	if hook != nil {
		err := hook()
		if err != nil {
			http.Error(responseWriter, err.Error(), http.StatusInternalServerError)

			return
		}
	}

	h.next.ServeHTTP(responseWriter, request)
}

func (h *taxonomyRequestHook) runOnce(method, path string, hook func() error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.method, h.path, h.hook, h.called = method, path, hook, false
}

func (h *taxonomyRequestHook) consume(method, path string) func() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if method != h.method || !strings.HasSuffix(path, h.path) {
		return nil
	}

	hook := h.hook
	h.method, h.path, h.hook, h.called = "", "", nil, true

	return hook
}

func (h *taxonomyRequestHook) wasCalled() bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	return h.called
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

type taxonomyResponseMutator struct {
	next http.Handler

	mu     sync.Mutex
	method string
	path   string
	locale string
	add    bool
}

func (m *taxonomyResponseMutator) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	recorder := httptest.NewRecorder()
	m.next.ServeHTTP(recorder, request)

	body := recorder.Body.Bytes()

	if recorder.Code >= http.StatusOK && recorder.Code < http.StatusMultipleChoices {
		locale, add, mutate := m.consumeMutation(request.Method, request.URL.Path)
		if mutate {
			if add {
				body = addEmptyLabelLocale(body, locale)
			} else {
				body = removePreferredLabelLocale(body, locale)
			}
		}
	}

	for key, values := range recorder.Header() {
		responseWriter.Header()[key] = append([]string(nil), values...)
	}

	responseWriter.Header().Del("Content-Length")
	responseWriter.WriteHeader(recorder.Code)

	_, _ = io.Copy(responseWriter, bytes.NewReader(body))
}

func (m *taxonomyResponseMutator) dropPreferredLabelOnce(method, path, locale string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.method, m.path, m.locale, m.add = method, path, locale, false
}

func (m *taxonomyResponseMutator) addEmptyLabelLocaleOnce(method, path, locale string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.method, m.path, m.locale, m.add = method, path, locale, true
}

func (m *taxonomyResponseMutator) consumeMutation(method, path string) (string, bool, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if method != m.method || !strings.Contains(path, m.path) {
		return "", false, false
	}

	locale := m.locale
	add := m.add
	m.method, m.path, m.locale, m.add = "", "", "", false

	return locale, add, true
}

func addEmptyLabelLocale(body []byte, locale string) []byte {
	document := map[string]json.RawMessage{}

	err := json.Unmarshal(body, &document)
	if err != nil {
		return body
	}

	for _, field := range []string{"altLabels", "hiddenLabels"} {
		labels := map[string][]string{}

		err = json.Unmarshal(document[field], &labels)
		if err != nil {
			return body
		}

		labels[locale] = []string{}

		document[field], err = json.Marshal(labels)
		if err != nil {
			return body
		}
	}

	result, err := json.Marshal(document)
	if err != nil {
		return body
	}

	return result
}

func removePreferredLabelLocale(body []byte, locale string) []byte {
	document := map[string]json.RawMessage{}

	err := json.Unmarshal(body, &document)
	if err != nil {
		return body
	}

	labels := map[string]string{}

	err = json.Unmarshal(document["prefLabel"], &labels)
	if err != nil {
		return body
	}

	delete(labels, locale)

	encodedLabels, err := json.Marshal(labels)
	if err != nil {
		return body
	}

	document["prefLabel"] = encodedLabels

	result, err := json.Marshal(document)
	if err != nil {
		return body
	}

	return result
}

func deleteTaxonomyConceptOutOfBand(t *testing.T, server *cmt.Server, organizationID, conceptID string) {
	t.Helper()

	err := deleteTaxonomyConceptRemote(server, organizationID, conceptID)
	if err != nil {
		t.Fatal(err)
	}
}

func deleteTaxonomyConceptSchemeOutOfBand(t *testing.T, server *cmt.Server, organizationID, schemeID string) {
	t.Helper()

	err := deleteTaxonomyConceptSchemeRemote(server, organizationID, schemeID)
	if err != nil {
		t.Fatal(err)
	}
}
