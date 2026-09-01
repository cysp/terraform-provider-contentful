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
	"github.com/stretchr/testify/require"
)

func TestWebhookTransformationRequestValues(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	valuePath := path.Root("transformation")

	known := func(values map[string]attr.Value) TypedObject[WebhookTransformationValue] {
		t.Helper()

		return DiagsNoErrorsMust(NewTypedObjectFromAttributes[WebhookTransformationValue](ctx, values))
	}

	nullChildren := map[string]attr.Value{
		"method":                 types.StringNull(),
		"content_type":           types.StringNull(),
		"include_content_length": types.BoolNull(),
		"body":                   jsontypes.NewNormalizedNull(),
	}

	tests := map[string]struct {
		value         TypedObject[WebhookTransformationValue]
		expected      cm.OptNilWebhookDefinitionDataTransformation
		expectedPaths []string
	}{
		"null is explicit null": {
			value:    NewTypedObjectNull[WebhookTransformationValue](),
			expected: cm.OptNilWebhookDefinitionDataTransformation{Set: true, Null: true},
		},
		"unknown object": {
			value:         NewTypedObjectUnknown[WebhookTransformationValue](),
			expectedPaths: []string{"transformation"},
		},
		"null children are omitted": {
			value: known(nullChildren),
			expected: cm.NewOptNilWebhookDefinitionDataTransformation(
				cm.WebhookDefinitionDataTransformation{},
			),
		},
		"known": {
			value: known(map[string]attr.Value{
				"method":                 types.StringValue("POST"),
				"content_type":           types.StringValue("application/json"),
				"include_content_length": types.BoolValue(true),
				"body":                   NewNormalizedJSONValue([]byte(`{"key":"value"}`)),
			}),
			expected: cm.NewOptNilWebhookDefinitionDataTransformation(
				cm.WebhookDefinitionDataTransformation{
					Method:               cm.NewOptString("POST"),
					ContentType:          cm.NewOptString("application/json"),
					IncludeContentLength: cm.NewOptBool(true),
					Body:                 []byte(`{"key":"value"}`),
				},
			),
		},
		"unknown method": {
			value: known(map[string]attr.Value{
				"method":                 types.StringUnknown(),
				"content_type":           types.StringNull(),
				"include_content_length": types.BoolNull(),
				"body":                   jsontypes.NewNormalizedNull(),
			}),
			expectedPaths: []string{"transformation.method"},
		},
		"unknown content type": {
			value: known(map[string]attr.Value{
				"method":                 types.StringNull(),
				"content_type":           types.StringUnknown(),
				"include_content_length": types.BoolNull(),
				"body":                   jsontypes.NewNormalizedNull(),
			}),
			expectedPaths: []string{"transformation.content_type"},
		},
		"unknown content length": {
			value: known(map[string]attr.Value{
				"method":                 types.StringNull(),
				"content_type":           types.StringNull(),
				"include_content_length": types.BoolUnknown(),
				"body":                   jsontypes.NewNormalizedNull(),
			}),
			expectedPaths: []string{"transformation.include_content_length"},
		},
		"unknown body": {
			value: known(map[string]attr.Value{
				"method":                 types.StringNull(),
				"content_type":           types.StringNull(),
				"include_content_length": types.BoolNull(),
				"body":                   jsontypes.NewNormalizedUnknown(),
			}),
			expectedPaths: []string{"transformation.body"},
		},
		"multiple unknown children fail atomically": {
			value: known(map[string]attr.Value{
				"method":                 types.StringValue("POST"),
				"content_type":           types.StringUnknown(),
				"include_content_length": types.BoolUnknown(),
				"body":                   jsontypes.NewNormalizedUnknown(),
			}),
			expectedPaths: []string{
				"transformation.content_type",
				"transformation.include_content_length",
				"transformation.body",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, diags := ToOptNilWebhookDefinitionDataTransformation(valuePath, test.value)

			if len(test.expectedPaths) > 0 {
				require.True(t, diags.HasError())
				assert.Equal(t, test.expectedPaths, attributeDiagnosticPaths(t, diags))
				assert.Equal(t, cm.OptNilWebhookDefinitionDataTransformation{}, actual)
			} else {
				require.False(t, diags.HasError())
				assert.Equal(t, test.expected, actual)
			}
		})
	}
}

func TestWebhookModelFailsClosedForNestedRequestError(t *testing.T) {
	t.Parallel()

	model := validWebhookRequestModel()
	model.Transformation = DiagsNoErrorsMust(NewTypedObjectFromAttributes[WebhookTransformationValue](t.Context(), map[string]attr.Value{
		"method":                 types.StringValue("POST"),
		"content_type":           types.StringUnknown(),
		"include_content_length": types.BoolNull(),
		"body":                   jsontypes.NewNormalizedNull(),
	}))

	actual, diags := model.ToWebhookDefinitionData(t.Context(), WebhookModel{}, path.Empty())

	require.True(t, diags.HasError())
	assert.Equal(t, []string{"transformation.content_type"}, attributeDiagnosticPaths(t, diags))
	assert.Equal(t, cm.WebhookDefinitionData{}, actual)
}
