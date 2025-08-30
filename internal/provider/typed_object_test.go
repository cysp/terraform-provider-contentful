package provider_test

import (
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestNewTypedObjectFromAttributes(t *testing.T) {
	t.Parallel()

	type testTypedObjectNestedStruct struct {
		string types.String `tfsdk:"unexported_string"`
		String types.String `tfsdk:"string"`
	}

	type testTypedObjectStruct struct {
		string       types.String                             `tfsdk:"unexported_string"`
		String       types.String                             `tfsdk:"string"`
		nestedObject TypedObject[testTypedObjectNestedStruct] `tfsdk:"unexported_nested_object"`
		NestedObject TypedObject[testTypedObjectNestedStruct] `tfsdk:"nested_object"`
	}

	t.Run("missing attributes", func(t *testing.T) {
		t.Parallel()

		expected := diag.Diagnostics{}
		expected.AddAttributeError(path.Root("string"), "invalid data", "attribute missing: string")

		object, objectDiags := NewTypedObjectFromAttributes[testTypedObjectNestedStruct](t.Context(), map[string]attr.Value{})
		assert.NotNil(t, object)
		assert.Equal(t, expected, objectDiags)
		assert.Empty(t, object.Value().string)
		assert.True(t, object.Value().String.IsNull())
	})

	t.Run("excess attributes", func(t *testing.T) {
		t.Parallel()

		expected := diag.Diagnostics{}
		expected.AddAttributeError(path.Root("unexported_string"), "invalid data", "unknown attribute: unexported_string")

		object, objectDiags := NewTypedObjectFromAttributes[testTypedObjectNestedStruct](t.Context(), map[string]attr.Value{
			"unexported_string": types.StringValue("test"),
			"string":            types.StringValue("test"),
		})
		assert.NotNil(t, object)
		assert.Equal(t, expected, objectDiags)
		assert.Empty(t, object.Value().string)
		assert.Equal(t, "test", object.Value().String.ValueString())
	})

	t.Run("valid", func(t *testing.T) {
		t.Parallel()

		nestedObject, nestedObjectDiags := NewTypedObjectFromAttributes[testTypedObjectNestedStruct](t.Context(), map[string]attr.Value{
			"string": types.StringValue("nested test"),
		})
		if nestedObjectDiags.HasError() {
			t.Fatalf("Failed to create nested typed object: %v", nestedObjectDiags)
		}

		object, objectDiags := NewTypedObjectFromAttributes[testTypedObjectStruct](t.Context(), map[string]attr.Value{
			"string":        types.StringValue("test"),
			"nested_object": nestedObject,
		})

		if objectDiags.HasError() {
			t.Fatalf("Failed to create typed object: %v", objectDiags)
		}

		assert.NotNil(t, object)

		assert.Empty(t, object.Value().string)
		assert.Equal(t, "test", object.Value().String.ValueString())
		assert.True(t, object.Value().nestedObject.IsNull())

		assert.False(t, object.Value().NestedObject.IsNull())
		assert.False(t, object.Value().NestedObject.IsUnknown())
		assert.Equal(t, "nested test", object.Value().NestedObject.Value().String.ValueString())
	})
}
