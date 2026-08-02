package provider_test

import (
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertExactlyOneKnownAlternative(t *testing.T) {
	t.Parallel()

	unionPath := path.Root("union")
	firstPath := unionPath.AtName("first")
	secondPath := unionPath.AtName("second")

	tests := map[string]struct {
		first           types.String
		second          types.String
		expected        string
		expectedPaths   []string
		expectedErrors  bool
		conversionError bool
	}{
		"first": {
			first:    types.StringValue("first"),
			second:   types.StringNull(),
			expected: "first",
		},
		"second": {
			first:    types.StringNull(),
			second:   types.StringValue("second"),
			expected: "second",
		},
		"neither": {
			first:          types.StringNull(),
			second:         types.StringNull(),
			expectedPaths:  []string{"union"},
			expectedErrors: true,
		},
		"both": {
			first:          types.StringValue("first"),
			second:         types.StringValue("second"),
			expectedPaths:  []string{"union"},
			expectedErrors: true,
		},
		"unknown": {
			first:          types.StringUnknown(),
			second:         types.StringNull(),
			expectedPaths:  []string{"union.first"},
			expectedErrors: true,
		},
		"conversion error": {
			first:           types.StringValue("first"),
			second:          types.StringNull(),
			expectedPaths:   []string{"union.first"},
			expectedErrors:  true,
			conversionError: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, diags := ConvertExactlyOneKnownAlternative(unionPath,
				KnownUnionAlternative[string]{
					Name:  "first",
					Path:  firstPath,
					Value: test.first,
					Convert: func() (string, diag.Diagnostics) {
						if test.conversionError {
							return "partial", diag.Diagnostics{diag.NewAttributeErrorDiagnostic(firstPath, "Conversion failed", "")}
						}

						return "first", nil
					},
				},
				KnownUnionAlternative[string]{
					Name:  "second",
					Path:  secondPath,
					Value: test.second,
					Convert: func() (string, diag.Diagnostics) {
						return "second", nil
					},
				},
			)

			assert.Equal(t, test.expected, actual)
			assert.Equal(t, test.expectedErrors, diags.HasError())

			if test.expectedErrors {
				require.NotEmpty(t, diags)
				assert.Equal(t, test.expectedPaths, diagnosticPaths(t, diags))
			}
		})
	}
}
