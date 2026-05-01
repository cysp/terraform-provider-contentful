package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedactForLoggingDirectPath(t *testing.T) {
	t.Parallel()

	request := cm.AppSigningSecretRequestData{
		Value: "secret",
	}

	actual, typeOk := provider.RedactForLogging(request, provider.RedactPath("value")).(cm.AppSigningSecretRequestData)
	require.True(t, typeOk)

	assert.Equal(t, "<redacted>", actual.Value)
	assert.Equal(t, "secret", request.Value)
}

func TestRedactForLoggingWrapperPath(t *testing.T) {
	t.Parallel()

	response := &cm.AppSigningSecretStatusCode{
		StatusCode: 200,
		Response: cm.AppSigningSecret{
			Sys:           cm.NewAppSigningSecretSys("organization-id", "app-definition-id"),
			RedactedValue: "encoded-secret",
		},
	}

	actual, typeOk := provider.RedactForLogging(response, provider.RedactPath("response.redactedValue")).(*cm.AppSigningSecretStatusCode)
	require.True(t, typeOk)

	assert.Equal(t, "<redacted>", actual.Response.RedactedValue)
	assert.Equal(t, "encoded-secret", response.Response.RedactedValue)
}

func TestRedactForLoggingNestedPath(t *testing.T) {
	t.Parallel()

	response := cm.PersonalAccessToken{
		Sys:    cm.NewPersonalAccessTokenSys("token-id"),
		Name:   "token",
		Scopes: []string{"content_management_read"},
		Token:  cm.NewOptString("CFPAT-secret"),
	}
	response.Sys.RedactedValue = cm.NewOptString("redacted-token")

	actual, typeOk := provider.RedactForLogging(response, provider.RedactPath("token"), provider.RedactPath("sys.redactedValue")).(cm.PersonalAccessToken)
	require.True(t, typeOk)

	assert.Equal(t, "<redacted>", actual.Token.Value)
	assert.Equal(t, "<redacted>", actual.Sys.RedactedValue.Value)
	assert.Equal(t, "CFPAT-secret", response.Token.Value)
	assert.Equal(t, "redacted-token", response.Sys.RedactedValue.Value)
}

func TestRedactForLoggingWildcardAndConditionalPath(t *testing.T) {
	t.Parallel()

	request := cm.WebhookDefinitionData{
		Name:              "webhook",
		URL:               "https://example.com/webhook",
		HttpBasicPassword: cm.NewOptNilString("basic-password"),
		Headers: cm.WebhookDefinitionHeaders{
			{
				Key:    "X-Secret",
				Value:  cm.NewOptString("secret-header"),
				Secret: cm.NewOptBool(true),
			},
			{
				Key:    "X-Public",
				Value:  cm.NewOptString("public-header"),
				Secret: cm.NewOptBool(false),
			},
		},
	}

	actual, typeOk := provider.RedactForLogging(
		request,
		provider.RedactPath("httpBasicPassword"),
		provider.RedactPathWhen("headers.*.value", "secret", true),
	).(cm.WebhookDefinitionData)
	require.True(t, typeOk)

	assert.Equal(t, "<redacted>", actual.HttpBasicPassword.Value)
	assert.Equal(t, "<redacted>", actual.Headers[0].Value.Value)
	assert.Equal(t, "public-header", actual.Headers[1].Value.Value)
	assert.Equal(t, "basic-password", request.HttpBasicPassword.Value)
	assert.Equal(t, "secret-header", request.Headers[0].Value.Value)
}

func TestRedactForLoggingUnknownPathLeavesValueUnchanged(t *testing.T) {
	t.Parallel()

	response := cm.PreviewApiKey{
		AccessToken: "preview-token",
	}

	actual := provider.RedactForLogging(response, provider.RedactPath("missing.accessToken"))

	assert.Equal(t, response, actual)
}

func TestRedactForLoggingMapSlicePointerAndInterfaceValues(t *testing.T) {
	t.Parallel()

	type item struct {
		Secret string `json:"secret"`
		Public string `json:"public"`
	}

	input := map[string]any{
		"items": []any{
			&item{
				Secret: "secret",
				Public: "public",
			},
		},
	}

	actual, typeOk := provider.RedactForLogging(input, provider.RedactPath("items.*.secret")).(map[string]any)
	require.True(t, typeOk)

	items, itemsOk := actual["items"].([]any)
	require.True(t, itemsOk)
	require.Len(t, items, 1)

	actualItem, actualItemOk := items[0].(*item)
	require.True(t, actualItemOk)

	originalItems, originalItemsOk := input["items"].([]any)
	require.True(t, originalItemsOk)

	originalItem, originalItemOk := originalItems[0].(*item)
	require.True(t, originalItemOk)

	assert.Equal(t, "<redacted>", actualItem.Secret)
	assert.Equal(t, "public", actualItem.Public)
	assert.Equal(t, "secret", originalItem.Secret)
}
