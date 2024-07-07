package provider_test

import (
	"context"
	"testing"

	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/stretchr/testify/assert"
)

func TestWebhookFilterValueKnownFromAttributesInvalid(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	attributes := map[string]attr.Value{
		"not":    provider.NewWebhookFilterNotValueNull(),
		"equals": provider.NewWebhookFilterEqualsValueNull(),
		"in":     provider.NewWebhookFilterInValueNull(),
		"regexp": provider.NewWebhookFilterRegexpValueNull(),
	}

	testcases := GenerateInvalidValueFromAttributesTestcases(t, attributes)

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, diags := provider.NewWebhookFilterValueKnownFromAttributes(ctx, testcase)
			assert.True(t, diags.HasError())
		})
	}
}
