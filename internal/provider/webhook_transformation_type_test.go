package provider_test

import (
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookTransformationTypeValueFromObject(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	typ := WebhookTransformationValue{}.ObjectType(ctx)

	t.Run("unknown", func(t *testing.T) {
		t.Parallel()

		value := types.ObjectUnknown(typ.AttrTypes)

		object, diags := WebhookTransformationType{}.ValueFromObject(ctx, value)

		assert.True(t, object.IsUnknown())
		assert.Empty(t, diags)
	})

	t.Run("null", func(t *testing.T) {
		t.Parallel()

		value := types.ObjectNull(typ.AttrTypes)

		object, diags := WebhookTransformationType{}.ValueFromObject(ctx, value)

		assert.True(t, object.IsNull())
		assert.Empty(t, diags)
	})

	t.Run("value", func(t *testing.T) {
		t.Parallel()

		value, diags := types.ObjectValue(typ.AttrTypes, map[string]attr.Value{
			"method":                 types.StringValue("method"),
			"content_type":           types.StringValue("content_type"),
			"include_content_length": types.BoolValue(true),
			"body":                   jsontypes.NewNormalizedValue("{}"),
		})
		require.Empty(t, diags)
		require.False(t, diags.HasError())

		object, diags := WebhookTransformationType{}.ValueFromObject(ctx, value)

		assert.False(t, diags.HasError())
		assert.False(t, object.IsNull())
		assert.False(t, object.IsUnknown())

		transformation, transformationOk := object.(WebhookTransformationValue)
		assert.True(t, transformationOk)
		assert.Equal(t, "method", transformation.Method.ValueString())
		assert.True(t, transformation.IncludeContentLength.ValueBool())
	})
}
