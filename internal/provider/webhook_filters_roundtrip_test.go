package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/stretchr/testify/assert"
)

//nolint:dupl // Keep independently authored fixtures and expectations explicit.
func TestWebhookFiltersRoundTrip(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	filters := webhookFiltersListForTesting(t)

	webhookDefinitionFilterArray, webhookDefinitionFilterArrayDiags := ToOptNilWebhookDefinitionFilterArray(ctx, path.Root("filters"), filters)
	assert.Empty(t, webhookDefinitionFilterArrayDiags)

	assert.Equal(t, webhookDefinitionFilterArray, cm.NewOptNilWebhookDefinitionFilterArray([]cm.WebhookDefinitionFilter{
		{
			Equals: cm.WebhookDefinitionFilterEquals{[]byte(testJSON(map[string]any{"doc": "sys.type"})), []byte(testJSON("abc"))},
		},
		{
			In: cm.WebhookDefinitionFilterIn{[]byte(testJSON(map[string]any{"doc": "sys.type"})), []byte(testJSON([]any{"abc", "def"}))},
		},
		{
			Regexp: cm.WebhookDefinitionFilterRegexp{[]byte(testJSON(map[string]any{"doc": "sys.type"})), []byte(testJSON(map[string]any{"pattern": "abc.*"}))},
		},
		{
			Not: cm.NewOptWebhookDefinitionFilterNot(cm.WebhookDefinitionFilterNot{
				Equals: cm.WebhookDefinitionFilterEquals{[]byte(testJSON(map[string]any{"doc": "sys.type"})), []byte(testJSON("abc"))},
			}),
		},
		{
			Not: cm.NewOptWebhookDefinitionFilterNot(cm.WebhookDefinitionFilterNot{
				In: cm.WebhookDefinitionFilterIn{[]byte(testJSON(map[string]any{"doc": "sys.type"})), []byte(testJSON([]any{"abc", "def"}))},
			}),
		},
		{
			Not: cm.NewOptWebhookDefinitionFilterNot(cm.WebhookDefinitionFilterNot{
				Regexp: cm.WebhookDefinitionFilterRegexp{[]byte(testJSON(map[string]any{"doc": "sys.type"})), []byte(testJSON(map[string]any{"pattern": "abc.*"}))},
			}),
		},
	}))

	filterValuesList, filterValuesListDiags := ReadWebhookFiltersListValueFromResponse(ctx, path.Root("filters"), webhookDefinitionFilterArray)
	assert.Empty(t, filterValuesListDiags)

	assert.True(t, filters.Equal(filterValuesList))
}
