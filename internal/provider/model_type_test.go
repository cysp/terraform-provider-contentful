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

func TestModelTypeEqual(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	type InequalType struct {
		attr.Type
	}

	//nolint:dupl
	types := []attr.Type{
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

	for aIndex, aType := range types {
		t.Run(aType.String(), func(t *testing.T) {
			t.Parallel()

			t.Run(aType.String(), func(t *testing.T) {
				t.Parallel()

				assert.True(t, aType.Equal(aType)) //nolint:gocritic
			})

			t.Run("inequal", func(t *testing.T) {
				t.Parallel()

				assert.False(t, aType.Equal(InequalType{aType}))
			})

			for bIndex, bType := range types {
				if aIndex == bIndex {
					continue
				}

				t.Run(bType.String(), func(t *testing.T) {
					t.Parallel()

					assert.False(t, aType.Equal(InequalType{aType}))
				})
			}
		})
	}
}

func TestModelTypeValueFromObject(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		NullValue    attr.Value
		UnknownValue attr.Value
	}{
		"ContentTypeField": {
			NullValue:    NewTypedObjectNull[ContentTypeFieldValue](),
			UnknownValue: NewTypedObjectUnknown[ContentTypeFieldValue](),
		},
		"ContentTypeFieldItems": {
			NullValue:    NewTypedObjectNull[ContentTypeFieldItemsValue](),
			UnknownValue: NewTypedObjectUnknown[ContentTypeFieldItemsValue](),
		},
		"ContentTypeFieldAllowedResourceItem": {
			NullValue:    NewTypedObjectNull[ContentTypeFieldAllowedResourceItemValue](),
			UnknownValue: NewTypedObjectUnknown[ContentTypeFieldAllowedResourceItemValue](),
		},
		"ContentTypeFieldAllowedResourceItemContentfulEntry": {
			NullValue:    NewTypedObjectNull[ContentTypeFieldAllowedResourceItemContentfulEntryValue](),
			UnknownValue: NewTypedObjectUnknown[ContentTypeFieldAllowedResourceItemContentfulEntryValue](),
		},
		"ContentTypeFieldAllowedResourceItemExternal": {
			NullValue:    NewTypedObjectNull[ContentTypeFieldAllowedResourceItemExternalValue](),
			UnknownValue: NewTypedObjectUnknown[ContentTypeFieldAllowedResourceItemExternalValue](),
		},
		"ContentTypeMetadataTaxonomyItem": {
			NullValue:    NewTypedObjectNull[ContentTypeMetadataTaxonomyItemValue](),
			UnknownValue: NewTypedObjectUnknown[ContentTypeMetadataTaxonomyItemValue](),
		},
		"ContentTypeMetadataTaxonomyItemConceptScheme": {
			NullValue:    NewTypedObjectNull[ContentTypeMetadataTaxonomyItemConceptSchemeValue](),
			UnknownValue: NewTypedObjectUnknown[ContentTypeMetadataTaxonomyItemConceptSchemeValue](),
		},
		"ContentTypeMetadataTaxonomyItemConcept": {
			NullValue:    NewTypedObjectNull[ContentTypeMetadataTaxonomyItemConceptValue](),
			UnknownValue: NewTypedObjectUnknown[ContentTypeMetadataTaxonomyItemConceptValue](),
		},
		"ContentTypeMetadata": {
			NullValue:    NewTypedObjectNull[ContentTypeMetadataValue](),
			UnknownValue: NewTypedObjectUnknown[ContentTypeMetadataValue](),
		},
		"EditorInterfaceControl": {
			NullValue:    NewTypedObjectNull[EditorInterfaceControlValue](),
			UnknownValue: NewTypedObjectUnknown[EditorInterfaceControlValue](),
		},
		"EditorInterfaceEditorLayoutItem": {
			NullValue:    NewTypedObjectNull[EditorInterfaceEditorLayoutItemValue](),
			UnknownValue: NewTypedObjectUnknown[EditorInterfaceEditorLayoutItemValue](),
		},
		"EditorInterfaceEditorLayoutItemGroup": {
			NullValue:    NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupValue](),
			UnknownValue: NewTypedObjectUnknown[EditorInterfaceEditorLayoutItemGroupValue](),
		},
		"EditorInterfaceEditorLayoutItemGroupItem": {
			NullValue:    NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemValue](),
			UnknownValue: NewTypedObjectUnknown[EditorInterfaceEditorLayoutItemGroupItemValue](),
		},
		"EditorInterfaceEditorLayoutItemGroupItemField": {
			NullValue:    NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemFieldValue](),
			UnknownValue: NewTypedObjectUnknown[EditorInterfaceEditorLayoutItemGroupItemFieldValue](),
		},
		"EditorInterfaceEditorLayoutItemGroupItemGroup": {
			NullValue:    NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemGroupValue](),
			UnknownValue: NewTypedObjectUnknown[EditorInterfaceEditorLayoutItemGroupItemGroupValue](),
		},
		"EditorInterfaceEditorLayoutItemGroupItemGroupItem": {
			NullValue:    NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemGroupItemValue](),
			UnknownValue: NewTypedObjectUnknown[EditorInterfaceEditorLayoutItemGroupItemGroupItemValue](),
		},
		"EditorInterfaceEditorLayoutItemGroupItemGroupItemField": {
			NullValue:    NewTypedObjectNull[EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue](),
			UnknownValue: NewTypedObjectUnknown[EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValue](),
		},
		"EditorInterfaceGroupControl": {
			NullValue:    NewTypedObjectNull[EditorInterfaceGroupControlValue](),
			UnknownValue: NewTypedObjectUnknown[EditorInterfaceGroupControlValue](),
		},
		"EditorInterfaceSidebar": {
			NullValue:    NewTypedObjectNull[EditorInterfaceSidebarValue](),
			UnknownValue: NewTypedObjectUnknown[EditorInterfaceSidebarValue](),
		},
		"RolePolicy": {
			NullValue:    NewTypedObjectNull[RolePolicyValue](),
			UnknownValue: NewTypedObjectUnknown[RolePolicyValue](),
		},
		"WebhookFilterEquals": {
			NullValue:    NewTypedObjectNull[WebhookFilterEqualsValue](),
			UnknownValue: NewTypedObjectUnknown[WebhookFilterEqualsValue](),
		},
		"WebhookFilterIn": {
			NullValue:    NewTypedObjectNull[WebhookFilterInValue](),
			UnknownValue: NewTypedObjectUnknown[WebhookFilterInValue](),
		},
		"WebhookFilterNot": {
			NullValue:    NewTypedObjectNull[WebhookFilterNotValue](),
			UnknownValue: NewTypedObjectUnknown[WebhookFilterNotValue](),
		},
		"WebhookFilterRegexp": {
			NullValue:    NewTypedObjectNull[WebhookFilterRegexpValue](),
			UnknownValue: NewTypedObjectUnknown[WebhookFilterRegexpValue](),
		},
		"WebhookFilter": {
			NullValue:    NewTypedObjectNull[WebhookFilterValue](),
			UnknownValue: NewTypedObjectUnknown[WebhookFilterValue](),
		},
		"WebhookHeader": {
			NullValue:    NewTypedObjectNull[WebhookHeaderValue](),
			UnknownValue: NewTypedObjectUnknown[WebhookHeaderValue](),
		},
		"WebhookTransformation": {
			NullValue:    NewTypedObjectNull[WebhookTransformationValue](),
			UnknownValue: NewTypedObjectUnknown[WebhookTransformationValue](),
		},
	}

	for _, testcase := range testcases {
		t.Run("unknown", func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()

			val, valOk := testcase.UnknownValue.(basetypes.ObjectValuable)
			require.True(t, valOk)

			typ, typOk := testcase.UnknownValue.Type(ctx).(basetypes.ObjectTypable)
			require.True(t, typOk)

			objval, objvalDiags := val.ToObjectValue(ctx)
			require.Empty(t, objvalDiags)

			actual, actualDiags := typ.ValueFromObject(ctx, objval)
			require.Empty(t, actualDiags)

			assert.True(t, actual.IsUnknown())
			assert.False(t, actual.IsNull())
		})

		t.Run("null", func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()

			val, valOk := testcase.NullValue.(basetypes.ObjectValuable)
			require.True(t, valOk)

			typ, typOk := testcase.NullValue.Type(ctx).(basetypes.ObjectTypable)
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

func TestModelTypeValueFromTerraform(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	//nolint:dupl
	types := []attr.Type{
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

	tfvalniltype := tftypes.NewValue(nil, nil)

	for _, typ := range types {
		tftyp := typ.TerraformType(ctx)

		t.Run("unknown", func(t *testing.T) {
			t.Parallel()

			tfvalunknown := tftypes.NewValue(tftyp, tftypes.UnknownValue)
			valueUnknown, err := typ.ValueFromTerraform(ctx, tfvalunknown)
			require.NoError(t, err)
			assert.True(t, valueUnknown.IsUnknown())
		})

		t.Run("nil", func(t *testing.T) {
			t.Parallel()

			valueNil, err := typ.ValueFromTerraform(ctx, tfvalniltype)
			require.NoError(t, err)
			assert.True(t, valueNil.IsNull())
		})

		t.Run("null", func(t *testing.T) {
			t.Parallel()

			tfvalnull := tftypes.NewValue(tftyp, nil)
			valueNull, err := typ.ValueFromTerraform(ctx, tfvalnull)
			require.NoError(t, err)
			assert.True(t, valueNull.IsNull())
		})
	}
}
