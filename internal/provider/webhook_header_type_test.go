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

func TestWebhookHeaderTypeEqual(t *testing.T) {
	t.Parallel()

	types := []attr.Type{
		provider.WebhookHeaderType{},
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

func TestWebhookHeaderTypeValueFromObject(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	typ := provider.WebhookHeaderValue{}.ObjectType(ctx)

	t.Run("unknown", func(t *testing.T) {
		t.Parallel()

		value := basetypes.NewObjectUnknown(typ.AttrTypes)

		object, diags := provider.WebhookHeaderType{}.ValueFromObject(ctx, value)

		assert.True(t, object.IsUnknown())
		assert.Empty(t, diags)
	})

	t.Run("null", func(t *testing.T) {
		t.Parallel()

		value := basetypes.NewObjectNull(typ.AttrTypes)

		object, diags := provider.WebhookHeaderType{}.ValueFromObject(ctx, value)

		assert.True(t, object.IsNull())
		assert.Empty(t, diags)
	})

	t.Run("value", func(t *testing.T) {
		t.Parallel()

		value, diags := basetypes.NewObjectValue(typ.AttrTypes, map[string]attr.Value{
			"value":  types.StringValue("value"),
			"secret": types.BoolValue(true),
		})
		assert.False(t, diags.HasError())

		object, diags := provider.WebhookHeaderType{}.ValueFromObject(ctx, value)

		assert.False(t, diags.HasError())
		assert.False(t, object.IsNull())
		assert.False(t, object.IsUnknown())

		header, headerOk := object.(provider.WebhookHeaderValue)
		assert.True(t, headerOk)
		assert.Equal(t, "value", header.Value.ValueString())
		assert.True(t, header.Secret.ValueBool())
	})
}

func TestWebhookHeaderTypeValueFromTerraform(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	types := []attr.Type{
		provider.WebhookHeaderType{},
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
