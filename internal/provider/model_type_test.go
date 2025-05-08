package provider_test

import (
	"testing"

	"github.com/cysp/terraform-provider-contentful/internal/provider"
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

	types := []attr.Type{
		provider.NewTypedListNull[types.String](ctx).Type(ctx),
		provider.NewTypedMapNull[types.String](ctx).Type(ctx),
		provider.ContentTypeFieldAllowedResourceItemContentfulEntryType{},
		provider.ContentTypeFieldAllowedResourceItemExternalType{},
		provider.ContentTypeFieldAllowedResourceItemType{},
		provider.ContentTypeFieldItemsType{},
		provider.ContentTypeFieldType{},
		provider.EditorInterfaceControlType{},
		provider.EditorInterfaceEditorLayoutItemType{},
		provider.EditorInterfaceEditorLayoutItemGroupType{},
		provider.EditorInterfaceEditorLayoutItemGroupItemType{},
		provider.EditorInterfaceEditorLayoutItemGroupItemFieldType{},
		provider.EditorInterfaceEditorLayoutItemGroupItemGroupType{},
		provider.EditorInterfaceEditorLayoutItemGroupItemGroupItemType{},
		provider.EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldType{},
		provider.EditorInterfaceGroupControlType{},
		provider.EditorInterfaceSidebarType{},
		provider.WebhookFilterEqualsType{},
		provider.WebhookFilterInType{},
		provider.WebhookFilterNotType{},
		provider.WebhookFilterRegexpType{},
		provider.WebhookFilterType{},
		provider.WebhookHeaderType{},
		provider.WebhookTransformationType{},
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
			NullValue:    provider.NewContentTypeFieldValueNull(),
			UnknownValue: provider.NewContentTypeFieldValueUnknown(),
		},
		"ContentTypeFieldItems": {
			NullValue:    provider.NewContentTypeFieldItemsValueNull(),
			UnknownValue: provider.NewContentTypeFieldItemsValueUnknown(),
		},
		"ContentTypeFieldAllowedResourceItem": {
			NullValue:    provider.NewContentTypeFieldAllowedResourceItemValueNull(),
			UnknownValue: provider.NewContentTypeFieldAllowedResourceItemValueUnknown(),
		},
		"ContentTypeFieldAllowedResourceItemContentfulEntry": {
			NullValue:    provider.NewContentTypeFieldAllowedResourceItemContentfulEntryValueNull(),
			UnknownValue: provider.NewContentTypeFieldAllowedResourceItemContentfulEntryValueUnknown(),
		},
		"ContentTypeFieldAllowedResourceItemExternal": {
			NullValue:    provider.NewContentTypeFieldAllowedResourceItemExternalValueNull(),
			UnknownValue: provider.NewContentTypeFieldAllowedResourceItemExternalValueUnknown(),
		},
		"EditorInterfaceControl": {
			NullValue:    provider.NewEditorInterfaceControlValueNull(),
			UnknownValue: provider.NewEditorInterfaceControlValueUnknown(),
		},
		"EditorInterfaceEditorLayoutItem": {
			NullValue:    provider.NewEditorInterfaceEditorLayoutItemValueNull(),
			UnknownValue: provider.NewEditorInterfaceEditorLayoutItemValueUnknown(),
		},
		"EditorInterfaceEditorLayoutItemGroup": {
			NullValue:    provider.NewEditorInterfaceEditorLayoutItemGroupValueNull(),
			UnknownValue: provider.NewEditorInterfaceEditorLayoutItemGroupValueUnknown(),
		},
		"EditorInterfaceEditorLayoutItemGroupItem": {
			NullValue:    provider.NewEditorInterfaceEditorLayoutItemGroupItemValueNull(),
			UnknownValue: provider.NewEditorInterfaceEditorLayoutItemGroupItemValueUnknown(),
		},
		"EditorInterfaceEditorLayoutItemGroupItemField": {
			NullValue:    provider.NewEditorInterfaceEditorLayoutItemGroupItemFieldValueNull(),
			UnknownValue: provider.NewEditorInterfaceEditorLayoutItemGroupItemFieldValueUnknown(),
		},
		"EditorInterfaceEditorLayoutItemGroupItemGroup": {
			NullValue:    provider.NewEditorInterfaceEditorLayoutItemGroupItemGroupValueNull(),
			UnknownValue: provider.NewEditorInterfaceEditorLayoutItemGroupItemGroupValueUnknown(),
		},
		"EditorInterfaceEditorLayoutItemGroupItemGroupItem": {
			NullValue:    provider.NewEditorInterfaceEditorLayoutItemGroupItemGroupItemValueNull(),
			UnknownValue: provider.NewEditorInterfaceEditorLayoutItemGroupItemGroupItemValueUnknown(),
		},
		"EditorInterfaceEditorLayoutItemGroupItemGroupItemField": {
			NullValue:    provider.NewEditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValueNull(),
			UnknownValue: provider.NewEditorInterfaceEditorLayoutItemGroupItemGroupItemFieldValueUnknown(),
		},
		"EditorInterfaceGroupControl": {
			NullValue:    provider.NewEditorInterfaceGroupControlValueNull(),
			UnknownValue: provider.NewEditorInterfaceGroupControlValueUnknown(),
		},
		"EditorInterfaceSidebar": {
			NullValue:    provider.NewEditorInterfaceSidebarValueNull(),
			UnknownValue: provider.NewEditorInterfaceSidebarValueUnknown(),
		},
		"WebhookFilterEquals": {
			NullValue:    provider.NewWebhookFilterEqualsValueNull(),
			UnknownValue: provider.NewWebhookFilterEqualsValueUnknown(),
		},
		"WebhookFilterIn": {
			NullValue:    provider.NewWebhookFilterInValueNull(),
			UnknownValue: provider.NewWebhookFilterInValueUnknown(),
		},
		"WebhookFilterNot": {
			NullValue:    provider.NewWebhookFilterNotValueNull(),
			UnknownValue: provider.NewWebhookFilterNotValueUnknown(),
		},
		"WebhookFilterRegexp": {
			NullValue:    provider.NewWebhookFilterRegexpValueNull(),
			UnknownValue: provider.NewWebhookFilterRegexpValueUnknown(),
		},
		"WebhookFilter": {
			NullValue:    provider.NewWebhookFilterValueNull(),
			UnknownValue: provider.NewWebhookFilterValueUnknown(),
		},
		"WebhookHeader": {
			NullValue:    provider.NewWebhookHeaderValueNull(),
			UnknownValue: provider.NewWebhookHeaderValueUnknown(),
		},
		"WebhookTransformation": {
			NullValue:    provider.NewWebhookTransformationValueNull(),
			UnknownValue: provider.NewWebhookTransformationValueUnknown(),
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

	types := []attr.Type{
		provider.NewTypedListNull[types.String](ctx).Type(ctx),
		provider.NewTypedMapNull[types.String](ctx).Type(ctx),
		provider.ContentTypeFieldAllowedResourceItemContentfulEntryType{},
		provider.ContentTypeFieldAllowedResourceItemExternalType{},
		provider.ContentTypeFieldAllowedResourceItemType{},
		provider.ContentTypeFieldItemsType{},
		provider.ContentTypeFieldType{},
		provider.EditorInterfaceControlType{},
		provider.EditorInterfaceEditorLayoutItemType{},
		provider.EditorInterfaceEditorLayoutItemGroupType{},
		provider.EditorInterfaceEditorLayoutItemGroupItemType{},
		provider.EditorInterfaceEditorLayoutItemGroupItemFieldType{},
		provider.EditorInterfaceEditorLayoutItemGroupItemGroupType{},
		provider.EditorInterfaceEditorLayoutItemGroupItemGroupItemType{},
		provider.EditorInterfaceEditorLayoutItemGroupItemGroupItemFieldType{},
		provider.EditorInterfaceGroupControlType{},
		provider.EditorInterfaceSidebarType{},
		provider.WebhookFilterEqualsType{},
		provider.WebhookFilterInType{},
		provider.WebhookFilterNotType{},
		provider.WebhookFilterRegexpType{},
		provider.WebhookFilterType{},
		provider.WebhookHeaderType{},
		provider.WebhookTransformationType{},
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
