package provider_test

import (
	"testing"

	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/assert"
)

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
