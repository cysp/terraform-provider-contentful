package provider_test

import (
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

type typedCodecSample[T attr.Value] struct {
	native T
	wire   tftypes.Value
}

// Fixed depth and at most three children per collection keep shrinking useful.
// Sibling states vary independently, unlike the single mixed graph in the
// example tests. The real Content Type items model includes opaque JSON: raw
// codecs must retain even invalid known JSON without running schema validation.
func TestTypedCodecStructuredValues(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	rapid.Check(t, func(t *rapid.T) {
		sample := typedCodecMapSamples().Draw(t, "value")

		// Neither direction uses the other as its oracle. Terraform types and
		// values below are literal protocol constructions, never codec output.
		encoded, err := sample.native.ToTerraformValue(ctx)
		require.NoError(t, err)
		require.True(t, sample.wire.Equal(encoded), "expected %s, got %s", sample.wire, encoded)

		decoded, err := sample.native.Type(ctx).ValueFromTerraform(ctx, sample.wire)
		require.NoError(t, err)
		// Compare concrete fields and states rather than the codec's Equal method.
		require.Equal(t, sample.native, decoded)
	})
}

func typedCodecStringSamples() *rapid.Generator[typedCodecSample[types.String]] {
	return rapid.Custom(func(t *rapid.T) typedCodecSample[types.String] {
		switch rapid.IntRange(0, 2).Draw(t, "state") {
		case 0:
			return typedCodecSample[types.String]{types.StringNull(), tftypes.NewValue(tftypes.String, nil)}
		case 1:
			return typedCodecSample[types.String]{types.StringUnknown(), tftypes.NewValue(tftypes.String, tftypes.UnknownValue)}
		default:
			value := rapid.StringN(0, 20, 40).Draw(t, "string")

			return typedCodecSample[types.String]{types.StringValue(value), tftypes.NewValue(tftypes.String, value)}
		}
	})
}

func typedCodecJSONSamples() *rapid.Generator[typedCodecSample[jsontypes.Normalized]] {
	return rapid.Custom(func(t *rapid.T) typedCodecSample[jsontypes.Normalized] {
		switch rapid.IntRange(0, 2).Draw(t, "state") {
		case 0:
			return typedCodecSample[jsontypes.Normalized]{jsontypes.NewNormalizedNull(), tftypes.NewValue(tftypes.String, nil)}
		case 1:
			return typedCodecSample[jsontypes.Normalized]{jsontypes.NewNormalizedUnknown(), tftypes.NewValue(tftypes.String, tftypes.UnknownValue)}
		default:
			// Include valid scalars/documents and deliberately invalid opaque text.
			value := rapid.SampledFrom([]string{"", "{", "[false,", "null", `""`, "{}", "[]", ` { "n": 9007199254740993 } `}).Draw(t, "json")

			return typedCodecSample[jsontypes.Normalized]{jsontypes.NewNormalizedValue(value), tftypes.NewValue(tftypes.String, value)}
		}
	})
}

func typedCodecValidationSamples() *rapid.Generator[typedCodecSample[TypedList[jsontypes.Normalized]]] {
	return rapid.Custom(func(t *rapid.T) typedCodecSample[TypedList[jsontypes.Normalized]] {
		typ := tftypes.List{ElementType: tftypes.String}

		switch rapid.IntRange(0, 5).Draw(t, "state") {
		case 0:
			return typedCodecSample[TypedList[jsontypes.Normalized]]{NewTypedListNull[jsontypes.Normalized](), tftypes.NewValue(typ, nil)}
		case 1:
			return typedCodecSample[TypedList[jsontypes.Normalized]]{NewTypedListUnknown[jsontypes.Normalized](), tftypes.NewValue(typ, tftypes.UnknownValue)}
		default:
			samples := rapid.SliceOfN(typedCodecJSONSamples(), 0, 3).Draw(t, "validations")
			native := make([]jsontypes.Normalized, len(samples))
			wire := make([]tftypes.Value, len(samples))

			for index, sample := range samples {
				native[index], wire[index] = sample.native, sample.wire
			}

			return typedCodecSample[TypedList[jsontypes.Normalized]]{NewTypedList(native), tftypes.NewValue(typ, wire)}
		}
	})
}

func typedCodecItemsType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"type": tftypes.String, "link_type": tftypes.String,
		"validations": tftypes.List{ElementType: tftypes.String},
	}}
}

func typedCodecObjectSamples() *rapid.Generator[typedCodecSample[TypedObject[ContentTypeFieldItemsValue]]] {
	return rapid.Custom(func(t *rapid.T) typedCodecSample[TypedObject[ContentTypeFieldItemsValue]] {
		typ := typedCodecItemsType()

		switch rapid.IntRange(0, 5).Draw(t, "state") {
		case 0:
			return typedCodecSample[TypedObject[ContentTypeFieldItemsValue]]{NewTypedObjectNull[ContentTypeFieldItemsValue](), tftypes.NewValue(typ, nil)}
		case 1:
			return typedCodecSample[TypedObject[ContentTypeFieldItemsValue]]{NewTypedObjectUnknown[ContentTypeFieldItemsValue](), tftypes.NewValue(typ, tftypes.UnknownValue)}
		default:
			itemType := typedCodecStringSamples().Draw(t, "type")
			linkType := typedCodecStringSamples().Draw(t, "link_type")
			validations := typedCodecValidationSamples().Draw(t, "validations")

			return typedCodecSample[TypedObject[ContentTypeFieldItemsValue]]{
				NewTypedObject(ContentTypeFieldItemsValue{ItemsType: itemType.native, LinkType: linkType.native, Validations: validations.native}),
				tftypes.NewValue(typ, map[string]tftypes.Value{"type": itemType.wire, "link_type": linkType.wire, "validations": validations.wire}),
			}
		}
	})
}

func typedCodecListSamples() *rapid.Generator[typedCodecSample[TypedList[TypedObject[ContentTypeFieldItemsValue]]]] {
	return rapid.Custom(func(t *rapid.T) typedCodecSample[TypedList[TypedObject[ContentTypeFieldItemsValue]]] {
		typ := tftypes.List{ElementType: typedCodecItemsType()}

		switch rapid.IntRange(0, 5).Draw(t, "state") {
		case 0:
			return typedCodecSample[TypedList[TypedObject[ContentTypeFieldItemsValue]]]{NewTypedListNull[TypedObject[ContentTypeFieldItemsValue]](), tftypes.NewValue(typ, nil)}
		case 1:
			return typedCodecSample[TypedList[TypedObject[ContentTypeFieldItemsValue]]]{NewTypedListUnknown[TypedObject[ContentTypeFieldItemsValue]](), tftypes.NewValue(typ, tftypes.UnknownValue)}
		default:
			samples := rapid.SliceOfN(typedCodecObjectSamples(), 0, 3).Draw(t, "items")
			native := make([]TypedObject[ContentTypeFieldItemsValue], len(samples))
			wire := make([]tftypes.Value, len(samples))

			for index, sample := range samples {
				native[index], wire[index] = sample.native, sample.wire
			}

			return typedCodecSample[TypedList[TypedObject[ContentTypeFieldItemsValue]]]{NewTypedList(native), tftypes.NewValue(typ, wire)}
		}
	})
}

func typedCodecMapSamples() *rapid.Generator[typedCodecSample[TypedMap[TypedList[TypedObject[ContentTypeFieldItemsValue]]]]] {
	return rapid.Custom(func(t *rapid.T) typedCodecSample[TypedMap[TypedList[TypedObject[ContentTypeFieldItemsValue]]]] {
		typ := tftypes.Map{ElementType: tftypes.List{ElementType: typedCodecItemsType()}}

		switch rapid.IntRange(0, 5).Draw(t, "state") {
		case 0:
			return typedCodecSample[TypedMap[TypedList[TypedObject[ContentTypeFieldItemsValue]]]]{NewTypedMapNull[TypedList[TypedObject[ContentTypeFieldItemsValue]]](), tftypes.NewValue(typ, nil)}
		case 1:
			return typedCodecSample[TypedMap[TypedList[TypedObject[ContentTypeFieldItemsValue]]]]{NewTypedMapUnknown[TypedList[TypedObject[ContentTypeFieldItemsValue]]](), tftypes.NewValue(typ, tftypes.UnknownValue)}
		default:
			samples := rapid.MapOfN(rapid.StringN(0, 10, 20), typedCodecListSamples(), 0, 3).Draw(t, "lookup")
			native := make(map[string]TypedList[TypedObject[ContentTypeFieldItemsValue]], len(samples))
			wire := make(map[string]tftypes.Value, len(samples))

			for key, sample := range samples {
				native[key], wire[key] = sample.native, sample.wire
			}

			return typedCodecSample[TypedMap[TypedList[TypedObject[ContentTypeFieldItemsValue]]]]{NewTypedMap(native), tftypes.NewValue(typ, wire)}
		}
	})
}
