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
		input TypedList[TypedObject[WebhookFilterValue]]
	}{
		"null": {
			input: NewTypedListNull[TypedObject[WebhookFilterValue]](),
		},
		"unknown": {
			input: NewTypedListUnknown[TypedObject[WebhookFilterValue]](),
		},
	}

	expected := cm.NewOptNilWebhookDefinitionFilterArrayNull()

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result, diags := ToOptNilWebhookDefinitionFilterArray(
				ctx,
				path.Root("test"),
				testcase.input,
			)

			assert.Equal(t, expected, result)
			assert.Empty(t, diags)
		})
	}
}
