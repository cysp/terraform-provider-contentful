package provider_test

import (
	"context"
	"testing"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider"
	webhookfilter "github.com/cysp/terraform-provider-contentful/internal/provider/webhook_filter"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestWebhookModelToCreateWebhookDefinitionReq(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	filterEquals := webhookfilter.NewWebhookFilterEqualsValueKnown()
	filterEquals.Doc = types.StringValue("sys.type")
	filterEquals.Value = types.StringValue("abc")

	filterNot := webhookfilter.NewWebhookFilterNotValueKnown()
	filterNot.Equals = filterEquals

	filter := webhookfilter.NewWebhookFilterValueKnown()
	// filter.Equals = filterEquals
	filter.Not = filterNot

	filters, filtersDiags := types.ListValueFrom(ctx, webhookfilter.WebhookFilterValue{}.CustomType(ctx), []attr.Value{
		filter,
	})

	assert.Empty(t, filtersDiags)
	// assert.False(t, filtersDiags.HasError())

	testcases := map[string]struct {
		model     provider.WebhookModel
		expected  contentfulManagement.CreateWebhookDefinitionReq
		expectErr bool
	}{
		// "basic": {
		// 	model: provider.WebhookModel{
		// 		Name:              types.StringValue("test-webhook"),
		// 		Active:            types.BoolValue(true),
		// 		Url:               types.StringValue("https://example.com/webhook"),
		// 		HttpBasicUsername: types.StringNull(),
		// 		HttpBasicPassword: types.StringNull(),
		// 		Topics: types.ListValueMust(types.StringType, []attr.Value{
		// 			types.StringValue("Entry.create"),
		// 			types.StringValue("Entry.delete"),
		// 		}),
		// 	},
		// 	expected: contentfulManagement.CreateWebhookDefinitionReq{
		// 		Name:              "test-webhook",
		// 		Active:            contentfulManagement.NewOptBool(true),
		// 		URL:               "https://example.com/webhook",
		// 		Headers:           contentfulManagement.WebhookDefinitionHeaders{},
		// 		HttpBasicUsername: contentfulManagement.NewOptNilStringNull(),
		// 		HttpBasicPassword: contentfulManagement.NewOptNilStringNull(),
		// 		Topics:            []string{"Entry.create", "Entry.delete"},
		// 		Filters:           contentfulManagement.NewOptNilWebhookDefinitionFilterArrayNull(),
		// 	},
		// 	expectErr: false,
		// },
		// "with auth": {
		// 	model: provider.WebhookModel{
		// 		Name:              types.StringValue("auth-webhook"),
		// 		Active:            types.BoolValue(true),
		// 		Url:               types.StringValue("https://example.com/webhook"),
		// 		HttpBasicUsername: types.StringValue("user"),
		// 		HttpBasicPassword: types.StringValue("pass"),
		// 		Topics: types.ListValueMust(types.StringType, []attr.Value{
		// 			types.StringValue("Entry.*"),
		// 		}),
		// 	},
		// 	expected: contentfulManagement.CreateWebhookDefinitionReq{
		// 		Name:              "auth-webhook",
		// 		Active:            contentfulManagement.NewOptBool(true),
		// 		URL:               "https://example.com/webhook",
		// 		Headers:           contentfulManagement.WebhookDefinitionHeaders{},
		// 		HttpBasicUsername: contentfulManagement.NewOptNilString("user"),
		// 		HttpBasicPassword: contentfulManagement.NewOptNilString("pass"),
		// 		Topics:            []string{"Entry.*"},
		// 		Filters:           contentfulManagement.NewOptNilWebhookDefinitionFilterArrayNull(),
		// 	},
		// 	expectErr: false,
		// },
		"with filters": {
			model: provider.WebhookModel{
				Name:              types.StringValue("auth-webhook"),
				Active:            types.BoolValue(true),
				Url:               types.StringValue("https://example.com/webhook"),
				HttpBasicUsername: types.StringNull(),
				HttpBasicPassword: types.StringNull(),
				Filters:           filters,
			},
			expected: contentfulManagement.CreateWebhookDefinitionReq{
				Name:              "auth-webhook",
				Active:            contentfulManagement.NewOptBool(true),
				URL:               "https://example.com/webhook",
				Headers:           contentfulManagement.WebhookDefinitionHeaders{},
				HttpBasicUsername: contentfulManagement.NewOptNilStringNull(),
				HttpBasicPassword: contentfulManagement.NewOptNilStringNull(),
				Filters: contentfulManagement.NewOptNilWebhookDefinitionFilterArray([]contentfulManagement.WebhookDefinitionFilter{
					[]byte(`{"equals":{"doc":"sys.type","value":"abc"}}`),
				}),
			},
			expectErr: false,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := testcase.model.ToCreateWebhookDefinitionReq(ctx)

			if testcase.expectErr {
				assert.True(t, diags.HasError())
			} else {
				assert.Empty(t, diags)
				assert.False(t, diags.HasError())
				assert.Equal(t, testcase.expected, got)
			}
		})
	}
}
