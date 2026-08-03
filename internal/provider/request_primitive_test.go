//nolint:testpackage // The package-private request conversion module is intentionally tested through its interface.
package provider

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestRequiredString(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		value         types.String
		expected      string
		expectedPaths []string
	}{
		"known":   {value: types.StringValue("value"), expected: "value", expectedPaths: []string{}},
		"null":    {value: types.StringNull(), expectedPaths: []string{"value"}},
		"unknown": {value: types.StringUnknown(), expectedPaths: []string{"value"}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, diags := requestRequiredString(test.value, path.Root("value"))

			assert.Equal(t, test.expected, actual)
			assert.Equal(t, test.expectedPaths, requestDiagnosticPaths(t, diags))
		})
	}
}

func TestRequestRequiredBool(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		value         types.Bool
		expected      bool
		expectedPaths []string
	}{
		"known true":  {value: types.BoolValue(true), expected: true, expectedPaths: []string{}},
		"known false": {value: types.BoolValue(false), expected: false, expectedPaths: []string{}},
		"null":        {value: types.BoolNull(), expectedPaths: []string{"value"}},
		"unknown":     {value: types.BoolUnknown(), expectedPaths: []string{"value"}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, diags := requestRequiredBool(test.value, path.Root("value"))

			assert.Equal(t, test.expected, actual)
			assert.Equal(t, test.expectedPaths, requestDiagnosticPaths(t, diags))
		})
	}
}

func TestRequestOptionalString(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		value         types.String
		expected      cm.OptString
		expectedPaths []string
	}{
		"known":   {value: types.StringValue("value"), expected: cm.NewOptString("value"), expectedPaths: []string{}},
		"null":    {value: types.StringNull(), expectedPaths: []string{}},
		"unknown": {value: types.StringUnknown(), expectedPaths: []string{"value"}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, diags := requestOptionalString(test.value, path.Root("value"))

			assert.Equal(t, test.expected, actual)
			assert.Equal(t, test.expectedPaths, requestDiagnosticPaths(t, diags))
		})
	}
}

func TestRequestOptionalBool(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		value         types.Bool
		expected      cm.OptBool
		expectedPaths []string
	}{
		"known":   {value: types.BoolValue(true), expected: cm.NewOptBool(true), expectedPaths: []string{}},
		"null":    {value: types.BoolNull(), expectedPaths: []string{}},
		"unknown": {value: types.BoolUnknown(), expectedPaths: []string{"value"}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, diags := requestOptionalBool(test.value, path.Root("value"))

			assert.Equal(t, test.expected, actual)
			assert.Equal(t, test.expectedPaths, requestDiagnosticPaths(t, diags))
		})
	}
}

func TestRequestPrimitiveErrorsReturnZeroValues(t *testing.T) {
	t.Parallel()

	stringValue, stringDiags := requestRequiredString(types.StringUnknown(), path.Root("string"))
	boolValue, boolDiags := requestRequiredBool(types.BoolUnknown(), path.Root("bool"))
	optionalString, optionalStringDiags := requestOptionalString(types.StringUnknown(), path.Root("optional_string"))
	optionalBool, optionalBoolDiags := requestOptionalBool(types.BoolUnknown(), path.Root("optional_bool"))

	require.True(t, stringDiags.HasError())
	require.True(t, boolDiags.HasError())
	require.True(t, optionalStringDiags.HasError())
	require.True(t, optionalBoolDiags.HasError())
	assert.Empty(t, stringValue)
	assert.False(t, boolValue)
	assert.Equal(t, cm.OptString{}, optionalString)
	assert.Equal(t, cm.OptBool{}, optionalBool)
}
