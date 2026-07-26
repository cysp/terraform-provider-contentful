package provider_test

import (
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExactlyOneKnownAlternative(t *testing.T) {
	t.Parallel()

	unionPath := path.Root("union")
	firstPath := unionPath.AtName("first")
	secondPath := unionPath.AtName("second")

	tests := map[string]struct {
		first          types.String
		second         types.String
		expected       string
		expectedPaths  []string
		expectedErrors bool
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
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, diags := ExactlyOneKnownAlternative(unionPath,
				KnownUnionAlternative{Name: "first", Path: firstPath, Value: test.first},
				KnownUnionAlternative{Name: "second", Path: secondPath, Value: test.second},
			)

			assert.Equal(t, test.expected, actual.Name)
			assert.Equal(t, test.expectedErrors, diags.HasError())

			if test.expectedErrors {
				require.NotEmpty(t, diags)
				assert.Equal(t, test.expectedPaths, diagnosticPaths(t, diags))
			}
		})
	}
}
