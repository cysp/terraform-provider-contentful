package provider_test

import (
	"encoding/json"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A known computed Plan ID is an artificial protocol input here: it constrains
// consistency but does not select a server-generated POST identity.
func TestGeneratedMutationIdentityRemainsResponseOwnedAcrossPlanConstraints(t *testing.T) {
	t.Parallel()

	cases := map[string]struct{ attribute, createPath, deletePath, id string }{
		"contentful_preview_environment": {"preview_environment_id", "/spaces/space/preview_environments", "/spaces/space/preview_environments/allocated", "space/allocated"},
		"contentful_role":                {"role_id", "/spaces/space/roles", "/spaces/space/roles/allocated", "space/allocated"},
		"contentful_webhook":             {"webhook_id", "/spaces/space/webhook_definitions", "/spaces/space/webhook_definitions/allocated", "space/allocated"},
		"contentful_locale":              {"locale_id", "/spaces/space/environments/environment/locales", "/spaces/space/environments/environment/locales/allocated", "space/environment/allocated"},
	}

	bodies := make(map[string]map[string]any)
	for _, fixture := range readEndpointCases(t) {
		bodies["contentful_"+fixture.name] = fixture.body
	}

	for _, fixture := range requiredPrivateVersionResourceCases(t) {
		test, ok := cases[fixture.typeName]
		if !ok || fixture.delete {
			continue
		}

		for _, constraint := range []struct {
			name          string
			value         types.String
			expectIDError bool
		}{
			{"known", types.StringValue("planned"), true},
			{"null", types.StringNull(), true},
			{"unknown", types.StringUnknown(), false},
		} {
			t.Run(fixture.typeName+"/"+constraint.name, func(t *testing.T) {
				t.Parallel()

				body := maps.Clone(bodies[fixture.typeName])
				sys, ok := body["sys"].(map[string]any)
				require.True(t, ok)

				sys = maps.Clone(sys)
				body["sys"] = sys
				sys["id"] = "allocated"
				sys["space"] = map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Space", "id": "space"}}
				sys["environment"] = map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Environment", "id": "environment"}}
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.Method {
					case http.MethodPost:
						assert.Equal(t, test.createPath, r.URL.Path)
						request, err := io.ReadAll(r.Body)
						assert.NoError(t, err)

						var fields map[string]any
						assert.NoError(t, json.Unmarshal(request, &fields))
						assert.NotContains(t, fields, "id")
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusCreated)
						_, _ = w.Write([]byte(testJSON(body)))
					case http.MethodDelete:
						assert.Equal(t, test.deletePath, r.URL.Path)
						w.WriteHeader(http.StatusNoContent)
					default:
						t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
						w.WriteHeader(http.StatusBadRequest)
					}
				}))
				t.Cleanup(server.Close)
				protocol, err := makeTestAccProtoV6ProviderFactories(WithContentfulURL(server.URL), WithAccessToken("test-token"))["contentful"]()
				require.NoError(t, err)
				config, err := providerConfigDynamicValue(map[string]any{"url": server.URL, "access_token": "test-token"})
				require.NoError(t, err)
				configured, err := protocol.ConfigureProvider(t.Context(), &tfprotov6.ConfigureProviderRequest{Config: &config})
				require.NoError(t, err)
				require.Empty(t, configured.Diagnostics)
				prior := nullResourceDynamicValue(t, fixture.resourceSchema)
				frameworkPlan := tfsdk.Plan{Schema: fixture.resourceSchema}
				require.Empty(t, frameworkPlan.Set(t.Context(), fixture.model))
				require.Empty(t, frameworkPlan.SetAttribute(t.Context(), path.Root("id"), types.StringUnknown()))
				require.Empty(t, frameworkPlan.SetAttribute(t.Context(), path.Root(test.attribute), constraint.value))
				planned, err := tfprotov6.NewDynamicValue(fixture.resourceSchema.Type().TerraformType(t.Context()), frameworkPlan.Raw)
				require.NoError(t, err)
				// Preview endpoint selection follows Config presence while effective Plan
				// values remain consistency constraints for the generated result.
				if fixture.typeName == "contentful_preview_environment" {
					require.Empty(t, frameworkPlan.SetAttribute(t.Context(), path.Root(test.attribute), types.StringNull()))
				}

				configuration, err := tfprotov6.NewDynamicValue(fixture.resourceSchema.Type().TerraformType(t.Context()), frameworkPlan.Raw)
				require.NoError(t, err)
				response, err := protocol.ApplyResourceChange(t.Context(), &tfprotov6.ApplyResourceChangeRequest{TypeName: fixture.typeName, PriorState: &prior, PlannedState: &planned, Config: &configuration})
				require.NoError(t, err)

				foundIDError := false

				for _, diagnostic := range response.Diagnostics {
					if diagnostic.Attribute != nil && diagnostic.Attribute.Equal(tftypes.NewAttributePath().WithAttributeName(test.attribute)) {
						foundIDError = true
					}
				}

				assert.Equal(t, constraint.expectIDError, foundIDError, "only unknown Plan IDs are unconstrained")
				require.NotNil(t, response.NewState)
				raw, err := response.NewState.Unmarshal(fixture.resourceSchema.Type().TerraformType(t.Context()))
				require.NoError(t, err)

				state := tfsdk.State{Schema: fixture.resourceSchema, Raw: raw}

				var id, allocated types.String
				require.Empty(t, state.GetAttribute(t.Context(), path.Root("id"), &id))
				require.Empty(t, state.GetAttribute(t.Context(), path.Root(test.attribute), &allocated))
				assert.Equal(t, types.StringValue(test.id), id)
				assert.Equal(t, types.StringValue("allocated"), allocated)

				var private map[string][]byte
				require.NoError(t, json.Unmarshal(response.Private, &private))
				assert.Equal(t, "73", string(private["version"]), "the generated object confirmed its endpoint")
				deleted, err := protocol.ApplyResourceChange(t.Context(), &tfprotov6.ApplyResourceChangeRequest{TypeName: fixture.typeName, PriorState: response.NewState, PlannedState: &prior, Config: &prior, PlannedPrivate: response.Private})
				require.NoError(t, err)
				assert.Empty(t, deleted.Diagnostics)
			})
		}
	}
}
