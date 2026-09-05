package provider_test

import (
	"net/http"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

func TestAccResourceTypeResourceLifecycle(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)

	configVariables := config.Variables{
		"organization_id":   config.StringVariable("organization-id"),
		"app_definition_id": config.StringVariable("app-definition-id"),
		"resource_type_id":  config.StringVariable("resource-provider:test"),
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
		CheckDestroy: func(_ *terraform.State) error {
			response, err := server.Handler().GetResourceType(t.Context(), cm.GetResourceTypeParams{OrganizationID: "organization-id", AppDefinitionID: "app-definition-id", ResourceTypeID: "resource-provider:test"})
			require.NoError(t, err)

			status, ok := response.(cm.StatusCodeResponse)
			require.True(t, ok, "unexpected response after destroy: %T", response)
			require.Equal(t, http.StatusNotFound, status.GetStatusCode())

			return nil
		},
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("contentful_resource_type.test", tfjsonpath.New("id"), knownvalue.StringExact("organization-id/app-definition-id/resource-provider:test")),
					statecheck.ExpectKnownValue("contentful_resource_type.test", tfjsonpath.New("default_field_mapping").AtMapKey("title"), knownvalue.StringExact("{ /name }")),
				},
			},
			{
				ConfigDirectory:  config.TestStepDirectory(),
				ConfigVariables:  configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("contentful_resource_type.test", plancheck.ResourceActionUpdate)}},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("contentful_resource_type.test", tfjsonpath.New("id"), knownvalue.StringExact("organization-id/app-definition-id/resource-provider:test")),
					statecheck.ExpectKnownValue("contentful_resource_type.test", tfjsonpath.New("default_field_mapping").AtMapKey("title"), knownvalue.StringExact("{ /name }")),
					statecheck.ExpectKnownValue("contentful_resource_type.test", tfjsonpath.New("default_field_mapping").AtMapKey("subtitle"), knownvalue.StringExact("{ /description }")),
					statecheck.ExpectKnownValue("contentful_resource_type.test", tfjsonpath.New("default_field_mapping").AtMapKey("image").AtMapKey("url"), knownvalue.StringExact("{ /image }")),
					statecheck.ExpectKnownValue("contentful_resource_type.test", tfjsonpath.New("default_field_mapping").AtMapKey("badge").AtMapKey("label"), knownvalue.StringExact("beta")),
				},
			},
		},
	})
}

func TestAccResourceTypeResourceImport(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)

	configVariables := config.Variables{
		"organization_id":   config.StringVariable("organization-id"),
		"app_definition_id": config.StringVariable("app-definition-id"),
		"resource_type_id":  config.StringVariable("resource-provider:test"),
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
					resource.TestCheckResourceAttr("contentful_resource_type.test", "id", "organization-id/app-definition-id/resource-provider:test"),
				),
			},
			{
				ConfigDirectory:   config.TestNameDirectory(),
				ConfigVariables:   configVariables,
				ImportState:       true,
				ImportStateVerify: true,
				ResourceName:      "contentful_resource_type.test",
			},
		},
	})
}
