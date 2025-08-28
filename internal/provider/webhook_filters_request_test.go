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
		input TypedList[WebhookFilterValue]
	}{
		"null": {
			input: NewTypedListNull[WebhookFilterValue](),
		},
		"unknown": {
			input: NewTypedListUnknown[WebhookFilterValue](),
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

func TestToWebhookDefinitionFilterNil(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	testcases := map[string]WebhookFilterValue{
		"null":    NewWebhookFilterValueNull(),
		"unknown": NewWebhookFilterValueUnknown(),
	}

	for name, value := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result, diags := ToWebhookDefinitionFilter(
				ctx,
				path.Root("test"),
				value,
			)

			assert.Empty(t, result)
			assert.Empty(t, diags)
		})
	}
}

func TestToWebhookDefinitionFilterNotNil(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	testcases := map[string]WebhookFilterNotValue{
		"null":    NewWebhookFilterNotValueNull(),
		"unknown": NewWebhookFilterNotValueUnknown(),
	}

	for name, value := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result, diags := ToWebhookDefinitionFilterNot(
				ctx,
				path.Root("test"),
				value,
			)

			assert.Empty(t, result)
			assert.Empty(t, diags)
		})
	}
}

func TestToWebhookDefinitionFilterEqualsNil(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	testcases := map[string]WebhookFilterEqualsValue{
		"null":    NewWebhookFilterEqualsValueNull(),
		"unknown": NewWebhookFilterEqualsValueUnknown(),
	}

	for name, value := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result, diags := ToWebhookDefinitionFilterEquals(
				ctx,
				path.Root("test"),
				value,
			)

			assert.Empty(t, result)
			assert.Empty(t, diags)
		})
	}
}

func TestToWebhookDefinitionFilterInValue(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	testcases := map[string]WebhookFilterInValue{
		"null":    NewWebhookFilterInValueNull(),
		"unknown": NewWebhookFilterInValueUnknown(),
	}

	for name, value := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result, diags := ToWebhookDefinitionFilterIn(
				ctx,
				path.Root("test"),
				value,
			)

			assert.Empty(t, result)
			assert.Empty(t, diags)
		})
	}
}

func TestToWebhookDefinitionFilterRegexpValue(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	testcases := map[string]WebhookFilterRegexpValue{
		"null":    NewWebhookFilterRegexpValueNull(),
		"unknown": NewWebhookFilterRegexpValueUnknown(),
	}

	for name, value := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result, diags := ToWebhookDefinitionFilterRegexp(
				ctx,
				path.Root("test"),
				value,
			)

			assert.Empty(t, result)
			assert.Empty(t, diags)
		})
	}
}
