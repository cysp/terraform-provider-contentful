package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookModelRejectsInvalidScalarRequestValues(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		setValue     func(*WebhookModel)
		expectedPath string
	}{
		"unknown name": {
			setValue:     func(model *WebhookModel) { model.Name = types.StringUnknown() },
			expectedPath: "name",
		},
		"null name": {
			setValue:     func(model *WebhookModel) { model.Name = types.StringNull() },
			expectedPath: "name",
		},
		"unknown URL": {
			setValue:     func(model *WebhookModel) { model.URL = types.StringUnknown() },
			expectedPath: "url",
		},
		"null URL": {
			setValue:     func(model *WebhookModel) { model.URL = types.StringNull() },
			expectedPath: "url",
		},
		"unknown active": {
			setValue:     func(model *WebhookModel) { model.Active = types.BoolUnknown() },
			expectedPath: "active",
		},
		"null active": {
			setValue:     func(model *WebhookModel) { model.Active = types.BoolNull() },
			expectedPath: "active",
		},
		"missing password": {
			setValue:     func(model *WebhookModel) { model.HTTPBasicUsername = types.StringValue("user") },
			expectedPath: "http_basic_password",
		},
		"missing username": {
			setValue:     func(model *WebhookModel) { model.HTTPBasicPassword = types.StringValue("password") },
			expectedPath: "http_basic_username",
		},
		"unknown username": {
			setValue:     func(model *WebhookModel) { model.HTTPBasicUsername = types.StringUnknown() },
			expectedPath: "http_basic_username",
		},
		"unknown password": {
			setValue:     func(model *WebhookModel) { model.HTTPBasicPassword = types.StringUnknown() },
			expectedPath: "http_basic_password",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			model := validWebhookRequestModel()
			test.setValue(&model)

			actual, diags := model.ToWebhookDefinitionData(t.Context(), path.Empty())

			require.True(t, diags.HasError())
			assert.Equal(t, []string{test.expectedPath}, attributeDiagnosticPaths(t, diags))
			assert.Equal(t, cm.WebhookDefinitionData{}, actual)
		})
	}
}

func TestWebhookModelTopicsRequestValues(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		topics        TypedList[types.String]
		expected      []string
		expectedPaths []string
	}{
		"null is omitted": {
			topics: NewTypedListNull[types.String](),
		},
		"empty is preserved": {
			topics:   NewTypedList([]types.String{}),
			expected: []string{},
		},
		"known": {
			topics: NewTypedList([]types.String{
				types.StringValue("Entry.create"),
				types.StringValue("Entry.delete"),
			}),
			expected: []string{"Entry.create", "Entry.delete"},
		},
		"unknown list": {
			topics:        NewTypedListUnknown[types.String](),
			expectedPaths: []string{"topics"},
		},
		"unknown and null children fail atomically": {
			topics: NewTypedList([]types.String{
				types.StringValue("Entry.create"),
				types.StringUnknown(),
				types.StringNull(),
			}),
			expectedPaths: []string{"topics[1]", "topics[2]"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			model := validWebhookRequestModel()
			model.Topics = test.topics

			actual, diags := model.ToWebhookDefinitionData(t.Context(), path.Empty())

			if len(test.expectedPaths) > 0 {
				require.True(t, diags.HasError())
				assert.Equal(t, test.expectedPaths, attributeDiagnosticPaths(t, diags))
				assert.Equal(t, cm.WebhookDefinitionData{}, actual)
			} else {
				require.False(t, diags.HasError())
				assert.Equal(t, test.expected, actual.Topics)
			}
		})
	}
}

func TestWebhookHeadersRequestValues(t *testing.T) {
	t.Parallel()

	header := func(value types.String, secret types.Bool) TypedObject[WebhookHeaderValue] {
		return NewTypedObject(WebhookHeaderValue{Value: value, Secret: secret})
	}

	tests := map[string]struct {
		headers       TypedMap[TypedObject[WebhookHeaderValue]]
		expected      cm.WebhookDefinitionHeaders
		expectedPaths []string
	}{
		"null is omitted": {
			headers: NewTypedMapNull[TypedObject[WebhookHeaderValue]](),
		},
		"unknown is rejected": {
			headers:       NewTypedMapUnknown[TypedObject[WebhookHeaderValue]](),
			expectedPaths: []string{"headers"},
		},
		"empty is preserved": {
			headers:  NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{}),
			expected: cm.WebhookDefinitionHeaders{},
		},
		"sorted keys preserve empty and omitted values": {
			headers: NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{
				"X-Retained":     header(types.StringNull(), types.BoolValue(true)),
				"X-Secret":       header(types.StringValue("token"), types.BoolValue(true)),
				"X-Empty-Secret": header(types.StringValue(""), types.BoolValue(true)),
				"X-Empty":        header(types.StringValue(""), types.BoolValue(false)),
			}),
			expected: cm.WebhookDefinitionHeaders{
				{Key: "X-Empty", Value: cm.NewOptString(""), Secret: cm.NewOptBool(false)},
				{Key: "X-Empty-Secret", Value: cm.NewOptString(""), Secret: cm.NewOptBool(true)},
				{Key: "X-Retained", Secret: cm.NewOptBool(true)},
				{Key: "X-Secret", Value: cm.NewOptString("token"), Secret: cm.NewOptBool(true)},
			},
		},
		"unknown object": {
			headers: NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{
				"X-Header": NewTypedObjectUnknown[WebhookHeaderValue](),
			}),
			expectedPaths: []string{"headers[\"X-Header\"]"},
		},
		"null object": {
			headers: NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{
				"X-Header": NewTypedObjectNull[WebhookHeaderValue](),
			}),
			expectedPaths: []string{"headers[\"X-Header\"]"},
		},
		"unknown value": {
			headers: NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{
				"X-Header": header(types.StringUnknown(), types.BoolValue(false)),
			}),
			expectedPaths: []string{"headers[\"X-Header\"].value"},
		},
		"null value": {
			headers: NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{
				"X-Header": header(types.StringNull(), types.BoolValue(false)),
			}),
			expectedPaths: []string{"headers[\"X-Header\"].value"},
		},
		"null secret": {
			headers: NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{
				"X-Header": header(types.StringValue("value"), types.BoolNull()),
			}),
			expectedPaths: []string{"headers[\"X-Header\"].secret"},
		},
		"mixed valid and unknown values fail atomically": {
			headers: NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{
				"A-Header": header(types.StringValue("valid"), types.BoolValue(false)),
				"X-Header": header(types.StringValue("value"), types.BoolUnknown()),
			}),
			expectedPaths: []string{"headers[\"X-Header\"].secret"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, diags := ToWebhookDefinitionHeaders(path.Root("headers"), test.headers)

			if len(test.expectedPaths) > 0 {
				require.True(t, diags.HasError())
				assert.Equal(t, test.expectedPaths, attributeDiagnosticPaths(t, diags))
				assert.Nil(t, actual)
			} else {
				require.False(t, diags.HasError())
				assert.Equal(t, test.expected, actual)
			}
		})
	}
}

func validWebhookRequestModel() WebhookModel {
	return WebhookModel{
		Name:              types.StringValue("webhook"),
		URL:               types.StringValue("https://example.com/webhook"),
		Active:            types.BoolValue(true),
		HTTPBasicUsername: types.StringNull(),
		HTTPBasicPassword: types.StringNull(),
		Topics:            NewTypedListNull[types.String](),
		Filters:           NewTypedListNull[TypedObject[WebhookFilterValue]](),
		Headers:           NewTypedMapNull[TypedObject[WebhookHeaderValue]](),
		Transformation:    NewTypedObjectNull[WebhookTransformationValue](),
	}
}
