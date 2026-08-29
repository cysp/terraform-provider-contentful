package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"
)

func TestAccEntryResourceImportedPendingDraftIsNotPublished(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")
	server.SetEntry("space", "environment", "article", "entry", cm.EntryRequest{
		Fields: cm.NewOptEntryFields(cm.EntryFields{"managed": jx.Raw(`{"en-US":"one"}`)}),
		Metadata: cm.NewOptEntryMetadata(cm.EntryMetadata{
			Concepts: []cm.TaxonomyConceptLink{},
			Tags:     []cm.TagLink{},
		}),
	})

	publishResponse, publishErr := server.Handler().PublishEntry(t.Context(), cm.PublishEntryParams{
		SpaceID: "space", EnvironmentID: "environment", EntryID: "entry", XContentfulVersion: 1,
	})
	require.NoError(t, publishErr)
	require.IsType(t, &cm.EntryStatusCode{}, publishResponse)

	entry := getTestEntry(t, server)
	putResponse, putErr := server.Handler().PutEntry(t.Context(), &cm.EntryRequest{
		Fields: entry.Fields, Metadata: entry.Metadata,
	}, cm.PutEntryParams{
		SpaceID: "space", EnvironmentID: "environment", EntryID: "entry", XContentfulVersion: cm.NewOptInt(entry.Sys.Version),
	})
	require.NoError(t, putErr)
	require.IsType(t, &cm.EntryStatusCode{}, putResponse)

	errorSink := new(entryFixtureErrorSink)
	recorder := newEntryMutationRecorder(server, errorSink)
	additionalCLIOptions := &resource.AdditionalCLIOptions{}
	config := `
resource "contentful_entry" "test" {
  space_id        = "space"
  environment_id  = "environment"
  entry_id        = "entry"
  content_type_id = "article"
  fields = { managed = jsonencode({ "en-US" = "one" }) }
}
`

	ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{AdditionalCLIOptions: additionalCLIOptions, Steps: []resource.TestStep{
		{
			Config:             config,
			ResourceName:       "contentful_entry.test",
			ImportState:        true,
			ImportStateId:      "space/environment/entry",
			ImportStatePersist: true,
		},
		{
			PreConfig: recorder.reset,
			Config:    config,
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionNoop),
			}},
			Check: func(*terraform.State) error {
				requireNoEntryMutations(t, recorder)
				entry := getTestEntry(t, server)
				require.True(t, entry.Sys.PublishedVersion.IsSet())
				require.Less(t, entry.Sys.PublishedVersion.Or(0), entry.Sys.Version)

				return nil
			},
		},
		{
			PreConfig: func() {
				additionalCLIOptions.Plan.NoRefresh = true

				recorder.reset()
			},
			Config: config,
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionNoop),
			}},
			Check: func(*terraform.State) error {
				requireNoEntryMutations(t, recorder, "import and Read must not serialize publication authority")

				return nil
			},
		},
	}})
}
