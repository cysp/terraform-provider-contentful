package provider_test

import (
	"maps"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccEnvironmentResource(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	server.RegisterSpaceEnvironment("space-id", "master")

	environmentID := "acctest-" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	configVariables := config.Variables{
		"space_id":            config.StringVariable("space-id"),
		"test_environment_id": config.StringVariable(environmentID),
		"environment_name":    config.StringVariable("Test Environment"),
	}

	configVariables1 := maps.Clone(configVariables)

	configVariables2 := maps.Clone(configVariables)
	configVariables2["environment_name"] = config.StringVariable("Updated Test Environment")

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
				ResourceName:    "contentful_environment.test",
			},
		},
	})
}

func TestAccEnvironmentImport(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	server.RegisterSpaceEnvironment("space-id", "master")

	server.SetEnvironment("space-id", "staging", "ready", cm.EnvironmentData{
		Name: "Staging Environment",
	})

	configVariables := config.Variables{
		"space_id":            config.StringVariable("space-id"),
		"test_environment_id": config.StringVariable("staging"),
		"environment_name":    config.StringVariable("Staging Environment"),
	}

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("contentful_environment.test", "id", "space-id/staging"),
					resource.TestCheckResourceAttr("contentful_environment.test", "space_id", "space-id"),
					resource.TestCheckResourceAttr("contentful_environment.test", "environment_id", "staging"),
					resource.TestCheckResourceAttr("contentful_environment.test", "name", "Staging Environment"),
					resource.TestCheckResourceAttr("contentful_environment.test", "status", "ready"),
				),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ImportState:     true,
				ResourceName:    "contentful_environment.test",
			},
		},
	})
}
