package provider_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreviewEnvironmentMutationConsistencyErrorCheckpointsResponseStateIdentityAndVersion(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		setup           func(*cmt.Server)
		method          string
		prior           *PreviewEnvironmentModel
		plan            PreviewEnvironmentModel
		config          PreviewEnvironmentModel
		plannedPrivate  []byte
		expectedID      string
		expectedVersion int
	}{
		"create": {
			method: http.MethodPost,
			plan: previewEnvironmentCheckpointModel(
				types.StringUnknown(),
				types.StringUnknown(),
				"Planned",
			),
			config: previewEnvironmentCheckpointModel(
				types.StringNull(),
				types.StringNull(),
				"Planned",
			),
			expectedVersion: 0,
		},
		"update": {
			setup: func(server *cmt.Server) {
				server.SetPreviewEnvironment("space", "preview", cm.PreviewEnvironmentData{
					Name:           "Prior",
					Description:    "",
					Configurations: []cm.PreviewEnvironmentConfigurationData{},
				})
			},
			method: http.MethodPut,
			prior: new(previewEnvironmentCheckpointModel(
				types.StringValue("space/preview"),
				types.StringValue("preview"),
				"Prior",
			)),
			plan: previewEnvironmentCheckpointModel(
				types.StringValue("space/preview"),
				types.StringValue("preview"),
				"Planned",
			),
			config: previewEnvironmentCheckpointModel(
				types.StringNull(),
				types.StringValue("preview"),
				"Planned",
			),
			plannedPrivate:  privateVersionBytes(t, 0),
			expectedID:      "preview",
			expectedVersion: 1,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
			require.NoError(t, err)
			server.RegisterSpaceEnvironment("space", "master")

			if test.setup != nil {
				test.setup(server)
			}

			adapter := &mutationJSONResponseAdapter{delegate: server}
			adapter.mutateNext(test.method, func(response map[string]any) {
				response["name"] = "Returned"
			})

			testServer := httptest.NewServer(adapter)
			t.Cleanup(testServer.Close)

			providerServer, err := makeTestAccProtoV6ProviderFactories(
				WithContentfulURL(testServer.URL),
				WithAccessToken(cmt.ValidAccessToken),
			)["contentful"]()
			require.NoError(t, err)

			providerConfig, err := providerConfigDynamicValue(map[string]any{
				"url":          testServer.URL,
				"access_token": cmt.ValidAccessToken,
			})
			require.NoError(t, err)
			configureResponse, err := providerServer.ConfigureProvider(t.Context(), &tfprotov6.ConfigureProviderRequest{Config: &providerConfig})
			require.NoError(t, err)
			require.Empty(t, configureResponse.Diagnostics)

			resourceSchema := PreviewEnvironmentResourceSchema(t.Context())
			plannedState := resourceModelDynamicValue(t, resourceSchema, test.plan)
			configValue := resourceModelDynamicValue(t, resourceSchema, test.config)

			priorState := nullResourceDynamicValue(t, resourceSchema)
			if test.prior != nil {
				priorState = resourceModelDynamicValue(t, resourceSchema, test.prior)
			}

			applyResponse, err := providerServer.ApplyResourceChange(t.Context(), &tfprotov6.ApplyResourceChangeRequest{
				TypeName:       "contentful_preview_environment",
				PriorState:     &priorState,
				PlannedState:   &plannedState,
				Config:         &configValue,
				PlannedPrivate: test.plannedPrivate,
			})
			require.NoError(t, err)
			require.Len(t, applyResponse.Diagnostics, 1)
			assert.Equal(t, "Contentful returned a different content preview platform name", applyResponse.Diagnostics[0].Summary)

			state := previewEnvironmentModelFromDynamicValue(t, *applyResponse.NewState)
			assert.Equal(t, types.StringValue("Returned"), state.Name)
			assert.Equal(t, types.StringValue("space"), state.SpaceID)

			if test.expectedID == "" {
				require.False(t, state.PreviewEnvironmentID.IsNull())
				require.False(t, state.PreviewEnvironmentID.IsUnknown())
				assert.NotEmpty(t, state.PreviewEnvironmentID.ValueString())
			} else {
				assert.Equal(t, types.StringValue(test.expectedID), state.PreviewEnvironmentID)
			}

			identity := previewEnvironmentIdentityFromApplyResponse(t, applyResponse)
			assert.Equal(t, state.SpaceID, identity.SpaceID)
			assert.Equal(t, state.PreviewEnvironmentID, identity.PreviewEnvironmentID)
			assertPreviewEnvironmentPrivateVersion(t, applyResponse.Private, test.expectedVersion)
		})
	}
}

func previewEnvironmentCheckpointModel(id, previewEnvironmentID types.String, name string) PreviewEnvironmentModel {
	return PreviewEnvironmentModel{
		IDIdentityModel: IDIdentityModel{ID: id},
		PreviewEnvironmentIdentityModel: PreviewEnvironmentIdentityModel{
			SpaceID:              types.StringValue("space"),
			PreviewEnvironmentID: previewEnvironmentID,
		},
		Name:                      types.StringValue(name),
		Description:               types.StringValue(""),
		ContentTypeConfigurations: NewTypedMap(map[string]TypedObject[PreviewEnvironmentContentTypeConfigurationValue]{}),
		Timeouts:                  TimeoutsNull(),
	}
}

func previewEnvironmentModelFromDynamicValue(t *testing.T, dynamicValue tfprotov6.DynamicValue) PreviewEnvironmentModel {
	t.Helper()

	ctx := t.Context()
	resourceSchema := PreviewEnvironmentResourceSchema(ctx)
	terraformValue, err := dynamicValue.Unmarshal(resourceSchema.Type().TerraformType(ctx))
	require.NoError(t, err)

	state := tfsdk.State{Raw: terraformValue, Schema: resourceSchema}

	var model PreviewEnvironmentModel

	diagnostics := state.Get(ctx, &model)
	require.False(t, diagnostics.HasError(), diagnostics.Errors())

	return model
}

func previewEnvironmentIdentityFromApplyResponse(
	t *testing.T,
	response *tfprotov6.ApplyResourceChangeResponse,
) PreviewEnvironmentIdentityModel {
	t.Helper()
	require.NotNil(t, response.NewIdentity)
	require.NotNil(t, response.NewIdentity.IdentityData)

	ctx := t.Context()
	identitySchema := identityschema.Schema{Attributes: map[string]identityschema.Attribute{
		"space_id":               identityschema.StringAttribute{RequiredForImport: true},
		"preview_environment_id": identityschema.StringAttribute{RequiredForImport: true},
	}}
	terraformValue, err := response.NewIdentity.IdentityData.Unmarshal(identitySchema.Type().TerraformType(ctx))
	require.NoError(t, err)

	identity := tfsdk.ResourceIdentity{Raw: terraformValue, Schema: identitySchema}

	var model PreviewEnvironmentIdentityModel

	diagnostics := identity.Get(ctx, &model)
	require.False(t, diagnostics.HasError(), diagnostics.Errors())

	return model
}

func assertPreviewEnvironmentPrivateVersion(t *testing.T, private []byte, expected int) {
	t.Helper()

	var values map[string][]byte
	require.NoError(t, json.Unmarshal(private, &values))
	assert.Equal(t, []byte(strconv.Itoa(expected)), values["version"])
}
