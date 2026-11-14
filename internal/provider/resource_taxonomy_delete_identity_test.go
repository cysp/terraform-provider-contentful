package provider_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
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
	httpServer := httptest.NewServer(recorder)
	t.Cleanup(httpServer.Close)
	runtime := newTerraformTestRuntime(t)
	runtime.providerURL = httpServer.URL

	runtime.writeConfig(t, fmt.Sprintf(`
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
`, kind))
	output, err := runtime.run(t.Context(), "apply", "-auto-approve", "-input=false", "-no-color")
	require.NoError(t, err, output)

	address := "contentful_taxonomy_" + kind + ".target"
	output, err = runtime.run(t.Context(), "taint", "-no-color", address)
	require.NoError(t, err, output)

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

	mutator.replaceIdentityOnce(http.MethodGet, endpoint, organizationID, resourceID)

	output, err = runtime.run(t.Context(), "apply", "-refresh=false", "-auto-approve", "-input=false", "-no-color")
	require.Error(t, err, output)
	assert.Contains(t, output, "Unexpected Contentful taxonomy read identity")
	assert.Empty(t, recorder.matchingRequests(http.MethodDelete, endpoint), "an unconfirmed GET version must not authorize deletion")

	output, err = runtime.run(t.Context(), "state", "pull")
	require.NoError(t, err, output)

	var state struct {
		Resources []struct {
			Name      string `json:"name"`
			Instances []struct {
				Status     string                     `json:"status"`
				Attributes map[string]json.RawMessage `json:"attributes"`
			} `json:"instances"`
		} `json:"resources"`
	}
	require.NoError(t, json.Unmarshal([]byte(output), &state))
	require.Len(t, state.Resources, 2)

	for _, resource := range state.Resources {
		require.Len(t, resource.Instances, 1)

		if resource.Name == "target" {
			assert.Equal(t, "tainted", resource.Instances[0].Status)
			assert.JSONEq(t, `"organization-id/target"`, string(resource.Instances[0].Attributes["id"]))
		}
	}

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

	output, err = runtime.run(t.Context(), "apply", "-refresh=false", "-auto-approve", "-input=false", "-no-color")
	require.NoError(t, err, output)

	deletes := recorder.matchingRequests(http.MethodDelete, endpoint)
	require.Len(t, deletes, 1, "a subsequent confirmed GET can authorize replacement")
	assert.Equal(t, "1", deletes[0].version)
	output, err = runtime.run(t.Context(), "destroy", "-auto-approve", "-input=false", "-no-color")
	require.NoError(t, err, output)
}
