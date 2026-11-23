package provider_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	returnedWebhookPassword    = "synthetic-returned-password"
	returnedWebhookHeaderValue = "synthetic-returned-header"
)

func webhookPublicationRuntime(t *testing.T) (*terraformTestRuntime, *mutationJSONResponseAdapter) {
	t.Helper()

	runtime := newTerraformTestRuntime(t)
	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")
	server.SetWebhookDefinition("space", "existing", cm.WebhookDefinitionData{
		Name: "Existing", URL: "https://example.com/webhook", Topics: []string{"Entry.publish"}, Active: cm.NewOptBool(true),
		HttpBasicUsername: cm.NewOptNilString("user"), HttpBasicPassword: cm.NewOptNilString("remote-password"),
		Headers: cm.WebhookDefinitionHeaders{{Key: "X-Secret", Secret: cm.NewOptBool(true), Value: cm.NewOptString("remote-header")}},
	})
	adapter := &mutationJSONResponseAdapter{delegate: server}
	httpServer := httptest.NewServer(adapter)
	t.Cleanup(httpServer.Close)
	runtime.providerURL = httpServer.URL
	runtime.logPath = filepath.Join(runtime.workingDirectory, "provider.log")
	runtime.writeConfig(t, webhookBasicPasswordConfig("Updated", "null", ""))

	return runtime, adapter
}

type webhookStateInstance struct {
	Status     string         `json:"status"`
	Attributes map[string]any `json:"attributes"`
	Private    []byte         `json:"private"`
}

func readWebhookStateInstances(t *testing.T, runtime *terraformTestRuntime) []webhookStateInstance {
	t.Helper()

	logs, err := os.ReadFile(runtime.logPath)
	require.NoError(t, err)
	assert.Contains(t, string(logs), "webhook.")
	assert.NotContains(t, string(logs), returnedWebhookPassword)
	assert.NotContains(t, string(logs), returnedWebhookHeaderValue)

	body, err := os.ReadFile(filepath.Join(runtime.workingDirectory, "terraform.tfstate"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	require.NoError(t, err)

	var state struct {
		Resources []struct {
			Instances []webhookStateInstance `json:"instances"`
		} `json:"resources"`
	}
	require.NoError(t, json.Unmarshal(body, &state))

	if len(state.Resources) == 0 {
		return nil
	}

	require.Len(t, state.Resources, 1)

	return state.Resources[0].Instances
}

func readableWebhookCredentials(doc map[string]any) {
	doc["httpBasicPassword"] = returnedWebhookPassword
	doc["headers"] = []map[string]any{{"key": "X-Secret", "secret": true, "value": returnedWebhookHeaderValue}}
}

func requireReturnedWebhookCredentials(t *testing.T, instance webhookStateInstance) {
	t.Helper()

	require.Equal(t, returnedWebhookPassword, instance.Attributes["http_basic_password"])
	require.Equal(t, map[string]any{"X-Secret": map[string]any{"secret": true, "value": returnedWebhookHeaderValue, "value_wo": nil}}, instance.Attributes["headers"])
}

func TestAccWebhookResourceImportReturnedCredentials(t *testing.T) {
	t.Parallel()
	runtime, adapter := webhookPublicationRuntime(t)
	adapter.mutateNext(http.MethodGet, readableWebhookCredentials)

	output, err := runtime.run(t.Context(), "import", "-input=false", "-no-color", "contentful_webhook.test", "space/existing")
	require.NoError(t, err, output)
	instances := readWebhookStateInstances(t, runtime)
	require.Len(t, instances, 1)
	requireReturnedWebhookCredentials(t, instances[0])
}

func TestAccWebhookResourceReturnedCredentialsInRecoveryState(t *testing.T) {
	t.Parallel()

	for _, method := range []string{http.MethodPost, http.MethodPut} {
		t.Run(method, func(t *testing.T) {
			t.Parallel()
			runtime, adapter := webhookPublicationRuntime(t)
			wantVersion := "0"

			if method == http.MethodPut {
				output, err := runtime.run(t.Context(), "import", "-input=false", "-no-color", "contentful_webhook.test", "space/existing")
				require.NoError(t, err, output)

				wantVersion = "1"
			}

			runtime.writeConfig(t, webhookBasicPasswordConfig("Updated", `"configured-password"`, `headers = {}`))
			adapter.mutateNext(method, readableWebhookCredentials)

			output, err := runtime.run(t.Context(), "apply", "-refresh=false", "-auto-approve", "-input=false", "-no-color")
			require.Error(t, err)
			assert.Contains(t, output, "Contentful returned different webhook headers")
			assert.Contains(t, output, "Contentful returned a different webhook Basic password")
			assert.NotContains(t, output, returnedWebhookPassword)
			assert.NotContains(t, output, returnedWebhookHeaderValue)

			instances := readWebhookStateInstances(t, runtime)
			require.Len(t, instances, 1)
			instance := instances[0]

			if method == http.MethodPost {
				requireReturnedWebhookCredentials(t, instance)
				assert.Equal(t, "tainted", instance.Status)
			} else {
				// Import left a redacted header without a verifier. The PUT
				// requested its removal and supplied an ordinary Basic password.
				assert.Equal(t, returnedWebhookPassword, instance.Attributes["http_basic_password"])
				assert.Equal(t, map[string]any{"X-Secret": map[string]any{"secret": true, "value": nil, "value_wo": nil}}, instance.Attributes["headers"])
			}

			webhookID, ok := instance.Attributes["webhook_id"].(string)
			require.True(t, ok)
			require.NotEmpty(t, webhookID)
			assert.Equal(t, "space/"+webhookID, instance.Attributes["id"])

			var private map[string][]byte
			require.NoError(t, json.Unmarshal(instance.Private, &private))
			assert.Equal(t, wantVersion, string(private["version"]))
		})
	}
}

func TestAccWebhookResourceEmptyAssignedID(t *testing.T) {
	t.Parallel()
	runtime, adapter := webhookPublicationRuntime(t)
	adapter.mutateNext(http.MethodPost, func(doc map[string]any) {
		sys, _ := doc["sys"].(map[string]any)
		sys["id"] = ""
	})

	output, err := runtime.run(t.Context(), "apply", "-auto-approve", "-input=false", "-no-color")
	assert.Contains(t, output, "Missing created webhook ID")
	assert.Empty(t, readWebhookStateInstances(t, runtime))
	require.Error(t, err)
}
