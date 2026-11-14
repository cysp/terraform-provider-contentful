package provider_test

import (
	"reflect"
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTypedTypeCopiesPreserveValueTypes(t *testing.T) {
	t.Parallel()

	type objectModel struct {
		Name types.String `tfsdk:"name"`
	}

	type nestedModel struct {
		Names TypedList[types.String]            `tfsdk:"names"`
		Items TypedMap[TypedObject[objectModel]] `tfsdk:"items"`
	}

	listType := NewTypedListNull[types.String]().CustomType(t.Context())
	mapType := NewTypedMapNull[types.String]().CustomType(t.Context())
	objectType := NewTypedObjectNull[objectModel]().CustomType(t.Context())
	nestedType := NewTypedObjectNull[nestedModel]().CustomType(t.Context())

	tests := map[string]struct {
		typ      attr.Type
		wireType tftypes.Type
		known    any
		expected attr.Value
	}{
		"list": {
			typ:      listType.WithElementType(types.StringType),
			wireType: tftypes.List{ElementType: tftypes.String},
			known:    []tftypes.Value{tftypes.NewValue(tftypes.String, "value")},
			expected: NewTypedList([]types.String{types.StringValue("value")}),
		},
		"map": {
			typ:      mapType.WithElementType(types.StringType),
			wireType: tftypes.Map{ElementType: tftypes.String},
			known:    map[string]tftypes.Value{"key": tftypes.NewValue(tftypes.String, "value")},
			expected: NewTypedMap(map[string]types.String{"key": types.StringValue("value")}),
		},
		"object": {
			typ:      objectType.WithAttributeTypes(map[string]attr.Type{"name": types.StringType}),
			wireType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{"name": tftypes.String}},
			known:    map[string]tftypes.Value{"name": tftypes.NewValue(tftypes.String, "value")},
			expected: NewTypedObject(objectModel{Name: types.StringValue("value")}),
		},
		"nested_object": {
			typ: nestedType.WithAttributeTypes(map[string]attr.Type{
				"names": NewTypedListNull[types.String]().CustomType(t.Context()),
				"items": NewTypedMapNull[TypedObject[objectModel]]().CustomType(t.Context()),
			}),
			wireType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
				"names": tftypes.List{ElementType: tftypes.String},
				"items": tftypes.Map{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{"name": tftypes.String}}},
			}},
			known: map[string]tftypes.Value{
				"names": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "value"), tftypes.NewValue(tftypes.String, nil), tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				}),
				"items": tftypes.NewValue(tftypes.Map{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{"name": tftypes.String}}}, map[string]tftypes.Value{
					"known":   tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"name": tftypes.String}}, map[string]tftypes.Value{"name": tftypes.NewValue(tftypes.String, "nested")}),
					"null":    tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"name": tftypes.String}}, nil),
					"unknown": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"name": tftypes.String}}, tftypes.UnknownValue),
				}),
			},
			expected: NewTypedObject(nestedModel{
				Names: NewTypedList([]types.String{types.StringValue("value"), types.StringNull(), types.StringUnknown()}),
				Items: NewTypedMap(map[string]TypedObject[objectModel]{
					"known":   NewTypedObject(objectModel{Name: types.StringValue("nested")}),
					"null":    NewTypedObjectNull[objectModel](),
					"unknown": NewTypedObjectUnknown[objectModel](),
				}),
			}),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var wireType tftypes.Type

			require.NotPanics(t, func() { wireType = test.typ.TerraformType(t.Context()) })
			assert.True(t, test.wireType.Equal(wireType))
			assert.IsType(t, test.expected, test.typ.ValueType(t.Context()))

			for state, value := range map[string]any{"null": nil, "unknown": tftypes.UnknownValue, "known": test.known} {
				t.Run(state, func(t *testing.T) {
					t.Parallel()

					wire := tftypes.NewValue(test.wireType, value)
					converted, err := test.typ.ValueFromTerraform(t.Context(), wire)
					require.NoError(t, err)
					assert.IsType(t, test.expected, converted)
					assert.Equal(t, wire.IsNull(), converted.IsNull())
					assert.Equal(t, !wire.IsKnown(), converted.IsUnknown())

					if wire.IsKnown() && !wire.IsNull() {
						assert.True(t, test.expected.Equal(converted))
					}

					encoded, err := converted.ToTerraformValue(t.Context())
					require.NoError(t, err)
					assert.True(t, wire.Equal(encoded))
				})
			}
		})
	}
}

func TestTypedTypeCopiesSupportFrameworkConversion(t *testing.T) {
	t.Parallel()

	type model struct {
		Name types.String `tfsdk:"name"`
	}

	type plainModel struct {
		Name string `tfsdk:"name"`
	}

	for name, test := range map[string]struct {
		typ      attr.Type
		input    any
		output   any
		expected attr.Value
	}{
		"list": {
			typ:   NewTypedListNull[types.String]().CustomType(t.Context()).WithElementType(types.StringType),
			input: []string{"value"}, output: new([]string), expected: NewTypedList([]types.String{types.StringValue("value")}),
		},
		"map": {
			typ:   NewTypedMapNull[types.String]().CustomType(t.Context()).WithElementType(types.StringType),
			input: map[string]string{"key": "value"}, output: new(map[string]string), expected: NewTypedMap(map[string]types.String{"key": types.StringValue("value")}),
		},
		"object": {
			typ:   NewTypedObjectNull[model]().CustomType(t.Context()).WithAttributeTypes(map[string]attr.Type{"name": types.StringType}),
			input: plainModel{Name: "value"}, output: new(plainModel), expected: NewTypedObject(model{Name: types.StringValue("value")}),
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var value attr.Value

			diags := tfsdk.ValueFrom(t.Context(), test.input, test.typ, &value)
			require.False(t, diags.HasError(), "%v", diags)
			assert.True(t, test.expected.Equal(value))

			diags = tfsdk.ValueAs(t.Context(), value, test.output)
			require.False(t, diags.HasError(), "%v", diags)
			assert.Equal(t, test.input, reflect.ValueOf(test.output).Elem().Interface())
		})
	}
}

func TestTypedTypeCopiesTraverseNestedDescriptors(t *testing.T) {
	t.Parallel()

	type model struct {
		Name types.String `tfsdk:"name"`
	}

	objectType := NewTypedObjectNull[model]().CustomType(t.Context())
	listType := NewTypedListNull[TypedObject[model]]().CustomType(t.Context())
	mapType := NewTypedMapNull[TypedList[TypedObject[model]]]().CustomType(t.Context())
	copied := mapType.WithElementType(listType.WithElementType(objectType.WithAttributeTypes(map[string]attr.Type{"name": types.StringType})))
	assert.True(t, mapType.Equal(copied))
	assert.True(t, copied.Equal(mapType))

	path := tftypes.NewAttributePath().WithElementKeyString("key").WithElementKeyInt(0).WithAttributeName("name")
	value, remaining, err := tftypes.WalkAttributePath(copied, path)
	require.NoError(t, err)
	assert.Empty(t, remaining.Steps())
	assert.Equal(t, types.StringType, value)

	missing := tftypes.NewAttributePath().WithAttributeName("missing")
	_, remaining, err = tftypes.WalkAttributePath(copied, tftypes.NewAttributePath().WithElementKeyString("key").WithElementKeyInt(0).WithAttributeName("missing"))
	require.Error(t, err)
	assert.True(t, missing.Equal(remaining))
}
