//nolint:testpackage // This test exercises internal secret preservation before response redaction.
package cmtesting

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookSecretValueIsPreservedAndRedacted(t *testing.T) {
	t.Parallel()

	webhook := NewWebhookDefinitionFromFields("space", "webhook", cm.WebhookDefinitionData{
		Name: "Webhook",
		URL:  "https://example.com/webhook",
		Headers: cm.WebhookDefinitionHeaders{
			{Key: "Authorization", Value: cm.NewOptString("existing-secret"), Secret: cm.NewOptBool(true)},
		},
	})

	UpdateWebhookDefinitionFromFields(&webhook, cm.WebhookDefinitionData{
		Name: "Updated webhook",
		URL:  "https://example.com/webhook",
		Headers: cm.WebhookDefinitionHeaders{
			{Key: "Authorization", Secret: cm.NewOptBool(true)},
		},
	})

	require.Len(t, webhook.Headers, 1)
	assert.Equal(t, cm.NewOptString("existing-secret"), webhook.Headers[0].Value)

	response := redactWebhookDefinitionSecrets(webhook)
	require.Len(t, response.Headers, 1)
	assert.False(t, response.Headers[0].Value.IsSet())
	assert.Equal(t, cm.NewOptString("existing-secret"), webhook.Headers[0].Value)
}
