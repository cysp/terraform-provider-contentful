package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

//nolint:paralleltest
func TestAccResourceTypeResource(t *testing.T) {
	server, _ := cmt.NewContentfulManagementServer()

	configVariables := config.Variables{
		"organization_id":   config.StringVariable("organization-id"),
		"app_definition_id": config.StringVariable("app-definition-id"),
		"resource_type_id":  config.StringVariable("resource-provider:test"),
	}

	server.SetAppDefinition("organization-id", "app-definition-id", cm.AppDefinitionData{
		Name: "Test App",
	})

	server.SetResourceProvider("organization-id", "app-definition-id", cm.ResourceProviderRequest{
		Sys: cm.ResourceProviderRequestSys{
			ID: "resource-provider",
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
func TestAccResourceTypeResourceImport(t *testing.T) {
	server, _ := cmt.NewContentfulManagementServer()

	configVariables := config.Variables{
		"organization_id":   config.StringVariable("organization-id"),
		"app_definition_id": config.StringVariable("app-definition-id"),
		"resource_type_id":  config.StringVariable("resource-provider:test"),
	}

	server.SetAppDefinition("organization-id", "app-definition-id", cm.AppDefinitionData{
		Name: "Test App",
	})

	server.SetResourceProvider("organization-id", "app-definition-id", cm.ResourceProviderRequest{
		Sys: cm.ResourceProviderRequestSys{
			ID: "resource-provider",
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
					resource.TestCheckResourceAttr("contentful_resource_type.test", "id", "organization-id/app-definition-id/resource-provider:test"),
				),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ImportState:     true,
				ResourceName:    "contentful_resource_type.test",
			},
		},
	})
}

//nolint:paralleltest
func TestAccResourceTypeResourceMovedFromAppDefinitionResourceType(t *testing.T) {
	server, _ := cmt.NewContentfulManagementServer()

	server.SetAppDefinition("organization-id", "app-definition-id", cm.AppDefinitionData{
		Name: "Test App",
	})

	server.SetResourceProvider("organization-id", "app-definition-id", cm.ResourceProviderRequest{
		Sys: cm.ResourceProviderRequestSys{
			ID: "resource-provider",
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
				ConfigVariables: config.Variables{
					"organization_id":      config.StringVariable("organization-id"),
					"app_definition_id":    config.StringVariable("app-definition-id"),
					"resource_provider_id": config.StringVariable("resource-provider"),
				},
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: config.Variables{
					"organization_id":   config.StringVariable("organization-id"),
					"app_definition_id": config.StringVariable("app-definition-id"),
					"resource_type_id":  config.StringVariable("resource-provider:test"),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_resource_type.test", plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}
