package provider

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func editorInterfaceSettingsValidationPlan(t *testing.T, attribute string, settings any) tfsdk.Plan {
	t.Helper()

	item := map[string]any{"widget_namespace": "builtin", "widget_id": "widget", "settings": settings}

	switch attribute {
	case "controls":
		item["field_id"] = "field"
	case "group_controls":
		item["group_id"] = "group"
	case "sidebar":
		item["disabled"] = false
	}

	return editorInterfaceValidationPlan(t, map[string]any{attribute: []any{item}})
}

func editorInterfaceValidationPlan(t *testing.T, values map[string]any) tfsdk.Plan {
	t.Helper()
	schema := EditorInterfaceResourceSchema(t.Context())
	values["space_id"] = "space"
	values["environment_id"] = "master"
	values["content_type_id"] = "type"
	raw, err := json.Marshal(values)
	require.NoError(t, err)
	value, err := tftypes.ValueFromJSONWithOpts(raw, schema.Type().TerraformType(t.Context()), tftypes.ValueFromJSONOpts{})
	require.NoError(t, err)

	return tfsdk.Plan{Schema: schema, Raw: value}
}

func TestEditorInterfaceSettingsValidation(t *testing.T) {
	t.Parallel()

	server, err := providerserver.NewProtocol6WithError(New("test"))()
	require.NoError(t, err)

	for _, attribute := range []string{"controls", "group_controls", "sidebar"} {
		t.Run(attribute, func(t *testing.T) {
			t.Parallel()

			for _, test := range []struct {
				name    string
				raw     any
				invalid bool
			}{
				{"omitted", nil, false},
				{"empty object", "{}", false},
				{"opaque object", `{ "future": {"value":9007199254740993},"empty":null,"enabled":false }`, false},
				{"JSON null", "null", true},
				{"array", "[]", true},
				{"string", `"value"`, true},
				{"number", "42", true},
				{"Boolean", "false", true},
			} {
				t.Run(test.name, func(t *testing.T) {
					t.Parallel()
					plan := editorInterfaceSettingsValidationPlan(t, attribute, test.raw)
					dynamic, err := tfprotov6.NewDynamicValue(plan.Schema.Type().TerraformType(t.Context()), plan.Raw)
					require.NoError(t, err)
					response, err := server.ValidateResourceConfig(t.Context(), &tfprotov6.ValidateResourceConfigRequest{TypeName: "contentful_editor_interface", Config: &dynamic})
					require.NoError(t, err)

					var model EditorInterfaceModel
					require.Empty(t, plan.Get(t.Context(), &model))

					converted, diags := model.ToEditorInterfaceData(t.Context())
					if test.invalid {
						require.Len(t, response.Diagnostics, 1)
						assert.Equal(t, tftypes.NewAttributePath().WithAttributeName(attribute).WithElementKeyInt(0).WithAttributeName("settings"), response.Diagnostics[0].Attribute)
						assert.Equal(t, "Invalid Editor Interface settings", response.Diagnostics[0].Summary)
						require.Equal(t, []string{attribute + "[0].settings"}, mutationDiagnosticPaths(t, diags))
					} else {
						require.Empty(t, response.Diagnostics)
						require.Empty(t, diags)

						var raw []byte

						switch attribute {
						case "controls":
							raw = converted.Controls.Value[0].Settings
						case "group_controls":
							raw = converted.GroupControls.Value[0].Settings
						case "sidebar":
							raw = converted.Sidebar.Value[0].Settings
						}

						if test.raw == nil {
							assert.Nil(t, raw)
						} else {
							assert.Equal(t, test.raw, string(raw))
						}
					}
				})
			}

			plan := editorInterfaceSettingsValidationPlan(t, attribute, "{}")
			require.Empty(t, plan.SetAttribute(t.Context(), path.Root(attribute).AtListIndex(0).AtName("settings"), types.StringUnknown()))
			dynamic, err := tfprotov6.NewDynamicValue(plan.Schema.Type().TerraformType(t.Context()), plan.Raw)
			require.NoError(t, err)
			response, err := server.ValidateResourceConfig(t.Context(), &tfprotov6.ValidateResourceConfigRequest{TypeName: "contentful_editor_interface", Config: &dynamic})
			require.NoError(t, err)
			require.Empty(t, response.Diagnostics)
		})
	}
}

func TestEditorInterfaceLayoutValidation(t *testing.T) {
	t.Parallel()

	server, err := providerserver.NewProtocol6WithError(New("test"))()
	require.NoError(t, err)

	for _, test := range []struct{ name, topID, nestedID, targetID, widget, namespace, wantPath string }{
		{"valid top tab", "top", "nested", "top", "topLevelTab", "builtin", ""},
		{"valid nested fieldset", "top", "nested", "nested", "fieldset", "builtin", ""},
		{"duplicate group", "top", "top", "top", "topLevelTab", "builtin", "editor_layout[0].group.items[0].group.group_id"},
		{"nested tab", "top", "nested", "nested", "topLevelTab", "builtin", "group_controls[0].widget_id"},
		{"top fieldset", "top", "nested", "top", "fieldset", "builtin", "group_controls[0].widget_id"},
		{"unknown target deferred", "top", "nested", "remote", "topLevelTab", "builtin", ""},
		{"future widget deferred", "top", "nested", "nested", "futureWidget", "builtin", ""},
		{"custom namespace deferred", "top", "nested", "nested", "topLevelTab", "app", ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			plan := editorInterfaceValidationPlan(t, map[string]any{
				"editor_layout":  []any{map[string]any{"group": map[string]any{"group_id": test.topID, "name": "Top", "items": []any{map[string]any{"group": map[string]any{"group_id": test.nestedID, "name": "Nested", "items": []any{}}}}}}},
				"group_controls": []any{map[string]any{"group_id": test.targetID, "widget_namespace": test.namespace, "widget_id": test.widget}},
			})
			dynamic, err := tfprotov6.NewDynamicValue(plan.Schema.Type().TerraformType(t.Context()), plan.Raw)
			require.NoError(t, err)
			response, err := server.ValidateResourceConfig(t.Context(), &tfprotov6.ValidateResourceConfigRequest{TypeName: "contentful_editor_interface", Config: &dynamic})
			require.NoError(t, err)

			var model EditorInterfaceModel
			require.Empty(t, plan.Get(t.Context(), &model))

			_, diags := model.ToEditorInterfaceData(t.Context())
			if test.wantPath == "" {
				require.Empty(t, response.Diagnostics)
				require.Empty(t, diags)
			} else {
				require.Len(t, response.Diagnostics, 1)
				require.Equal(t, []string{test.wantPath}, mutationDiagnosticPaths(t, diags))
			}
		})
	}
}

func TestEditorInterfaceLayoutValidationDefersUnknownParticipants(t *testing.T) {
	t.Parallel()
	plan := editorInterfaceValidationPlan(t, map[string]any{})
	require.Empty(t, plan.SetAttribute(t.Context(), path.Root("editor_layout"), NewTypedListUnknown[TypedObject[EditorInterfaceEditorLayoutItemValue]]()))

	var model EditorInterfaceModel
	require.Empty(t, plan.Get(t.Context(), &model))
	require.Empty(t, validateEditorInterfaceLayout(model))
}

func TestEditorInterfaceSettingsRequestRejectsMalformedJSON(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{"", "{", "{} {}", "invalid json"} {
		t.Run(raw, func(t *testing.T) {
			t.Parallel()

			valuePath := path.Root("controls").AtListIndex(0).AtName("settings")
			converted, diags := editorInterfaceOptionalSettings(jsontypes.NewNormalizedValue(raw), valuePath)
			require.Nil(t, converted)
			require.Equal(t, []string{"controls[0].settings"}, mutationDiagnosticPaths(t, diags))
		})
	}
}
