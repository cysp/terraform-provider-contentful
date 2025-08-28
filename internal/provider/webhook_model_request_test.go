package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestWebhookModelToWebhookDefinitionFields(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	filters := webhookFiltersListForTesting(t)

	testcases := map[string]struct {
		model     WebhookModel
		expected  cm.WebhookDefinitionFields
		expectErr bool
	}{
		"basic": {
			model: WebhookModel{
				Name:              types.StringValue("test-webhook"),
				Active:            types.BoolValue(true),
				URL:               types.StringValue("https://example.com/webhook"),
				HTTPBasicUsername: types.StringNull(),
				HTTPBasicPassword: types.StringNull(),
				Topics: NewTypedList([]types.String{
					types.StringValue("Entry.create"),
					types.StringValue("Entry.delete"),
				}),
			},
			expected: cm.WebhookDefinitionFields{
				Name:              "test-webhook",
				Active:            cm.NewOptBool(true),
				URL:               "https://example.com/webhook",
				HttpBasicUsername: cm.NewOptNilStringNull(),
				HttpBasicPassword: cm.NewOptNilStringNull(),
				Topics:            []string{"Entry.create", "Entry.delete"},
				Filters:           cm.NewOptNilWebhookDefinitionFilterArrayNull(),
				Transformation:    cm.OptNilWebhookDefinitionFieldsTransformation{Set: true, Null: true},
			},
			expectErr: false,
		},
		"with auth": {
			model: WebhookModel{
				Name:              types.StringValue("auth-webhook"),
				Active:            types.BoolValue(true),
				URL:               types.StringValue("https://example.com/webhook"),
				HTTPBasicUsername: types.StringValue("user"),
				HTTPBasicPassword: types.StringValue("pass"),
				Topics: NewTypedList([]types.String{
					types.StringValue("Entry.*"),
				}),
			},
			expected: cm.WebhookDefinitionFields{
				Name:              "auth-webhook",
				Active:            cm.NewOptBool(true),
				URL:               "https://example.com/webhook",
				HttpBasicUsername: cm.NewOptNilString("user"),
				HttpBasicPassword: cm.NewOptNilString("pass"),
				Topics:            []string{"Entry.*"},
				Filters:           cm.NewOptNilWebhookDefinitionFilterArrayNull(),
				Transformation:    cm.OptNilWebhookDefinitionFieldsTransformation{Set: true, Null: true},
			},
			expectErr: false,
		},
		"with headers": {
			model: WebhookModel{
				Name:   types.StringValue("headers-webhook"),
				Active: types.BoolValue(true),
				URL:    types.StringValue("https://example.com/webhook"),
				Headers: NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{
					"X-Header": DiagsNoErrorsMust(NewTypedObjectFromAttributes[WebhookHeaderValue](ctx, map[string]attr.Value{
						"value":  types.StringValue("value"),
						"secret": types.BoolValue(false),
					})),
				}),
			},
			expected: cm.WebhookDefinitionFields{
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
				Transformation: cm.OptNilWebhookDefinitionFieldsTransformation{Set: true, Null: true},
			},
		},
		"with transformation": {
			model: WebhookModel{
				Name:   types.StringValue("headers-webhook"),
				Active: types.BoolValue(true),
				URL:    types.StringValue("https://example.com/webhook"),
				Transformation: DiagsNoErrorsMust(NewTypedObjectFromAttributes[WebhookTransformationValue](ctx, map[string]attr.Value{
					"method":                 types.StringValue("POST"),
					"content_type":           types.StringValue("application/json"),
					"include_content_length": types.BoolValue(true),
					"body":                   jsontypes.NewNormalizedNull(),
				})),
			},
			expected: cm.WebhookDefinitionFields{
				Name:              "headers-webhook",
				Active:            cm.NewOptBool(true),
				URL:               "https://example.com/webhook",
				HttpBasicUsername: cm.NewOptNilStringNull(),
				HttpBasicPassword: cm.NewOptNilStringNull(),
				Filters:           cm.NewOptNilWebhookDefinitionFilterArrayNull(),
				Transformation: cm.OptNilWebhookDefinitionFieldsTransformation{Set: true, Value: cm.WebhookDefinitionFieldsTransformation{
					Method:               cm.NewOptString("POST"),
					ContentType:          cm.NewOptString("application/json"),
					IncludeContentLength: cm.NewOptBool(true),
				}},
			},
		},
		"with transformation body": {
			model: WebhookModel{
				Name:   types.StringValue("headers-webhook"),
				Active: types.BoolValue(true),
				URL:    types.StringValue("https://example.com/webhook"),
				Transformation: DiagsNoErrorsMust(NewTypedObjectFromAttributes[WebhookTransformationValue](ctx, map[string]attr.Value{
					"method":                 types.StringValue("POST"),
					"content_type":           types.StringValue("application/json"),
					"include_content_length": types.BoolValue(true),
					"body":                   jsontypes.NewNormalizedValue("{\"key\":\"value\"}"),
				})),
			},
			expected: cm.WebhookDefinitionFields{
				Name:              "headers-webhook",
				Active:            cm.NewOptBool(true),
				URL:               "https://example.com/webhook",
				HttpBasicUsername: cm.NewOptNilStringNull(),
				HttpBasicPassword: cm.NewOptNilStringNull(),
				Filters:           cm.NewOptNilWebhookDefinitionFilterArrayNull(),
				Transformation: cm.OptNilWebhookDefinitionFieldsTransformation{Set: true, Value: cm.WebhookDefinitionFieldsTransformation{
					Method:               cm.NewOptString("POST"),
					ContentType:          cm.NewOptString("application/json"),
					IncludeContentLength: cm.NewOptBool(true),
					Body:                 []byte("{\"key\":\"value\"}"),
				}},
			},
		},
		"with filters": {
			model: WebhookModel{
				Name:              types.StringValue("filters-webhook"),
				Active:            types.BoolValue(true),
				URL:               types.StringValue("https://example.com/webhook"),
				HTTPBasicUsername: types.StringNull(),
				HTTPBasicPassword: types.StringNull(),
				Filters:           filters,
			},
			expected: cm.WebhookDefinitionFields{
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
				Transformation: cm.OptNilWebhookDefinitionFieldsTransformation{Set: true, Null: true},
			},
			expectErr: false,
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := testcase.model.ToWebhookDefinitionFields(ctx, path.Empty())

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
