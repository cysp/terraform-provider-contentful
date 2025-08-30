package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/stretchr/testify/assert"
)

func TestWebhookFiltersRoundtrip(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	filters := webhookFiltersListForTesting(t)

	webhookDefinitionFilterArray, webhookDefinitionFilterArrayDiags := ToOptNilWebhookDefinitionFilterArray(ctx, path.Root("filters"), filters)
	assert.Empty(t, webhookDefinitionFilterArrayDiags)

	assert.Equal(t, webhookDefinitionFilterArray, cm.NewOptNilWebhookDefinitionFilterArray([]cm.WebhookDefinitionFilter{
		{
			Equals: cm.WebhookDefinitionFilterEquals{[]byte(`{"doc":"sys.type"}`), []byte(`"abc"`)},
		},
		{
			In: cm.WebhookDefinitionFilterIn{[]byte(`{"doc":"sys.type"}`), []byte(`["abc","def"]`)},
		},
		{
			Regexp: cm.WebhookDefinitionFilterRegexp{[]byte(`{"doc":"sys.type"}`), []byte(`{"pattern":"abc.*"}`)},
		},
		{
			Not: cm.NewOptWebhookDefinitionFilterNot(cm.WebhookDefinitionFilterNot{
				Equals: cm.WebhookDefinitionFilterEquals{[]byte(`{"doc":"sys.type"}`), []byte(`"abc"`)},
			}),
		},
		{
			Not: cm.NewOptWebhookDefinitionFilterNot(cm.WebhookDefinitionFilterNot{
				In: cm.WebhookDefinitionFilterIn{[]byte(`{"doc":"sys.type"}`), []byte(`["abc","def"]`)},
			}),
		},
		{
			Not: cm.NewOptWebhookDefinitionFilterNot(cm.WebhookDefinitionFilterNot{
				Regexp: cm.WebhookDefinitionFilterRegexp{[]byte(`{"doc":"sys.type"}`), []byte(`{"pattern":"abc.*"}`)},
			}),
		},
	}))

	filterValuesList, filterValuesListDiags := ReadWebhookFiltersListValueFromResponse(ctx, path.Root("filters"), webhookDefinitionFilterArray)
	assert.Empty(t, filterValuesListDiags)

	assert.True(t, filters.Equal(filterValuesList))
}
