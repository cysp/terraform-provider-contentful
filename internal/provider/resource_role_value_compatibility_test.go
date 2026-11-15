package provider_test

import (
	"maps"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoleUnknownContainersStopBeforeMutation(t *testing.T) {
	t.Parallel()

	stringList := tftypes.List{ElementType: tftypes.String}
	permissionsType := tftypes.Map{ElementType: stringList}
	policyType := roleCompatibilityPolicyType()
	policiesType := tftypes.List{ElementType: policyType}
	cases := []struct {
		name      string
		attribute string
		value     tftypes.Value
		path      *tftypes.AttributePath
	}{
		{
			name: "permissions map", attribute: "permissions",
			value: tftypes.NewValue(permissionsType, tftypes.UnknownValue),
			path:  tftypes.NewAttributePath().WithAttributeName("permissions"),
		},
		{
			name: "permission list", attribute: "permissions",
			value: tftypes.NewValue(permissionsType, map[string]tftypes.Value{
				"ContentModel": tftypes.NewValue(stringList, tftypes.UnknownValue),
			}),
			path: tftypes.NewAttributePath().WithAttributeName("permissions").WithElementKeyString("ContentModel"),
		},
		{
			name: "policies list", attribute: "policies",
			value: tftypes.NewValue(policiesType, tftypes.UnknownValue),
			path:  tftypes.NewAttributePath().WithAttributeName("policies"),
		},
		{
			name: "policy object", attribute: "policies",
			value: tftypes.NewValue(policiesType, []tftypes.Value{tftypes.NewValue(policyType, tftypes.UnknownValue)}),
			path:  tftypes.NewAttributePath().WithAttributeName("policies").WithElementKeyInt(0),
		},
		{
			name: "policy actions list", attribute: "policies",
			value: tftypes.NewValue(policiesType, []tftypes.Value{roleCompatibilityPolicy(tftypes.UnknownValue, nil, "allow")}),
			path:  tftypes.NewAttributePath().WithAttributeName("policies").WithElementKeyInt(0).WithAttributeName("actions"),
		},
	}

	for _, create := range []bool{true, false} {
		operation := "update"
		if create {
			operation = "create"
		}

		for _, test := range cases {
			t.Run(operation+"/"+test.name, func(t *testing.T) {
				t.Parallel()

				var requests atomic.Int64

				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					requests.Add(1)
					w.WriteHeader(http.StatusBadRequest)
				}))
				t.Cleanup(server.Close)
				providerServer, err := makeTestAccProtoV6ProviderFactories(WithContentfulURL(server.URL), WithAccessToken("test-token"))["contentful"]()
				require.NoError(t, err)
				providerConfig, err := providerConfigDynamicValue(map[string]any{"url": server.URL, "access_token": "test-token"})
				require.NoError(t, err)
				configured, err := providerServer.ConfigureProvider(t.Context(), &tfprotov6.ConfigureProviderRequest{Config: &providerConfig})
				require.NoError(t, err)
				require.Empty(t, configured.Diagnostics)

				wireType := roleCompatibilityResourceType()
				priorValues := roleCompatibilityItem(map[string]tftypes.Value{}, []tftypes.Value{})
				priorValues["id"] = tftypes.NewValue(tftypes.String, "space/role")
				priorValues["space_id"] = tftypes.NewValue(tftypes.String, "space")
				priorValues["timeouts"] = tftypes.NewValue(wireType.AttributeTypes["timeouts"], nil)
				plannedValues := maps.Clone(priorValues)
				plannedValues[test.attribute] = test.value

				priorRaw := tftypes.NewValue(wireType, priorValues)
				if create {
					priorRaw = tftypes.NewValue(wireType, nil)
					plannedValues["id"] = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
					plannedValues["role_id"] = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
				}

				prior, err := tfprotov6.NewDynamicValue(wireType, priorRaw)
				require.NoError(t, err)
				planned, err := tfprotov6.NewDynamicValue(wireType, tftypes.NewValue(wireType, plannedValues))
				require.NoError(t, err)

				configValues := maps.Clone(plannedValues)
				configValues["id"] = tftypes.NewValue(tftypes.String, nil)
				configValues["role_id"] = tftypes.NewValue(tftypes.String, nil)
				config, err := tfprotov6.NewDynamicValue(wireType, tftypes.NewValue(wireType, configValues))
				require.NoError(t, err)

				response, err := providerServer.ApplyResourceChange(t.Context(), &tfprotov6.ApplyResourceChangeRequest{
					TypeName: "contentful_role", PriorState: &prior, PlannedState: &planned,
					Config: &config, PlannedPrivate: privateVersionBytes(t, 17),
				})
				require.NoError(t, err)
				require.Len(t, response.Diagnostics, 1)
				assert.Equal(t, tfprotov6.DiagnosticSeverityError, response.Diagnostics[0].Severity)
				assert.Equal(t, test.path, response.Diagnostics[0].Attribute)
				assert.Zero(t, requests.Load(), "unresolved containers must never reach Contentful")
			})
		}
	}
}
