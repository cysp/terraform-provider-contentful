package provider_test

import (
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The registry exercises reflection metadata for each production model. Generic
// type equality and value conversion behavior have focused tests in typed_*.
func TestModelTypesPreserveNullAndUnknown(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	modelTypes := []attr.Type{
		NewTypedListNull[types.String]().Type(ctx),
		NewTypedMapNull[types.String]().Type(ctx),
		NewTypedObjectNull[ContentTypeFieldAllowedResourceItemContentfulEntryValue]().Type(ctx),
		NewTypedObjectNull[ContentTypeFieldAllowedResourceItemExternalValue]().Type(ctx),
		NewTypedObjectNull[ContentTypeFieldAllowedResourceItemValue]().Type(ctx),
		NewTypedObjectNull[ContentTypeFieldItemsValue]().Type(ctx),
		NewTypedObjectNull[ContentTypeFieldValue]().Type(ctx),
		NewTypedObjectNull[ContentTypeMetadataTaxonomyItemConceptSchemeValue]().Type(ctx),
		NewTypedObjectNull[ContentTypeMetadataTaxonomyItemConceptValue]().Type(ctx),
		NewTypedObjectNull[ContentTypeMetadataTaxonomyItemValue]().Type(ctx),
		NewTypedObjectNull[ContentTypeMetadataValue]().Type(ctx),
		NewTypedObjectNull[EditorInterfaceControlValue]().Type(ctx),
		NewTypedObjectNull[EditorInterfaceEditorLayoutItemValue]().Type(ctx),
		NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupValue]().Type(ctx),
		NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemValue]().Type(ctx),
		NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemFieldValue]().Type(ctx),
		NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemGroupValue]().Type(ctx),
		NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemGroupItemValue]().Type(ctx),
		NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue]().Type(ctx),
		NewTypedObjectNull[EditorInterfaceGroupControlValue]().Type(ctx),
		NewTypedObjectNull[EditorInterfaceSidebarValue]().Type(ctx),
		NewTypedObjectNull[RolePolicyValue]().Type(ctx),
		NewTypedObjectNull[WebhookFilterEqualsValue]().Type(ctx),
		NewTypedObjectNull[WebhookFilterInValue]().Type(ctx),
		NewTypedObjectNull[WebhookFilterNotValue]().Type(ctx),
		NewTypedObjectNull[WebhookFilterRegexpValue]().Type(ctx),
		NewTypedObjectNull[WebhookFilterValue]().Type(ctx),
		NewTypedObjectNull[WebhookHeaderValue]().Type(ctx),
		NewTypedObjectNull[WebhookTransformationValue]().Type(ctx),
	}

	for _, typ := range modelTypes {
		t.Run(typ.String(), func(t *testing.T) {
			t.Parallel()

			for name, test := range map[string]struct {
				value   tftypes.Value
				unknown bool
				null    bool
			}{
				"unknown":   {value: tftypes.NewValue(typ.TerraformType(t.Context()), tftypes.UnknownValue), unknown: true},
				"null type": {value: tftypes.NewValue(nil, nil), null: true},
				"null":      {value: tftypes.NewValue(typ.TerraformType(t.Context()), nil), null: true},
			} {
				t.Run(name, func(t *testing.T) {
					t.Parallel()

					value, err := typ.ValueFromTerraform(t.Context(), test.value)
					require.NoError(t, err)
					assert.IsType(t, typ.ValueType(t.Context()), value)
					assert.Equal(t, test.unknown, value.IsUnknown())
					assert.Equal(t, test.null, value.IsNull())

					switch typ.(type) {
					case TypedListType[types.String], TypedMapType[types.String]:
						return
					default:
					}

					objectType, ok := typ.(basetypes.ObjectTypable)
					require.True(t, ok, "registered object models must support ObjectTypable")

					object, ok := value.(basetypes.ObjectValuable)
					require.True(t, ok)
					objectValue, diags := object.ToObjectValue(t.Context())
					require.Empty(t, diags)
					assert.Equal(t, test.unknown, objectValue.IsUnknown())
					assert.Equal(t, test.null, objectValue.IsNull())

					roundTrip, diags := objectType.ValueFromObject(t.Context(), objectValue)
					require.Empty(t, diags)
					assert.Equal(t, value, roundTrip)
				})
			}
		})
	}
}
