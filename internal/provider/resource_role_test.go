package provider_test

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"sync"
	"sync/atomic"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
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
				ConfigDirectory:  config.TestNameDirectory(),
				ConfigVariables:  configVariables,
				ResourceName:     "contentful_role.author",
				ImportState:      true,
				ImportStateId:    "0p38pssr0fi3/2EZrF9oDqi4AnsLNy21n6z",
				ImportStateCheck: testAccImportAttributes(map[string]string{"id": "0p38pssr0fi3/2EZrF9oDqi4AnsLNy21n6z", "space_id": "0p38pssr0fi3", "role_id": "2EZrF9oDqi4AnsLNy21n6z"}),
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

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("0p38pssr0fi3", "master")

	var (
		roleID       string
		requestMutex sync.Mutex
		mutations    []string
	)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			requestMutex.Lock()

			mutations = append(mutations, r.Method+" "+r.URL.Path+" version="+r.Header.Get("X-Contentful-Version"))
			requestMutex.Unlock()
		}

		server.ServeHTTP(w, r)
	})
	identity := statecheck.CompareValue(compare.ValuesSame())
	configVariables := config.Variables{"space_id": config.StringVariable("0p38pssr0fi3")}

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{
		CheckDestroy: func(_ *terraform.State) error {
			response, err := server.Handler().GetRole(t.Context(), cm.GetRoleParams{SpaceID: "0p38pssr0fi3", RoleID: roleID})
			require.NoError(t, err)

			status, ok := response.(cm.StatusCodeResponse)
			require.True(t, ok, "unexpected response after destroy: %T", response)
			require.Equal(t, http.StatusNotFound, status.GetStatusCode())

			return nil
		},
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/TestAccRoleResourceCreateUpdateDelete/1"),
				ConfigVariables: configVariables,
				ConfigStateChecks: []statecheck.StateCheck{
					identity.AddStateValue("contentful_role.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("contentful_role.test", tfjsonpath.New("permissions"), knownvalue.MapExact(map[string]knownvalue.Check{})),
					statecheck.ExpectKnownValue("contentful_role.test", tfjsonpath.New("policies"), knownvalue.ListExact([]knownvalue.Check{})),
				},
				Check: func(state *terraform.State) error {
					roleID = state.RootModule().Resources["contentful_role.test"].Primary.Attributes["role_id"]
					require.NotEmpty(t, roleID)
					response, err := server.Handler().GetRole(t.Context(), cm.GetRoleParams{SpaceID: "0p38pssr0fi3", RoleID: roleID})
					require.NoError(t, err)

					role, ok := response.(*cm.Role)
					require.True(t, ok, "unexpected role response: %T", response)
					require.Equal(t, "Test", role.Name)

					return nil
				},
			},
			{
				ConfigDirectory:   config.StaticDirectory("testdata/TestAccRoleResourceCreateUpdateDelete/1"),
				ConfigVariables:   configVariables,
				ResourceName:      "contentful_role.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ConfigDirectory:  config.StaticDirectory("testdata/TestAccRoleResourceCreateUpdateDelete/2"),
				ConfigVariables:  configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("contentful_role.test", plancheck.ResourceActionUpdate)}},
				ConfigStateChecks: []statecheck.StateCheck{
					identity.AddStateValue("contentful_role.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("contentful_role.test", tfjsonpath.New("permissions"), knownvalue.MapExact(map[string]knownvalue.Check{
						"ContentDelivery":    knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("all")}),
						"ContentModel":       knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("read")}),
						"EnvironmentAliases": knownvalue.ListExact([]knownvalue.Check{}),
						"Environments":       knownvalue.ListExact([]knownvalue.Check{}),
						"Settings":           knownvalue.ListExact([]knownvalue.Check{}),
						"Tags":               knownvalue.ListExact([]knownvalue.Check{}),
					})),
				},
			},
		},
	})

	requestMutex.Lock()
	defer requestMutex.Unlock()

	require.Equal(t, []string{
		"POST /spaces/0p38pssr0fi3/roles version=",
		"PUT /spaces/0p38pssr0fi3/roles/" + roleID + " version=0",
		"DELETE /spaces/0p38pssr0fi3/roles/" + roleID + " version=",
	}, mutations)
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
	assert.Equal(t, tfprotov6.DiagnosticSeverityError, response.Diagnostics[0].Severity)
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

func TestAccRoleResourceImportedVersionUsedForUpdate(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("0p38pssr0fi3", "master")
	server.SetRole("0p38pssr0fi3", "imported", cm.RoleData{Name: "Test"})
	response, err := server.Handler().GetRole(t.Context(), cm.GetRoleParams{SpaceID: "0p38pssr0fi3", RoleID: "imported"})
	require.NoError(t, err)

	role, ok := response.(*cm.Role)
	require.True(t, ok)
	// Seed before serving any requests; the mock does not model version increments.
	role.Sys.Version = 37

	var puts, posts atomic.Int64

	var updateVersion atomic.Value

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut && r.URL.Path == "/spaces/0p38pssr0fi3/roles/imported" {
			puts.Add(1)
			updateVersion.Store(r.Header.Get("X-Contentful-Version"))
		}

		if r.Method == http.MethodPost {
			posts.Add(1)
		}

		server.ServeHTTP(w, r)
	})
	variables := config.Variables{"space_id": config.StringVariable("0p38pssr0fi3")}
	// Use the version persisted by import, without a planning refresh replacing it.
	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
		Steps: []resource.TestStep{
			{
				ConfigDirectory:    config.StaticDirectory("testdata/TestAccRoleResourceCreateUpdateDelete/1"),
				ConfigVariables:    variables,
				ResourceName:       "contentful_role.test",
				ImportState:        true,
				ImportStateId:      "0p38pssr0fi3/imported",
				ImportStatePersist: true,
				ImportStateCheck:   testAccImportAttributes(map[string]string{"role_id": "imported", "name": "Test"}),
			},
			{
				ConfigDirectory:   config.StaticDirectory("testdata/TestAccRoleResourceCreateUpdateDelete/2"),
				ConfigVariables:   variables,
				ConfigPlanChecks:  resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("contentful_role.test", plancheck.ResourceActionUpdate)}},
				ConfigStateChecks: []statecheck.StateCheck{statecheck.ExpectKnownValue("contentful_role.test", tfjsonpath.New("role_id"), knownvalue.StringExact("imported"))},
			},
		},
	})
	require.Equal(t, int64(1), puts.Load())
	require.Zero(t, posts.Load())
	require.Equal(t, "37", updateVersion.Load())
}
