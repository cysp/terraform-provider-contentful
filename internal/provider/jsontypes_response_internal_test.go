package provider

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponseJSONNormalization(t *testing.T) {
	t.Parallel()

	for _, projector := range []struct {
		name    string
		path    path.Path
		project func([]byte) (jsontypes.Normalized, diag.Diagnostics)
	}{
		{
			name: "app installation",
			path: path.Root("parameters"),
			project: func(raw []byte) (jsontypes.Normalized, diag.Diagnostics) {
				model, diags := NewAppInstallationResourceModelFromResponse(cm.AppInstallation{Parameters: raw}, AppInstallationModel{Marketplace: types.SetNull(types.StringType)})

				return model.Parameters, diags
			},
		},
		{
			name: "extension",
			path: path.Root("parameters"),
			project: func(raw []byte) (jsontypes.Normalized, diag.Diagnostics) {
				model, diags := NewExtensionModelFromResponse(t.Context(), cm.Extension{Parameters: raw})

				return model.Parameters, diags
			},
		},
		{
			name: "role policy",
			path: path.Root("policies").AtListIndex(2).AtName("constraint"),
			project: func(raw []byte) (jsontypes.Normalized, diag.Diagnostics) {
				model, diags := NewPoliciesValueFromResponse(t.Context(), path.Root("policies").AtListIndex(2), cm.RolePoliciesItem{
					Effect: cm.RolePoliciesItemEffectAllow, Actions: cm.NewStringRolePoliciesItemActions("read"), Constraint: raw,
				})

				return model.Value().Constraint, diags
			},
		},
		{
			name: "editor control",
			path: path.Root("controls").AtListIndex(2).AtName("settings"),
			project: func(raw []byte) (jsontypes.Normalized, diag.Diagnostics) {
				model, diags := NewEditorInterfaceControlValueFromResponse(path.Root("controls").AtListIndex(2), cm.EditorInterfaceControlsItem{Settings: raw})

				return model.Value().Settings, diags
			},
		},
		{
			name: "editor group control",
			path: path.Root("group_controls").AtListIndex(2).AtName("settings"),
			project: func(raw []byte) (jsontypes.Normalized, diag.Diagnostics) {
				model, diags := NewEditorInterfaceGroupControlValueFromResponse(path.Root("group_controls").AtListIndex(2), cm.EditorInterfaceGroupControlsItem{Settings: raw})

				return model.Value().Settings, diags
			},
		},
		{
			name: "editor sidebar",
			path: path.Root("sidebar").AtListIndex(2).AtName("settings"),
			project: func(raw []byte) (jsontypes.Normalized, diag.Diagnostics) {
				model, diags := NewEditorInterfaceSidebarValueFromResponse(path.Root("sidebar").AtListIndex(2), cm.EditorInterfaceSidebarItem{Settings: raw})

				return model.Value().Settings, diags
			},
		},
	} {
		t.Run(projector.name, func(t *testing.T) {
			t.Parallel()

			for _, test := range []struct {
				name     string
				raw      []byte
				expected string
				isNull   bool
				hasError bool
			}{
				{name: "omitted", isNull: true},
				{name: "JSON null", raw: []byte("null"), expected: "null"},
				{name: "empty object", raw: []byte("{}"), expected: "{}"},
				{name: "number precision and escaping", raw: []byte(`{"z":[9007199254740993,0.123456789012345678901,-0,1.2300e+100],"a<b":"<&>"}`), expected: `{"a\u003cb":"\u003c\u0026\u003e","z":[9007199254740993,0.123456789012345678901,-0,1.2300e+100]}`},
				{name: "invalid UTF-8 key collision", raw: []byte("{\"\xff\":\"invalid\",\"�\":\"replacement\"}"), expected: `{"�":"replacement"}`},
				{name: "unpaired surrogate before valid pair", raw: []byte(`{"value":"\ud800\ud800\udc00"}`), expected: `{"value":"�𐀀"}`},
				{name: "empty input", raw: []byte{}, hasError: true},
				{name: "invalid JSON", raw: []byte(`{"key":`), hasError: true},
				{name: "trailing JSON", raw: []byte(`true false`), hasError: true},
				{name: "trailing garbage", raw: []byte(`true trailing`), hasError: true},
			} {
				t.Run(test.name, func(t *testing.T) {
					t.Parallel()

					value, diags := projector.project(test.raw)
					if test.hasError {
						require.Len(t, diags, 1)
						require.True(t, diags.HasError(), diags)
						assertDiagnosticPath(t, diags[0], projector.path)
						assert.True(t, value.IsNull())
						assert.False(t, value.IsUnknown())

						return
					}

					require.Empty(t, diags)
					assert.Equal(t, test.isNull, value.IsNull())
					assert.False(t, value.IsUnknown())
					assert.Equal(t, test.expected, value.ValueString())
				})
			}
		})
	}
}
