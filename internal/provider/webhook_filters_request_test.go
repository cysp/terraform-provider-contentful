package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/stretchr/testify/assert"
)

func TestToOptNilWebhookDefinitionFilterArrayNil(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	testcases := map[string]struct {
		input       TypedList[TypedObject[WebhookFilterValue]]
		expected    cm.OptNilWebhookDefinitionFilterArray
		expectError bool
	}{
		"null": {
			input:    NewTypedListNull[TypedObject[WebhookFilterValue]](),
			expected: cm.NewOptNilWebhookDefinitionFilterArrayNull(),
		},
		"unknown": {
			input:       NewTypedListUnknown[TypedObject[WebhookFilterValue]](),
			expected:    cm.OptNilWebhookDefinitionFilterArray{},
			expectError: true,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result, diags := ToOptNilWebhookDefinitionFilterArray(
				ctx,
				path.Root("test"),
				testcase.input,
			)

			assert.Equal(t, testcase.expected, result)
			assert.Equal(t, testcase.expectError, diags.HasError())
		})
	}
}
