package provider_test

import (
	"testing"

	provider "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookTransformationValueUnknown(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	value := provider.NewWebhookTransformationValueUnknown()
	assert.True(t, value.IsUnknown())
	assert.False(t, value.IsNull())

	objectValue, objectValueDiags := value.ToObjectValue(ctx)
	assert.Empty(t, objectValueDiags)

	assert.True(t, objectValue.IsUnknown())
	assert.False(t, objectValue.IsNull())

	tfvalue, tfvalueErr := value.ToTerraformValue(ctx)
	require.NoError(t, tfvalueErr)

	assert.False(t, tfvalue.IsKnown())
	assert.False(t, tfvalue.IsNull())
}

func TestWebhookTransformationValueNull(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	value := provider.NewWebhookTransformationValueNull()
	assert.False(t, value.IsUnknown())
	assert.True(t, value.IsNull())

	objectValue, objectValueDiags := value.ToObjectValue(ctx)
	assert.Empty(t, objectValueDiags)

	assert.False(t, objectValue.IsUnknown())
	assert.True(t, objectValue.IsNull())

	tfvalue, tfvalueErr := value.ToTerraformValue(ctx)
	require.NoError(t, tfvalueErr)

	assert.True(t, tfvalue.IsKnown())
	assert.True(t, tfvalue.IsNull())
}

func TestWebhookTransformationValueInvalid(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	testcases := map[string]map[string]attr.Value{
		"invalid": {
			"method":                 types.DynamicNull(),
			"content_type":           types.DynamicNull(),
			"include_content_length": types.DynamicNull(),
			"body":                   types.DynamicNull(),
		},
	}

	for name, attributes := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			value, diags := provider.NewWebhookTransformationValueKnownFromAttributes(ctx, attributes)

			assert.False(t, value.IsUnknown())
			assert.False(t, value.IsNull())

			assert.NotEmpty(t, diags)
			assert.True(t, diags.HasError())
		})
	}
}

func TestWebhookTransformationValueConversion(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	values := []provider.AttrValueWithToObjectValue{
		provider.NewWebhookTransformationValueKnown(),
		DiagsNoErrorsMust(provider.NewWebhookTransformationValueKnownFromAttributes(ctx, map[string]attr.Value{
			"method":                 types.StringUnknown(),
			"content_type":           types.StringUnknown(),
			"include_content_length": types.BoolUnknown(),
			"body":                   jsontypes.NewNormalizedUnknown(),
		})),
		DiagsNoErrorsMust(provider.NewWebhookTransformationValueKnownFromAttributes(ctx, map[string]attr.Value{
			"method":                 types.StringNull(),
			"content_type":           types.StringNull(),
			"include_content_length": types.BoolNull(),
			"body":                   jsontypes.NewNormalizedNull(),
		})),
		DiagsNoErrorsMust(provider.NewWebhookTransformationValueKnownFromAttributes(ctx, map[string]attr.Value{
			"method":                 types.StringValue("method"),
			"content_type":           types.StringValue("content_type"),
			"include_content_length": types.BoolValue(true),
			"body":                   jsontypes.NewNormalizedValue("{}"),
		})),
	}

	for _, value := range values {
		t.Run("ToObjectValue: "+value.String(), func(t *testing.T) {
			t.Parallel()

			objectValue, objectValueDiags := value.ToObjectValue(ctx)
			assert.Empty(t, objectValueDiags)

			assert.False(t, objectValue.IsUnknown())
			assert.False(t, objectValue.IsNull())
		})

		t.Run("ToTerraformValue: "+value.String(), func(t *testing.T) {
			t.Parallel()

			tfvalue, tfvalueErr := value.ToTerraformValue(ctx)
			require.NoError(t, tfvalueErr)

			assert.True(t, tfvalue.IsKnown())
			assert.False(t, tfvalue.IsNull())
		})
	}
}
