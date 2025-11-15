package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

//nolint:paralleltest
func TestAccMarketplaceAppDefinitionDataSource(t *testing.T) {
	server, _ := cmt.NewContentfulManagementServer()

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

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.contentful_marketplace_app_definition.test", "id", "5EJGHo8tYJcjnEhYWDxivp/5KySdUzG7OWuCE2V3fgtIa"),
				),
			},
		},
	})
}
