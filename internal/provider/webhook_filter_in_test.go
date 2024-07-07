package provider_test

import (
	"context"
	"testing"

	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestWebhookFilterInValueObjectRoundtrip(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	value, valueDiags := provider.NewWebhookFilterInValueKnownFromAttributes(ctx, map[string]attr.Value{
		"doc":    types.StringValue("doc"),
		"values": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("value")}),
	})
	assert.Empty(t, valueDiags)

	objectValue, objectValueDiags := value.ToObjectValue(ctx)
	assert.Empty(t, objectValueDiags)

	valueFromObject, valueFromObjectDiags := provider.WebhookFilterInType{}.ValueFromObject(ctx, objectValue)
	assert.Empty(t, valueFromObjectDiags)

	assert.True(t, value.Equal(valueFromObject))
}
