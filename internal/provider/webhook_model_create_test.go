package provider_test

import (
	"context"
	"testing"

	contentfulManagement "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestWebhookModelToCreateWebhookDefinitionReq(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	filters := webhookFiltersListForTesting(t)

	testcases := map[string]struct {
		model     provider.WebhookModel
		expected  contentfulManagement.CreateWebhookDefinitionReq
		expectErr bool
	}{
		"basic": {
			model: provider.WebhookModel{
				Name:              types.StringValue("test-webhook"),
				Active:            types.BoolValue(true),
				Url:               types.StringValue("https://example.com/webhook"),
				HttpBasicUsername: types.StringNull(),
				HttpBasicPassword: types.StringNull(),
				Topics: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("Entry.create"),
					types.StringValue("Entry.delete"),
				}),
			},
			expected: contentfulManagement.CreateWebhookDefinitionReq{
				Name:              "test-webhook",
				Active:            contentfulManagement.NewOptBool(true),
				URL:               "https://example.com/webhook",
				HttpBasicUsername: contentfulManagement.NewOptNilStringNull(),
				HttpBasicPassword: contentfulManagement.NewOptNilStringNull(),
				Topics:            []string{"Entry.create", "Entry.delete"},
				Filters:           contentfulManagement.NewOptNilWebhookDefinitionFilterArrayNull(),
				Transformation:    contentfulManagement.OptNilCreateWebhookDefinitionReqTransformation{Set: true, Null: true},
			},
			expectErr: false,
		},
		"with auth": {
			model: provider.WebhookModel{
				Name:              types.StringValue("auth-webhook"),
				Active:            types.BoolValue(true),
				Url:               types.StringValue("https://example.com/webhook"),
				HttpBasicUsername: types.StringValue("user"),
				HttpBasicPassword: types.StringValue("pass"),
				Topics: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("Entry.*"),
				}),
			},
			expected: contentfulManagement.CreateWebhookDefinitionReq{
				Name:              "auth-webhook",
				Active:            contentfulManagement.NewOptBool(true),
				URL:               "https://example.com/webhook",
				HttpBasicUsername: contentfulManagement.NewOptNilString("user"),
				HttpBasicPassword: contentfulManagement.NewOptNilString("pass"),
				Topics:            []string{"Entry.*"},
				Filters:           contentfulManagement.NewOptNilWebhookDefinitionFilterArrayNull(),
				Transformation:    contentfulManagement.OptNilCreateWebhookDefinitionReqTransformation{Set: true, Null: true},
			},
			expectErr: false,
		},
		"with headers": {
			model: provider.WebhookModel{
				Name:   types.StringValue("headers-webhook"),
				Active: types.BoolValue(true),
				Url:    types.StringValue("https://example.com/webhook"),
				Headers: types.MapValueMust(provider.HeadersValue{}.Type(ctx), map[string]attr.Value{
					"X-Header": provider.NewHeadersValueKnownFromAttributesMust(ctx, map[string]attr.Value{
						"value":  types.StringValue("value"),
						"secret": types.BoolValue(false),
					}),
				}),
			},
			expected: contentfulManagement.CreateWebhookDefinitionReq{
				Name:              "headers-webhook",
				Active:            contentfulManagement.NewOptBool(true),
				URL:               "https://example.com/webhook",
				HttpBasicUsername: contentfulManagement.NewOptNilStringNull(),
				HttpBasicPassword: contentfulManagement.NewOptNilStringNull(),
				Headers: []contentfulManagement.WebhookDefinitionHeader{
					{
						Key:    "X-Header",
						Value:  contentfulManagement.NewOptString("value"),
						Secret: contentfulManagement.NewOptBool(false),
					},
				},
				Filters:        contentfulManagement.NewOptNilWebhookDefinitionFilterArrayNull(),
				Transformation: contentfulManagement.OptNilCreateWebhookDefinitionReqTransformation{Set: true, Null: true},
			},
		},
		"with transformation": {
			model: provider.WebhookModel{
				Name:   types.StringValue("headers-webhook"),
				Active: types.BoolValue(true),
				Url:    types.StringValue("https://example.com/webhook"),
				Transformation: provider.NewTransformationValueMust(provider.TransformationValue{}.AttributeTypes(ctx), map[string]attr.Value{
					"method":                 types.StringValue("POST"),
					"content_type":           types.StringValue("application/json"),
					"include_content_length": types.BoolValue(true),
					"body":                   types.StringNull(),
				}),
			},
			expected: contentfulManagement.CreateWebhookDefinitionReq{
				Name:              "headers-webhook",
				Active:            contentfulManagement.NewOptBool(true),
				URL:               "https://example.com/webhook",
				HttpBasicUsername: contentfulManagement.NewOptNilStringNull(),
				HttpBasicPassword: contentfulManagement.NewOptNilStringNull(),
				Filters:           contentfulManagement.NewOptNilWebhookDefinitionFilterArrayNull(),
				Transformation: contentfulManagement.OptNilCreateWebhookDefinitionReqTransformation{Set: true, Value: contentfulManagement.CreateWebhookDefinitionReqTransformation{
					Method:               contentfulManagement.NewOptString("POST"),
					ContentType:          contentfulManagement.NewOptString("application/json"),
					IncludeContentLength: contentfulManagement.NewOptBool(true),
				}},
			},
		},
		"with transformation body": {
			model: provider.WebhookModel{
				Name:   types.StringValue("headers-webhook"),
				Active: types.BoolValue(true),
				Url:    types.StringValue("https://example.com/webhook"),
				Transformation: provider.NewTransformationValueMust(provider.TransformationValue{}.AttributeTypes(ctx), map[string]attr.Value{
					"method":                 types.StringValue("POST"),
					"content_type":           types.StringValue("application/json"),
					"include_content_length": types.BoolValue(true),
					"body":                   types.StringValue("{\"key\":\"value\"}"),
				}),
			},
			expected: contentfulManagement.CreateWebhookDefinitionReq{
				Name:              "headers-webhook",
				Active:            contentfulManagement.NewOptBool(true),
				URL:               "https://example.com/webhook",
				HttpBasicUsername: contentfulManagement.NewOptNilStringNull(),
				HttpBasicPassword: contentfulManagement.NewOptNilStringNull(),
				Filters:           contentfulManagement.NewOptNilWebhookDefinitionFilterArrayNull(),
				Transformation: contentfulManagement.OptNilCreateWebhookDefinitionReqTransformation{Set: true, Value: contentfulManagement.CreateWebhookDefinitionReqTransformation{
					Method:               contentfulManagement.NewOptString("POST"),
					ContentType:          contentfulManagement.NewOptString("application/json"),
					IncludeContentLength: contentfulManagement.NewOptBool(true),
					Body:                 []byte("{\"key\":\"value\"}"),
				}},
			},
		},
		"with filters": {
			model: provider.WebhookModel{
				Name:              types.StringValue("filters-webhook"),
				Active:            types.BoolValue(true),
				Url:               types.StringValue("https://example.com/webhook"),
				HttpBasicUsername: types.StringNull(),
				HttpBasicPassword: types.StringNull(),
				Filters:           filters,
			},
			expected: contentfulManagement.CreateWebhookDefinitionReq{
				Name:              "filters-webhook",
				Active:            contentfulManagement.NewOptBool(true),
				URL:               "https://example.com/webhook",
				HttpBasicUsername: contentfulManagement.NewOptNilStringNull(),
				HttpBasicPassword: contentfulManagement.NewOptNilStringNull(),
				Filters: contentfulManagement.NewOptNilWebhookDefinitionFilterArray([]contentfulManagement.WebhookDefinitionFilter{
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
				}),
				Transformation: contentfulManagement.OptNilCreateWebhookDefinitionReqTransformation{Set: true, Null: true},
			},
			expectErr: false,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := testcase.model.ToCreateWebhookDefinitionReq(ctx, path.Empty())

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
