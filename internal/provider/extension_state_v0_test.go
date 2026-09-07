package provider_test

import (
	"encoding/json"
	"os"
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtensionStateV0MatchesReleasedSchema(t *testing.T) {
	t.Parallel()

	// Captured from `terraform providers schema -json` with the actual Registry
	// v0.0.62 binary, independently of the current provider's schema helpers.
	encodedType, err := os.ReadFile("testdata/extension_state_v0_type.json")
	require.NoError(t, err)

	implementation, ok := NewExtensionResource().(resource.ResourceWithUpgradeState)
	require.True(t, ok)
	priorSchema := implementation.UpgradeState(t.Context())[0].PriorSchema
	require.NotNil(t, priorSchema)
	assert.EqualValues(t, 0, priorSchema.Version)
	actualType, err := json.Marshal(priorSchema.Type().TerraformType(t.Context()))
	require.NoError(t, err)
	assert.JSONEq(t, string(encodedType), string(actualType))
}

func TestExtensionStateV0UpgradePreservesData(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name         string
		legacySrc    any
		legacySrcDoc any
		wantSrc      any
		wantSrcDoc   any
	}{
		{name: "URL", legacySrc: "https://example.com/extension.html", legacySrcDoc: "", wantSrc: "https://example.com/extension.html"},
		{name: "inline HTML", legacySrc: "", legacySrcDoc: "<title>Extension</title>", wantSrcDoc: "<title>Extension</title>"},
		{name: "empty inline HTML", legacySrc: "", legacySrcDoc: "", wantSrcDoc: ""},
		{name: "null sources"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			implementation := NewExtensionResource()
			upgrading, ok := implementation.(resource.ResourceWithUpgradeState)
			require.True(t, ok)
			upgrader := upgrading.UpgradeState(t.Context())[0]

			var schemaResponse resource.SchemaResponse
			implementation.Schema(t.Context(), resource.SchemaRequest{}, &schemaResponse)

			legacy := extensionStateV0Fixture(test.legacySrc, test.legacySrcDoc)
			want := extensionStateV0Fixture(test.wantSrc, test.wantSrcDoc)
			priorRaw := extensionStateV0Raw(t, legacy, upgrader.PriorSchema.Type().TerraformType(t.Context()))
			response := resource.UpgradeStateResponse{State: tfsdk.State{Schema: schemaResponse.Schema}}

			upgrader.StateUpgrader(t.Context(), resource.UpgradeStateRequest{
				State: &tfsdk.State{Raw: priorRaw, Schema: *upgrader.PriorSchema},
			}, &response)
			require.False(t, response.Diagnostics.HasError(), response.Diagnostics.Errors())
			expected := extensionStateV0Raw(t, want, schemaResponse.Schema.Type().TerraformType(t.Context()))
			assert.True(t, response.State.Raw.Equal(expected), "upgraded state differs outside the expected source normalization")
		})
	}
}

func TestExtensionStateV0MissingConfigurationDiagnostic(t *testing.T) {
	t.Parallel()

	implementation := NewExtensionResource()
	upgrading, ok := implementation.(resource.ResourceWithUpgradeState)
	require.True(t, ok)
	upgrader := upgrading.UpgradeState(t.Context())[0]

	var schemaResponse resource.SchemaResponse
	implementation.Schema(t.Context(), resource.SchemaRequest{}, &schemaResponse)

	legacy := extensionStateV0Fixture("", "")
	legacy["extension"] = nil
	priorRaw := extensionStateV0Raw(t, legacy, upgrader.PriorSchema.Type().TerraformType(t.Context()))
	response := resource.UpgradeStateResponse{State: tfsdk.State{Schema: schemaResponse.Schema}}

	upgrader.StateUpgrader(t.Context(), resource.UpgradeStateRequest{
		State: &tfsdk.State{Raw: priorRaw, Schema: *upgrader.PriorSchema},
	}, &response)
	require.Len(t, response.Diagnostics.Errors(), 1)
	assert.Equal(t, "Missing extension configuration", response.Diagnostics.Errors()[0].Summary())
	assert.Equal(t, "The prior state does not contain the extension configuration required for a state upgrade.", response.Diagnostics.Errors()[0].Detail())
}

func extensionStateV0Raw(t *testing.T, state map[string]any, valueType tftypes.Type) tftypes.Value {
	t.Helper()

	data, err := json.Marshal(state)
	require.NoError(t, err)
	value, err := (tfprotov6.RawState{JSON: data}).Unmarshal(valueType)
	require.NoError(t, err)

	return value
}

func extensionStateV0Fixture(src, srcdoc any) map[string]any {
	return map[string]any{
		"id": "space/environment/extension", "space_id": "space", "environment_id": "environment", "extension_id": "extension",
		"parameters": `{"theme":"dark","enabled":false}`,
		"timeouts":   map[string]any{"create": "1m", "read": "2m", "update": "3m", "delete": "4m"},
		"extension": map[string]any{
			"name": "Historical Extension", "src": src, "srcdoc": srcdoc, "sidebar": false,
			"field_types": []any{
				map[string]any{"type": "Array", "link_type": nil, "items": map[string]any{"type": "Link", "link_type": "Asset"}},
				map[string]any{"type": "Symbol", "link_type": nil, "items": nil},
			},
			"parameters": map[string]any{
				"installation": []any{
					map[string]any{
						"id": "theme", "type": "Enum", "name": "Theme", "description": "", "required": false,
						"default": `"light"`, "options": []string{`"light"`, `"dark"`}, "labels": nil,
					},
					map[string]any{
						"id": "flag", "type": "Boolean", "name": "Flag", "description": nil, "required": true,
						"default": "false", "options": nil,
						"labels": map[string]any{"empty": "Choose", "true": "On", "false": "Off"},
					},
				},
				"instance": []any{},
			},
		},
	}
}
