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
		t.Context(),
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
			result, diags := ToWebhookDefinitionHeaders(t.Context(), path.Root("headers"), headers)
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
	result, diags := ToWebhookDefinitionHeaders(t.Context(), path.Root("headers"), headers)
	require.False(t, diags.HasError(), diags.Errors())
	assert.Equal(t, cm.WebhookDefinitionHeaders{
		{Key: "a", Value: cm.NewOptString("first"), Secret: cm.NewOptBool(true)},
		{Key: "z", Value: cm.NewOptString("last"), Secret: cm.NewOptBool(false)},
	}, result)

	headers.Set("broken", NewTypedObject(WebhookHeaderValue{Value: types.StringUnknown(), Secret: types.BoolValue(false)}))
	result, diags = ToWebhookDefinitionHeaders(t.Context(), path.Root("headers"), headers)
	assert.Nil(t, result)
	require.True(t, diags.HasError())
	assert.Equal(t, []string{`headers["broken"].value`}, diagnosticPaths(t, diags))
}
