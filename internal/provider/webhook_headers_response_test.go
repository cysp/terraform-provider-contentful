package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestWebhookResponseUsesExactKeyForOmittedSecretValues(t *testing.T) {
	t.Parallel()

	fallback := WebhookModel{Headers: NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{
		"X-Token":   NewTypedObject(WebhookHeaderValue{Secret: types.BoolValue(false), Value: types.StringValue("previous")}),
		"X-Empty":   NewTypedObject(WebhookHeaderValue{Secret: types.BoolValue(true), Value: types.StringValue("")}),
		"X-Unknown": NewTypedObject(WebhookHeaderValue{Secret: types.BoolValue(true), Value: types.StringUnknown()}),
	})}
	response := completeWebhookMutationResponse()
	response.Headers = cm.WebhookDefinitionHeaders{
		{Key: "X-Token", Secret: cm.NewOptBool(true)},
		{Key: "x-token", Secret: cm.NewOptBool(true)},
		{Key: "X-Empty", Secret: cm.NewOptBool(true)},
		{Key: "X-Unknown", Secret: cm.NewOptBool(true)},
	}

	state, diags := NewWebhookResourceModelFromResponse(t.Context(), response, fallback)
	require.Empty(t, diags)
	require.Equal(t, map[string]TypedObject[WebhookHeaderValue]{
		"X-Token":   NewTypedObject(WebhookHeaderValue{Secret: types.BoolValue(true), Value: types.StringNull()}),
		"x-token":   NewTypedObject(WebhookHeaderValue{Secret: types.BoolValue(true), Value: types.StringNull()}),
		"X-Empty":   NewTypedObject(WebhookHeaderValue{Secret: types.BoolValue(true), Value: types.StringValue("")}),
		"X-Unknown": NewTypedObject(WebhookHeaderValue{Secret: types.BoolValue(true), Value: types.StringNull()}),
	}, state.Headers.Elements())
}
