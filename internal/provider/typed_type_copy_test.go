package provider_test

import (
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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

	listType := NewTypedListNull[types.String]().CustomType(t.Context())
	mapType := NewTypedMapNull[types.String]().CustomType(t.Context())
	objectType := NewTypedObjectNull[objectModel]().CustomType(t.Context())

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
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			require.IsType(t, test.expected, test.typ.ValueType(t.Context()))
			assert.True(t, test.wireType.Equal(test.typ.TerraformType(t.Context())))

			for state, value := range map[string]any{"null": nil, "unknown": tftypes.UnknownValue, "known": test.known} {
				t.Run(state, func(t *testing.T) {
					t.Parallel()

					wire := tftypes.NewValue(test.wireType, value)
					converted, err := test.typ.ValueFromTerraform(t.Context(), wire)
					require.NoError(t, err)
					require.IsType(t, test.expected, converted)
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
