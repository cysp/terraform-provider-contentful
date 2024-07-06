package provider_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

//nolint:paralleltest
func TestAccAppDefinitionDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				data "contentful_app_definition" "test" {
					organization_id = "2zuSjSO4A0e6GKBrhJRe2m"
					app_definition_id = "1WkQ2J9LERPtbMTdUfSHka"
				}
				`,
			},
			{
				RefreshState: true,
			},
		},
	})
}

//nolint:paralleltest
func TestAccAppDefinitionDataSourceNotFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				data "contentful_app_definition" "test" {
					organization_id = "2zuSjSO4A0e6GKBrhJRe2m"
					app_definition_id = "12345"
				}
				`,
				ExpectError: regexp.MustCompile(`Failed to read app definition`),
			},
		},
	})
}
