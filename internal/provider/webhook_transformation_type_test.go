package provider_test

import (
	"testing"

	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookTransformationTypeEqual(t *testing.T) {
	t.Parallel()

	types := []attr.Type{
		provider.WebhookTransformationType{},
	}

	for aIndex, aType := range types {
		t.Run(aType.String(), func(t *testing.T) {
			t.Parallel()

			for bIndex, bType := range types {
				t.Run(bType.String(), func(t *testing.T) {
					t.Parallel()

					assert.Equal(t, aIndex == bIndex, aType.Equal(bType))
				})
			}
		})
	}
}

func TestWebhookTransformationTypeValueFromObject(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	typ := provider.WebhookTransformationValue{}.ObjectType(ctx)

	t.Run("unknown", func(t *testing.T) {
		t.Parallel()

		value := basetypes.NewObjectUnknown(typ.AttrTypes)

		object, diags := provider.WebhookTransformationType{}.ValueFromObject(ctx, value)

		assert.True(t, object.IsUnknown())
		assert.Empty(t, diags)
	})

	t.Run("null", func(t *testing.T) {
		t.Parallel()

		value := basetypes.NewObjectNull(typ.AttrTypes)

		object, diags := provider.WebhookTransformationType{}.ValueFromObject(ctx, value)

		assert.True(t, object.IsNull())
		assert.Empty(t, diags)
	})

	t.Run("value", func(t *testing.T) {
		t.Parallel()

		value, diags := basetypes.NewObjectValue(typ.AttrTypes, map[string]attr.Value{
			"method":                 types.StringValue("method"),
			"content_type":           types.StringValue("content_type"),
			"include_content_length": types.BoolValue(true),
			"body":                   jsontypes.NewNormalizedValue("{}"),
		})
		require.Empty(t, diags)
		require.False(t, diags.HasError())

		object, diags := provider.WebhookTransformationType{}.ValueFromObject(ctx, value)

		assert.False(t, diags.HasError())
		assert.False(t, object.IsNull())
		assert.False(t, object.IsUnknown())

		transformation, transformationOk := object.(provider.WebhookTransformationValue)
		assert.True(t, transformationOk)
		assert.Equal(t, "method", transformation.Method.ValueString())
		assert.True(t, transformation.IncludeContentLength.ValueBool())
	})
}

//nolint:dupl
func TestWebhookTransformationTypeValueFromTerraform(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	types := []attr.Type{
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
