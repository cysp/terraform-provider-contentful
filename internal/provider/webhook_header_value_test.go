package provider_test

import (
	"testing"

	provider "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookHeaderValueUnknown(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	value := provider.NewWebhookHeaderValueUnknown()
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

func TestWebhookHeaderValueNull(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	value := provider.NewWebhookHeaderValueNull()
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

func TestWebhookHeaderValueInvalid(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	testcases := map[string]map[string]attr.Value{
		"invalid": {
			"value":  types.DynamicNull(),
			"secret": types.DynamicNull(),
		},
		"invalid value": {
			"value":  types.DynamicNull(),
			"secret": types.BoolNull(),
		},
		"invalid secret": {
			"value":  types.StringNull(),
			"secret": types.DynamicNull(),
		},
	}

	for name, attributes := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			value, diags := provider.NewWebhookHeaderValueKnownFromAttributes(ctx, attributes)

			assert.False(t, value.IsUnknown())
			assert.False(t, value.IsNull())

			assert.NotEmpty(t, diags)
			assert.True(t, diags.HasError())
		})
	}
}

func TestWebhookHeaderValueConversion(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	values := []provider.AttrValueWithToObjectValue{
		provider.NewWebhookHeaderValueKnown(),
		provider.NewWebhookHeaderValueKnownFromAttributesMust(ctx, map[string]attr.Value{
			"value":  types.StringUnknown(),
			"secret": types.BoolUnknown(),
		}),
		provider.NewWebhookHeaderValueKnownFromAttributesMust(ctx, map[string]attr.Value{
			"value":  types.StringNull(),
			"secret": types.BoolNull(),
		}),
		provider.NewWebhookHeaderValueKnownFromAttributesMust(ctx, map[string]attr.Value{
			"value":  types.StringValue("value"),
			"secret": types.BoolValue(false),
		}),
		provider.NewWebhookHeaderValueKnownFromAttributesMust(ctx, map[string]attr.Value{
			"value":  types.StringValue("value"),
			"secret": types.BoolValue(true),
		}),
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
