package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

func TestAccMarketplaceAppDefinitionDataSourceRead(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)

	configVariables := config.Variables{
		"app_definition_id": config.StringVariable("5KySdUzG7OWuCE2V3fgtIa"),
	}

	server.SetMarketplaceAppDefinition("5EJGHo8tYJcjnEhYWDxivp", "5KySdUzG7OWuCE2V3fgtIa", cm.AppDefinitionData{
		Name: "Bynder",
		Locations: []cm.AppDefinitionDataLocationsItem{
			{
				Location: "app-config",
			},
			{
				Location: "entry-field",
				FieldTypes: []cm.AppDefinitionDataLocationsItemFieldTypesItem{
					{
						Type: "Object",
					},
				},
			},
			{
				Location: "dialog",
			},
		},
		Parameters: cm.NewOptAppDefinitionParameters(cm.AppDefinitionParameters{}),
	})

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.contentful_marketplace_app_definition.test", tfjsonpath.New("id"), knownvalue.StringExact("5EJGHo8tYJcjnEhYWDxivp/5KySdUzG7OWuCE2V3fgtIa")),
					statecheck.ExpectKnownValue("data.contentful_marketplace_app_definition.test", tfjsonpath.New("name"), knownvalue.StringExact("Bynder")),
					statecheck.ExpectKnownValue("data.contentful_marketplace_app_definition.test", tfjsonpath.New("locations"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{"location": knownvalue.StringExact("app-config")}),
						knownvalue.ObjectPartial(map[string]knownvalue.Check{"location": knownvalue.StringExact("entry-field"), "field_types": knownvalue.SetExact([]knownvalue.Check{knownvalue.ObjectPartial(map[string]knownvalue.Check{"type": knownvalue.StringExact("Object")})})}),
						knownvalue.ObjectPartial(map[string]knownvalue.Check{"location": knownvalue.StringExact("dialog")}),
					})),
				},
			},
		},
	})
}
