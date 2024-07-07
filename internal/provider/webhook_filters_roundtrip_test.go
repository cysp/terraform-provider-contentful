package provider_test

import (
	"context"
	"testing"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/stretchr/testify/assert"
)

func TestWebhookFiltersRoundtrip(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	filters := webhookFiltersListForTesting(t)

	webhookDefinitionFilterArray, webhookDefinitionFilterArrayDiags := provider.ToOptNilWebhookDefinitionFilterArray(ctx, path.Root("filters"), filters)
	assert.Empty(t, webhookDefinitionFilterArrayDiags)

	assert.EqualValues(t, webhookDefinitionFilterArray, contentfulManagement.NewOptNilWebhookDefinitionFilterArray([]contentfulManagement.WebhookDefinitionFilter{
		{
			Equals: contentfulManagement.WebhookDefinitionFilterEquals{[]byte(`{"doc":"sys.type"}`), []byte(`"abc"`)},
		},
		{
			In: contentfulManagement.WebhookDefinitionFilterIn{[]byte(`{"doc":"sys.type"}`), []byte(`["abc","def"]`)},
		},
		{
			Regexp: contentfulManagement.WebhookDefinitionFilterRegexp{[]byte(`{"doc":"sys.type"}`), []byte(`{"pattern":"abc.*"}`)},
		},
		{
			Not: contentfulManagement.NewOptWebhookDefinitionFilterNot(contentfulManagement.WebhookDefinitionFilterNot{
				Equals: contentfulManagement.WebhookDefinitionFilterEquals{[]byte(`{"doc":"sys.type"}`), []byte(`"abc"`)},
			}),
		},
		{
			Not: contentfulManagement.NewOptWebhookDefinitionFilterNot(contentfulManagement.WebhookDefinitionFilterNot{
				In: contentfulManagement.WebhookDefinitionFilterIn{[]byte(`{"doc":"sys.type"}`), []byte(`["abc","def"]`)},
			}),
		},
		{
			Not: contentfulManagement.NewOptWebhookDefinitionFilterNot(contentfulManagement.WebhookDefinitionFilterNot{
				Regexp: contentfulManagement.WebhookDefinitionFilterRegexp{[]byte(`{"doc":"sys.type"}`), []byte(`{"pattern":"abc.*"}`)},
			}),
		},
	}))

	filterValuesList, filterValuesListDiags := provider.ReadWebhookFiltersListValueFromResponse(ctx, path.Root("filters"), webhookDefinitionFilterArray)
	assert.Empty(t, filterValuesListDiags)

	assert.True(t, filters.Equal(filterValuesList))
}
