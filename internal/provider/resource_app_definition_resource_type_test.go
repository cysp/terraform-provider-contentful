package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

//nolint:paralleltest
func TestAccAppDefinitionResourceTypeResource(t *testing.T) {
	server, _ := cmt.NewContentfulManagementServer()

	configVariables := config.Variables{
		"organization_id":      config.StringVariable("organization-id"),
		"app_definition_id":    config.StringVariable("app-definition-id"),
		"resource_provider_id": config.StringVariable("resource-provider-id"),
	}

	server.SetAppDefinition("organization-id", "app-definition-id", cm.AppDefinitionFields{
		Name: "Test App",
	})

	server.SetAppDefinitionResourceProvider("organization-id", "app-definition-id", cm.ResourceProviderRequest{
		Sys: cm.ResourceProviderRequestSys{
			ID: "resource-provider-id",
		},
		Type: cm.ResourceProviderRequestTypeFunction,
		Function: cm.FunctionLink{
			Sys: cm.FunctionLinkSys{
				Type:     cm.FunctionLinkSysTypeLink,
				LinkType: cm.FunctionLinkSysLinkTypeFunction,
				ID:       "function-id",
			},
		},
	})

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
			},
		},
	})
}

//nolint:paralleltest
func TestAccAppDefinitionResourceTypeImport(t *testing.T) {
	server, _ := cmt.NewContentfulManagementServer()

	configVariables := config.Variables{
		"organization_id":      config.StringVariable("organization-id"),
		"app_definition_id":    config.StringVariable("app-definition-id"),
		"resource_provider_id": config.StringVariable("resource-provider-id"),
	}

	server.SetAppDefinition("organization-id", "app-definition-id", cm.AppDefinitionFields{
		Name: "Test App",
	})

	server.SetAppDefinitionResourceProvider("organization-id", "app-definition-id", cm.ResourceProviderRequest{
		Sys: cm.ResourceProviderRequestSys{
			ID: "resource-provider-id",
		},
		Type: cm.ResourceProviderRequestTypeFunction,
		Function: cm.FunctionLink{
			Sys: cm.FunctionLinkSys{
				Type:     cm.FunctionLinkSysTypeLink,
				LinkType: cm.FunctionLinkSysLinkTypeFunction,
				ID:       "function-id",
			},
		},
	})

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("contentful_app_definition_resource_type.test", "id", "organization-id/app-definition-id/resource-provider-id:test"),
				),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ImportState:     true,
				ResourceName:    "contentful_app_definition_resource_type.test",
			},
		},
	})
}
