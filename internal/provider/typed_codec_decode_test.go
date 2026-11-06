package provider_test

import (
	"context"
	"fmt"
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type typedCodecConversionCounts struct {
	decoded int
	encoded int
}

type typedCodecCountsContextKey struct{}

type typedCodecCountedString struct{ basetypes.StringValue }

type typedCodecCountedStringType struct{ basetypes.StringType }

//nolint:ireturn
func (value typedCodecCountedString) Type(context.Context) attr.Type {
	return typedCodecCountedStringType{}
}

func (value typedCodecCountedString) Equal(other attr.Value) bool {
	otherString, ok := other.(typedCodecCountedString)

	return ok && value.StringValue.Equal(otherString.StringValue)
}

func (value typedCodecCountedString) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	if counts, ok := ctx.Value(typedCodecCountsContextKey{}).(*typedCodecConversionCounts); ok {
		counts.encoded++
	}

	wire, err := value.StringValue.ToTerraformValue(ctx)
	if err != nil {
		return wire, fmt.Errorf("encoding counted string: %w", err)
	}

	return wire, nil
}

func (typ typedCodecCountedStringType) Equal(other attr.Type) bool {
	_, ok := other.(typedCodecCountedStringType)

	return ok
}

//nolint:ireturn
func (typ typedCodecCountedStringType) ValueType(context.Context) attr.Value {
	return typedCodecCountedString{}
}

//nolint:ireturn
func (typ typedCodecCountedStringType) ValueFromTerraform(ctx context.Context, wire tftypes.Value) (attr.Value, error) {
	if counts, ok := ctx.Value(typedCodecCountsContextKey{}).(*typedCodecConversionCounts); ok {
		counts.decoded++
	}

	value, err := typ.StringType.ValueFromTerraform(ctx, wire)
	if err != nil {
		return nil, fmt.Errorf("decoding counted string: %w", err)
	}

	stringValue, ok := value.(basetypes.StringValue)
	if !ok {
		//nolint:err113
		return nil, fmt.Errorf("unexpected counted string value %T", value)
	}

	return typedCodecCountedString{StringValue: stringValue}, nil
}

type typedCodecCountedChild struct {
	Leaf typedCodecCountedString `tfsdk:"leaf"`
}

type typedCodecCountedRoot struct {
	Items TypedList[TypedMap[typedCodecCountedString]] `tfsdk:"items"`
}

// A raw decode should visit each leaf once. Re-encoding the decoded Framework
// container before adapting it to the generic type repeats work at every level.
func TestTypedCodecDecodesNestedLeafOnce(t *testing.T) {
	t.Parallel()

	leaf := typedCodecCountedString{StringValue: basetypes.NewStringValue("leaf")}
	leafWire := tftypes.NewValue(tftypes.String, "leaf")
	child := NewTypedObject(typedCodecCountedChild{Leaf: leaf})
	childType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"leaf": tftypes.String}}
	childWire := tftypes.NewValue(childType, map[string]tftypes.Value{"leaf": leafWire})
	childListType := tftypes.List{ElementType: childType}
	childMapType := tftypes.Map{ElementType: childType}
	leafMapType := tftypes.Map{ElementType: tftypes.String}
	leafMapListType := tftypes.List{ElementType: leafMapType}

	for name, test := range map[string]struct {
		expected attr.Value
		wire     tftypes.Value
	}{
		"list of map of object": {
			expected: NewTypedList([]TypedMap[TypedObject[typedCodecCountedChild]]{
				NewTypedMap(map[string]TypedObject[typedCodecCountedChild]{"child": child}),
			}),
			wire: tftypes.NewValue(tftypes.List{ElementType: childMapType}, []tftypes.Value{
				tftypes.NewValue(childMapType, map[string]tftypes.Value{"child": childWire}),
			}),
		},
		"map of list of object": {
			expected: NewTypedMap(map[string]TypedList[TypedObject[typedCodecCountedChild]]{
				"children": NewTypedList([]TypedObject[typedCodecCountedChild]{child}),
			}),
			wire: tftypes.NewValue(tftypes.Map{ElementType: childListType}, map[string]tftypes.Value{
				"children": tftypes.NewValue(childListType, []tftypes.Value{childWire}),
			}),
		},
		"object with list of map": {
			expected: NewTypedObject(typedCodecCountedRoot{
				Items: NewTypedList([]TypedMap[typedCodecCountedString]{
					NewTypedMap(map[string]typedCodecCountedString{"leaf": leaf}),
				}),
			}),
			wire: tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"items": leafMapListType}}, map[string]tftypes.Value{
				"items": tftypes.NewValue(leafMapListType, []tftypes.Value{
					tftypes.NewValue(leafMapType, map[string]tftypes.Value{"leaf": leafWire}),
				}),
			}),
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			counts := &typedCodecConversionCounts{}
			ctx := context.WithValue(t.Context(), typedCodecCountsContextKey{}, counts)
			decoded, err := test.expected.Type(ctx).ValueFromTerraform(ctx, test.wire)
			require.NoError(t, err)
			assert.True(t, test.expected.Equal(decoded), "expected %#v, got %#v", test.expected, decoded)
			assert.Equal(t, 1, counts.decoded, "raw decoding must decode the leaf exactly once")
			assert.Zero(t, counts.encoded, "raw decoding must not re-encode a decoded leaf")
		})
	}
}

type typedCodecOpaqueJSON struct {
	Value jsontypes.Normalized `tfsdk:"value"`
}

// Raw conversion preserves the custom value; validation belongs to the
// Framework adapters, not encoding or decoding the Terraform wire value.
func TestTypedCodecRawConversionDoesNotValidateCustomValues(t *testing.T) {
	t.Parallel()

	const invalidJSON = "{"

	for name, test := range map[string]struct {
		expected attr.Value
		wire     tftypes.Value
	}{
		"list": {
			expected: NewTypedList([]jsontypes.Normalized{jsontypes.NewNormalizedValue(invalidJSON)}),
			wire: tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, invalidJSON),
			}),
		},
		"map": {
			expected: NewTypedMap(map[string]jsontypes.Normalized{"value": jsontypes.NewNormalizedValue(invalidJSON)}),
			wire: tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{
				"value": tftypes.NewValue(tftypes.String, invalidJSON),
			}),
		},
		"object": {
			expected: NewTypedObject(typedCodecOpaqueJSON{Value: jsontypes.NewNormalizedValue(invalidJSON)}),
			wire: tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{"value": tftypes.String}}, map[string]tftypes.Value{
				"value": tftypes.NewValue(tftypes.String, invalidJSON),
			}),
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			encoded, err := test.expected.ToTerraformValue(ctx)
			require.NoError(t, err)
			assert.True(t, test.wire.Equal(encoded), "expected %s, got %s", test.wire, encoded)

			decoded, err := test.expected.Type(ctx).ValueFromTerraform(ctx, test.wire)
			require.NoError(t, err)
			assert.True(t, test.expected.Equal(decoded), "expected %#v, got %#v", test.expected, decoded)
		})
	}
}
