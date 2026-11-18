package provider

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/stretchr/testify/require"
)

// On import, the returned header descriptors are the only redaction evidence.
// Both orderings must reject a readable case alias before map projection.
func TestWebhookRedactedResponseDuplicates(t *testing.T) {
	t.Parallel()

	hidden := cm.WebhookDefinitionHeader{Key: "X-Secret", Secret: cm.NewOptBool(true)}

	readable := cm.WebhookDefinitionHeader{Key: "x-secret", Secret: cm.NewOptBool(false), Value: cm.NewOptString("unexpected-readable-value")}
	for _, headers := range []cm.WebhookDefinitionHeaders{{hidden, readable}, {readable, hidden}} {
		response := cm.WebhookDefinition{Headers: headers}
		sanitized, diags := redactWebhookResponse(response, WebhookModel{}, WebhookModel{}, WebhookModel{})
		require.True(t, diags.HasError())

		for _, header := range sanitized.Headers {
			require.False(t, header.Value.IsSet())
		}

		require.Contains(t, response.Headers, readable)
	}
}
