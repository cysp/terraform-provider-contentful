package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppInstallationStateV0Parameters(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name     string
		legacy   any
		cause    string
		expected map[string]string
	}{
		{name: "absent"},
		{name: "empty", legacy: "{}", expected: map[string]string{}},
		{name: "member null and numeric tokens", legacy: `{"n":9007199254740993123456789,"v":null,"obj":{"b":1.2300e+21,"a":-0}}`, expected: map[string]string{"n": "9007199254740993123456789", "v": "null", "obj": `{"a":-0,"b":1.2300e+21}`}},
		{name: "escaped nested Unicode", legacy: `{"field":{"e\u0301":"e\u0301🌸"}}`, expected: map[string]string{"field": `{"e\u0301":"e\u0301\ud83c\udf38"}`}},
		{name: "literal nested Unicode", legacy: "{\"field\":{\"e\u0301\":\"e\u0301🌸\"}}", expected: map[string]string{"field": `{"e\u0301":"e\u0301\ud83c\udf38"}`}},
		{name: "NFC strings", legacy: `{"é":"é🌸"}`, expected: map[string]string{"é": `"é🌸"`}},
		{name: "JSON null", legacy: " null "},
		{name: "scalar", legacy: "false", cause: "Expected null or a JSON object keyed by parameter ID."},
		{name: "array", legacy: "[42]", cause: "Expected null or a JSON object keyed by parameter ID."},
		{name: "invalid document", legacy: `{} private-sentinel`, cause: "Expected null or a JSON object keyed by parameter ID."},
		{name: "single non-NFC key", legacy: `{"e\u0301":"private-sentinel"}`, cause: "Terraform would change a parameter ID during Unicode normalization."},
		{name: "colliding keys", legacy: `{"é":1,"e\u0301":2}`, cause: "Terraform would change a parameter ID during Unicode normalization."},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			upgrader := (&appInstallationResource{}).UpgradeState(t.Context())[0]
			prior := map[string]any{"id": "space/master/app", "space_id": "space", "environment_id": "master", "app_definition_id": "app", "marketplace": []string{"terms"}, "timeouts": map[string]any{"update": "3m"}, "parameters": test.legacy}
			raw, err := tftypes.ValueFromJSONWithOpts([]byte(testJSON(prior)), upgrader.PriorSchema.Type().TerraformType(t.Context()), tftypes.ValueFromJSONOpts{})
			require.NoError(t, err)
			response := resource.UpgradeStateResponse{State: tfsdk.State{Schema: AppInstallationResourceSchema(t.Context())}}
			upgrader.StateUpgrader(t.Context(), resource.UpgradeStateRequest{State: &tfsdk.State{Schema: *upgrader.PriorSchema, Raw: raw}}, &response)

			if test.cause != "" {
				require.Len(t, response.Diagnostics.Errors(), 1)
				assertDiagnosticPath(t, response.Diagnostics[0], path.Root("parameters"))
				assert.Contains(t, response.Diagnostics[0].Detail(), test.cause)
				assert.Contains(t, response.Diagnostics[0].Detail(), "refresh-only")
				assert.NotContains(t, response.Diagnostics[0].Detail(), "private-sentinel")
				assert.True(t, response.State.Raw.IsNull())

				return
			}

			require.Empty(t, response.Diagnostics)

			prior["parameters"] = test.expected
			want, err := tftypes.ValueFromJSONWithOpts([]byte(testJSON(prior)), response.State.Schema.Type().TerraformType(t.Context()), tftypes.ValueFromJSONOpts{})
			require.NoError(t, err)
			assert.True(t, want.Equal(response.State.Raw), "upgraded state")
		})
	}
}
