package provider

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func appDefinitionParametersTestPlan(t *testing.T, parameters string) tfsdk.Plan {
	t.Helper()
	schema := AppDefinitionResourceSchema(t.Context())
	raw, err := tftypes.ValueFromJSONWithOpts([]byte(testJSON(map[string]any{"id": "org/app", "organization_id": "org", "app_definition_id": "app", "name": "App", "locations": []any{}, "parameters": json.RawMessage(parameters)})), schema.Type().TerraformType(t.Context()), tftypes.ValueFromJSONOpts{})
	require.NoError(t, err)

	return tfsdk.Plan{Schema: schema, Raw: raw}
}

func TestAppDefinitionParameterLiveShapes(t *testing.T) {
	t.Parallel()

	server, err := providerserver.NewProtocol6WithError(New("test"))()
	require.NoError(t, err)

	installation := tftypes.NewAttributePath().WithAttributeName("parameters").WithAttributeName("installation").WithElementKeyInt(0)

	instance := tftypes.NewAttributePath().WithAttributeName("parameters").WithAttributeName("instance").WithElementKeyInt(0)
	for _, test := range []struct {
		name, parameters string
		attribute        *tftypes.AttributePath
		detail           string
	}{
		{"absent", testJSON(nil), nil, ""},
		{"instance Secret", testJSON(map[string]any{"instance": []any{map[string]any{"id": "x", "name": "X", "type": "Secret"}}}), nil, ""},
		{"Number string", testJSON(map[string]any{"installation": []any{map[string]any{"id": "x", "name": "X", "type": "Number", "default": "\"42\""}}}), nil, ""},
		{"Symbol Boolean", testJSON(map[string]any{"installation": []any{map[string]any{"id": "x", "name": "X", "type": "Symbol", "default": "false"}}}), nil, ""},
		{"Boolean number", testJSON(map[string]any{"installation": []any{map[string]any{"id": "x", "name": "X", "type": "Boolean", "default": "9007199254740993"}}}), nil, ""},
		{"Enum no options", testJSON(map[string]any{"installation": []any{map[string]any{"id": "x", "name": "X", "type": "Enum", "default": "\"dark\""}}}), nil, ""},
		{"nonmember and mixed options", testJSON(map[string]any{"installation": []any{map[string]any{"id": "x", "name": "X", "type": "Enum", "default": "\"dark\"", "options": []any{"\"light\"", "{\"other\":1}", "\"light\""}}}}), nil, ""},
		{"nonEnum options", testJSON(map[string]any{"instance": []any{map[string]any{"id": "x", "name": "X", "type": "Symbol", "options": []any{"{\"light\":null}"}}}}), nil, ""},
		{"object default", testJSON(map[string]any{"installation": []any{map[string]any{"id": "x", "name": "X", "type": "Number", "default": "{}"}}}), installation.WithAttributeName("default"), "A parameter default must be a JSON string, number, or Boolean. Omit default to leave it unspecified."},
		{"null default", testJSON(map[string]any{"instance": []any{map[string]any{"id": "x", "name": "X", "type": "Number", "default": "null"}}}), instance.WithAttributeName("default"), "A parameter default must be a JSON string, number, or Boolean. Omit default to leave it unspecified."},
		{"empty options", testJSON(map[string]any{"installation": []any{map[string]any{"id": "x", "name": "X", "type": "Enum", "options": []any{}}}}), installation.WithAttributeName("options"), "When configured, options must contain at least one option. Omit options to leave them unspecified."},
		{"Boolean option", testJSON(map[string]any{"installation": []any{map[string]any{"id": "x", "name": "X", "type": "Enum", "options": []any{"false"}}}}), installation.WithAttributeName("options").WithElementKeyInt(0), "An option must be a nonempty JSON string or a JSON object with exactly one member. The member value is preserved without coercion."},
		{"empty option string", testJSON(map[string]any{"installation": []any{map[string]any{"id": "x", "name": "X", "type": "Enum", "options": []any{"\"\""}}}}), installation.WithAttributeName("options").WithElementKeyInt(0), "An option must be a nonempty JSON string or a JSON object with exactly one member. The member value is preserved without coercion."},
		{"empty object", testJSON(map[string]any{"installation": []any{map[string]any{"id": "x", "name": "X", "type": "Enum", "options": []any{"{}"}}}}), installation.WithAttributeName("options").WithElementKeyInt(0), "An option must be a nonempty JSON string or a JSON object with exactly one member. The member value is preserved without coercion."},
		{"multiple members", testJSON(map[string]any{"installation": []any{map[string]any{"id": "x", "name": "X", "type": "Enum", "options": []any{"{\"a\":1,\"b\":2}"}}}}), installation.WithAttributeName("options").WithElementKeyInt(0), "An option must be a nonempty JSON string or a JSON object with exactly one member. The member value is preserved without coercion."},
		{"duplicate keys deferred", testJSON(map[string]any{"installation": []any{map[string]any{"id": "x", "name": "X", "type": "Enum", "options": []any{"{\"a\":1,\"a\":2}"}}}}), nil, ""},
		{"future type", testJSON(map[string]any{"installation": []any{map[string]any{"id": "x", "name": "X", "type": "FutureType", "default": "{}", "options": []any{}}}}), nil, ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			plan := appDefinitionParametersTestPlan(t, test.parameters)
			require.Empty(t, plan.SetAttribute(t.Context(), path.Root("id"), types.StringNull()))
			require.Empty(t, plan.SetAttribute(t.Context(), path.Root("app_definition_id"), types.StringNull()))
			config, err := tfprotov6.NewDynamicValue(plan.Schema.Type().TerraformType(t.Context()), plan.Raw)
			require.NoError(t, err)
			response, err := server.ValidateResourceConfig(t.Context(), &tfprotov6.ValidateResourceConfigRequest{TypeName: "contentful_app_definition", Config: &config})
			require.NoError(t, err)

			if test.attribute == nil {
				require.Empty(t, response.Diagnostics)

				return
			}

			require.Len(t, response.Diagnostics, 1)
			assert.Equal(t, tfprotov6.DiagnosticSeverityError, response.Diagnostics[0].Severity)
			assert.Equal(t, "Invalid App Definition parameter configuration", response.Diagnostics[0].Summary)

			assert.Equal(t, test.attribute, response.Diagnostics[0].Attribute)
			assert.Equal(t, test.detail, response.Diagnostics[0].Detail)
		})
	}
}

func TestAppDefinitionParametersUnknownAndCount(t *testing.T) {
	t.Parallel()
	plan := appDefinitionParametersTestPlan(t, testJSON(map[string]any{"installation": []any{map[string]any{"id": "x", "name": "X", "type": "Number", "default": "{}"}, map[string]any{"id": "y", "name": "Y", "type": "Enum"}}}))
	unknown := path.Root("parameters").AtName("installation").AtListIndex(1).AtName("type")
	require.Empty(t, plan.SetAttribute(t.Context(), unknown, types.StringUnknown()))

	var value types.Object
	require.Empty(t, plan.GetAttribute(t.Context(), path.Root("parameters"), &value))
	diags := validateAppDefinitionParameters(value, path.Root("parameters"))
	require.Len(t, diags, 1)
	assert.Equal(t, "parameters.installation[0].default", appDefinitionParameterDiagnosticPath(t, diags[0]).String())
	diags = validateAppDefinitionPlannedParameters(t.Context(), plan)
	require.Len(t, diags, 1)
	assert.Equal(t, unknown, appDefinitionParameterDiagnosticPath(t, diags[0]))

	for _, count := range []int{8, 9} {
		entries := make([]map[string]string, count)
		for i := range entries {
			entries[i] = map[string]string{"id": "x", "name": "X", "type": "Symbol"}
		}

		raw, err := json.Marshal(map[string]any{"instance": entries})
		require.NoError(t, err)
		plan = appDefinitionParametersTestPlan(t, string(raw))
		require.Empty(t, plan.GetAttribute(t.Context(), path.Root("parameters"), &value))

		diags = validateAppDefinitionParameters(value, path.Root("parameters"))
		if count == 8 {
			assert.Empty(t, diags)
		} else {
			require.Len(t, diags, 1)
			assert.Equal(t, "parameters.instance", appDefinitionParameterDiagnosticPath(t, diags[0]).String())
		}
	}
}

func TestAppDefinitionParametersMutationUsesPlan(t *testing.T) {
	t.Parallel()

	for _, method := range []string{http.MethodPost, http.MethodPut} {
		for _, test := range []struct {
			name, parameters, body string
			accepted               bool
		}{
			{"invalid", testJSON(map[string]any{"installation": []any{map[string]any{"id": "x", "name": "X", "type": "Number", "default": "{}"}}}), "", false},
			{"preserved", testJSON(map[string]any{"installation": []any{map[string]any{"id": "x", "name": "X", "type": "Number", "default": "\"42\"", "options": []any{"{\"light\":false}"}}}}), testJSON(map[string]any{"name": "App", "locations": []any{}, "parameters": map[string]any{"installation": []any{map[string]any{"id": "x", "name": "X", "type": "Number", "default": "42", "options": []any{map[string]any{"light": false}}}}}}), true},
		} {
			t.Run(method+test.name, func(t *testing.T) {
				t.Parallel()
				implementation, calls := appDefinitionLocationsTestResource(t, method, test.body, testJSON(map[string]any{
					"sys":  map[string]any{"type": "AppDefinition", "id": "app", "organization": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Organization", "id": "org"}}},
					"name": "App", "locations": []any{}, "parameters": map[string]any{"installation": []any{map[string]any{"id": "x", "name": "X", "type": "Number", "default": "42", "options": []any{map[string]any{"light": false}}}}},
				}))
				plan := appDefinitionParametersTestPlan(t, test.parameters)
				config := tfsdk.Config(appDefinitionParametersTestPlan(t, testJSON(nil)))
				state := tfsdk.State(appDefinitionParametersTestPlan(t, testJSON(nil)))
				identitySchema := resourceIdentitySchema(appDefinitionIdentityAttributeNames())
				identity := &tfsdk.ResourceIdentity{Schema: identitySchema, Raw: tftypes.NewValue(identitySchema.Type().TerraformType(t.Context()), nil)}

				var diags diag.Diagnostics

				if method == http.MethodPost {
					response := resource.CreateResponse{State: tfsdk.State{Schema: plan.Schema}, Identity: identity}
					implementation.Create(t.Context(), resource.CreateRequest{Plan: plan, Config: config}, &response)
					diags = response.Diagnostics
				} else {
					response := resource.UpdateResponse{State: state, Identity: identity}
					implementation.Update(t.Context(), resource.UpdateRequest{Plan: plan, Config: config, State: state}, &response)
					diags = response.Diagnostics
				}

				if test.accepted {
					require.Empty(t, diags)
					assert.EqualValues(t, 1, calls.Load())
				} else {
					require.True(t, diags.HasError())
					assert.Zero(t, calls.Load())
				}
			})
		}
	}
}

func appDefinitionParameterDiagnosticPath(t *testing.T, diagnostic diag.Diagnostic) path.Path {
	t.Helper()

	withPath, ok := diagnostic.(diag.DiagnosticWithPath)
	require.True(t, ok)

	return withPath.Path()
}

func TestAppDefinitionParameterJSONBoundaries(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{"9007199254740993", "1.234567890123456789", "1e100", "-0"} {
		decoded, valid := appDefinitionParameterJSON(jsontypes.NewNormalizedValue(raw))
		require.True(t, valid)
		require.Equal(t, json.Number(raw), decoded)
	}

	for _, raw := range []string{"{", "1 2"} {
		_, valid := appDefinitionParameterJSON(jsontypes.NewNormalizedValue(raw))
		require.False(t, valid)

		parameter := map[string]any{"installation": []map[string]string{{"id": "x", "name": "X", "type": "Number", "default": raw}}}
		encoded, err := json.Marshal(parameter)
		require.NoError(t, err)
		plan := appDefinitionParametersTestPlan(t, string(encoded))
		require.Empty(t, plan.SetAttribute(t.Context(), path.Root("id"), types.StringNull()))
		require.Empty(t, plan.SetAttribute(t.Context(), path.Root("app_definition_id"), types.StringNull()))
		config, err := tfprotov6.NewDynamicValue(plan.Schema.Type().TerraformType(t.Context()), plan.Raw)
		require.NoError(t, err)
		server, err := providerserver.NewProtocol6WithError(New("test"))()
		require.NoError(t, err)
		response, err := server.ValidateResourceConfig(t.Context(), &tfprotov6.ValidateResourceConfigRequest{TypeName: "contentful_app_definition", Config: &config})
		require.NoError(t, err)
		require.Len(t, response.Diagnostics, 1)
		require.Equal(t, "Invalid JSON String Value", response.Diagnostics[0].Summary)
		diags := validateAppDefinitionPlannedParameters(t.Context(), plan)
		require.Len(t, diags, 1)
		require.Equal(t, "parameters.installation[0].default", appDefinitionParameterDiagnosticPath(t, diags[0]).String())
		require.Equal(t, "This value must contain exactly one valid JSON value.", diags[0].Detail())
	}
}

func TestAppDefinitionParameterOptionsPreflight(t *testing.T) {
	t.Parallel()

	optionsPath := path.Root("parameters").AtName("installation").AtListIndex(0).AtName("options")
	for _, method := range []string{http.MethodPost, http.MethodPut} {
		for _, test := range []struct {
			name            string
			options         TypedList[jsontypes.Normalized]
			attribute       path.Path
			summary, detail string
		}{
			{name: "preserved object value", options: NewTypedList([]jsontypes.Normalized{jsontypes.NewNormalizedValue(`{"choice":false}`)})},
			{name: "unknown list", options: NewTypedListUnknown[jsontypes.Normalized](), attribute: optionsPath, summary: "Unknown App Definition parameter value", detail: "This value must be known before sending the App Definition to Contentful."},
			{name: "unknown element", options: NewTypedList([]jsontypes.Normalized{jsontypes.NewNormalizedUnknown()}), attribute: optionsPath.AtListIndex(0), summary: "Unknown App Definition parameter value", detail: "This value must be known before sending the App Definition to Contentful."},
			{name: "null element", options: NewTypedList([]jsontypes.Normalized{jsontypes.NewNormalizedNull()}), attribute: optionsPath.AtListIndex(0), summary: "Invalid App Definition parameter configuration", detail: "List elements must not be null."},
			{name: "malformed element", options: NewTypedList([]jsontypes.Normalized{jsontypes.NewNormalizedValue(`{"VALUE_DO_NOT_ECHO":`)}), attribute: optionsPath.AtListIndex(0), summary: "Invalid App Definition parameter configuration", detail: "This value must contain exactly one valid JSON value."},
			{name: "trailing JSON", options: NewTypedList([]jsontypes.Normalized{jsontypes.NewNormalizedValue(`{"VALUE_DO_NOT_ECHO":true} false`)}), attribute: optionsPath.AtListIndex(0), summary: "Invalid App Definition parameter configuration", detail: "This value must contain exactly one valid JSON value."},
			{name: "invalid shape", options: NewTypedList([]jsontypes.Normalized{jsontypes.NewNormalizedValue(`false`)}), attribute: optionsPath.AtListIndex(0), summary: "Invalid App Definition parameter configuration", detail: "An option must be a nonempty JSON string or a JSON object with exactly one member. The member value is preserved without coercion."},
		} {
			t.Run(method+"/"+test.name, func(t *testing.T) {
				t.Parallel()
				implementation, calls := appDefinitionLocationsTestResource(t, method,
					testJSON(map[string]any{"name": "App", "locations": []any{}, "parameters": map[string]any{"installation": []any{map[string]any{"id": "x", "name": "X", "type": "Enum", "options": []any{map[string]any{"choice": false}}}}}}),
					testJSON(map[string]any{"sys": map[string]any{"type": "AppDefinition", "id": "app", "organization": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Organization", "id": "org"}}}, "name": "App", "locations": []any{}, "parameters": map[string]any{"installation": []any{map[string]any{"id": "x", "name": "X", "type": "Enum", "options": []any{map[string]any{"choice": false}}}}}}))
				plan := appDefinitionParametersTestPlan(t, testJSON(map[string]any{"installation": []any{map[string]any{"id": "x", "name": "X", "type": "Enum"}}}))
				config := tfsdk.Config(plan)
				state := tfsdk.State(plan)
				require.Empty(t, plan.SetAttribute(t.Context(), optionsPath, test.options))

				identitySchema := resourceIdentitySchema(appDefinitionIdentityAttributeNames())
				identity := &tfsdk.ResourceIdentity{Schema: identitySchema, Raw: tftypes.NewValue(identitySchema.Type().TerraformType(t.Context()), nil)}

				var diags diag.Diagnostics

				if method == http.MethodPost {
					response := resource.CreateResponse{State: tfsdk.State{Schema: plan.Schema}, Identity: identity}
					implementation.Create(t.Context(), resource.CreateRequest{Plan: plan, Config: config}, &response)
					diags = response.Diagnostics
				} else {
					response := resource.UpdateResponse{State: state, Identity: identity}
					implementation.Update(t.Context(), resource.UpdateRequest{Plan: plan, Config: config, State: state}, &response)
					diags = response.Diagnostics
				}

				if test.summary == "" {
					require.Empty(t, diags)
					assert.EqualValues(t, 1, calls.Load())

					return
				}

				assert.Zero(t, calls.Load())
				require.Len(t, diags, 1)
				assert.Equal(t, diag.SeverityError, diags[0].Severity())
				assert.Equal(t, test.attribute, appDefinitionParameterDiagnosticPath(t, diags[0]))
				assert.Equal(t, test.summary, diags[0].Summary())
				assert.Equal(t, test.detail, diags[0].Detail())
				assert.NotContains(t, diags[0].Detail(), "VALUE_DO_NOT_ECHO")
			})
		}
	}
}
