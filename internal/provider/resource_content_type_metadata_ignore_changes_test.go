package provider_test

import (
	"encoding/json"
	"fmt"
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

func TestAccContentTypeResourceMetadataIgnoreChanges(t *testing.T) {
	t.Parallel()

	for _, ignored := range []string{"metadata", "metadata.annotations"} {
		t.Run(ignored, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
			require.NoError(t, err)
			server.RegisterSpaceEnvironment("space", "environment")
			handler := &contentTypeActivationTestHandler{delegate: server}

			testAccMockedResource(t, handler, resource.TestCase{Steps: []resource.TestStep{
				{Config: contentTypeMetadataIgnoreChangesConfig("Original", "{ ContentType = [] }", `taxonomy = [{ taxonomy_concept = { id = "concept" } }]`, ignored)},
				{
					PreConfig: handler.resetRequestHistory,
					Config:    contentTypeMetadataIgnoreChangesConfig("Updated", "{ ContentTypeField = {} }", "", ignored),
					ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue("contentful_content_type.test", tfjsonpath.New("metadata").AtMapKey("annotations"), knownvalue.StringExact(`{"ContentType":[]}`)),
					}},
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue("contentful_content_type.test", tfjsonpath.New("metadata").AtMapKey("annotations"), knownvalue.StringExact(`{"ContentType":[]}`)),
					},
					Check: contentTypeActivationRequestAndVersionsCheck(handler, 1, 1, []int64{3}),
				},
			}})
		})
	}
}

func contentTypeMetadataIgnoreChangesConfig(name, annotations, taxonomy, ignored string) string {
	return fmt.Sprintf(`
resource "contentful_content_type" "test" {
  space_id        = "space"
  environment_id  = "environment"
  content_type_id = "metadata-ignore-changes"
  name            = %[1]q
  description     = "Metadata ignore_changes test"
  display_field   = "title"
  fields = [{
    id = "title"
    name = "Title"
    type = "Symbol"
    localized = false
    required = false
  }]
  metadata = {
    annotations = jsonencode(%[2]s)
    %[3]s
  }
  lifecycle {
    ignore_changes = [%[4]s]
  }
}
`, name, annotations, taxonomy, ignored)
}

func TestAccContentTypeResourceOmittedTaxonomyDoesNotAuthorizeContradictoryDraft(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")
	handler := &contentTypeActivationTestHandler{delegate: server}

	testAccMockedResource(t, handler, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
		Steps: []resource.TestStep{
			{Config: contentTypeMetadataIgnoreChangesConfig("Original", "{ ContentType = [] }", `taxonomy = [{ taxonomy_concept = { id = "concept" } }]`, "metadata.annotations")},
			{
				PreConfig: func() {
					handler.resetRequestHistory()
					handler.mutatePutResponse = contentTypeResponseMutator(t, func(response map[string]any) {
						mustContentTypeResponseObject(response["metadata"])["taxonomy"] = []any{map[string]any{
							"sys":      map[string]any{"type": "Link", "id": "contradictory", "linkType": "TaxonomyConcept"},
							"required": false,
						}}
					})
				},
				Config:      contentTypeMetadataIgnoreChangesConfig("Updated", "{ ContentType = [] }", "", "metadata.annotations"),
				ExpectError: regexp.MustCompile(`Unexpected Contentful content type response`),
			},
			{
				PreConfig: func() {
					require.Equal(t, 1, handler.eventCount(contentTypeOperationPut))
					require.Equal(t, 0, handler.eventCount(contentTypeOperationActivate))

					var request map[string]any
					require.NoError(t, json.Unmarshal(handler.lastPutBody(), &request))
					require.Equal(t, []any{map[string]any{
						"sys":      map[string]any{"type": "Link", "id": "concept", "linkType": "TaxonomyConcept"},
						"required": false,
					}}, mustContentTypeResponseObject(request["metadata"])["taxonomy"])
					handler.resetRequestHistory()
				},
				Config: contentTypeMetadataIgnoreChangesConfig("Updated", "{ ContentType = [] }", "", "metadata.annotations"),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					expectResponseStateInPlan{
						address:   "contentful_content_type.test",
						valuePath: tfjsonpath.New("metadata").AtMapKey("taxonomy").AtSliceIndex(0).AtMapKey("taxonomy_concept").AtMapKey("id"),
						value:     "contradictory",
					},
				}},
				Check: contentTypeActivationRequestCheck(handler, 0, 0),
			},
		},
	})
}
