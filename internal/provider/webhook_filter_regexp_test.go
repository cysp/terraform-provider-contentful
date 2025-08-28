//nolint:dupl
package provider_test

import (
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestWebhookFilterRegexpValueObjectRoundtrip(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	value, valueDiags := NewTypedObjectFromAttributes[WebhookFilterRegexpValue](ctx, map[string]attr.Value{
		"doc":     types.StringValue("doc"),
		"pattern": types.StringValue("pattern"),
	})
	assert.Empty(t, valueDiags)

	objectValue, objectValueDiags := value.ToObjectValue(ctx)
	assert.Empty(t, objectValueDiags)

	valueFromObject, valueFromObjectDiags := value.CustomType(ctx).ValueFromObject(ctx, objectValue)
	assert.Empty(t, valueFromObjectDiags)

	assert.True(t, value.Equal(valueFromObject))
}

func TestWebhookFilterRegexpValueKnownFromAttributesInvalid(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	attributes := map[string]attr.Value{
		"doc":     types.StringNull(),
		"pattern": types.StringNull(),
	}

	testcases := GenerateInvalidValueFromAttributesTestcases(t, attributes)

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, diags := NewTypedObjectFromAttributes[WebhookFilterRegexpValue](ctx, testcase)
			assert.True(t, diags.HasError())
		})
	}
}
