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
		ContentTypeFieldAllowedResourceItemContentfulEntryType{},
		ContentTypeFieldAllowedResourceItemExternalType{},
		ContentTypeFieldAllowedResourceItemType{},
		ContentTypeFieldItemsType{},
		ContentTypeFieldType{},
		ContentTypeMetadataTaxonomyItemConceptSchemeType{},
		ContentTypeMetadataTaxonomyItemConceptType{},
		ContentTypeMetadataTaxonomyItemType{},
		ContentTypeMetadataType{},
		EditorInterfaceControlType{},
		EditorInterfaceEditorLayoutItemType{},
		EditorInterfaceEditorLayoutItemGroupType{},
		EditorInterfaceEditorLayoutItemGroupItemType{},
		EditorInterfaceEditorLayoutItemGroupItemFieldType{},
		EditorInterfaceEditorLayoutItemGroupItemGroupType{},
		EditorInterfaceEditorLayoutItemGroupItemGroupItemType{},
		EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldType{},
		EditorInterfaceGroupControlType{},
		EditorInterfaceSidebarType{},
		NewTypedObjectNull[RolePolicyValue]().Type(ctx),
		WebhookFilterEqualsType{},
		WebhookFilterInType{},
		WebhookFilterNotType{},
		WebhookFilterRegexpType{},
		WebhookFilterType{},
		WebhookHeaderType{},
		WebhookTransformationType{},
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
			NullValue:    NewContentTypeFieldValueNull(),
			UnknownValue: NewContentTypeFieldValueUnknown(),
		},
		"ContentTypeFieldItems": {
			NullValue:    NewContentTypeFieldItemsValueNull(),
			UnknownValue: NewContentTypeFieldItemsValueUnknown(),
		},
		"ContentTypeFieldAllowedResourceItem": {
			NullValue:    NewContentTypeFieldAllowedResourceItemValueNull(),
			UnknownValue: NewContentTypeFieldAllowedResourceItemValueUnknown(),
		},
		"ContentTypeFieldAllowedResourceItemContentfulEntry": {
			NullValue:    NewContentTypeFieldAllowedResourceItemContentfulEntryValueNull(),
			UnknownValue: NewContentTypeFieldAllowedResourceItemContentfulEntryValueUnknown(),
		},
		"ContentTypeFieldAllowedResourceItemExternal": {
			NullValue:    NewContentTypeFieldAllowedResourceItemExternalValueNull(),
			UnknownValue: NewContentTypeFieldAllowedResourceItemExternalValueUnknown(),
		},
		"ContentTypeMetadataTaxonomyItem": {
			NullValue:    NewContentTypeMetadataTaxonomyItemValueNull(),
			UnknownValue: NewContentTypeMetadataTaxonomyItemValueUnknown(),
		},
		"ContentTypeMetadataTaxonomyItemConceptScheme": {
			NullValue:    NewContentTypeMetadataTaxonomyItemConceptSchemeValueNull(),
			UnknownValue: NewContentTypeMetadataTaxonomyItemConceptSchemeValueUnknown(),
		},
		"ContentTypeMetadataTaxonomyItemConcept": {
			NullValue:    NewContentTypeMetadataTaxonomyItemConceptValueNull(),
			UnknownValue: NewContentTypeMetadataTaxonomyItemConceptValueUnknown(),
		},
		"ContentTypeMetadata": {
			NullValue:    NewContentTypeMetadataValueNull(),
			UnknownValue: NewContentTypeMetadataValueUnknown(),
		},
		"EditorInterfaceControl": {
			NullValue:    NewEditorInterfaceControlValueNull(),
			UnknownValue: NewEditorInterfaceControlValueUnknown(),
		},
		"EditorInterfaceEditorLayoutItem": {
			NullValue:    NewEditorInterfaceEditorLayoutItemValueNull(),
			UnknownValue: NewEditorInterfaceEditorLayoutItemValueUnknown(),
		},
		"EditorInterfaceEditorLayoutItemGroup": {
			NullValue:    NewEditorInterfaceEditorLayoutItemGroupValueNull(),
			UnknownValue: NewEditorInterfaceEditorLayoutItemGroupValueUnknown(),
		},
		"EditorInterfaceEditorLayoutItemGroupItem": {
			NullValue:    NewEditorInterfaceEditorLayoutItemGroupItemValueNull(),
			UnknownValue: NewEditorInterfaceEditorLayoutItemGroupItemValueUnknown(),
		},
		"EditorInterfaceEditorLayoutItemGroupItemField": {
			NullValue:    NewEditorInterfaceEditorLayoutItemGroupItemFieldValueNull(),
			UnknownValue: NewEditorInterfaceEditorLayoutItemGroupItemFieldValueUnknown(),
		},
		"EditorInterfaceEditorLayoutItemGroupItemGroup": {
			NullValue:    NewEditorInterfaceEditorLayoutItemGroupItemGroupValueNull(),
			UnknownValue: NewEditorInterfaceEditorLayoutItemGroupItemGroupValueUnknown(),
		},
		"EditorInterfaceEditorLayoutItemGroupItemGroupItem": {
			NullValue:    NewEditorInterfaceEditorLayoutItemGroupItemGroupItemValueNull(),
			UnknownValue: NewEditorInterfaceEditorLayoutItemGroupItemGroupItemValueUnknown(),
		},
		"EditorInterfaceEditorLayoutItemGroupItemGroupItemField": {
			NullValue:    NewEditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValueNull(),
			UnknownValue: NewEditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValueUnknown(),
		},
		"EditorInterfaceGroupControl": {
			NullValue:    NewEditorInterfaceGroupControlValueNull(),
			UnknownValue: NewEditorInterfaceGroupControlValueUnknown(),
		},
		"EditorInterfaceSidebar": {
			NullValue:    NewEditorInterfaceSidebarValueNull(),
			UnknownValue: NewEditorInterfaceSidebarValueUnknown(),
		},
		"RolePolicy": {
			NullValue:    NewTypedObjectNull[RolePolicyValue](),
			UnknownValue: NewTypedObjectUnknown[RolePolicyValue](),
		},
		"WebhookFilterEquals": {
			NullValue:    NewWebhookFilterEqualsValueNull(),
			UnknownValue: NewWebhookFilterEqualsValueUnknown(),
		},
		"WebhookFilterIn": {
			NullValue:    NewWebhookFilterInValueNull(),
			UnknownValue: NewWebhookFilterInValueUnknown(),
		},
		"WebhookFilterNot": {
			NullValue:    NewWebhookFilterNotValueNull(),
			UnknownValue: NewWebhookFilterNotValueUnknown(),
		},
		"WebhookFilterRegexp": {
			NullValue:    NewWebhookFilterRegexpValueNull(),
			UnknownValue: NewWebhookFilterRegexpValueUnknown(),
		},
		"WebhookFilter": {
			NullValue:    NewWebhookFilterValueNull(),
			UnknownValue: NewWebhookFilterValueUnknown(),
		},
		"WebhookHeader": {
			NullValue:    NewWebhookHeaderValueNull(),
			UnknownValue: NewWebhookHeaderValueUnknown(),
		},
		"WebhookTransformation": {
			NullValue:    NewWebhookTransformationValueNull(),
			UnknownValue: NewWebhookTransformationValueUnknown(),
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
		ContentTypeFieldAllowedResourceItemContentfulEntryType{},
		ContentTypeFieldAllowedResourceItemExternalType{},
		ContentTypeFieldAllowedResourceItemType{},
		ContentTypeFieldItemsType{},
		ContentTypeFieldType{},
		ContentTypeMetadataTaxonomyItemConceptSchemeType{},
		ContentTypeMetadataTaxonomyItemConceptType{},
		ContentTypeMetadataTaxonomyItemType{},
		ContentTypeMetadataType{},
		EditorInterfaceControlType{},
		EditorInterfaceEditorLayoutItemType{},
		EditorInterfaceEditorLayoutItemGroupType{},
		EditorInterfaceEditorLayoutItemGroupItemType{},
		EditorInterfaceEditorLayoutItemGroupItemFieldType{},
		EditorInterfaceEditorLayoutItemGroupItemGroupType{},
		EditorInterfaceEditorLayoutItemGroupItemGroupItemType{},
		EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldType{},
		EditorInterfaceGroupControlType{},
		EditorInterfaceSidebarType{},
		NewTypedObjectNull[RolePolicyValue]().Type(ctx),
		WebhookFilterEqualsType{},
		WebhookFilterInType{},
		WebhookFilterNotType{},
		WebhookFilterRegexpType{},
		WebhookFilterType{},
		WebhookHeaderType{},
		WebhookTransformationType{},
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
