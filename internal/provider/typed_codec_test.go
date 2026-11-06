package provider_test

import (
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type typedCodecChild struct {
	Name types.String         `tfsdk:"name"`
	JSON jsontypes.Normalized `tfsdk:"json"`
}

type typedCodecModel struct {
	Child  TypedObject[typedCodecChild]            `tfsdk:"child"`
	Items  TypedList[TypedObject[typedCodecChild]] `tfsdk:"items"`
	Lookup TypedMap[TypedObject[typedCodecChild]]  `tfsdk:"lookup"`
}

// Literal protocol values are independent of the typed codec under test. Check
// each direction separately so a matching encode/decode mistake cannot hide.
func TestTypedCodecPreservesProtocolValues(t *testing.T) {
	t.Parallel()

	childType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"name": tftypes.String, "json": tftypes.String,
	}}
	listType := tftypes.List{ElementType: childType}
	mapType := tftypes.Map{ElementType: childType}
	objectType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"child": childType, "items": listType, "lookup": mapType,
	}}
	child := NewTypedObject(typedCodecChild{
		Name: types.StringUnknown(), JSON: jsontypes.NewNormalizedValue(`{"n":9007199254740993}`),
	})
	childWire := tftypes.NewValue(childType, map[string]tftypes.Value{
		"name": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"json": tftypes.NewValue(tftypes.String, `{"n":9007199254740993}`),
	})
	list := NewTypedList([]TypedObject[typedCodecChild]{child, NewTypedObjectNull[typedCodecChild](), NewTypedObjectUnknown[typedCodecChild]()})
	listWire := tftypes.NewValue(listType, []tftypes.Value{childWire, tftypes.NewValue(childType, nil), tftypes.NewValue(childType, tftypes.UnknownValue)})
	lookup := NewTypedMap(map[string]TypedObject[typedCodecChild]{"known": child, "null": NewTypedObjectNull[typedCodecChild](), "unknown": NewTypedObjectUnknown[typedCodecChild]()})
	lookupWire := tftypes.NewValue(mapType, map[string]tftypes.Value{"known": childWire, "null": tftypes.NewValue(childType, nil), "unknown": tftypes.NewValue(childType, tftypes.UnknownValue)})

	for name, test := range map[string]struct {
		value attr.Value
		wire  tftypes.Value
	}{
		"nested object": {
			value: NewTypedObject(typedCodecModel{Child: child, Items: list, Lookup: lookup}),
			wire: tftypes.NewValue(objectType, map[string]tftypes.Value{
				"child": childWire, "items": listWire, "lookup": lookupWire,
			}),
		},
		"null object":    {NewTypedObjectNull[typedCodecModel](), tftypes.NewValue(objectType, nil)},
		"unknown object": {NewTypedObjectUnknown[typedCodecModel](), tftypes.NewValue(objectType, tftypes.UnknownValue)},
		"list":           {list, listWire},
		"null list":      {NewTypedListNull[TypedObject[typedCodecChild]](), tftypes.NewValue(listType, nil)},
		"unknown list":   {NewTypedListUnknown[TypedObject[typedCodecChild]](), tftypes.NewValue(listType, tftypes.UnknownValue)},
		"nil list":       {NewTypedList[TypedObject[typedCodecChild]](nil), tftypes.NewValue(listType, []tftypes.Value{})},
		"empty list":     {NewTypedList([]TypedObject[typedCodecChild]{}), tftypes.NewValue(listType, []tftypes.Value{})},
		"map":            {lookup, lookupWire},
		"null map":       {NewTypedMapNull[TypedObject[typedCodecChild]](), tftypes.NewValue(mapType, nil)},
		"unknown map":    {NewTypedMapUnknown[TypedObject[typedCodecChild]](), tftypes.NewValue(mapType, tftypes.UnknownValue)},
		"nil map":        {NewTypedMap[TypedObject[typedCodecChild]](nil), tftypes.NewValue(mapType, map[string]tftypes.Value{})},
		"empty map":      {NewTypedMap(map[string]TypedObject[typedCodecChild]{}), tftypes.NewValue(mapType, map[string]tftypes.Value{})},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			typ := test.value.Type(ctx)
			assert.True(t, typ.TerraformType(ctx).Equal(test.wire.Type()))
			encoded, err := test.value.ToTerraformValue(ctx)
			require.NoError(t, err)
			assert.True(t, test.wire.Equal(encoded), "expected %s, got %s", test.wire, encoded)

			// Framework adapters must preserve the same container states, including
			// known nil-backed collections, while validating their custom elements.
			switch value := test.value.(type) {
			case basetypes.ListValuable:
				base, diags := value.ToListValue(ctx)
				require.False(t, diags.HasError(), diags)

				adapted, err := base.ToTerraformValue(ctx)
				require.NoError(t, err)
				assert.True(t, test.wire.Equal(adapted), "expected %s, got %s", test.wire, adapted)
			case basetypes.MapValuable:
				base, diags := value.ToMapValue(ctx)
				require.False(t, diags.HasError(), diags)

				adapted, err := base.ToTerraformValue(ctx)
				require.NoError(t, err)
				assert.True(t, test.wire.Equal(adapted), "expected %s, got %s", test.wire, adapted)
			}

			decoded, err := typ.ValueFromTerraform(ctx, test.wire)
			require.NoError(t, err)
			assert.IsType(t, test.value, decoded)
			assert.True(t, test.value.Equal(decoded), "expected %#v, got %#v", test.value, decoded)

			_, err = typ.ValueFromTerraform(ctx, tftypes.NewValue(tftypes.Bool, true))
			assert.Error(t, err)
		})
	}
}
