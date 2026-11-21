package provider_test

import (
	"net/http"
	"sync/atomic"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookHeadersUnknownContainerIsOmitted(t *testing.T) {
	t.Parallel()

	headers, diags := ToWebhookDefinitionHeaders(
		path.Root("headers"),
		NewTypedMapUnknown[TypedObject[WebhookHeaderValue]](),
		nil,
	)

	assert.Nil(t, headers)
	assert.Empty(t, diags)
}

func TestWebhookHeadersFailWithoutPartialOutputAndSortKeys(t *testing.T) {
	t.Parallel()

	headers := NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{
		"z": NewTypedObject(WebhookHeaderValue{Value: types.StringValue("last"), Secret: types.BoolValue(false)}),
		"a": NewTypedObject(WebhookHeaderValue{Value: types.StringValue("first"), Secret: types.BoolValue(true)}),
	})
	result, diags := ToWebhookDefinitionHeaders(path.Root("headers"), headers, nil)
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
	result, diags = ToWebhookDefinitionHeaders(path.Root("headers"), headers, nil)
	assert.Nil(t, result)
	require.True(t, diags.HasError())
	assert.Equal(t, []string{`headers["broken"].value`}, attributeDiagnosticPaths(t, diags))
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
	}, model.Headers.Elements(), path.Empty())

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

	actual, diags := ToWebhookDefinitionHeaders(path.Root("headers"), headers, nil)

	require.False(t, diags.HasError(), diags.Errors())
	require.Len(t, actual, 2)

	for _, header := range actual {
		assert.True(t, header.Value.IsSet(), "configured empty value for %q must be present on the wire", header.Key)
		assert.Empty(t, header.Value.Or("not-empty"))
	}
}

// Apply the plan directly so planning diagnostics cannot mask a missing mutation check.
func TestWebhookApplyRequiresExistingSecretAtExactKey(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, priorKey string
		priorSecret    bool
	}{
		{name: "create"},
		{name: "renamed", priorKey: "x-token", priorSecret: true},
		{name: "ordinary to secret", priorKey: "X-Token"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var calls atomic.Int64

			protocol := configuredProtocolServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				calls.Add(1)
				w.WriteHeader(http.StatusBadRequest)
			}))
			schema := WebhookResourceSchema(t.Context())
			model := validWebhookRequestModel()
			model.SpaceID = types.StringValue("space")
			model.WebhookID = types.StringValue("webhook")
			model.Topics = NewTypedList([]types.String{types.StringValue("Entry.publish")})
			model.Timeouts = TimeoutsNull()
			prior := nullResourceDynamicValue(t, schema)

			if test.priorKey != "" {
				model.Headers = NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{test.priorKey: NewTypedObject(WebhookHeaderValue{Secret: types.BoolValue(test.priorSecret), Value: types.StringValue("prior")})})
				prior = resourceModelDynamicValue(t, schema, model)
			}

			model.Headers = NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{"X-Token": NewTypedObject(WebhookHeaderValue{Secret: types.BoolValue(true)})})
			plan := resourceModelDynamicValue(t, schema, model)
			response, err := protocol.ApplyResourceChange(t.Context(), &tfprotov6.ApplyResourceChangeRequest{
				TypeName: "contentful_webhook", Config: &plan, PlannedState: &plan, PriorState: &prior, PlannedPrivate: privateVersionBytes(t, 7),
			})
			require.NoError(t, err)
			require.Len(t, response.Diagnostics, 1)
			assert.Equal(t, "Missing webhook header value", response.Diagnostics[0].Summary)
			assert.Equal(t, tftypes.NewAttributePath().WithAttributeName("headers").WithElementKeyString("X-Token").WithAttributeName("value"), response.Diagnostics[0].Attribute)
			assert.Zero(t, calls.Load())
		})
	}
}
