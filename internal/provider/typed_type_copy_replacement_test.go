package provider_test

import (
	"context"
	"fmt"
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTypedTypeCopiesReplaceDescriptors(t *testing.T) {
	t.Parallel()

	type objectModel struct {
		Name types.String `tfsdk:"name"`
	}

	type parameterizedModel struct {
		Name parameterizedStringValue `tfsdk:"name"`
	}

	type otherObjectModel struct {
		Name types.String `tfsdk:"name"`
	}

	ctx := t.Context()
	child := parameterizedStringType{parameter: "replacement"}
	tests := map[string]struct {
		typ          attr.Type
		expectedType attr.Type
		known        tftypes.Value
	}{
		"list_changed_element": {
			typ:          NewTypedListNull[types.String]().CustomType(ctx).WithElementType(types.BoolType),
			expectedType: basetypes.ListType{ElemType: types.BoolType},
			known:        tftypes.NewValue(tftypes.List{ElementType: tftypes.Bool}, []tftypes.Value{tftypes.NewValue(tftypes.Bool, true)}),
		},
		"map_changed_element": {
			typ:          NewTypedMapNull[types.String]().CustomType(ctx).WithElementType(types.BoolType),
			expectedType: basetypes.MapType{ElemType: types.BoolType},
			known:        tftypes.NewValue(tftypes.Map{ElementType: tftypes.Bool}, map[string]tftypes.Value{"key": tftypes.NewValue(tftypes.Bool, true)}),
		},
		"list_same_wire_type_changed_model": {
			typ:          NewTypedListNull[TypedObject[objectModel]]().CustomType(ctx).WithElementType(NewTypedObjectNull[otherObjectModel]().CustomType(ctx)),
			expectedType: basetypes.ListType{ElemType: NewTypedObjectNull[otherObjectModel]().CustomType(ctx)},
			known: tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{"name": tftypes.String}}}, []tftypes.Value{
				tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"name": tftypes.String}}, map[string]tftypes.Value{"name": tftypes.NewValue(tftypes.String, "value")}),
			}),
		},
		"object_empty_attributes": {
			typ:          NewTypedObjectNull[objectModel]().CustomType(ctx).WithAttributeTypes(map[string]attr.Type{}),
			expectedType: basetypes.ObjectType{AttrTypes: map[string]attr.Type{}},
			known:        tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{}}, map[string]tftypes.Value{}),
		},
		"object_changed_attributes": {
			typ:          NewTypedObjectNull[objectModel]().CustomType(ctx).WithAttributeTypes(map[string]attr.Type{"enabled": types.BoolType}),
			expectedType: basetypes.ObjectType{AttrTypes: map[string]attr.Type{"enabled": types.BoolType}},
			known:        tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"enabled": tftypes.Bool}}, map[string]tftypes.Value{"enabled": tftypes.NewValue(tftypes.Bool, true)}),
		},
		"list_same_value_type_changed_parameter": {
			typ:          NewTypedListNull[parameterizedStringValue]().CustomType(ctx).WithElementType(child),
			expectedType: basetypes.ListType{ElemType: child},
			known:        tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{tftypes.NewValue(tftypes.String, "value")}),
		},
		"map_same_value_type_changed_parameter": {
			typ:          NewTypedMapNull[parameterizedStringValue]().CustomType(ctx).WithElementType(child),
			expectedType: basetypes.MapType{ElemType: child},
			known:        tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{"key": tftypes.NewValue(tftypes.String, "value")}),
		},
		"object_same_value_type_changed_parameter": {
			typ:          NewTypedObjectNull[parameterizedModel]().CustomType(ctx).WithAttributeTypes(map[string]attr.Type{"name": child}),
			expectedType: basetypes.ObjectType{AttrTypes: map[string]attr.Type{"name": child}},
			known:        tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"name": tftypes.String}}, map[string]tftypes.Value{"name": tftypes.NewValue(tftypes.String, "value")}),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.True(t, test.expectedType.Equal(test.typ), "replacement descriptor")
			assert.True(t, test.known.Type().Equal(test.typ.TerraformType(t.Context())), "replacement wire type")

			for state, wire := range map[string]tftypes.Value{
				"null":    tftypes.NewValue(test.known.Type(), nil),
				"unknown": tftypes.NewValue(test.known.Type(), tftypes.UnknownValue),
				"known":   test.known,
			} {
				t.Run(state, func(t *testing.T) {
					t.Parallel()

					value, err := test.typ.ValueFromTerraform(t.Context(), wire)
					require.NoError(t, err)
					assert.True(t, test.expectedType.Equal(value.Type(t.Context())), "decoded descriptor")
					assert.Equal(t, wire.IsNull(), value.IsNull())
					assert.Equal(t, !wire.IsKnown(), value.IsUnknown())
					encoded, err := value.ToTerraformValue(t.Context())
					require.NoError(t, err)
					assert.True(t, wire.Equal(encoded), "exact wire round trip")
				})
			}
		})
	}
}

// A valid custom type can carry parameters without changing its Go value type
// or Terraform wire type. Retain the parameter even for null and unknown values.
type parameterizedStringType struct {
	basetypes.StringType

	parameter string
}

type parameterizedStringValue struct {
	basetypes.StringValue

	parameter string
}

//nolint:ireturn
func (v parameterizedStringValue) Type(context.Context) attr.Type {
	return parameterizedStringType{parameter: v.parameter}
}

func (v parameterizedStringValue) Equal(other attr.Value) bool {
	o, ok := other.(parameterizedStringValue)

	return ok && v.parameter == o.parameter && v.StringValue.Equal(o.StringValue)
}

func (t parameterizedStringType) Equal(other attr.Type) bool {
	o, ok := other.(parameterizedStringType)

	return ok && t.parameter == o.parameter
}

//nolint:ireturn
func (t parameterizedStringType) ValueType(context.Context) attr.Value {
	return parameterizedStringValue{parameter: t.parameter}
}

//nolint:ireturn
func (t parameterizedStringType) ValueFromTerraform(ctx context.Context, wire tftypes.Value) (attr.Value, error) {
	value, err := t.StringType.ValueFromTerraform(ctx, wire)
	if err != nil {
		return nil, fmt.Errorf("decode parameterized string: %w", err)
	}

	stringValue, ok := value.(basetypes.StringValue)
	if !ok {
		//nolint:err113
		return nil, fmt.Errorf("unexpected string value type %T", value)
	}

	return parameterizedStringValue{StringValue: stringValue, parameter: t.parameter}, nil
}

func TestParameterizedStringDescriptorRoundTrip(t *testing.T) {
	t.Parallel()

	typ := parameterizedStringType{parameter: "replacement"}

	for name, raw := range map[string]any{"null": nil, "unknown": tftypes.UnknownValue, "known": "value"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			wire := tftypes.NewValue(tftypes.String, raw)
			value, err := typ.ValueFromTerraform(t.Context(), wire)
			require.NoError(t, err)
			assert.True(t, typ.Equal(value.Type(t.Context())))
			encoded, err := value.ToTerraformValue(t.Context())
			require.NoError(t, err)
			assert.True(t, wire.Equal(encoded))
		})
	}
}
