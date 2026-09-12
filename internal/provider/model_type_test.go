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

func TestModelType(t *testing.T) {
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

	type differentType struct {
		attr.Type
	}

	for index, typ := range modelTypes {
		t.Run(typ.String(), func(t *testing.T) {
			t.Parallel()

			t.Run("Equal/different Go type", func(t *testing.T) {
				t.Parallel()

				assert.False(t, typ.Equal(differentType{typ}))
			})

			for otherIndex, otherType := range modelTypes {
				t.Run("Equal/"+otherType.String(), func(t *testing.T) {
					t.Parallel()

					assert.Equal(t, index == otherIndex, typ.Equal(otherType))
				})
			}

			for name, test := range map[string]struct {
				value   tftypes.Value
				unknown bool
				null    bool
			}{
				"unknown": {
					value:   tftypes.NewValue(typ.TerraformType(t.Context()), tftypes.UnknownValue),
					unknown: true,
				},
				"null type": {
					value: tftypes.NewValue(nil, nil),
					null:  true,
				},
				"null": {
					value: tftypes.NewValue(typ.TerraformType(t.Context()), nil),
					null:  true,
				},
			} {
				t.Run("ValueFromTerraform/"+name, func(t *testing.T) {
					t.Parallel()

					value, err := typ.ValueFromTerraform(t.Context(), test.value)
					require.NoError(t, err)

					assert.Equal(t, test.unknown, value.IsUnknown())
					assert.Equal(t, test.null, value.IsNull())
				})
			}
		})
	}
}

func TestModelTypeValueFromObject(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		nullValue    attr.Value
		unknownValue attr.Value
	}{
		"ContentTypeField": {
			nullValue:    NewTypedObjectNull[ContentTypeFieldValue](),
			unknownValue: NewTypedObjectUnknown[ContentTypeFieldValue](),
		},
		"ContentTypeFieldItems": {
			nullValue:    NewTypedObjectNull[ContentTypeFieldItemsValue](),
			unknownValue: NewTypedObjectUnknown[ContentTypeFieldItemsValue](),
		},
		"ContentTypeFieldAllowedResourceItem": {
			nullValue:    NewTypedObjectNull[ContentTypeFieldAllowedResourceItemValue](),
			unknownValue: NewTypedObjectUnknown[ContentTypeFieldAllowedResourceItemValue](),
		},
		"ContentTypeFieldAllowedResourceItemContentfulEntry": {
			nullValue:    NewTypedObjectNull[ContentTypeFieldAllowedResourceItemContentfulEntryValue](),
			unknownValue: NewTypedObjectUnknown[ContentTypeFieldAllowedResourceItemContentfulEntryValue](),
		},
		"ContentTypeFieldAllowedResourceItemExternal": {
			nullValue:    NewTypedObjectNull[ContentTypeFieldAllowedResourceItemExternalValue](),
			unknownValue: NewTypedObjectUnknown[ContentTypeFieldAllowedResourceItemExternalValue](),
		},
		"ContentTypeMetadataTaxonomyItem": {
			nullValue:    NewTypedObjectNull[ContentTypeMetadataTaxonomyItemValue](),
			unknownValue: NewTypedObjectUnknown[ContentTypeMetadataTaxonomyItemValue](),
		},
		"ContentTypeMetadataTaxonomyItemConceptScheme": {
			nullValue:    NewTypedObjectNull[ContentTypeMetadataTaxonomyItemConceptSchemeValue](),
			unknownValue: NewTypedObjectUnknown[ContentTypeMetadataTaxonomyItemConceptSchemeValue](),
		},
		"ContentTypeMetadataTaxonomyItemConcept": {
			nullValue:    NewTypedObjectNull[ContentTypeMetadataTaxonomyItemConceptValue](),
			unknownValue: NewTypedObjectUnknown[ContentTypeMetadataTaxonomyItemConceptValue](),
		},
		"ContentTypeMetadata": {
			nullValue:    NewTypedObjectNull[ContentTypeMetadataValue](),
			unknownValue: NewTypedObjectUnknown[ContentTypeMetadataValue](),
		},
		"EditorInterfaceControl": {
			nullValue:    NewTypedObjectNull[EditorInterfaceControlValue](),
			unknownValue: NewTypedObjectUnknown[EditorInterfaceControlValue](),
		},
		"EditorInterfaceEditorLayoutItem": {
			nullValue:    NewTypedObjectNull[EditorInterfaceEditorLayoutItemValue](),
			unknownValue: NewTypedObjectUnknown[EditorInterfaceEditorLayoutItemValue](),
		},
		"EditorInterfaceEditorLayoutItemGroup": {
			nullValue:    NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupValue](),
			unknownValue: NewTypedObjectUnknown[EditorInterfaceEditorLayoutItemGroupValue](),
		},
		"EditorInterfaceEditorLayoutItemGroupItem": {
			nullValue:    NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemValue](),
			unknownValue: NewTypedObjectUnknown[EditorInterfaceEditorLayoutItemGroupItemValue](),
		},
		"EditorInterfaceEditorLayoutItemGroupItemField": {
			nullValue:    NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemFieldValue](),
			unknownValue: NewTypedObjectUnknown[EditorInterfaceEditorLayoutItemGroupItemFieldValue](),
		},
		"EditorInterfaceEditorLayoutItemGroupItemGroup": {
			nullValue:    NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemGroupValue](),
			unknownValue: NewTypedObjectUnknown[EditorInterfaceEditorLayoutItemGroupItemGroupValue](),
		},
		"EditorInterfaceEditorLayoutItemGroupItemGroupItem": {
			nullValue:    NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemGroupItemValue](),
			unknownValue: NewTypedObjectUnknown[EditorInterfaceEditorLayoutItemGroupItemGroupItemValue](),
		},
		"EditorInterfaceEditorLayoutItemGroupItemGroupItemField": {
			nullValue:    NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue](),
			unknownValue: NewTypedObjectUnknown[EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue](),
		},
		"EditorInterfaceGroupControl": {
			nullValue:    NewTypedObjectNull[EditorInterfaceGroupControlValue](),
			unknownValue: NewTypedObjectUnknown[EditorInterfaceGroupControlValue](),
		},
		"EditorInterfaceSidebar": {
			nullValue:    NewTypedObjectNull[EditorInterfaceSidebarValue](),
			unknownValue: NewTypedObjectUnknown[EditorInterfaceSidebarValue](),
		},
		"RolePolicy": {
			nullValue:    NewTypedObjectNull[RolePolicyValue](),
			unknownValue: NewTypedObjectUnknown[RolePolicyValue](),
		},
		"WebhookFilterEquals": {
			nullValue:    NewTypedObjectNull[WebhookFilterEqualsValue](),
			unknownValue: NewTypedObjectUnknown[WebhookFilterEqualsValue](),
		},
		"WebhookFilterIn": {
			nullValue:    NewTypedObjectNull[WebhookFilterInValue](),
			unknownValue: NewTypedObjectUnknown[WebhookFilterInValue](),
		},
		"WebhookFilterNot": {
			nullValue:    NewTypedObjectNull[WebhookFilterNotValue](),
			unknownValue: NewTypedObjectUnknown[WebhookFilterNotValue](),
		},
		"WebhookFilterRegexp": {
			nullValue:    NewTypedObjectNull[WebhookFilterRegexpValue](),
			unknownValue: NewTypedObjectUnknown[WebhookFilterRegexpValue](),
		},
		"WebhookFilter": {
			nullValue:    NewTypedObjectNull[WebhookFilterValue](),
			unknownValue: NewTypedObjectUnknown[WebhookFilterValue](),
		},
		"WebhookHeader": {
			nullValue:    NewTypedObjectNull[WebhookHeaderValue](),
			unknownValue: NewTypedObjectUnknown[WebhookHeaderValue](),
		},
		"WebhookTransformation": {
			nullValue:    NewTypedObjectNull[WebhookTransformationValue](),
			unknownValue: NewTypedObjectUnknown[WebhookTransformationValue](),
		},
	}

	for name, testcase := range testcases {
		t.Run(name+"/unknown", func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()

			val, valOk := testcase.unknownValue.(basetypes.ObjectValuable)
			require.True(t, valOk)

			typ, typOk := testcase.unknownValue.Type(ctx).(basetypes.ObjectTypable)
			require.True(t, typOk)

			objval, objvalDiags := val.ToObjectValue(ctx)
			require.Empty(t, objvalDiags)

			actual, actualDiags := typ.ValueFromObject(ctx, objval)
			require.Empty(t, actualDiags)

			assert.True(t, actual.IsUnknown())
			assert.False(t, actual.IsNull())
		})

		t.Run(name+"/null", func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()

			val, valOk := testcase.nullValue.(basetypes.ObjectValuable)
			require.True(t, valOk)

			typ, typOk := testcase.nullValue.Type(ctx).(basetypes.ObjectTypable)
			require.True(t, typOk)

			objval, objvalDiags := val.ToObjectValue(ctx)
			require.Empty(t, objvalDiags)

			actual, actualDiags := typ.ValueFromObject(ctx, objval)
			require.Empty(t, actualDiags)

			assert.False(t, actual.IsUnknown())
			assert.True(t, actual.IsNull())
		})
	}
}
