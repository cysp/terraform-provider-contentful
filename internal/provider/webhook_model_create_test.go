package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestWebhookModelToCreateWebhookDefinitionReq(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	filters := webhookFiltersListForTesting(t)

	testcases := map[string]struct {
		model     provider.WebhookModel
		expected  cm.CreateWebhookDefinitionReq
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
			expected: cm.CreateWebhookDefinitionReq{
				Name:              "test-webhook",
				Active:            cm.NewOptBool(true),
				URL:               "https://example.com/webhook",
				HttpBasicUsername: cm.NewOptNilStringNull(),
				HttpBasicPassword: cm.NewOptNilStringNull(),
				Topics:            []string{"Entry.create", "Entry.delete"},
				Filters:           cm.NewOptNilWebhookDefinitionFilterArrayNull(),
				Transformation:    cm.OptNilCreateWebhookDefinitionReqTransformation{Set: true, Null: true},
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
			expected: cm.CreateWebhookDefinitionReq{
				Name:              "auth-webhook",
				Active:            cm.NewOptBool(true),
				URL:               "https://example.com/webhook",
				HttpBasicUsername: cm.NewOptNilString("user"),
				HttpBasicPassword: cm.NewOptNilString("pass"),
				Topics:            []string{"Entry.*"},
				Filters:           cm.NewOptNilWebhookDefinitionFilterArrayNull(),
				Transformation:    cm.OptNilCreateWebhookDefinitionReqTransformation{Set: true, Null: true},
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
			expected: cm.CreateWebhookDefinitionReq{
				Name:              "headers-webhook",
				Active:            cm.NewOptBool(true),
				URL:               "https://example.com/webhook",
				HttpBasicUsername: cm.NewOptNilStringNull(),
				HttpBasicPassword: cm.NewOptNilStringNull(),
				Headers: []cm.WebhookDefinitionHeader{
					{
						Key:    "X-Header",
						Value:  cm.NewOptString("value"),
						Secret: cm.NewOptBool(false),
					},
				},
				Filters:        cm.NewOptNilWebhookDefinitionFilterArrayNull(),
				Transformation: cm.OptNilCreateWebhookDefinitionReqTransformation{Set: true, Null: true},
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
			expected: cm.CreateWebhookDefinitionReq{
				Name:              "headers-webhook",
				Active:            cm.NewOptBool(true),
				URL:               "https://example.com/webhook",
				HttpBasicUsername: cm.NewOptNilStringNull(),
				HttpBasicPassword: cm.NewOptNilStringNull(),
				Filters:           cm.NewOptNilWebhookDefinitionFilterArrayNull(),
				Transformation: cm.OptNilCreateWebhookDefinitionReqTransformation{Set: true, Value: cm.CreateWebhookDefinitionReqTransformation{
					Method:               cm.NewOptString("POST"),
					ContentType:          cm.NewOptString("application/json"),
					IncludeContentLength: cm.NewOptBool(true),
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
			expected: cm.CreateWebhookDefinitionReq{
				Name:              "headers-webhook",
				Active:            cm.NewOptBool(true),
				URL:               "https://example.com/webhook",
				HttpBasicUsername: cm.NewOptNilStringNull(),
				HttpBasicPassword: cm.NewOptNilStringNull(),
				Filters:           cm.NewOptNilWebhookDefinitionFilterArrayNull(),
				Transformation: cm.OptNilCreateWebhookDefinitionReqTransformation{Set: true, Value: cm.CreateWebhookDefinitionReqTransformation{
					Method:               cm.NewOptString("POST"),
					ContentType:          cm.NewOptString("application/json"),
					IncludeContentLength: cm.NewOptBool(true),
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
			expected: cm.CreateWebhookDefinitionReq{
				Name:              "filters-webhook",
				Active:            cm.NewOptBool(true),
				URL:               "https://example.com/webhook",
				HttpBasicUsername: cm.NewOptNilStringNull(),
				HttpBasicPassword: cm.NewOptNilStringNull(),
				Filters: cm.NewOptNilWebhookDefinitionFilterArray([]cm.WebhookDefinitionFilter{
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
				}),
				Transformation: cm.OptNilCreateWebhookDefinitionReqTransformation{Set: true, Null: true},
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
