package provider_test

import (
	"context"
	"testing"

	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestWebhookFilterRegexpValueObjectRoundtrip(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	value, valueDiags := provider.NewWebhookFilterRegexpValueKnownFromAttributes(ctx, map[string]attr.Value{
		"doc":     types.StringValue("doc"),
		"pattern": types.StringValue("pattern"),
	})
	assert.Empty(t, valueDiags)

	objectValue, objectValueDiags := value.ToObjectValue(ctx)
	assert.Empty(t, objectValueDiags)

	valueFromObject, valueFromObjectDiags := provider.WebhookFilterRegexpType{}.ValueFromObject(ctx, objectValue)
	assert.Empty(t, valueFromObjectDiags)

	assert.True(t, value.Equal(valueFromObject))
}
