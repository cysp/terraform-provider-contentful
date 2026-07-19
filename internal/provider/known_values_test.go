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
