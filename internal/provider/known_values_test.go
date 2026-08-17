package provider_test

import (
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type knownObjectTestValue struct {
	Name types.String `tfsdk:"name"`
}

func TestRequireKnownObject(t *testing.T) {
	t.Parallel()

	known := knownObjectTestValue{Name: types.StringValue("known")}

	actual, diags := RequireKnownObject(NewTypedObject(known), path.Root("value"))
	require.False(t, diags.HasError())
	assert.Equal(t, known, actual)

	for name, value := range map[string]TypedObject[knownObjectTestValue]{
		"null":    NewTypedObjectNull[knownObjectTestValue](),
		"unknown": NewTypedObjectUnknown[knownObjectTestValue](),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, diags := RequireKnownObject(value, path.Root("value"))
			require.True(t, diags.HasError())
			diagnostic, ok := diags.Errors()[0].(diag.DiagnosticWithPath)
			require.True(t, ok)
			assert.Equal(t, path.Root("value"), diagnostic.Path())
		})
	}
}

func TestRequireKnownString(t *testing.T) {
	t.Parallel()

	actual, diags := RequireKnownString(types.StringValue("known"), path.Root("value"))
	require.False(t, diags.HasError())
	assert.Equal(t, "known", actual)

	for name, value := range map[string]types.String{
		"null":    types.StringNull(),
		"unknown": types.StringUnknown(),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, diags := RequireKnownString(value, path.Root("value"))
			require.True(t, diags.HasError())
			diagnostic, ok := diags.Errors()[0].(diag.DiagnosticWithPath)
			require.True(t, ok)
			assert.Equal(t, path.Root("value"), diagnostic.Path())
		})
	}
}

func TestRequireKnownStringListPreservesElementPathsAndFailsClosed(t *testing.T) {
	t.Parallel()

	valuePath := path.Root("permissions").AtMapKey("Entry")
	actual, diags := RequireKnownStringList(types.ListValueMust(types.StringType, []attr.Value{
		types.StringValue("read"),
		types.StringNull(),
		types.StringUnknown(),
	}), valuePath)

	assert.Nil(t, actual)
	require.True(t, diags.HasError())
	assert.Equal(t, []string{
		`permissions["Entry"][1]`,
		`permissions["Entry"][2]`,
	}, diagnosticPaths(t, diags))
}

func diagnosticPaths(t *testing.T, diags diag.Diagnostics) []string {
	t.Helper()

	paths := make([]string, 0, len(diags.Errors()))
	for _, diagnostic := range diags.Errors() {
		withPath, ok := diagnostic.(diag.DiagnosticWithPath)
		require.True(t, ok)

		paths = append(paths, withPath.Path().String())
	}

	return paths
}

func TestRequireKnownStringListPreservesKnownEmpty(t *testing.T) {
	t.Parallel()

	actual, diags := RequireKnownStringList(types.ListValueMust(types.StringType, []attr.Value{}), path.Root("values"))

	require.False(t, diags.HasError())
	require.NotNil(t, actual)
	assert.Empty(t, actual)
}

func TestRequireKnownStringListMapPreservesNestedPathsAndFailsClosed(t *testing.T) {
	t.Parallel()

	value := types.MapValueMust(types.ListType{ElemType: types.StringType}, map[string]attr.Value{
		"en-US": types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("known"),
			types.StringNull(),
		}),
	})

	actual, diags := RequireKnownStringListMap(value, path.Root("alt_labels"))

	assert.Nil(t, actual)
	require.True(t, diags.HasError())
	assert.Equal(t, []string{`alt_labels["en-US"][1]`}, diagnosticPaths(t, diags))
}

func TestRequireKnownStringMapPreservesMapKeyPathsAndDeterministicOrder(t *testing.T) {
	t.Parallel()

	value := types.MapValueMust(types.StringType, map[string]attr.Value{
		"z-locale": types.StringUnknown(),
		"a-locale": types.StringNull(),
	})

	actual, diags := RequireKnownStringMap(value, path.Root("pref_label"))

	assert.Nil(t, actual)
	require.True(t, diags.HasError())
	assert.Equal(t, []string{
		`pref_label["a-locale"]`,
		`pref_label["z-locale"]`,
	}, diagnosticPaths(t, diags))
}

func TestRequireKnownStringMapPreservesKnownEmpty(t *testing.T) {
	t.Parallel()

	actual, diags := RequireKnownStringMap(
		types.MapValueMust(types.StringType, map[string]attr.Value{}),
		path.Root("pref_label"),
	)

	require.False(t, diags.HasError())
	require.NotNil(t, actual)
	assert.Empty(t, actual)
}
