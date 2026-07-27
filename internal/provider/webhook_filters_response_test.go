package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadWebhookFilterValueFromResponsePreservesRepresentableAlternatives(t *testing.T) {
	t.Parallel()

	filterPath := path.Root("filters").AtListIndex(0)

	t.Run("empty", func(t *testing.T) {
		t.Parallel()

		actual, diags := ReadWebhookFilterValueFromResponse(t.Context(), filterPath, cm.WebhookDefinitionFilter{})

		require.False(t, diags.HasError())
		require.False(t, actual.IsNull())
		assert.True(t, actual.Value().Not.IsNull())
		assert.True(t, actual.Value().Equals.IsNull())
		assert.True(t, actual.Value().In.IsNull())
		assert.True(t, actual.Value().Regexp.IsNull())
	})

	t.Run("multiple top-level alternatives", func(t *testing.T) {
		t.Parallel()

		input := cm.WebhookDefinitionFilter{
			Equals: cm.WebhookDefinitionFilterEquals{[]byte(`{"doc":"sys.type"}`), []byte(`"Entry"`)},
			In:     cm.WebhookDefinitionFilterIn{[]byte(`{"doc":"sys.id"}`), []byte(`["entry"]`)},
		}

		actual, diags := ReadWebhookFilterValueFromResponse(t.Context(), filterPath, input)

		require.False(t, diags.HasError())
		require.False(t, actual.IsNull())
		assert.False(t, actual.Value().Equals.IsNull())
		assert.False(t, actual.Value().In.IsNull())
		assert.Equal(t, "Entry", actual.Value().Equals.Value().Value.ValueString())
		assert.Equal(t, []types.String{types.StringValue("entry")}, actual.Value().In.Value().Values.Elements())
	})

	t.Run("multiple negated alternatives", func(t *testing.T) {
		t.Parallel()

		input := cm.WebhookDefinitionFilter{
			Not: cm.NewOptWebhookDefinitionFilterNot(cm.WebhookDefinitionFilterNot{
				Equals: cm.WebhookDefinitionFilterEquals{[]byte(`{"doc":"sys.type"}`), []byte(`"Entry"`)},
				In:     cm.WebhookDefinitionFilterIn{[]byte(`{"doc":"sys.id"}`), []byte(`["entry"]`)},
			}),
		}

		actual, diags := ReadWebhookFilterValueFromResponse(t.Context(), filterPath, input)

		require.False(t, diags.HasError())
		require.False(t, actual.IsNull())
		require.False(t, actual.Value().Not.IsNull())
		assert.False(t, actual.Value().Not.Value().Equals.IsNull())
		assert.False(t, actual.Value().Not.Value().In.IsNull())
	})
}

func TestReadWebhookFilterValueFromResponseReturnsNullAfterConversionError(t *testing.T) {
	t.Parallel()

	filterPath := path.Root("filters").AtListIndex(0)
	tests := map[string]struct {
		input        cm.WebhookDefinitionFilter
		expectedPath path.Path
	}{
		"malformed binary array": {
			input: cm.WebhookDefinitionFilter{
				Equals: cm.WebhookDefinitionFilterEquals{[]byte(`{"doc":"sys.type"}`)},
			},
			expectedPath: filterPath.AtName("equals"),
		},
		"incompatible term": {
			input: cm.WebhookDefinitionFilter{
				Equals: cm.WebhookDefinitionFilterEquals{[]byte(`{"doc":"sys.type"}`), []byte(`123`)},
			},
			expectedPath: filterPath.AtName("equals").AtName("value"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, diags := ReadWebhookFilterValueFromResponse(t.Context(), filterPath, test.input)

			require.True(t, diags.HasError())
			assert.True(t, actual.IsNull())
			assert.Contains(t, webhookDiagnosticPaths(t, diags), test.expectedPath.String())
		})
	}
}

func TestReadWebhookFiltersListValueFromResponseReturnsNullAfterElementError(t *testing.T) {
	t.Parallel()

	input := cm.NewOptNilWebhookDefinitionFilterArray([]cm.WebhookDefinitionFilter{
		{
			Equals: cm.WebhookDefinitionFilterEquals{[]byte(`{"doc":"sys.type"}`), []byte(`"Entry"`)},
		},
		{
			Equals: cm.WebhookDefinitionFilterEquals{[]byte(`{"doc":"sys.type"}`)},
		},
	})

	actual, diags := ReadWebhookFiltersListValueFromResponse(t.Context(), path.Root("filters"), input)

	require.True(t, diags.HasError())
	assert.True(t, actual.IsNull())
	assert.Contains(t, webhookDiagnosticPaths(t, diags), "filters[1].equals")
}

func TestReadWebhookDefinitionFilterTermString(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	testcases := map[string]struct {
		input       []byte
		expectError bool
	}{
		"valid": {
			input:       []byte(`"abc"`),
			expectError: false,
		},
		"invalid json": {
			input:       []byte(`{invalid`),
			expectError: true,
		},
		"wrong type": {
			input:       []byte(`123`),
			expectError: true,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, diags := ReadWebhookDefinitionFilterTermString(
				ctx,
				path.Root("test"),
				testcase.input,
			)

			if testcase.expectError {
				assert.True(t, diags.HasError())
			} else {
				assert.False(t, diags.HasError())
			}
		})
	}
}

func webhookDiagnosticPaths(t *testing.T, diags diag.Diagnostics) []string {
	t.Helper()

	paths := make([]string, 0, len(diags.Errors()))
	for _, diagnostic := range diags.Errors() {
		withPath, ok := diagnostic.(diag.DiagnosticWithPath)
		require.True(t, ok)

		paths = append(paths, withPath.Path().String())
	}

	return paths
}

func TestReadWebhookDefinitionFilterTermStringArray(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	testcases := map[string]struct {
		input       []byte
		expectError bool
	}{
		"valid": {
			input:       []byte(`["abc"]`),
			expectError: false,
		},
		"invalid json": {
			input:       []byte(`{invalid`),
			expectError: true,
		},
		"wrong type": {
			input:       []byte(`"abc"`),
			expectError: true,
		},
		"wrong element type": {
			input:       []byte(`["abc",123]`),
			expectError: true,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, diags := ReadWebhookDefinitionFilterTermStringArray(
				ctx,
				path.Root("test"),
				testcase.input,
			)

			if testcase.expectError {
				assert.True(t, diags.HasError())
			} else {
				assert.False(t, diags.HasError())
			}
		})
	}
}

func TestReadWebhookDefinitionFilterTermStringObject(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	testcases := map[string]struct {
		input       []byte
		name        string
		expectNull  bool
		expectValue string
		expectError bool
	}{
		"valid": {
			input:       []byte(`{"value":"abc"}`),
			name:        "value",
			expectValue: "abc",
		},
		"valid with excess": {
			input:       []byte(`{"a":"b","value":"abc","c":"d"}`),
			name:        "value",
			expectValue: "abc",
		},
		"value absent": {
			input:      []byte(`{"a":"b"}`),
			name:       "value",
			expectNull: true,
		},
		"value wrong type": {
			input:       []byte(`{"value":123}`),
			name:        "value",
			expectError: true,
		},
		"invalid json": {
			input:       []byte(`{invalid`),
			name:        "value",
			expectError: true,
		},
		"wrong type": {
			input:       []byte(`123`),
			name:        "value",
			expectError: true,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			value, diags := ReadWebhookDefinitionFilterTermStringObject(
				ctx,
				path.Root("test"),
				testcase.name,
				testcase.input,
			)

			if testcase.expectError {
				assert.True(t, diags.HasError())

				return
			}

			require.False(t, diags.HasError())

			if testcase.expectNull {
				assert.True(t, value.IsNull())
			} else {
				assert.Equal(t, testcase.expectValue, value.ValueString())
			}
		})
	}
}
