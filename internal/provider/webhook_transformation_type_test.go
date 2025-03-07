package provider_test

import (
	"testing"

	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
