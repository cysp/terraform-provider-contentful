package provider_test

import (
	"encoding/json"
	"io"
	"maps"
	"net/http"
	"sync/atomic"
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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

	cases := []struct {
		typeName, attribute, createPath, deletePath, id string
		schema                                          schema.Schema
		plan, body                                      map[string]any
		omitConfiguredID                                bool
	}{
		{
			typeName: "contentful_preview_environment", attribute: "preview_environment_id", schema: PreviewEnvironmentResourceSchema(t.Context()),
			createPath: "/spaces/space/preview_environments", deletePath: "/spaces/space/preview_environments/allocated", id: "space/allocated",
			plan:             map[string]any{"space_id": "space", "name": "Preview", "description": "", "content_type_configurations": map[string]any{"page": map[string]any{"url": "https://preview.invalid/page"}}},
			body:             readEndpointBody("PreviewEnvironment", map[string]any{"name": "Returned", "description": "", "configurations": []any{}}),
			omitConfiguredID: true,
		},
		{
			typeName: "contentful_role", attribute: "role_id", schema: RoleResourceSchema(t.Context()),
			createPath: "/spaces/space/roles", deletePath: "/spaces/space/roles/allocated", id: "space/allocated",
			plan: map[string]any{"space_id": "space", "name": "Role", "permissions": map[string]any{"ContentModel": []any{"read"}}, "policies": []any{map[string]any{"actions": []any{"read"}, "effect": "allow"}}},
			body: readEndpointBody("Role", map[string]any{"name": "Returned", "permissions": map[string]any{}}),
		},
		{
			typeName: "contentful_webhook", attribute: "webhook_id", schema: WebhookResourceSchema(t.Context()),
			createPath: "/spaces/space/webhook_definitions", deletePath: "/spaces/space/webhook_definitions/allocated", id: "space/allocated",
			plan: map[string]any{"space_id": "space", "name": "Webhook", "url": "https://example.com/webhook", "active": true},
			body: readEndpointBody("WebhookDefinition", map[string]any{"name": "Returned", "url": "https://example.test/returned", "topics": []any{"Entry.delete"}}),
		},
		{
			typeName: "contentful_locale", attribute: "locale_id", schema: LocaleResourceSchema(t.Context()),
			createPath: "/spaces/space/environments/environment/locales", deletePath: "/spaces/space/environments/environment/locales/allocated", id: "space/environment/allocated",
			plan: map[string]any{"space_id": "space", "environment_id": "environment", "name": "English (Australia)", "code": "en-AU", "content_delivery_api": true, "content_management_api": true, "optional": false, "default": false},
			body: readEndpointBody("Locale", map[string]any{"name": "Returned", "code": "en-US", "fallbackCode": nil, "default": false, "contentDeliveryApi": true, "contentManagementApi": true, "optional": true}),
		},
	}

	for _, test := range cases {
		for _, constraint := range []struct {
			name          string
			value         types.String
			expectIDError bool
		}{
			{"known", types.StringValue("planned"), true},
			{"null", types.StringNull(), true},
			{"unknown", types.StringUnknown(), false},
		} {
			t.Run(test.typeName+"/"+constraint.name, func(t *testing.T) {
				t.Parallel()

				var creates, deletes atomic.Int64

				body := maps.Clone(test.body)
				sys, ok := body["sys"].(map[string]any)
				require.True(t, ok)

				sys = maps.Clone(sys)
				body["sys"] = sys
				sys["id"] = "allocated"
				sys["space"] = map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Space", "id": "space"}}
				sys["environment"] = map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Environment", "id": "environment"}}
				protocol := configuredProtocolServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.Method {
					case http.MethodPost:
						creates.Add(1)
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
						deletes.Add(1)
						assert.Equal(t, test.deletePath, r.URL.Path)
						w.WriteHeader(http.StatusNoContent)
					default:
						t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
						w.WriteHeader(http.StatusBadRequest)
					}
				}))
				prior := nullResourceDynamicValue(t, test.schema)
				input := mutationEndpointDynamicValue(t, test.schema, test.plan)
				raw, err := input.Unmarshal(test.schema.Type().TerraformType(t.Context()))
				require.NoError(t, err)

				frameworkPlan := tfsdk.Plan{Schema: test.schema, Raw: raw}
				require.Empty(t, frameworkPlan.SetAttribute(t.Context(), path.Root("id"), types.StringUnknown()))
				require.Empty(t, frameworkPlan.SetAttribute(t.Context(), path.Root(test.attribute), constraint.value))
				planned, err := tfprotov6.NewDynamicValue(test.schema.Type().TerraformType(t.Context()), frameworkPlan.Raw)
				require.NoError(t, err)
				// Preview endpoint selection follows Config presence while effective Plan
				// values remain consistency constraints for the generated result.
				if test.omitConfiguredID {
					require.Empty(t, frameworkPlan.SetAttribute(t.Context(), path.Root(test.attribute), types.StringNull()))
				}

				configuration, err := tfprotov6.NewDynamicValue(test.schema.Type().TerraformType(t.Context()), frameworkPlan.Raw)
				require.NoError(t, err)
				response, err := protocol.ApplyResourceChange(t.Context(), &tfprotov6.ApplyResourceChangeRequest{TypeName: test.typeName, PriorState: &prior, PlannedState: &planned, Config: &configuration})
				require.NoError(t, err)
				require.EqualValues(t, 1, creates.Load(), response.Diagnostics)

				foundIDError := false

				for _, diagnostic := range response.Diagnostics {
					if diagnostic.Attribute != nil && diagnostic.Attribute.Equal(tftypes.NewAttributePath().WithAttributeName(test.attribute)) {
						foundIDError = true
					}
				}

				assert.Equal(t, constraint.expectIDError, foundIDError, "only unknown Plan IDs are unconstrained")
				require.NotNil(t, response.NewState)
				raw, err = response.NewState.Unmarshal(test.schema.Type().TerraformType(t.Context()))
				require.NoError(t, err)

				state := tfsdk.State{Schema: test.schema, Raw: raw}

				var id, allocated types.String
				require.Empty(t, state.GetAttribute(t.Context(), path.Root("id"), &id))
				require.Empty(t, state.GetAttribute(t.Context(), path.Root(test.attribute), &allocated))
				assert.Equal(t, types.StringValue(test.id), id)
				assert.Equal(t, types.StringValue("allocated"), allocated)

				var private map[string][]byte
				require.NoError(t, json.Unmarshal(response.Private, &private))
				assert.Equal(t, "73", string(private["version"]), "the generated object confirmed its endpoint")
				deleted, err := protocol.ApplyResourceChange(t.Context(), &tfprotov6.ApplyResourceChangeRequest{TypeName: test.typeName, PriorState: response.NewState, PlannedState: &prior, Config: &prior, PlannedPrivate: response.Private})
				require.NoError(t, err)
				assert.Empty(t, deleted.Diagnostics)
				assert.EqualValues(t, 1, deletes.Load(), "the allocated object must be deleted")
			})
		}
	}
}
