package provider_test

import (
	"testing"

	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRolePolicyValueUnknown(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	value := provider.NewRolePolicyValueUnknown()
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

func TestRolePolicyValueNull(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	value := provider.NewRolePolicyValueNull()
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

func TestRolePolicyValueInvalid(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	testcases := map[string]map[string]attr.Value{
		"invalid": {
			"actions":    types.DynamicNull(),
			"constraint": types.DynamicNull(),
			"effect":     types.DynamicNull(),
		},
	}

	for name, attributes := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			value, diags := provider.NewRolePolicyValueKnownFromAttributes(ctx, attributes)

			assert.False(t, value.IsUnknown())
			assert.False(t, value.IsNull())

			assert.NotEmpty(t, diags)
			assert.True(t, diags.HasError())
		})
	}
}
