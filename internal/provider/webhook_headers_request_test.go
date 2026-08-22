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
		NewTypedMapNull[TypedObject[WebhookHeaderValue]](),
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
			result, diags := ToWebhookDefinitionHeaders(path.Root("headers"), headers, headers)
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
	result, diags := ToWebhookDefinitionHeaders(path.Root("headers"), headers, headers)
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
	result, diags = ToWebhookDefinitionHeaders(path.Root("headers"), headers, headers)
	assert.Nil(t, result)
	require.True(t, diags.HasError())
	assert.Equal(t, []string{`headers["broken"].value`}, diagnosticPaths(t, diags))
}

func TestWebhookRequestPreservesResponseOwnedSecretWithoutSendingAValue(t *testing.T) {
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

	require.False(t, diags.HasError(), diags.Errors())
	require.Len(t, actual.Headers, 1)
	assert.Equal(t, "authorization", actual.Headers[0].Key)
	assert.Equal(t, cm.NewOptBool(true), actual.Headers[0].Secret)
	assert.False(t, actual.Headers[0].Value.IsSet())
}

func TestWebhookRequestPreservesConfiguredEmptyHeaderValues(t *testing.T) {
	t.Parallel()

	headers := NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{
		"ordinary": NewTypedObject(WebhookHeaderValue{Value: types.StringValue(""), Secret: types.BoolValue(false)}),
		"secret":   NewTypedObject(WebhookHeaderValue{Value: types.StringValue(""), Secret: types.BoolValue(true)}),
	})

	actual, diags := ToWebhookDefinitionHeaders(path.Root("headers"), headers, headers)

	require.False(t, diags.HasError(), diags.Errors())
	require.Len(t, actual, 2)
	for _, header := range actual {
		assert.True(t, header.Value.IsSet(), "configured empty value for %q must be present on the wire", header.Key)
		assert.Empty(t, header.Value.Or("not-empty"))
	}
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
