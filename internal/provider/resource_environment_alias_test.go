package provider_test

import (
	"maps"
	"testing"

	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccEnvironmentAliasResource(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	server.RegisterSpaceEnvironment("space-id", "master")
	server.RegisterSpaceEnvironment("space-id", "staging")

	environmentAliasID := "acctest-" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	configVariables := config.Variables{
		"space_id":                  config.StringVariable("space-id"),
		"test_environment_alias_id": config.StringVariable(environmentAliasID),
		"target_environment_id":     config.StringVariable("staging"),
	}

	configVariables1 := maps.Clone(configVariables)

	configVariables2 := maps.Clone(configVariables)
	configVariables2["target_environment_id"] = config.StringVariable("master")

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables1,
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables2,
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables2,
				ImportState:     true,
				ResourceName:    "contentful_environment_alias.test",
			},
		},
	})
}

func TestAccEnvironmentAliasImport(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	server.RegisterSpaceEnvironment("space-id", "master-1970-01-01")

	configVariables := config.Variables{
		"space_id":                  config.StringVariable("space-id"),
		"test_environment_alias_id": config.StringVariable("master"),
		"target_environment_id":     config.StringVariable("master-1970-01-01"),
	}

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("contentful_environment_alias.test", "id", "space-id/master"),
					resource.TestCheckResourceAttr("contentful_environment_alias.test", "space_id", "space-id"),
					resource.TestCheckResourceAttr("contentful_environment_alias.test", "environment_alias_id", "master"),
					resource.TestCheckResourceAttr("contentful_environment_alias.test", "target_environment_id", "master-1970-01-01"),
				),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ImportState:     true,
				ResourceName:    "contentful_environment_alias.test",
			},
		},
	})
}
