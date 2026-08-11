package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookModelRejectsUnresolvedScalarRequestValues(t *testing.T) {
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

			actual, diags := model.ToWebhookDefinitionData(t.Context(), WebhookModel{}, path.Empty())

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

			actual, diags := model.ToWebhookDefinitionData(t.Context(), WebhookModel{}, path.Empty())

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

	ctx := t.Context()

	validHeader := DiagsNoErrorsMust(NewTypedObjectFromAttributes[WebhookHeaderValue](ctx, map[string]attr.Value{
		"value":    types.StringValue("value"),
		"value_wo": types.StringNull(),
		"secret":   types.BoolValue(false),
	}))
	unknownValueHeader := DiagsNoErrorsMust(NewTypedObjectFromAttributes[WebhookHeaderValue](ctx, map[string]attr.Value{
		"value":    types.StringUnknown(),
		"value_wo": types.StringNull(),
		"secret":   types.BoolValue(false),
	}))
	unknownSecretHeader := DiagsNoErrorsMust(NewTypedObjectFromAttributes[WebhookHeaderValue](ctx, map[string]attr.Value{
		"value":    types.StringValue("value"),
		"value_wo": types.StringNull(),
		"secret":   types.BoolUnknown(),
	}))
	nullValueHeader := DiagsNoErrorsMust(NewTypedObjectFromAttributes[WebhookHeaderValue](ctx, map[string]attr.Value{
		"value":    types.StringNull(),
		"value_wo": types.StringNull(),
		"secret":   types.BoolValue(false),
	}))
	nullSecretHeader := DiagsNoErrorsMust(NewTypedObjectFromAttributes[WebhookHeaderValue](ctx, map[string]attr.Value{
		"value":    types.StringValue("value"),
		"value_wo": types.StringNull(),
		"secret":   types.BoolNull(),
	}))

	tests := map[string]struct {
		headers       TypedMap[TypedObject[WebhookHeaderValue]]
		expected      cm.WebhookDefinitionHeaders
		expectedPaths []string
	}{
		"null is omitted": {
			headers: NewTypedMapNull[TypedObject[WebhookHeaderValue]](),
		},
		"computed unknown is omitted": {
			headers: NewTypedMapUnknown[TypedObject[WebhookHeaderValue]](),
		},
		"empty is preserved": {
			headers:  NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{}),
			expected: cm.WebhookDefinitionHeaders{},
		},
		"known values are sorted": {
			headers: NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{
				"Z-Header": validHeader,
				"A-Header": validHeader,
			}),
			expected: cm.WebhookDefinitionHeaders{
				{Key: "A-Header", Value: cm.NewOptString("value"), Secret: cm.NewOptBool(false)},
				{Key: "Z-Header", Value: cm.NewOptString("value"), Secret: cm.NewOptBool(false)},
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
				"X-Header": unknownValueHeader,
			}),
			expectedPaths: []string{"headers[\"X-Header\"].value"},
		},
		"null value": {
			headers: NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{
				"X-Header": nullValueHeader,
			}),
			expectedPaths: []string{"headers[\"X-Header\"].value"},
		},
		"null secret": {
			headers: NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{
				"X-Header": nullSecretHeader,
			}),
			expectedPaths: []string{"headers[\"X-Header\"].secret"},
		},
		"mixed valid and unknown values fail atomically": {
			headers: NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{
				"A-Header": validHeader,
				"X-Header": unknownSecretHeader,
			}),
			expectedPaths: []string{"headers[\"X-Header\"].secret"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, diags := ToWebhookDefinitionHeaders(path.Root("headers"), test.headers, test.headers)

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
