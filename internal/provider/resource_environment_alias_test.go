package provider_test

import (
	"maps"
	"net/http"
	"sync/atomic"
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccEnvironmentAliasResourceLifecycle(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
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

	testAccMockedResource(t, server, resource.TestCase{
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
				ConfigDirectory: config.StaticDirectory("testdata/environment_alias_resource"),
				ConfigVariables: configVariables1,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("contentful_environment_alias.test", tfjsonpath.New("id"), knownvalue.StringExact("space-id/"+environmentAliasID)),
				},
			},
			{
				ConfigDirectory:  config.StaticDirectory("testdata/environment_alias_resource"),
				ConfigVariables:  configVariables2,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("contentful_environment_alias.test", plancheck.ResourceActionUpdate)}},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("contentful_environment_alias.test", tfjsonpath.New("id"), knownvalue.StringExact("space-id/"+environmentAliasID)),
				},
			},
			{
				ConfigDirectory:   config.StaticDirectory("testdata/environment_alias_resource"),
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

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)

	server.RegisterSpaceEnvironment("space-id", "master-1970-01-01")
	server.SetEnvironmentAlias(cmt.NewEnvironmentAliasFromEnvironmentAliasData("space-id", "master", cm.EnvironmentAliasData{
		Environment: cm.NewEnvironmentLink("master-1970-01-01"),
	}))

	configVariables := config.Variables{
		"space_id":                  config.StringVariable("space-id"),
		"test_environment_alias_id": config.StringVariable("master"),
		"target_environment_id":     config.StringVariable("master-1970-01-01"),
	}

	var mutations atomic.Int64

	handler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			mutations.Add(1)
		}

		server.ServeHTTP(writer, request)
	})

	testAccMockedResource(t, handler, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory:    config.StaticDirectory("testdata/environment_alias_resource"),
				ConfigVariables:    configVariables,
				ImportState:        true,
				ImportStatePersist: true,
				ImportStateId:      "space-id/master",
				ResourceName:       "contentful_environment_alias.test",
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					assert.Zero(t, mutations.Load(), "import must only read the seeded environment alias")

					return testAccImportAttributes(map[string]string{
						"id":                    "space-id/master",
						"space_id":              "space-id",
						"environment_alias_id":  "master",
						"target_environment_id": "master-1970-01-01",
					})(states)
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/environment_alias_resource"),
				ConfigVariables: configVariables,
				PlanOnly:        true,
				// PlanOnly runs this phase against the imported state before refresh.
				ConfigPlanChecks: resource.ConfigPlanChecks{PostApplyPreRefresh: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("contentful_environment_alias.test", plancheck.ResourceActionNoop),
				}},
			},
		},
	})
}
