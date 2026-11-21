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
	require.Nil(t, UpdateWebhookDefinitionFromFields(&webhook, fields))
	require.Equal(t, cm.NewOptBool(false), webhook.Active)

	fields.Active.Reset()
	require.Nil(t, UpdateWebhookDefinitionFromFields(&webhook, fields))
	require.Equal(t, cm.NewOptBool(true), webhook.Active)
}

func TestWebhookSecretValueIsPreservedAndRedacted(t *testing.T) {
	t.Parallel()

	webhook := NewWebhookDefinitionFromFields("space", "webhook", cm.WebhookDefinitionData{
		Name:              "Webhook",
		URL:               "https://example.com/webhook",
		HttpBasicUsername: cm.NewOptNilString("user"),
		HttpBasicPassword: cm.NewOptNilString("basic-password-test-sentinel"),
		Headers: cm.WebhookDefinitionHeaders{
			{Key: "Authorization", Value: cm.NewOptString("existing-secret"), Secret: cm.NewOptBool(true)},
		},
	})

	require.Nil(t, UpdateWebhookDefinitionFromFields(&webhook, cm.WebhookDefinitionData{
		Name:              "Updated webhook",
		URL:               "https://example.com/webhook",
		HttpBasicUsername: cm.NewOptNilString("user"),
		HttpBasicPassword: cm.NewOptNilString("basic-password-test-sentinel"),
		Headers: cm.WebhookDefinitionHeaders{
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
