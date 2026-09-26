package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

func TestAccAppDefinitionDataSourceArrayItems(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)

	server.SetAppDefinition("organization-id", "app-definition-id", cm.AppDefinitionData{
		Name:   "Nested field app",
		Src:    cm.NewOptString("https://example.invalid/app.js"),
		Bundle: cm.NewOptAppBundleLink(cm.NewAppBundleLink("bundle-id")),
		Locations: []cm.AppDefinitionDataLocationsItem{
			{
				Location: "entry-field",
				FieldTypes: []cm.AppDefinitionDataLocationsItemFieldTypesItem{
					{Type: "Array", Items: cm.NewOptAppDefinitionDataLocationsItemFieldTypesItemItems(cm.AppDefinitionDataLocationsItemFieldTypesItemItems{Type: "Symbol"})},
					{Type: "Array", Items: cm.NewOptAppDefinitionDataLocationsItemFieldTypesItemItems(cm.AppDefinitionDataLocationsItemFieldTypesItemItems{Type: "Link", LinkType: cm.NewOptString("Asset")})},
					{Type: "Link", LinkType: cm.NewOptString("Entry")},
				},
			},
			{
				Location: "page",
				NavigationItem: cm.NewOptAppDefinitionDataLocationsItemNavigationItem(cm.AppDefinitionDataLocationsItemNavigationItem{
					Name: "App page", Path: "/app-page",
				}),
			},
			{Location: "dialog", FieldTypes: []cm.AppDefinitionDataLocationsItemFieldTypesItem{}},
		},
		Parameters: cm.NewOptAppDefinitionParameters(cm.AppDefinitionParameters{
			Installation: []cm.AppDefinitionParameter{{ID: "client-id", Type: "Symbol", Name: "Client ID", Required: cm.NewOptBool(true)}},
			Instance:     []cm.AppDefinitionParameter{},
		}),
	})

	const config = `
data "contentful_app_definition" "test" {
  organization_id   = "organization-id"
  app_definition_id = "app-definition-id"
}

output "symbol_item_type" {
  value = data.contentful_app_definition.test.locations[0].field_types[0].items[0].type
}

output "asset_item_link_type" {
  value = data.contentful_app_definition.test.locations[0].field_types[1].items[0].link_type
}
`

	const address = "data.contentful_app_definition.test"

	locations := tfjsonpath.New("locations")
	fieldTypes := locations.AtSliceIndex(0).AtMapKey("field_types")
	parameters := tfjsonpath.New("parameters")

	testAccMockedResource(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: config,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(address, tfjsonpath.New("id"), knownvalue.StringExact("organization-id/app-definition-id")),
					statecheck.ExpectKnownValue(address, tfjsonpath.New("organization_id"), knownvalue.StringExact("organization-id")),
					statecheck.ExpectKnownValue(address, tfjsonpath.New("app_definition_id"), knownvalue.StringExact("app-definition-id")),
					statecheck.ExpectKnownValue(address, tfjsonpath.New("name"), knownvalue.StringExact("Nested field app")),
					statecheck.ExpectKnownValue(address, tfjsonpath.New("src"), knownvalue.StringExact("https://example.invalid/app.js")),
					statecheck.ExpectKnownValue(address, tfjsonpath.New("bundle_id"), knownvalue.StringExact("bundle-id")),
					statecheck.ExpectKnownValue(address, locations.AtSliceIndex(0).AtMapKey("location"), knownvalue.StringExact("entry-field")),
					statecheck.ExpectKnownValue(address, fieldTypes.AtSliceIndex(0).AtMapKey("type"), knownvalue.StringExact("Array")),
					statecheck.ExpectKnownValue(address, fieldTypes.AtSliceIndex(0).AtMapKey("items"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{"type": knownvalue.StringExact("Symbol"), "link_type": knownvalue.Null()}),
					})),
					statecheck.ExpectKnownValue(address, fieldTypes.AtSliceIndex(1).AtMapKey("items"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{"type": knownvalue.StringExact("Link"), "link_type": knownvalue.StringExact("Asset")}),
					})),
					statecheck.ExpectKnownValue(address, fieldTypes.AtSliceIndex(2).AtMapKey("type"), knownvalue.StringExact("Link")),
					statecheck.ExpectKnownValue(address, fieldTypes.AtSliceIndex(2).AtMapKey("link_type"), knownvalue.StringExact("Entry")),
					statecheck.ExpectKnownValue(address, fieldTypes.AtSliceIndex(2).AtMapKey("items"), knownvalue.Null()),
					statecheck.ExpectKnownValue(address, locations.AtSliceIndex(1).AtMapKey("location"), knownvalue.StringExact("page")),
					statecheck.ExpectKnownValue(address, locations.AtSliceIndex(1).AtMapKey("field_types"), knownvalue.Null()),
					statecheck.ExpectKnownValue(address, locations.AtSliceIndex(1).AtMapKey("navigation_item").AtMapKey("name"), knownvalue.StringExact("App page")),
					statecheck.ExpectKnownValue(address, locations.AtSliceIndex(1).AtMapKey("navigation_item").AtMapKey("path"), knownvalue.StringExact("/app-page")),
					statecheck.ExpectKnownValue(address, locations.AtSliceIndex(2).AtMapKey("location"), knownvalue.StringExact("dialog")),
					statecheck.ExpectKnownValue(address, locations.AtSliceIndex(2).AtMapKey("field_types"), knownvalue.ListExact([]knownvalue.Check{})),
					statecheck.ExpectKnownValue(address, parameters.AtMapKey("installation").AtSliceIndex(0).AtMapKey("id"), knownvalue.StringExact("client-id")),
					statecheck.ExpectKnownValue(address, parameters.AtMapKey("installation").AtSliceIndex(0).AtMapKey("type"), knownvalue.StringExact("Symbol")),
					statecheck.ExpectKnownValue(address, parameters.AtMapKey("installation").AtSliceIndex(0).AtMapKey("name"), knownvalue.StringExact("Client ID")),
					statecheck.ExpectKnownValue(address, parameters.AtMapKey("installation").AtSliceIndex(0).AtMapKey("required"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(address, parameters.AtMapKey("instance"), knownvalue.ListExact([]knownvalue.Check{})),
				},
			},
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("symbol_item_type", "Symbol"),
					resource.TestCheckOutput("asset_item_link_type", "Asset"),
				),
			},
		},
	})
}

func TestAccMarketplaceAppDefinitionDataSourceArrayItems(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)

	server.SetMarketplaceAppDefinition("marketplace-organization", "marketplace-app", cm.AppDefinitionData{
		Name: "Marketplace nested fields",
		Locations: []cm.AppDefinitionDataLocationsItem{{
			Location: "entry-field",
			FieldTypes: []cm.AppDefinitionDataLocationsItemFieldTypesItem{
				{Type: "Array", Items: cm.NewOptAppDefinitionDataLocationsItemFieldTypesItemItems(cm.AppDefinitionDataLocationsItemFieldTypesItemItems{Type: "Link", LinkType: cm.NewOptString("Asset")})},
				{Type: "Symbol"},
			},
		}},
	})

	const config = `
data "contentful_marketplace_app_definition" "test" {
  app_definition_id = "marketplace-app"
}

output "marketplace_item_link_type" {
  value = data.contentful_marketplace_app_definition.test.locations[0].field_types[0].items[0].link_type
}
`

	const address = "data.contentful_marketplace_app_definition.test"

	fieldTypes := tfjsonpath.New("locations").AtSliceIndex(0).AtMapKey("field_types")

	testAccMockedResource(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: config,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(address, tfjsonpath.New("id"), knownvalue.StringExact("marketplace-organization/marketplace-app")),
					statecheck.ExpectKnownValue(address, tfjsonpath.New("organization_id"), knownvalue.StringExact("marketplace-organization")),
					statecheck.ExpectKnownValue(address, tfjsonpath.New("name"), knownvalue.StringExact("Marketplace nested fields")),
					statecheck.ExpectKnownValue(address, fieldTypes.AtSliceIndex(0).AtMapKey("items"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{"type": knownvalue.StringExact("Link"), "link_type": knownvalue.StringExact("Asset")}),
					})),
					statecheck.ExpectKnownValue(address, fieldTypes.AtSliceIndex(1).AtMapKey("items"), knownvalue.Null()),
				},
			},
			{
				Config: config,
				Check:  resource.TestCheckOutput("marketplace_item_link_type", "Asset"),
			},
		},
	})
}
