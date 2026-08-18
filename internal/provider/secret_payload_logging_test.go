package provider_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecretBearingOperationLogsExcludePayloads(t *testing.T) {
	t.Parallel()

	logMessages := map[string][]string{
		"resource_app_installation.go": {
			"app_installation.create",
			"app_installation.read",
			"app_installation.update",
		},
		"resource_app_signing_secret.go": {
			"app_signing_secret.create",
			"app_signing_secret.read",
			"app_signing_secret.update",
		},
		"resource_delivery_api_key.go": {
			"delivery_api_key.create",
			"delivery_api_key.read",
			"delivery_api_key.update",
		},
		"data_source_preview_api_key.go": {
			"preview_api_key.read",
		},
		"resource_personal_access_token.go": {
			"personal_access_token.create",
			"personal_access_token.read",
			"personal_access_token.delete",
		},
		"resource_webhook.go": {
			"webhook.create",
			"webhook.read",
			"webhook.update",
		},
	}
	requestPayloadMessages := map[string]struct{}{
		"app_installation.create":      {},
		"app_installation.update":      {},
		"app_signing_secret.create":    {},
		"app_signing_secret.update":    {},
		"delivery_api_key.create":      {},
		"delivery_api_key.update":      {},
		"personal_access_token.create": {},
		"webhook.create":               {},
		"webhook.update":               {},
	}

	for filename, messages := range logMessages {
		t.Run(filename, func(t *testing.T) {
			t.Parallel()

			source, err := os.ReadFile(filename)
			require.NoError(t, err)

			for _, message := range messages {
				fields := directInfoLogFields(t, string(source), message)

				assert.Contains(t, fields, `// "response": response, omitted to avoid logging sensitive values`, message)

				if _, ok := requestPayloadMessages[message]; ok {
					assert.Contains(t, fields, `// "request": request, omitted to avoid logging sensitive values`, message)
				}

				activeFields := uncommentedLogFields(fields)
				assert.NotContains(t, activeFields, `"request":`, "%s must not log its request payload", message)
				assert.NotContains(t, activeFields, `"response":`, "%s must not log its response payload", message)
				assert.Contains(t, activeFields, `"err":`, "%s must retain error context", message)
			}
		})
	}
}

func directInfoLogFields(t *testing.T, source string, message string) string {
	t.Helper()

	prefix := fmt.Sprintf("tflog.Info(ctx, %q, map[string]any{", message)
	_, fields, found := strings.Cut(source, prefix)
	require.True(t, found, "%s log call not found", message)

	fields, _, found = strings.Cut(fields, "\n\t})")
	require.True(t, found, "%s log fields end not found", message)

	return fields
}

func uncommentedLogFields(fields string) string {
	lines := strings.Split(fields, "\n")
	activeLines := lines[:0]

	for _, line := range lines {
		if !strings.HasPrefix(strings.TrimSpace(line), "//") {
			activeLines = append(activeLines, line)
		}
	}

	return strings.Join(activeLines, "\n")
}
