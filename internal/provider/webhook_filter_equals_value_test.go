package provider_test

import (
	"context"
	"testing"

	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestWebhookFilterEqualsValueKnownFromAttributesInvalid(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	attributes := map[string]attr.Value{
		"doc":   types.StringNull(),
		"value": types.StringNull(),
	}

	testcases := GenerateInvalidValueFromAttributesTestcases(t, attributes)

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, diags := provider.NewWebhookFilterEqualsValueKnownFromAttributes(ctx, testcase)
			assert.True(t, diags.HasError())
		})
	}
}
