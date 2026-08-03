//nolint:testpackage // The package-private request conversion module is intentionally tested through its interface.
package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertKnownObjectListElements(t *testing.T) {
	t.Parallel()

	t.Run("converts known elements with exact paths", func(t *testing.T) {
		t.Parallel()

		actualPaths := []string{}
		actual, diags := convertKnownObjectListElements(
			t.Context(),
			path.Root("items"),
			[]TypedObject[string]{NewTypedObject("first"), NewTypedObject("second")},
			func(_ context.Context, valuePath path.Path, value string) (string, diag.Diagnostics) {
				actualPaths = append(actualPaths, valuePath.String())

				return value, nil
			},
		)

		require.False(t, diags.HasError(), diags.Errors())
		assert.Equal(t, []string{"first", "second"}, actual)
		assert.Equal(t, []string{"items[0]", "items[1]"}, actualPaths)
	})

	t.Run("rejects unresolved elements without partial output", func(t *testing.T) {
		t.Parallel()

		actual, diags := convertKnownObjectListElements(
			t.Context(),
			path.Root("items"),
			[]TypedObject[string]{
				NewTypedObject("first"),
				NewTypedObjectNull[string](),
				NewTypedObjectUnknown[string](),
			},
			func(_ context.Context, _ path.Path, value string) (string, diag.Diagnostics) {
				return value, nil
			},
		)

		assert.Nil(t, actual)
		assert.Equal(t, []string{"items[1]", "items[2]"}, diagnosticPathStrings(t, diags))
	})

	t.Run("discards converted values after a child conversion error", func(t *testing.T) {
		t.Parallel()

		actual, diags := convertKnownObjectListElements(
			t.Context(),
			path.Root("items"),
			[]TypedObject[string]{NewTypedObject("first"), NewTypedObject("second")},
			func(_ context.Context, valuePath path.Path, value string) (string, diag.Diagnostics) {
				if value == "second" {
					return "partial", diag.Diagnostics{diag.NewAttributeErrorDiagnostic(valuePath, "Conversion failed", "test failure")}
				}

				return value, nil
			},
		)

		assert.Nil(t, actual)
		assert.Equal(t, []string{"items[1]"}, diagnosticPathStrings(t, diags))
	})
}

func diagnosticPathStrings(t *testing.T, diags diag.Diagnostics) []string {
	t.Helper()

	paths := make([]string, 0, len(diags.Errors()))

	for _, diagnostic := range diags.Errors() {
		withPath, ok := diagnostic.(diag.DiagnosticWithPath)
		require.True(t, ok)

		paths = append(paths, withPath.Path().String())
	}

	return paths
}
