package provider_test

import (
	"context"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppSigningSecretModelWithWriteOnlySecret(t *testing.T) {
	t.Parallel()

	model, values, diags := AppSigningSecretModelWithWriteOnlySecrets(
		AppSigningSecretModel{},
		AppSigningSecretModel{
			ValueWO: types.StringValue("write-only-secret"),
		},
	)

	require.False(t, diags.HasError(), diags)
	assert.Equal(t, "write-only-secret", model.Value.ValueString())
	assert.True(t, model.ValueWO.IsNull())
	assert.Len(t, values, 1)
}

func TestAppSigningSecretRequestRejectsUnknownSecret(t *testing.T) {
	t.Parallel()

	_, diags := (&AppSigningSecretModel{Value: types.StringUnknown()}).ToAppSigningSecretRequest(t.Context(), path.Empty())

	assert.True(t, diags.HasError())
}

func TestWebhookModelWithWriteOnlyHeaderValue(t *testing.T) {
	t.Parallel()

	plan := WebhookModel{
		Headers: NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{
			`x-"secret"\key`: NewTypedObject(WebhookHeaderValue{
				Value:  types.StringNull(),
				Secret: types.BoolValue(true),
			}),
		}),
		HeaderValuesWO: NewTypedMapNull[types.String](),
	}

	config := WebhookModel{
		Headers: NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{
			`x-"secret"\key`: NewTypedObject(WebhookHeaderValue{
				Value:  types.StringNull(),
				Secret: types.BoolValue(true),
			}),
		}),
		HeaderValuesWO: NewTypedMap(map[string]types.String{
			`x-"secret"\key`: types.StringValue("write-only-secret"),
		}),
	}

	model, values, diags := WebhookModelWithWriteOnlySecrets(plan, config)

	require.False(t, diags.HasError(), diags)

	header := model.Headers.Elements()[`x-"secret"\key`].Value()
	assert.Equal(t, "write-only-secret", header.Value.ValueString())
	assert.True(t, model.HeaderValuesWO.IsNull())
	assert.Len(t, values, 1)
	assert.Equal(t, path.Root("header_values_wo").AtMapKey(`x-"secret"\key`).String(), values[0].Path.String())
}

func TestWebhookWriteOnlyHeaderValueDoesNotEnterResponseState(t *testing.T) {
	t.Parallel()

	const headerName = "X-Secret"

	planHeaders := NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{
		headerName: NewTypedObject(WebhookHeaderValue{
			Value:  types.StringNull(),
			Secret: types.BoolValue(true),
		}),
	})
	plan := WebhookModel{
		Headers:        planHeaders,
		HeaderValuesWO: NewTypedMapNull[types.String](),
	}
	config := WebhookModel{
		Headers: planHeaders,
		HeaderValuesWO: NewTypedMap(map[string]types.String{
			headerName: types.StringValue("write-only-secret"),
		}),
	}

	requestModel, _, requestDiags := WebhookModelWithWriteOnlySecrets(plan, config)
	require.False(t, requestDiags.HasError(), requestDiags)
	assert.Equal(t, "write-only-secret", requestModel.Headers.Elements()[headerName].Value().Value.ValueString())

	stateHeaders, responseDiags := ReadHeaderValueMapFromResponse(
		context.Background(),
		path.Root("headers"),
		[]cm.WebhookDefinitionHeader{{
			Key:   headerName,
			Value: cm.NewOptString("write-only-secret-returned-by-api"),
		}},
		plan.Headers,
	)
	require.False(t, responseDiags.HasError(), responseDiags)
	assert.True(t, stateHeaders.Elements()[headerName].Value().Value.IsNull())
	assert.True(t, stateHeaders.Elements()[headerName].Value().Secret.ValueBool())
}

func TestWebhookModelDefersUnknownWriteOnlyHeaderValuesMap(t *testing.T) {
	t.Parallel()

	const headerName = "X-Secret"

	headers := NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{
		headerName: NewTypedObject(WebhookHeaderValue{
			Value:  types.StringNull(),
			Secret: types.BoolValue(true),
		}),
	})
	model, values, diags := WebhookModelWithWriteOnlySecrets(
		WebhookModel{Headers: headers},
		WebhookModel{
			Headers:        headers,
			HeaderValuesWO: NewTypedMapUnknown[types.String](),
		},
	)

	require.False(t, diags.HasError(), diags)
	assert.True(t, model.Headers.Elements()[headerName].Value().Value.IsUnknown())
	require.Len(t, values, 1)
	assert.True(t, values[0].Value.IsUnknown())
}

func TestWebhookModelRejectsWriteOnlyValueWithoutMatchingHeader(t *testing.T) {
	t.Parallel()

	_, _, diags := WebhookModelWithWriteOnlySecrets(
		WebhookModel{Headers: NewTypedMapNull[TypedObject[WebhookHeaderValue]]()},
		WebhookModel{
			HeaderValuesWO: NewTypedMap(map[string]types.String{
				"X-Secret": types.StringValue("write-only-secret"),
			}),
		},
	)

	assert.True(t, diags.HasError())
}

func TestWebhookModelRejectsWriteOnlyValueForNonSecretHeader(t *testing.T) {
	t.Parallel()

	headers := NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{
		"X-Not-Secret": NewTypedObject(WebhookHeaderValue{
			Value:  types.StringNull(),
			Secret: types.BoolValue(false),
		}),
	})
	_, _, diags := WebhookModelWithWriteOnlySecrets(
		WebhookModel{Headers: headers},
		WebhookModel{
			Headers: headers,
			HeaderValuesWO: NewTypedMap(map[string]types.String{
				"X-Not-Secret": types.StringValue("write-only-value"),
			}),
		},
	)

	assert.True(t, diags.HasError())
}

func TestWebhookModelPreservesComputedRedactedSecretHeader(t *testing.T) {
	t.Parallel()

	const headerName = "X-Imported-Secret"

	model, values, diags := WebhookModelWithWriteOnlySecrets(
		WebhookModel{
			Headers: NewTypedMap(map[string]TypedObject[WebhookHeaderValue]{
				headerName: NewTypedObject(WebhookHeaderValue{
					Value:  types.StringNull(),
					Secret: types.BoolValue(true),
				}),
			}),
		},
		WebhookModel{
			Headers:        NewTypedMapNull[TypedObject[WebhookHeaderValue]](),
			HeaderValuesWO: NewTypedMapNull[types.String](),
		},
	)

	require.False(t, diags.HasError(), diags)
	assert.Empty(t, values)
	assert.True(t, model.Headers.Elements()[headerName].Value().Value.IsNull())
}
