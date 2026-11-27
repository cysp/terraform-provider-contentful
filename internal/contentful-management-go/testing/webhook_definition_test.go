//nolint:testpackage // This test exercises internal secret preservation before response redaction.
package cmtesting

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookDefinitionActiveDefault(t *testing.T) {
	t.Parallel()

	fields := cm.WebhookDefinitionData{Name: "Webhook"}
	webhook := NewWebhookDefinitionFromFields("space", "webhook", fields)
	require.Equal(t, cm.NewOptBool(true), webhook.Active)

	fields.Active = cm.NewOptBool(false)
	require.Nil(t, applyWebhookDefinitionUpdate(&webhook, fields))
	require.Equal(t, cm.NewOptBool(false), webhook.Active)

	fields.Active.Reset()
	require.Nil(t, applyWebhookDefinitionUpdate(&webhook, fields))
	require.Equal(t, cm.NewOptBool(true), webhook.Active)
}

func TestWebhookSecretValueIsPreservedAndRedacted(t *testing.T) {
	t.Parallel()

	webhook := NewWebhookDefinitionFromFields("space", "webhook", cm.WebhookDefinitionData{
		Name:              "Webhook",
		URL:               "https://example.com/webhook",
		HttpBasicUsername: cm.NewOptNilString("user"),
		HttpBasicPassword: cm.NewOptNilString("basic-password-test-sentinel"),
		Headers: cm.WebhookDefinitionRequestHeaders{
			{Key: "Authorization", Value: cm.NewOptNilString("existing-secret"), Secret: cm.NewOptBool(true)},
		},
	})

	require.Nil(t, applyWebhookDefinitionUpdate(&webhook, cm.WebhookDefinitionData{
		Name:              "Updated webhook",
		URL:               "https://example.com/webhook",
		HttpBasicUsername: cm.NewOptNilString("user"),
		HttpBasicPassword: cm.NewOptNilString("basic-password-test-sentinel"),
		Headers: cm.WebhookDefinitionRequestHeaders{
			{Key: "Authorization", Secret: cm.NewOptBool(true)},
		},
	}))

	require.Len(t, webhook.Headers, 1)
	assert.Equal(t, cm.NewOptString("existing-secret"), webhook.Headers[0].Value)

	response := redactWebhookDefinitionSecrets(webhook)
	require.Len(t, response.Headers, 1)
	assert.False(t, response.Headers[0].Value.IsSet())
	assert.True(t, response.HttpBasicPassword.IsEmpty())
	assert.Equal(t, cm.NewOptNilString("basic-password-test-sentinel"), webhook.HttpBasicPassword)
	assert.Equal(t, cm.NewOptString("existing-secret"), webhook.Headers[0].Value)
}

func TestWebhookUpdateRetainsExistingHeaderFlags(t *testing.T) {
	t.Parallel()

	ordinary := cm.WebhookDefinitionRequestHeader{Key: "X-Readable", Secret: cm.NewOptBool(false), Value: cm.NewOptNilString("first")}
	secret := cm.WebhookDefinitionRequestHeader{Key: "X-Secret", Secret: cm.NewOptBool(true), Value: cm.NewOptNilString("second")}
	fields := cm.WebhookDefinitionData{Headers: cm.WebhookDefinitionRequestHeaders{ordinary, secret}}
	webhook := NewWebhookDefinitionFromFields("space", "webhook", fields)

	fields.Headers = cm.WebhookDefinitionRequestHeaders{
		{Key: ordinary.Key, Secret: cm.NewOptBool(true), Value: cm.NewOptNilString("new-readable")},
		{Key: secret.Key, Secret: cm.NewOptBool(false), Value: cm.NewOptNilString("new-secret")},
	}
	require.Nil(t, applyWebhookDefinitionUpdate(&webhook, fields))
	response := redactWebhookDefinitionSecrets(webhook)
	require.Equal(t, cm.WebhookDefinitionHeaders{
		{Key: ordinary.Key, Secret: cm.NewOptBool(false), Value: cm.NewOptString("new-readable")},
		{Key: secret.Key, Secret: cm.NewOptBool(true)},
	}, response.Headers)
}
