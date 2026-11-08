package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
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

func extensionValidationPlan(t *testing.T, extension map[string]any) tfsdk.Plan {
	t.Helper()
	schema := ExtensionResourceSchema(t.Context())
	raw, err := json.Marshal(map[string]any{"space_id": "space", "environment_id": "master", "extension_id": "extension", "extension": extension})
	require.NoError(t, err)
	value, err := tftypes.ValueFromJSONWithOpts(raw, schema.Type().TerraformType(t.Context()), tftypes.ValueFromJSONOpts{})
	require.NoError(t, err)

	return tfsdk.Plan{Schema: schema, Raw: value}
}

func extensionValidationConfiguration() map[string]any {
	return map[string]any{"name": "Extension", "srcdoc": "", "sidebar": false, "field_types": []any{}}
}

func TestExtensionFieldTypeValidation(t *testing.T) {
	t.Parallel()

	server, err := providerserver.NewProtocol6WithError(New("test"))()
	require.NoError(t, err)

	for _, test := range []struct {
		name      string
		field     map[string]any
		attribute string
	}{
		{"scalar", map[string]any{"type": "Symbol"}, ""},
		{"Link", map[string]any{"type": "Link", "link_type": "Entry"}, ""},
		{"future link target", map[string]any{"type": "Link", "link_type": "Future"}, ""},
		{"future type", map[string]any{"type": "Future", "items": map[string]any{"type": "Number"}}, ""},
		{"future type with incomplete Link items", map[string]any{"type": "Future", "items": map[string]any{"type": "Link"}}, ""},
		{"Array Symbol", map[string]any{"type": "Array", "items": map[string]any{"type": "Symbol"}}, ""},
		{"Array Link", map[string]any{"type": "Array", "items": map[string]any{"type": "Link", "link_type": "Asset"}}, ""},
		{"Array future", map[string]any{"type": "Array", "items": map[string]any{"type": "Future"}}, ""},
		{"Link missing target", map[string]any{"type": "Link"}, "link_type"},
		{"Link empty target", map[string]any{"type": "Link", "link_type": ""}, "link_type"},
		{"Array missing items", map[string]any{"type": "Array"}, "items"},
		{"scalar items", map[string]any{"type": "Symbol", "items": map[string]any{"type": "Symbol"}}, "items"},
		{"scalar link", map[string]any{"type": "Number", "link_type": "Entry"}, "link_type"},
		{"Array outer link", map[string]any{"type": "Array", "link_type": "Entry", "items": map[string]any{"type": "Symbol"}}, "link_type"},
		{"Array Number", map[string]any{"type": "Array", "items": map[string]any{"type": "Number"}}, "items.type"},
		{"Array nested Array", map[string]any{"type": "Array", "items": map[string]any{"type": "Array"}}, "items.type"},
		{"Array Link missing target", map[string]any{"type": "Array", "items": map[string]any{"type": "Link"}}, "items.link_type"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			configuration := extensionValidationConfiguration()
			configuration["field_types"] = []any{test.field}
			plan := extensionValidationPlan(t, configuration)
			dynamic, err := tfprotov6.NewDynamicValue(plan.Schema.Type().TerraformType(t.Context()), plan.Raw)
			require.NoError(t, err)
			response, err := server.ValidateResourceConfig(t.Context(), &tfprotov6.ValidateResourceConfigRequest{TypeName: "contentful_extension", Config: &dynamic})
			require.NoError(t, err)

			diags := validateExtensionPlannedConfiguration(t.Context(), plan)
			if test.attribute == "" {
				require.Empty(t, response.Diagnostics)
				require.Empty(t, diags)

				return
			}

			require.Len(t, response.Diagnostics, 1)
			assert.Equal(t, "Invalid Extension field type configuration", response.Diagnostics[0].Summary)

			diagnosticPath := tftypes.NewAttributePath().WithAttributeName("extension").WithAttributeName("field_types").WithElementKeyInt(0)
			for name := range strings.SplitSeq(test.attribute, ".") {
				diagnosticPath = diagnosticPath.WithAttributeName(name)
			}

			assert.Equal(t, diagnosticPath, response.Diagnostics[0].Attribute)

			expected := "extension.field_types[0]." + test.attribute
			assert.Equal(t, []string{expected}, mutationDiagnosticPaths(t, diags))
			assert.Equal(t, diags[0].Detail(), response.Diagnostics[0].Detail)
		})
	}
}

func TestExtensionParameterValidation(t *testing.T) {
	t.Parallel()

	server, err := providerserver.NewProtocol6WithError(New("test"))()
	require.NoError(t, err)

	for _, scope := range []string{"installation", "instance"} {
		t.Run(scope, func(t *testing.T) {
			t.Parallel()

			for _, test := range []struct {
				name, kind string
				value      any
				options    any
				attribute  string
			}{
				{"omitted Symbol default", "Symbol", nil, nil, ""},
				{"empty Symbol default", "Symbol", `""`, nil, ""},
				{"Number precision", "Number", "9007199254740993", nil, ""},
				{"Boolean false", "Boolean", "false", nil, ""},
				{"Secret string", "Secret", `"secret"`, nil, ""},
				{"future type", "Future", `{"future":null}`, []any{`42`}, ""},
				{"Symbol numeric", "Symbol", "42", nil, "default"},
				{"Number string", "Number", `"42"`, nil, "default"},
				{"Boolean numeric", "Boolean", "0", nil, "default"},
				{"Secret numeric", "Secret", "42", nil, "default"},
				{"object default", "Symbol", "{}", nil, "default"},
				{"array default", "Symbol", "[]", nil, "default"},
				{"JSON null default", "Symbol", "null", nil, "default"},
				{"Enum string options", "Enum", `"a"`, []any{`"a"`, `"b"`}, ""},
				{"Enum labeled options", "Enum", `"a"`, []any{`{ "a": "Label" }`}, ""},
				{"Enum empty label", "Enum", `"a"`, []any{`{"a":""}`}, ""},
				{"Enum no default", "Enum", nil, []any{`"a"`}, ""},
				{"Enum no options", "Enum", nil, nil, "options"},
				{"Enum empty options", "Enum", nil, []any{}, "options"},
				{"Enum numeric default", "Enum", "42", []any{`"a"`}, "default"},
				{"Enum nonmember default", "Enum", `"b"`, []any{`"a"`}, "default"},
				{"Enum label default", "Enum", `"Label"`, []any{`{"a":"Label"}`}, "default"},
				{"Enum mixed options", "Enum", nil, []any{`"a"`, `{"b":"B"}`}, "options[1]"},
				{"Enum numeric label", "Enum", nil, []any{`{"a":42}`}, "options[0]"},
				{"Enum null label", "Enum", nil, []any{`{"a":null}`}, "options[0]"},
				{"Enum empty string", "Enum", nil, []any{`""`}, "options[0]"},
				{"Enum empty object", "Enum", nil, []any{`{}`}, "options[0]"},
				{"Enum multiple labels", "Enum", nil, []any{`{"a":"A","b":"B"}`}, "options[0]"},
				{"Enum duplicate keys with distinct members", "Enum", nil, []any{`{"a":"A","b":"B","a":"Z"}`}, "options[0]"},
				{"Enum array option", "Enum", nil, []any{`[]`}, "options[0]"},
				{"Enum Boolean option", "Enum", nil, []any{`true`}, "options[0]"},
				{"Symbol options", "Symbol", nil, []any{`"a"`}, "options"},
				{"Number options", "Number", nil, []any{`"a"`}, "options"},
				{"Boolean options", "Boolean", nil, []any{`"a"`}, "options"},
				{"Secret options", "Secret", nil, []any{`"a"`}, "options"},
			} {
				t.Run(test.name, func(t *testing.T) {
					t.Parallel()

					configuration := extensionValidationConfiguration()
					configuration["parameters"] = map[string]any{scope: []any{map[string]any{"id": "value", "name": "Value", "type": test.kind, "default": test.value, "options": test.options}}}
					plan := extensionValidationPlan(t, configuration)
					dynamic, err := tfprotov6.NewDynamicValue(plan.Schema.Type().TerraformType(t.Context()), plan.Raw)
					require.NoError(t, err)
					response, err := server.ValidateResourceConfig(t.Context(), &tfprotov6.ValidateResourceConfigRequest{TypeName: "contentful_extension", Config: &dynamic})
					require.NoError(t, err)

					diags := validateExtensionPlannedConfiguration(t.Context(), plan)
					if test.attribute != "" {
						require.Len(t, response.Diagnostics, 1)
						assert.Equal(t, "Invalid Extension parameter configuration", response.Diagnostics[0].Summary)

						diagnosticPath := tftypes.NewAttributePath().WithAttributeName("extension").WithAttributeName("parameters").WithAttributeName(scope).WithElementKeyInt(0)

						switch test.attribute {
						case "options[0]":
							diagnosticPath = diagnosticPath.WithAttributeName("options").WithElementKeyInt(0)
						case "options[1]":
							diagnosticPath = diagnosticPath.WithAttributeName("options").WithElementKeyInt(1)
						default:
							diagnosticPath = diagnosticPath.WithAttributeName(test.attribute)
						}

						assert.Equal(t, diagnosticPath, response.Diagnostics[0].Attribute)
						assert.Equal(t, []string{"extension.parameters." + scope + "[0]." + test.attribute}, mutationDiagnosticPaths(t, diags))
						assert.Equal(t, diags[0].Detail(), response.Diagnostics[0].Detail)

						return
					}

					require.Empty(t, response.Diagnostics)
					require.Empty(t, diags)
					assertExtensionParameterRequestValues(t, plan, scope, test.value, test.options)
				})
			}
		})
	}
}

func TestExtensionValidationUnknownSiblings(t *testing.T) {
	t.Parallel()

	configuration := extensionValidationConfiguration()
	configuration["field_types"] = []any{map[string]any{"type": "Link"}}
	configuration["parameters"] = map[string]any{"installation": []any{map[string]any{"id": "value", "name": "Value", "type": "Enum", "default": `"missing"`, "options": []any{`"a"`, `false`}}}}
	plan := extensionValidationPlan(t, configuration)
	require.Empty(t, plan.SetAttribute(t.Context(), path.Root("extension").AtName("parameters").AtName("installation").AtListIndex(0).AtName("options").AtListIndex(0), jsontypes.NewNormalizedUnknown()))
	require.Empty(t, plan.SetAttribute(t.Context(), path.Root("extension").AtName("name"), types.StringUnknown()))

	var extension types.Object
	require.Empty(t, plan.GetAttribute(t.Context(), path.Root("extension"), &extension))
	diags := validateExtensionConfiguration(extension, path.Root("extension"))
	assert.Equal(t, []string{"extension.field_types[0].link_type", "extension.parameters.installation[0].options[1]"}, mutationDiagnosticPaths(t, diags))
	// An unresolved option can still become the default, so membership must defer.
	require.Empty(t, plan.SetAttribute(t.Context(), path.Root("extension").AtName("field_types").AtListIndex(0).AtName("link_type"), types.StringUnknown()))
	require.Empty(t, plan.SetAttribute(t.Context(), path.Root("extension").AtName("parameters").AtName("installation").AtListIndex(0).AtName("options").AtListIndex(1), jsontypes.NewNormalizedValue(`"b"`)))
	require.Empty(t, plan.GetAttribute(t.Context(), path.Root("extension"), &extension))
	require.Empty(t, validateExtensionConfiguration(extension, path.Root("extension")))

	server, err := providerserver.NewProtocol6WithError(New("test"))()
	require.NoError(t, err)
	dynamic, err := tfprotov6.NewDynamicValue(plan.Schema.Type().TerraformType(t.Context()), plan.Raw)
	require.NoError(t, err)
	response, err := server.ValidateResourceConfig(t.Context(), &tfprotov6.ValidateResourceConfigRequest{TypeName: "contentful_extension", Config: &dynamic})
	require.NoError(t, err)
	require.Empty(t, response.Diagnostics)
}

func TestExtensionMutationsRejectInvalidConfigurationPlan(t *testing.T) {
	t.Parallel()

	var calls atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { calls.Add(1) }))
	t.Cleanup(server.Close)
	client, err := cm.NewClient(server.URL, cm.NewAccessTokenSecuritySource("test-token"))
	require.NoError(t, err)

	implementation := &extensionResource{providerData: ContentfulProviderData{client: client}}
	valid := extensionValidationPlan(t, extensionValidationConfiguration())
	configuration := extensionValidationConfiguration()
	configuration["parameters"] = map[string]any{"installation": []any{map[string]any{"id": "value", "name": "Value", "type": "Number", "default": `"42"`}}}
	invalid := extensionValidationPlan(t, configuration)
	state := tfsdk.State(valid)
	create := resource.CreateResponse{State: tfsdk.State{Schema: invalid.Schema}}
	implementation.Create(t.Context(), resource.CreateRequest{Plan: invalid, Config: tfsdk.Config(valid)}, &create)
	assert.Equal(t, []string{"extension.parameters.installation[0].default"}, mutationDiagnosticPaths(t, create.Diagnostics))

	update := resource.UpdateResponse{State: state}
	implementation.Update(t.Context(), resource.UpdateRequest{Plan: invalid, Config: tfsdk.Config(valid), State: state}, &update)
	assert.Equal(t, []string{"extension.parameters.installation[0].default"}, mutationDiagnosticPaths(t, update.Diagnostics))
	assert.True(t, update.State.Raw.Equal(state.Raw))
	assert.Zero(t, calls.Load())
}

func TestExtensionPlannedParametersRejectMalformedJSON(t *testing.T) {
	t.Parallel()

	for _, attribute := range []string{"default", "options"} {
		for _, raw := range []string{"", "{", "{} {}", "invalid json"} {
			t.Run(attribute+raw, func(t *testing.T) {
				t.Parallel()

				parameter := map[string]any{"id": "value", "name": "Value", "type": "Symbol"}
				expected := "extension.parameters.installation[0].default"

				if attribute == "default" {
					parameter[attribute] = raw
				} else {
					parameter["type"] = "Enum"
					parameter[attribute] = []any{raw}
					expected = "extension.parameters.installation[0].options[0]"
				}

				configuration := extensionValidationConfiguration()
				configuration["parameters"] = map[string]any{"installation": []any{parameter}}
				plan := extensionValidationPlan(t, configuration)
				diags := validateExtensionPlannedConfiguration(t.Context(), plan)
				assert.Equal(t, []string{expected}, mutationDiagnosticPaths(t, diags))
			})
		}
	}
}

func assertExtensionParameterRequestValues(t *testing.T, plan tfsdk.Plan, scope string, defaultValue, optionValues any) {
	t.Helper()

	var model ExtensionModel
	require.Empty(t, plan.Get(t.Context(), &model))
	converted, diags := model.ToExtensionData(model, path.Empty())
	require.Empty(t, diags)

	parameters := converted.Extension.Parameters.Value.Installation
	if scope == "instance" {
		parameters = converted.Extension.Parameters.Value.Instance
	}

	require.Len(t, parameters, 1)

	if defaultValue != nil {
		assert.Equal(t, defaultValue, string(parameters[0].Default))
	}

	if options, ok := optionValues.([]any); ok {
		require.Len(t, parameters[0].Options, len(options))

		for index, option := range options {
			assert.Equal(t, option, string(parameters[0].Options[index]))
		}
	}
}
