package provider_test

import (
	"maps"
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
		CheckDestroy: func(_ *terraform.State) error {
			response, err := server.Handler().GetResourceProvider(t.Context(), cm.GetResourceProviderParams{OrganizationID: "organization-id", AppDefinitionID: "app-definition-id"})
			require.NoError(t, err)

			status, ok := response.(cm.StatusCodeResponse)
			require.True(t, ok, "unexpected response after destroy: %T", response)
			require.Equal(t, http.StatusNotFound, status.GetStatusCode())

			return nil
		},
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: stepVariables1,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("contentful_resource_provider.test", tfjsonpath.New("id"), knownvalue.StringExact("organization-id/app-definition-id")),
					statecheck.ExpectKnownValue("contentful_resource_provider.test", tfjsonpath.New("function_id"), knownvalue.StringExact("resourceProvider")),
				},
			},
			{
				ConfigDirectory:  config.TestNameDirectory(),
				ConfigVariables:  stepVariables2,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("contentful_resource_provider.test", plancheck.ResourceActionUpdate)}},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("contentful_resource_provider.test", tfjsonpath.New("id"), knownvalue.StringExact("organization-id/app-definition-id")),
					statecheck.ExpectKnownValue("contentful_resource_provider.test", tfjsonpath.New("function_id"), knownvalue.StringExact("resourceProviderTwo")),
				},
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
				ConfigDirectory:   config.TestNameDirectory(),
				ConfigVariables:   configVariables,
				ImportState:       true,
				ImportStateVerify: true,
				ResourceName:      "contentful_resource_provider.test",
			},
		},
	})
}
