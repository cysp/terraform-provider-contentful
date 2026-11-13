//nolint:testpackage // Same-package access is required to test the unexported request conversion helpers directly.
package provider

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestRequestRequiredString(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		value         types.String
		expected      string
		expectedError bool
		expectedPaths []string
	}{
		"known":       {value: types.StringValue("value"), expected: "value", expectedPaths: []string{}},
		"known empty": {value: types.StringValue(""), expected: "", expectedPaths: []string{}},
		"null":        {value: types.StringNull(), expectedError: true, expectedPaths: []string{"value"}},
		"unknown":     {value: types.StringUnknown(), expectedError: true, expectedPaths: []string{"value"}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, diags := requestRequiredString(test.value, path.Root("value"))

			assert.Equal(t, test.expected, actual)
			assert.Equal(t, test.expectedError, diags.HasError())
			assert.Equal(t, test.expectedPaths, attributeDiagnosticPaths(t, diags))
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
			assert.Equal(t, test.expectedPaths, attributeDiagnosticPaths(t, diags))
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
			assert.Equal(t, test.expectedPaths, attributeDiagnosticPaths(t, diags))
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
			assert.Equal(t, test.expectedPaths, attributeDiagnosticPaths(t, diags))
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
			assert.Equal(t, test.expectedPaths, attributeDiagnosticPaths(t, diags))
		})
	}
}
