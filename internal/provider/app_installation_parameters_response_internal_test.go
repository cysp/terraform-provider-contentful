package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppInstallationParametersReconciliation(t *testing.T) {
	t.Parallel()

	for _, test := range []struct{ name, response, reference, expected string }{
		{"same length", testJSON(map[string]any{"secret": "*****"}), testJSON(map[string]any{"secret": "alpha"}), testJSON(map[string]any{"secret": "alpha"})},
		{"short mask", testJSON(map[string]any{"secret": "*"}), testJSON(map[string]any{"secret": "alpha"}), testJSON(map[string]any{"secret": "alpha"})},
		{"long mask", testJSON(map[string]any{"secret": "************"}), testJSON(map[string]any{"secret": "alpha"}), testJSON(map[string]any{"secret": "alpha"})},
		{"Unicode", testJSON(map[string]any{"secret": "*"}), testJSON(map[string]any{"secret": "é🌸"}), testJSON(map[string]any{"secret": "é🌸"})},
		{"ordinary asterisks are indistinguishable", testJSON(map[string]any{"ordinary": "**"}), testJSON(map[string]any{"ordinary": "readable"}), testJSON(map[string]any{"ordinary": "readable"})},
		{"readable contradictions", testJSON(map[string]any{"secret": "other", "visible": "new"}), testJSON(map[string]any{"secret": "alpha", "visible": "old"}), testJSON(map[string]any{"secret": "other", "visible": "new"})},
		{"partial response", testJSON(map[string]any{"secret": "*", "added": true}), testJSON(map[string]any{"secret": "alpha", "omitted": "old"}), testJSON(map[string]any{"added": true, "secret": "alpha"})},
		{"unusable references", testJSON(map[string]any{"absent": "*", "empty": "*", "null": "*", "number": "*", "boolean": "*", "object": "*", "array": "*", "stars": "***"}), testJSON(map[string]any{"empty": "", "null": nil, "number": 1, "boolean": true, "object": map[string]any{}, "array": []any{}, "stars": "*"}), testJSON(map[string]any{"absent": "*", "empty": "*", "null": "*", "number": "*", "boolean": "*", "object": "*", "array": "*", "stars": "***"})},
		{"nonmask responses", testJSON(map[string]any{"empty": "", "null": nil, "number": 1, "boolean": true, "object": map[string]any{}, "array": []any{}, "partial": "*x", "unicode": "＊"}), testJSON(map[string]any{"empty": "alpha", "null": "alpha", "number": "alpha", "boolean": "alpha", "object": "alpha", "array": "alpha", "partial": "alpha", "unicode": "alpha"}), testJSON(map[string]any{"empty": "", "null": nil, "number": 1, "boolean": true, "object": map[string]any{}, "array": []any{}, "partial": "*x", "unicode": "＊"})},
		{"nested values not reconciled", testJSON(map[string]any{"secret": "*", "array": []any{"*"}, "object": map[string]any{"secret": "*"}}), testJSON(map[string]any{"secret": "alpha", "array": []any{"alpha"}, "object": map[string]any{"secret": "alpha"}}), testJSON(map[string]any{"array": []any{"*"}, "object": map[string]any{"secret": "*"}, "secret": "alpha"})},
		{"number tokens and escaping", `{"z":900719925474099312345,"secret":"*","n":1.2300e+04,"minus":-0,"escape":"\u003c\"\\\n","null":null,"empty":[],"object":{}}`, `{"secret":"a\"\\\nb"}`, `{"empty":[],"escape":"\u003c\"\\\n","minus":-0,"n":1.2300e+04,"null":null,"object":{},"secret":"a\"\\\nb","z":900719925474099312345}`},
		{"empty object", testJSON(map[string]any{}), testJSON(map[string]any{"secret": "alpha"}), testJSON(map[string]any{})},
		{"null parameters", "null", testJSON(map[string]any{"secret": "alpha"}), ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			parametersPath := path.Root("parameters")
			response, diags := decodeInstallationParameters(parametersPath, []byte(test.response))
			require.Empty(t, diags)
			reference, diags := decodeInstallationParameters(parametersPath, []byte(test.reference))
			require.Empty(t, diags)

			actual := reconcileAppInstallationParameters(response, reference)
			raw, diags := encodeInstallationParameters(parametersPath, actual)
			require.Empty(t, diags)
			assert.Equal(t, test.expected, string(raw))
		})
	}
}
