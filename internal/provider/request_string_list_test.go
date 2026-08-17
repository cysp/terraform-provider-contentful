//nolint:testpackage // The package-private request conversion module is intentionally tested through its interface.
package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequireKnownStringListElements(t *testing.T) {
	t.Parallel()

	t.Run("known", func(t *testing.T) {
		t.Parallel()

		actual, diags := knownStringListElements(path.Root("values"), []types.String{
			types.StringValue("first"),
			types.StringValue("second"),
		})

		require.False(t, diags.HasError(), diags.Errors())
		assert.Equal(t, []string{"first", "second"}, actual)
	})

	t.Run("known empty", func(t *testing.T) {
		t.Parallel()

		actual, diags := knownStringListElements(path.Root("values"), []types.String{})

		require.False(t, diags.HasError(), diags.Errors())
		require.NotNil(t, actual)
		assert.Empty(t, actual)
	})

	t.Run("invalid elements fail closed with exact paths", func(t *testing.T) {
		t.Parallel()

		actual, diags := knownStringListElements(path.Root("values"), []types.String{
			types.StringValue("first"),
			types.StringNull(),
			types.StringUnknown(),
			types.StringValue("last"),
		})

		assert.Nil(t, actual)
		require.True(t, diags.HasError())
		assert.Equal(t, []string{"values[1]", "values[2]"}, requestDiagnosticPaths(t, diags))
	})
}

func TestKnownOptionalStringSetElements(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		value               types.Set
		expected            []string
		expectedDiagnostics []string
	}{
		"known": {
			value: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("first"),
				types.StringValue("second"),
			}),
			expected: []string{"first", "second"},
		},
		"known empty": {
			value:    types.SetValueMust(types.StringType, []attr.Value{}),
			expected: []string{},
		},
		"null container": {
			value:    types.SetNull(types.StringType),
			expected: []string{},
		},
		"unknown container": {
			value:               types.SetUnknown(types.StringType),
			expectedDiagnostics: []string{"values"},
		},
		"invalid children fail closed with set value paths": {
			value: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("first"),
				types.StringNull(),
				types.StringUnknown(),
				types.StringValue("last"),
			}),
			expectedDiagnostics: []string{"values[Value(<null>)]", "values[Value(<unknown>)]"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, diags := knownOptionalStringSetElements(path.Root("values"), test.value)

			if len(test.expectedDiagnostics) == 0 {
				require.False(t, diags.HasError(), diags.Errors())
				require.NotNil(t, actual)
				assert.ElementsMatch(t, test.expected, actual)

				return
			}

			assert.Nil(t, actual)
			require.True(t, diags.HasError())
			assert.Equal(t, test.expectedDiagnostics, requestDiagnosticPaths(t, diags))
		})
	}
}

func requestDiagnosticPaths(t *testing.T, diags diag.Diagnostics) []string {
	t.Helper()

	paths := make([]string, 0, len(diags.Errors()))

	for _, diagnostic := range diags.Errors() {
		withPath, ok := diagnostic.(diag.DiagnosticWithPath)
		require.True(t, ok)

		paths = append(paths, withPath.Path().String())
	}

	return paths
}
