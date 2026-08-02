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

func TestKnownObjectValue(t *testing.T) {
	t.Parallel()

	known := knownObjectTestValue{Name: types.StringValue("known")}

	actual, diags := KnownObjectValue(NewTypedObject(known), path.Root("value"))
	require.False(t, diags.HasError())
	assert.Equal(t, known, actual)

	for name, value := range map[string]TypedObject[knownObjectTestValue]{
		"null":    NewTypedObjectNull[knownObjectTestValue](),
		"unknown": NewTypedObjectUnknown[knownObjectTestValue](),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, diags := KnownObjectValue(value, path.Root("value"))
			require.True(t, diags.HasError())
			diagnostic, ok := diags.Errors()[0].(diag.DiagnosticWithPath)
			require.True(t, ok)
			assert.Equal(t, path.Root("value"), diagnostic.Path())
		})
	}
}

func TestKnownStringValue(t *testing.T) {
	t.Parallel()

	actual, diags := KnownStringValue(types.StringValue("known"), path.Root("value"))
	require.False(t, diags.HasError())
	assert.Equal(t, "known", actual)

	for name, value := range map[string]types.String{
		"null":    types.StringNull(),
		"unknown": types.StringUnknown(),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, diags := KnownStringValue(value, path.Root("value"))
			require.True(t, diags.HasError())
			diagnostic, ok := diags.Errors()[0].(diag.DiagnosticWithPath)
			require.True(t, ok)
			assert.Equal(t, path.Root("value"), diagnostic.Path())
		})
	}
}

func TestKnownBoolValue(t *testing.T) {
	t.Parallel()

	actual, diags := KnownBoolValue(types.BoolValue(true), path.Root("value"))
	require.False(t, diags.HasError())
	assert.True(t, actual)

	for name, value := range map[string]types.Bool{
		"null":    types.BoolNull(),
		"unknown": types.BoolUnknown(),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, diags := KnownBoolValue(value, path.Root("value"))
			require.True(t, diags.HasError())
			diagnostic, ok := diags.Errors()[0].(diag.DiagnosticWithPath)
			require.True(t, ok)
			assert.Equal(t, path.Root("value"), diagnostic.Path())
		})
	}
}

func TestKnownStringValuesPreservesElementPathsAndFailsClosed(t *testing.T) {
	t.Parallel()

	valuePath := path.Root("permissions").AtMapKey("Entry")
	actual, diags := KnownStringValues([]types.String{
		types.StringValue("read"),
		types.StringNull(),
		types.StringUnknown(),
	}, valuePath)

	assert.Nil(t, actual)
	require.True(t, diags.HasError())
	assert.Equal(t, []string{
		`permissions["Entry"][1]`,
		`permissions["Entry"][2]`,
	}, diagnosticPaths(t, diags))
}

func TestKnownStringValuesPreservesKnownEmpty(t *testing.T) {
	t.Parallel()

	actual, diags := KnownStringValues([]types.String{}, path.Root("values"))

	require.False(t, diags.HasError())
	require.NotNil(t, actual)
	assert.Empty(t, actual)
}

func TestKnownStringListMapPreservesNestedPathsAndFailsClosed(t *testing.T) {
	t.Parallel()

	value := types.MapValueMust(types.ListType{ElemType: types.StringType}, map[string]attr.Value{
		"en-US": types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("known"),
			types.StringNull(),
		}),
	})

	actual, diags := KnownStringListMap(value, path.Root("alt_labels"))

	assert.Nil(t, actual)
	require.True(t, diags.HasError())
	assert.Equal(t, []string{`alt_labels["en-US"][1]`}, diagnosticPaths(t, diags))
}

func TestKnownStringMapPreservesMapKeyPathsAndDeterministicOrder(t *testing.T) {
	t.Parallel()

	value := types.MapValueMust(types.StringType, map[string]attr.Value{
		"z-locale": types.StringUnknown(),
		"a-locale": types.StringNull(),
	})

	actual, diags := KnownStringMap(value, path.Root("pref_label"))

	assert.Nil(t, actual)
	require.True(t, diags.HasError())
	assert.Equal(t, []string{
		`pref_label["a-locale"]`,
		`pref_label["z-locale"]`,
	}, diagnosticPaths(t, diags))
}

func TestKnownStringMapPreservesKnownEmpty(t *testing.T) {
	t.Parallel()

	actual, diags := KnownStringMap(
		types.MapValueMust(types.StringType, map[string]attr.Value{}),
		path.Root("pref_label"),
	)

	require.False(t, diags.HasError())
	require.NotNil(t, actual)
	assert.Empty(t, actual)
}
