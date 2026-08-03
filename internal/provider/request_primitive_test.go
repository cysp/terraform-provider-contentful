//nolint:testpackage // Same-package access is required to test the unexported request conversion helpers directly.
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

func TestRequestNullableString(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		value         types.String
		expected      cm.OptNilString
		expectedPaths []string
	}{
		"known":       {value: types.StringValue("value"), expected: cm.NewOptNilString("value"), expectedPaths: []string{}},
		"known empty": {value: types.StringValue(""), expected: cm.NewOptNilString(""), expectedPaths: []string{}},
		"null":        {value: types.StringNull(), expected: cm.NewOptNilStringNull(), expectedPaths: []string{}},
		"unknown":     {value: types.StringUnknown(), expectedPaths: []string{"value"}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, diags := requestNullableString(test.value, path.Root("value"))

			assert.Equal(t, test.expected, actual)
			assert.Equal(t, test.expectedPaths, requestDiagnosticPaths(t, diags))
		})
	}
}

func TestRequestOmittableString(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		value         types.String
		expected      cm.OptString
		expectedPaths []string
	}{
		"known":       {value: types.StringValue("value"), expected: cm.NewOptString("value"), expectedPaths: []string{}},
		"known empty": {value: types.StringValue(""), expected: cm.NewOptString(""), expectedPaths: []string{}},
		"null":        {value: types.StringNull(), expectedPaths: []string{}},
		"unknown":     {value: types.StringUnknown(), expectedPaths: []string{"value"}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, diags := requestOmittableString(test.value, path.Root("value"))

			assert.Equal(t, test.expected, actual)
			assert.Equal(t, test.expectedPaths, requestDiagnosticPaths(t, diags))
		})
	}
}

func TestRequestOmittableBool(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		value         types.Bool
		expected      cm.OptBool
		expectedPaths []string
	}{
		"known true":  {value: types.BoolValue(true), expected: cm.NewOptBool(true), expectedPaths: []string{}},
		"known false": {value: types.BoolValue(false), expected: cm.NewOptBool(false), expectedPaths: []string{}},
		"null":        {value: types.BoolNull(), expectedPaths: []string{}},
		"unknown":     {value: types.BoolUnknown(), expectedPaths: []string{"value"}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, diags := requestOmittableBool(test.value, path.Root("value"))

			assert.Equal(t, test.expected, actual)
			assert.Equal(t, test.expectedPaths, requestDiagnosticPaths(t, diags))
		})
	}
}

func TestRequestPrimitiveErrorsReturnZeroValues(t *testing.T) {
	t.Parallel()

	stringValue, stringDiags := requestRequiredString(types.StringUnknown(), path.Root("string"))
	boolValue, boolDiags := requestRequiredBool(types.BoolUnknown(), path.Root("bool"))
	nullableString, nullableStringDiags := requestNullableString(types.StringUnknown(), path.Root("nullable_string"))
	omittableString, omittableStringDiags := requestOmittableString(types.StringUnknown(), path.Root("omittable_string"))
	omittableBool, omittableBoolDiags := requestOmittableBool(types.BoolUnknown(), path.Root("omittable_bool"))

	require.True(t, stringDiags.HasError())
	require.True(t, boolDiags.HasError())
	require.True(t, nullableStringDiags.HasError())
	require.True(t, omittableStringDiags.HasError())
	require.True(t, omittableBoolDiags.HasError())
	assert.Empty(t, stringValue)
	assert.False(t, boolValue)
	assert.Equal(t, cm.OptNilString{}, nullableString)
	assert.Equal(t, cm.OptString{}, omittableString)
	assert.Equal(t, cm.OptBool{}, omittableBool)
}
