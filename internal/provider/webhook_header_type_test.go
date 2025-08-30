package provider_test

import (
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestWebhookHeaderTypeValueFromObject(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	typ := WebhookHeaderValue{}.ObjectType(ctx)

	t.Run("unknown", func(t *testing.T) {
		t.Parallel()

		value := types.ObjectUnknown(typ.AttrTypes)

		object, diags := WebhookHeaderType{}.ValueFromObject(ctx, value)

		assert.True(t, object.IsUnknown())
		assert.Empty(t, diags)
	})

	t.Run("null", func(t *testing.T) {
		t.Parallel()

		value := types.ObjectNull(typ.AttrTypes)

		object, diags := WebhookHeaderType{}.ValueFromObject(ctx, value)

		assert.True(t, object.IsNull())
		assert.Empty(t, diags)
	})

	t.Run("value", func(t *testing.T) {
		t.Parallel()

		value, diags := types.ObjectValue(typ.AttrTypes, map[string]attr.Value{
			"value":  types.StringValue("value"),
			"secret": types.BoolValue(true),
		})
		assert.False(t, diags.HasError())

		object, diags := WebhookHeaderType{}.ValueFromObject(ctx, value)

		assert.False(t, diags.HasError())
		assert.False(t, object.IsNull())
		assert.False(t, object.IsUnknown())

		header, headerOk := object.(WebhookHeaderValue)
		assert.True(t, headerOk)
		assert.Equal(t, "value", header.Value.ValueString())
		assert.True(t, header.Secret.ValueBool())
	})
}
