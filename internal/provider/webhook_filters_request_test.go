package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestToOptNilWebhookDefinitionFilterArrayNil(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	testcases := map[string]struct {
		input types.List
	}{
		"null": {
			input: types.ListNull(types.StringType),
		},
		"unknown": {
			input: types.ListUnknown(types.StringType),
		},
	}

	expected := cm.NewOptNilWebhookDefinitionFilterArrayNull()

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result, diags := provider.ToOptNilWebhookDefinitionFilterArray(
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

	testcases := map[string]provider.WebhookFilterValue{
		"null":    provider.NewWebhookFilterValueNull(),
		"unknown": provider.NewWebhookFilterValueUnknown(),
	}

	for name, value := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result, diags := provider.ToWebhookDefinitionFilter(
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

	testcases := map[string]provider.WebhookFilterNotValue{
		"null":    provider.NewWebhookFilterNotValueNull(),
		"unknown": provider.NewWebhookFilterNotValueUnknown(),
	}

	for name, value := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result, diags := provider.ToWebhookDefinitionFilterNot(
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

	testcases := map[string]provider.WebhookFilterEqualsValue{
		"null":    provider.NewWebhookFilterEqualsValueNull(),
		"unknown": provider.NewWebhookFilterEqualsValueUnknown(),
	}

	for name, value := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result, diags := provider.ToWebhookDefinitionFilterEquals(
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

	testcases := map[string]provider.WebhookFilterInValue{
		"null":    provider.NewWebhookFilterInValueNull(),
		"unknown": provider.NewWebhookFilterInValueUnknown(),
	}

	for name, value := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result, diags := provider.ToWebhookDefinitionFilterIn(
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

	testcases := map[string]provider.WebhookFilterRegexpValue{
		"null":    provider.NewWebhookFilterRegexpValueNull(),
		"unknown": provider.NewWebhookFilterRegexpValueUnknown(),
	}

	for name, value := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result, diags := provider.ToWebhookDefinitionFilterRegexp(
				ctx,
				path.Root("test"),
				value,
			)

			assert.Empty(t, result)
			assert.Empty(t, diags)
		})
	}
}
