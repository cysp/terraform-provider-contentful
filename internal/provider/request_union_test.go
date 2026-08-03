//nolint:testpackage // Same-package access is required to test the unexported union conversion helper directly.
package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestConvertExactlyOneKnownAlternative(t *testing.T) {
	t.Parallel()

	unionPath := path.Root("union")
	firstPath := unionPath.AtName("first")
	secondPath := unionPath.AtName("second")

	tests := map[string]struct {
		first         types.String
		second        types.String
		firstError    bool
		expected      string
		expectedPaths []string
		expectedCalls []string
	}{
		"first": {
			first:         types.StringValue("first"),
			second:        types.StringNull(),
			expected:      "first",
			expectedPaths: []string{},
			expectedCalls: []string{"first"},
		},
		"second": {
			first:         types.StringNull(),
			second:        types.StringValue("second"),
			expected:      "second",
			expectedPaths: []string{},
			expectedCalls: []string{"second"},
		},
		"neither": {
			first:         types.StringNull(),
			second:        types.StringNull(),
			expectedPaths: []string{"union"},
		},
		"both": {
			first:         types.StringValue("first"),
			second:        types.StringValue("second"),
			expectedPaths: []string{"union"},
		},
		"unknown alternative": {
			first:         types.StringUnknown(),
			second:        types.StringNull(),
			expectedPaths: []string{"union.first"},
		},
		"known and unknown alternatives": {
			first:         types.StringValue("first"),
			second:        types.StringUnknown(),
			expectedPaths: []string{"union.second"},
		},
		"selected conversion error": {
			first:         types.StringValue("first"),
			second:        types.StringNull(),
			firstError:    true,
			expectedPaths: []string{"union.first"},
			expectedCalls: []string{"first"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var calls []string

			actual, diags := convertExactlyOneKnownAlternative(
				unionPath,
				knownUnionAlternative[string]{
					Name:  "first",
					Path:  firstPath,
					Value: test.first,
					Convert: func() (string, diag.Diagnostics) {
						calls = append(calls, "first")

						if test.firstError {
							return "partial", diag.Diagnostics{diag.NewAttributeErrorDiagnostic(firstPath, "Conversion failed", "test failure")}
						}

						return "first", nil
					},
				},
				knownUnionAlternative[string]{
					Name:  "second",
					Path:  secondPath,
					Value: test.second,
					Convert: func() (string, diag.Diagnostics) {
						calls = append(calls, "second")

						return "second", nil
					},
				},
			)

			assert.Equal(t, test.expected, actual)
			assert.Equal(t, test.expectedPaths, diagnosticPathStrings(t, diags))
			assert.Equal(t, test.expectedCalls, calls)
		})
	}
}
