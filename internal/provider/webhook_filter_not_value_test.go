package provider_test

import (
	"testing"

	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/stretchr/testify/assert"
)

func TestWebhookFilterNotValueKnownFromAttributesInvalid(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	attributes := map[string]attr.Value{
		"equals": provider.NewWebhookFilterEqualsValueNull(),
		"in":     provider.NewWebhookFilterInValueNull(),
		"regexp": provider.NewWebhookFilterRegexpValueNull(),
	}

	testcases := GenerateInvalidValueFromAttributesTestcases(t, attributes)

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, diags := provider.NewWebhookFilterNotValueKnownFromAttributes(ctx, testcase)
			assert.True(t, diags.HasError())
		})
	}
}
