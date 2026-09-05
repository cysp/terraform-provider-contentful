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

func TestAccEnvironmentAliasResourceLifecycle(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)

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
		CheckDestroy: func(_ *terraform.State) error {
			response, err := server.Handler().GetEnvironmentAlias(t.Context(), cm.GetEnvironmentAliasParams{SpaceID: "space-id", EnvironmentAliasID: environmentAliasID})
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
					statecheck.ExpectKnownValue("contentful_environment_alias.test", tfjsonpath.New("id"), knownvalue.StringExact("space-id/"+environmentAliasID)),
					statecheck.ExpectKnownValue("contentful_environment_alias.test", tfjsonpath.New("target_environment_id"), knownvalue.StringExact("staging")),
				},
			},
			{
				ConfigDirectory:  config.TestNameDirectory(),
				ConfigVariables:  configVariables2,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("contentful_environment_alias.test", plancheck.ResourceActionUpdate)}},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("contentful_environment_alias.test", tfjsonpath.New("id"), knownvalue.StringExact("space-id/"+environmentAliasID)),
					statecheck.ExpectKnownValue("contentful_environment_alias.test", tfjsonpath.New("target_environment_id"), knownvalue.StringExact("master")),
				},
			},
			{
				ConfigDirectory:   config.TestNameDirectory(),
				ConfigVariables:   configVariables2,
				ImportState:       true,
				ImportStateVerify: true,
				ResourceName:      "contentful_environment_alias.test",
			},
		},
	})
}

func TestAccEnvironmentAliasResourceImport(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)

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
				ConfigDirectory:   config.TestNameDirectory(),
				ConfigVariables:   configVariables,
				ImportState:       true,
				ImportStateVerify: true,
				ResourceName:      "contentful_environment_alias.test",
			},
		},
	})
}
