package provider_test

import (
	"testing"

	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestWebhookFilterInValueObjectRoundtrip(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	value, valueDiags := provider.NewWebhookFilterInValueKnownFromAttributes(ctx, map[string]attr.Value{
		"doc":    types.StringValue("doc"),
		"values": DiagsNoErrorsMust(provider.NewTypedList(ctx, []types.String{types.StringValue("value")})),
	})
	assert.Empty(t, valueDiags)

	objectValue, objectValueDiags := value.ToObjectValue(ctx)
	assert.Empty(t, objectValueDiags)

	valueFromObject, valueFromObjectDiags := provider.WebhookFilterInType{}.ValueFromObject(ctx, objectValue)
	assert.Empty(t, valueFromObjectDiags)

	assert.True(t, value.Equal(valueFromObject))
}
