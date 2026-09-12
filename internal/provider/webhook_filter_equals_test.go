//nolint:dupl
package provider_test

import (
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookFilterEqualsValueObjectRoundTrip(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	value, valueDiags := NewTypedObjectFromAttributes[WebhookFilterEqualsValue](ctx, map[string]attr.Value{
		"doc":   types.StringValue("doc"),
		"value": types.StringValue("value"),
	})
	require.Empty(t, valueDiags)

	objectValue, objectValueDiags := value.ToObjectValue(ctx)
	require.Empty(t, objectValueDiags)

	valueFromObject, valueFromObjectDiags := value.CustomType(ctx).ValueFromObject(ctx, objectValue)
	require.Empty(t, valueFromObjectDiags)

	assert.True(t, value.Equal(valueFromObject))
}

func TestWebhookFilterEqualsValueKnownFromAttributesInvalid(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	attributes := map[string]attr.Value{
		"doc":   types.StringNull(),
		"value": types.StringNull(),
	}

	testcases := GenerateInvalidValueFromAttributesTestcases(t, attributes)

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, diags := NewTypedObjectFromAttributes[WebhookFilterEqualsValue](ctx, testcase)
			assert.True(t, diags.HasError())
		})
	}
}
