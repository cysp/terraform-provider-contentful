package provider_test

import (
	"maps"
	"net/http"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

func TestAccEnvironmentResourceLifecycle(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)

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
		CheckDestroy: func(_ *terraform.State) error {
			response, err := server.Handler().GetEnvironment(t.Context(), cm.GetEnvironmentParams{SpaceID: "space-id", EnvironmentID: environmentID})
			require.NoError(t, err)

			status, ok := response.(cm.StatusCodeResponse)
			require.True(t, ok, "unexpected response after destroy: %T", response)
			require.Equal(t, http.StatusNotFound, status.GetStatusCode())

			return nil
		},
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables1,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("contentful_environment.test", tfjsonpath.New("id"), knownvalue.StringExact("space-id/"+environmentID)),
					statecheck.ExpectKnownValue("contentful_environment.test", tfjsonpath.New("name"), knownvalue.StringExact("Test Environment")),
					statecheck.ExpectKnownValue("contentful_environment.test", tfjsonpath.New("environment_id"), knownvalue.StringExact(environmentID)),
				},
			},
			{
				ConfigDirectory:  config.TestNameDirectory(),
				ConfigVariables:  configVariables2,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("contentful_environment.test", plancheck.ResourceActionUpdate)}},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("contentful_environment.test", tfjsonpath.New("id"), knownvalue.StringExact("space-id/"+environmentID)),
					statecheck.ExpectKnownValue("contentful_environment.test", tfjsonpath.New("name"), knownvalue.StringExact("Updated Test Environment")),
					statecheck.ExpectKnownValue("contentful_environment.test", tfjsonpath.New("environment_id"), knownvalue.StringExact(environmentID)),
				},
			},
			{
				ConfigDirectory:   config.TestNameDirectory(),
				ConfigVariables:   configVariables2,
				ImportState:       true,
				ImportStateVerify: true,
				ResourceName:      "contentful_environment.test",
			},
		},
	})
}

func TestAccEnvironmentResourceImport(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)

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
				ConfigDirectory:   config.TestNameDirectory(),
				ConfigVariables:   configVariables,
				ImportState:       true,
				ImportStateVerify: true,
				ResourceName:      "contentful_environment.test",
			},
		},
	})
}
