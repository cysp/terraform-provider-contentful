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

func TestWebhookHeadersUnknownContainerIsOmitted(t *testing.T) {
	t.Parallel()

	headers, diags := ToWebhookDefinitionHeaders(
		path.Root("headers"),
		NewTypedMapUnknown[TypedObject[WebhookHeaderValue]](),
	)

	assert.Nil(t, headers)
	assert.Empty(t, diags)
}

func TestWebhookHeadersRejectNullAndUnknownObjects(t *testing.T) {
	t.Parallel()

	for name, value := range map[string]TypedObject[WebhookHeaderValue]{
		"null":    NewTypedObjectNull[WebhookHeaderValue](),
		"unknown": NewTypedObjectUnknown[WebhookHeaderValue](),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			headers := NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{"authorization": value})
			result, diags := ToWebhookDefinitionHeaders(path.Root("headers"), headers)
			assert.Nil(t, result)
			require.True(t, diags.HasError())
			assert.Equal(t, []string{`headers["authorization"]`}, diagnosticPaths(t, diags))
		})
	}
}

func TestWebhookHeadersFailWithoutPartialOutputAndSortKeys(t *testing.T) {
	t.Parallel()

	headers := NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{
		"z": NewTypedObject(WebhookHeaderValue{Value: types.StringValue("last"), Secret: types.BoolValue(false)}),
		"a": NewTypedObject(WebhookHeaderValue{Value: types.StringValue("first"), Secret: types.BoolValue(true)}),
	})
	result, diags := ToWebhookDefinitionHeaders(path.Root("headers"), headers)
	require.False(t, diags.HasError(), diags.Errors())
	assert.Equal(t, cm.WebhookDefinitionHeaders{
		{Key: "a", Value: cm.NewOptString("first"), Secret: cm.NewOptBool(true)},
		{Key: "z", Value: cm.NewOptString("last"), Secret: cm.NewOptBool(false)},
	}, result)

	headers = NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{
		"z":      NewTypedObject(WebhookHeaderValue{Value: types.StringValue("last"), Secret: types.BoolValue(false)}),
		"a":      NewTypedObject(WebhookHeaderValue{Value: types.StringValue("first"), Secret: types.BoolValue(true)}),
		"broken": NewTypedObject(WebhookHeaderValue{Value: types.StringUnknown(), Secret: types.BoolValue(false)}),
	})
	result, diags = ToWebhookDefinitionHeaders(path.Root("headers"), headers)
	assert.Nil(t, result)
	require.True(t, diags.HasError())
	assert.Equal(t, []string{`headers["broken"].value`}, diagnosticPaths(t, diags))
}

func TestWebhookRequestDoesNotReplaceUnavailableSecretWithEmptyValue(t *testing.T) {
	t.Parallel()

	model := validWebhookRequestModel()
	model.Headers = NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{
		"authorization": NewTypedObject(WebhookHeaderValue{
			Value:  types.StringNull(),
			Secret: types.BoolValue(true),
		}),
	})

	actual, diags := model.ToWebhookDefinitionData(t.Context(), WebhookModel{
		Headers: NewTypedMapNull[TypedObject[WebhookHeaderValue]](),
	}, path.Empty())

	assert.Equal(t, cm.WebhookDefinitionData{}, actual)
	require.True(t, diags.HasError())
	assert.Equal(t, []string{`headers["authorization"].value`}, diagnosticPaths(t, diags))
}

func TestWebhookSecretResponseUsesOnlyKnownFallback(t *testing.T) {
	t.Parallel()

	responseHeader := cm.WebhookDefinitionHeader{
		Key:    "authorization",
		Secret: cm.NewOptBool(true),
	}

	t.Run("configured secret is preserved when Contentful redacts it", func(t *testing.T) {
		t.Parallel()

		fallback := NewTypedObject(WebhookHeaderValue{
			Value:  types.StringValue("configured-secret"),
			Secret: types.BoolValue(true),
		})
		actual, diags := NewWebhookHeaderValueFromResponse(t.Context(), path.Root("headers").AtMapKey("authorization"), responseHeader, fallback)

		require.False(t, diags.HasError(), diags.Errors())

		value, ok := actual.GetValue()
		require.True(t, ok)
		assert.Equal(t, "configured-secret", value.Value.ValueString())
	})

	t.Run("imported redacted secret remains unavailable rather than empty", func(t *testing.T) {
		t.Parallel()

		actual, diags := NewWebhookHeaderValueFromResponse(
			t.Context(),
			path.Root("headers").AtMapKey("authorization"),
			responseHeader,
			NewTypedObjectNull[WebhookHeaderValue](),
		)

		require.False(t, diags.HasError(), diags.Errors())

		value, ok := actual.GetValue()
		require.True(t, ok)
		assert.True(t, value.Value.IsNull())
		assert.NotEqual(t, types.StringValue(""), value.Value)
	})
}
