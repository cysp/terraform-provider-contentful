package provider_test

import (
	"maps"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

func TestAccResourceProviderResourceLifecycle(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)

	configVariables := config.Variables{
		"organization_id":   config.StringVariable("organization-id"),
		"app_definition_id": config.StringVariable("app-definition-id"),
	}

	server.SetAppDefinition("organization-id", "app-definition-id", cm.AppDefinitionData{
		Name: "Test App",
	})

	stepVariables1 := maps.Clone(configVariables)
	stepVariables1["function_id"] = config.StringVariable("resourceProvider")

	stepVariables2 := maps.Clone(configVariables)
	stepVariables2["function_id"] = config.StringVariable("resourceProviderTwo")

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: stepVariables1,
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: stepVariables2,
			},
		},
	})
}

func TestAccResourceProviderResourceImport(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)

	configVariables := config.Variables{
		"organization_id":   config.StringVariable("organization-id"),
		"app_definition_id": config.StringVariable("app-definition-id"),
	}

	server.SetAppDefinition("organization-id", "app-definition-id", cm.AppDefinitionData{
		Name: "Test App",
	})

	server.SetResourceProvider("organization-id", "app-definition-id", cm.ResourceProviderRequest{
		Sys:      cm.NewResourceProviderRequestSys("resource-provider"),
		Type:     cm.ResourceProviderRequestTypeFunction,
		Function: cm.NewFunctionLink("function-id"),
	})

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("contentful_resource_provider.test", "id", "organization-id/app-definition-id"),
				),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ImportState:     true,
				ResourceName:    "contentful_resource_provider.test",
			},
		},
	})
}
