package provider_test

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

func TestAccEntryResourceRegistryV0074LocalizedFields(t *testing.T) {
	fixture := newEntryAcceptanceFixture(t)
	server := httptest.NewServer(fixture.recorder)
	t.Cleanup(server.Close)
	t.Setenv("CONTENTFUL_URL", server.URL)
	t.Setenv("CONTENTFUL_MANAGEMENT_ACCESS_TOKEN", "CFPAT-12345")
	t.Setenv(resource.EnvTfAccProviderNamespace, "cysp")

	const address = "contentful_entry.test"

	const config = `
terraform {
  required_providers {
    contentful = { source = "cysp/contentful" }
  }
}

resource "contentful_entry" "test" {
  space_id        = "space"
  environment_id  = "environment"
  entry_id        = "entry"
  content_type_id = "article"
  fields = { title = %s }
}
`

	factories := makeTestAccProtoV6ProviderFactories(testProviderOptionsWithHTTPServer(server)...)
	// Keep refresh disabled across both candidate steps, so a GET cannot
	// supply the private version needed by the first post-upgrade update.
	cliOptions := &resource.AdditionalCLIOptions{}

	resource.Test(t, validateAcceptanceTestCase(t, resource.TestCase{
		AdditionalCLIOptions: cliOptions,
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"contentful": {
						Source:            "registry.terraform.io/cysp/contentful",
						VersionConstraint: "= 0.0.74",
					},
				},
				Config: fmt.Sprintf(config, `jsonencode({ "en-US" = "hello", "fr-FR" = "salut" })`),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(address, tfjsonpath.New("fields").AtMapKey("title"), knownvalue.StringExact(`{"en-US":"hello","fr-FR":"salut"}`)),
					statecheck.ExpectKnownValue(address, tfjsonpath.New("published_version"), knownvalue.Int64Exact(1)),
				},
			},
			{
				PreConfig: func() {
					cliOptions.Plan.NoRefresh = true

					fixture.recorder.reset()
				},
				Config:                   fmt.Sprintf(config, `{ "en-US" = jsonencode("hello"), "fr-FR" = jsonencode("salut") }`),
				ProtoV6ProviderFactories: factories,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(address, plancheck.ResourceActionNoop),
				}},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(address, tfjsonpath.New("fields").AtMapKey("title").AtMapKey("fr-FR"), knownvalue.StringExact(`"salut"`)),
				},
				Check: func(*terraform.State) error {
					requireNoEntryMutations(t, fixture.recorder)

					return nil
				},
			},
			{
				PreConfig:                fixture.recorder.reset,
				Config:                   fmt.Sprintf(config, `{ "en-US" = jsonencode("hi"), "fr-FR" = jsonencode("salut") }`),
				ProtoV6ProviderFactories: factories,
				Check: func(*terraform.State) error {
					draft, publish := requireEntryUpdateThenPublish(t, fixture.recorder.snapshot())
					require.Equal(t, "2", draft.version)
					require.Equal(t, "3", publish.version)
					require.JSONEq(t, `{"en-US":"hi","fr-FR":"salut"}`, string(draft.fields["title"]))

					return nil
				},
			},
		},
	}))
}
