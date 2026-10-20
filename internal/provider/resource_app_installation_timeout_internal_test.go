package provider

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppInstallationTimeoutOnlyUpdate(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, updateTimeout string
		parameters          jsontypes.Normalized
		marketplace         types.Set
		diagnostic          string
	}{
		{name: "normalized JSON and unordered acknowledgments", updateTimeout: "3m", parameters: jsontypes.NewNormalizedValue(`{ "visible": "one", "secret": "alpha" }`), marketplace: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("terms"), types.StringValue("privacy")})},
		{name: "invalid timeout", updateTimeout: "invalid", parameters: jsontypes.NewNormalizedValue(`{"secret":"alpha","visible":"one"}`), marketplace: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("privacy"), types.StringValue("terms")}), diagnostic: "Timeout Cannot Be Parsed"},
		{name: "unknown parameters", updateTimeout: "3m", parameters: jsontypes.NewNormalizedUnknown(), marketplace: types.SetNull(types.StringType), diagnostic: "Unexpected unknown app installation parameters"},
		{name: "unknown acknowledgments", updateTimeout: "3m", parameters: jsontypes.NewNormalizedValue(`{"secret":"alpha","visible":"one"}`), marketplace: types.SetUnknown(types.StringType), diagnostic: "Unexpected unknown string set"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var calls atomic.Int64

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				calls.Add(1)
				w.WriteHeader(http.StatusInternalServerError)
			}))
			t.Cleanup(server.Close)
			client, err := cm.NewClient(server.URL, cm.NewAccessTokenSecuritySource("test-token"))
			require.NoError(t, err)

			implementation := appInstallationResource{providerData: ContentfulProviderData{client: client}}
			state := tfsdk.State(appInstallationTestPlan(t, `{"secret":"alpha","visible":"one"}`))
			marketplace := types.SetValueMust(types.StringType, []attr.Value{types.StringValue("privacy"), types.StringValue("terms")})
			require.Empty(t, state.SetAttribute(t.Context(), path.Root("marketplace"), marketplace))
			plan := tfsdk.Plan(state)
			require.Empty(t, plan.SetAttribute(t.Context(), path.Root("parameters"), test.parameters))
			require.Empty(t, plan.SetAttribute(t.Context(), path.Root("marketplace"), test.marketplace))
			timeouts := types.ObjectValueMust(map[string]attr.Type{"create": types.StringType, "read": types.StringType, "update": types.StringType, "delete": types.StringType}, map[string]attr.Value{"create": types.StringNull(), "read": types.StringNull(), "update": types.StringValue(test.updateTimeout), "delete": types.StringNull()})
			require.Empty(t, plan.SetAttribute(t.Context(), path.Root("timeouts"), timeouts))

			identitySchema := resourceIdentitySchema(appInstallationIdentityAttributeNames())
			identity := &tfsdk.ResourceIdentity{Schema: identitySchema, Raw: tftypes.NewValue(identitySchema.Type().TerraformType(t.Context()), nil)}
			response := resource.UpdateResponse{State: state, Identity: identity}
			implementation.Update(t.Context(), resource.UpdateRequest{Plan: plan, State: state}, &response)
			assert.Zero(t, calls.Load())

			if test.diagnostic != "" {
				require.True(t, response.Diagnostics.HasError())
				require.Len(t, response.Diagnostics, 1)
				assert.Equal(t, test.diagnostic, response.Diagnostics[0].Summary())
				assert.True(t, state.Raw.Equal(response.State.Raw))

				return
			}

			require.Empty(t, response.Diagnostics)

			var actual AppInstallationModel
			require.Empty(t, response.State.Get(t.Context(), &actual))
			assert.JSONEq(t, `{"secret":"alpha","visible":"one"}`, actual.Parameters.ValueString())
			assert.True(t, actual.Marketplace.Equal(marketplace))
			assert.True(t, actual.Timeouts.Object.Equal(timeouts))
		})
	}
}
