package provider_test

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"sync/atomic"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:paralleltest
func TestAccRoleResourceImport(t *testing.T) {
	parallelWhenMocked(t)

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)

	configVariables := config.Variables{
		"space_id": config.StringVariable("0p38pssr0fi3"),
	}

	server.SetRole("0p38pssr0fi3", "2EZrF9oDqi4AnsLNy21n6z", cm.RoleData{
		Name: "author",
	})

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory:    config.TestNameDirectory(),
				ConfigVariables:    configVariables,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				ConfigDirectory:    config.TestNameDirectory(),
				ConfigVariables:    configVariables,
				ResourceName:       "contentful_role.author",
				ImportState:        true,
				ImportStateId:      "0p38pssr0fi3/2EZrF9oDqi4AnsLNy21n6z",
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

//nolint:paralleltest
func TestAccRoleResourceImportNotFound(t *testing.T) {
	parallelWhenMocked(t)

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)

	configVariables := config.Variables{
		"space_id": config.StringVariable("0p38pssr0fi3"),
	}

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory:    config.TestNameDirectory(),
				ConfigVariables:    configVariables,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ResourceName:    "contentful_role.test",
				ImportState:     true,
				ImportStateId:   "0p38pssr0fi3/nonexistent",
				ExpectError:     regexp.MustCompile(`Cannot import non-existent remote object`),
			},
		},
	})
}

func TestAccRoleResourceCreateUpdateDelete(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)

	server.RegisterSpaceEnvironment("0p38pssr0fi3", "master")

	configVariables := config.Variables{
		"space_id": config.StringVariable("0p38pssr0fi3"),
	}

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

func TestRoleUpdateRequestConversionErrorStopsBeforeMutation(t *testing.T) {
	t.Parallel()

	var requestCount atomic.Int64

	testServer := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		requestCount.Add(1)
		http.Error(response, "unexpected Contentful request", http.StatusInternalServerError)
	}))
	t.Cleanup(testServer.Close)

	providerServer, err := makeTestAccProtoV6ProviderFactories(
		WithContentfulURL(testServer.URL),
		WithAccessToken("test-token"),
	)["contentful"]()
	require.NoError(t, err)

	providerConfig, err := providerConfigDynamicValue(map[string]any{
		"url":          tftypes.UnknownValue,
		"access_token": tftypes.UnknownValue,
	})
	require.NoError(t, err)

	configureResponse, err := providerServer.ConfigureProvider(t.Context(), &tfprotov6.ConfigureProviderRequest{
		Config: &providerConfig,
	})
	require.NoError(t, err)
	require.Empty(t, configureResponse.Diagnostics)

	priorModel := validRoleRequestModel()
	priorModel.ID = types.StringValue("space/role")
	priorModel.SpaceID = types.StringValue("space")
	priorModel.RoleID = types.StringValue("role")
	priorModel.Description = types.StringNull()
	priorModel.Timeouts = TimeoutsNull()

	plannedModel := priorModel
	plannedModel.Policies = rolePoliciesWith(RolePolicyValue{
		Actions:    NewTypedList([]types.String{types.StringValue("all"), types.StringValue("read")}),
		Constraint: jsontypes.NewNormalizedNull(),
		Effect:     types.StringValue("allow"),
	})

	configModel := priorModel
	configModel.ID = types.StringNull()
	configModel.RoleID = types.StringNull()

	priorState := resourceModelDynamicValue(t, RoleResourceSchema(t.Context()), priorModel)
	plannedState := resourceModelDynamicValue(t, RoleResourceSchema(t.Context()), plannedModel)
	configValue := resourceModelDynamicValue(t, RoleResourceSchema(t.Context()), configModel)

	response, err := providerServer.ApplyResourceChange(t.Context(), &tfprotov6.ApplyResourceChangeRequest{
		TypeName:       "contentful_role",
		PriorState:     &priorState,
		PlannedState:   &plannedState,
		Config:         &configValue,
		PlannedPrivate: privateVersionBytes(t, 1),
	})
	require.NoError(t, err)
	require.Len(t, response.Diagnostics, 1)
	assert.Equal(t, "Invalid policy actions", response.Diagnostics[0].Summary)
	assert.Equal(t, `"all" must be specified by itself. Remove "all" or the other policy actions from this list.`, response.Diagnostics[0].Detail)
	assert.Equal(t,
		tftypes.NewAttributePath().WithAttributeName("policies").WithElementKeyInt(0).WithAttributeName("actions"),
		response.Diagnostics[0].Attribute,
	)
	assert.Zero(t, requestCount.Load())
}

func TestAccRoleResourceDeleted(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)

	server.RegisterSpaceEnvironment("0p38pssr0fi3", "master")

	server.SetRole("0p38pssr0fi3", "test", cm.RoleData{
		Name: "Test",
	})

	configVariables := config.Variables{
		"space_id": config.StringVariable("0p38pssr0fi3"),
	}

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
			{
				ConfigDirectory:    config.TestStepDirectory(),
				ConfigVariables:    configVariables,
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_role.test", plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}
