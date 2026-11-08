package provider_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/require"
)

func TestAccWebhookResourceTransformationValidationRecovery(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")
	server.SetWebhookDefinition("space", "webhook", cm.WebhookDefinitionData{
		Name: "Webhook", URL: "https://example.invalid/webhook", Topics: []string{"Entry.publish"}, Active: cm.NewOptBool(true),
		Transformation: cm.NewOptNilWebhookDefinitionDataTransformation(cm.WebhookDefinitionDataTransformation{Method: cm.NewOptString("HEAD")}),
	})

	var writes, deletes atomic.Int64

	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			deletes.Add(1)
		} else if r.Method != http.MethodGet {
			writes.Add(1)
		}

		server.ServeHTTP(w, r)
	}))
	t.Cleanup(httpServer.Close)
	runtime := newTerraformTestRuntime(t)
	runtime.providerURL = httpServer.URL
	configuration := func(name, method string) string {
		return fmt.Sprintf(`
 resource "contentful_webhook" "test" {
 space_id = "space"
 name = %q
 url = "https://example.invalid/webhook"
 topics = ["Entry.publish"]
 transformation = { method = %q }
 lifecycle { ignore_changes = [transformation] }
 }
 `, name, method)
	}
	runtime.writeConfig(t, configuration("Webhook", "POST"))
	output, err := runtime.run(t.Context(), "import", "-input=false", "-no-color", "contentful_webhook.test", "space/webhook")
	require.NoError(t, err, output)
	output, err = runtime.run(t.Context(), "plan", "-input=false", "-no-color", "-detailed-exitcode")
	require.NoError(t, err, output)
	runtime.writeConfig(t, configuration("Changed", "POST"))
	output, err = runtime.run(t.Context(), "apply", "-auto-approve", "-input=false", "-no-color")
	require.Error(t, err, output)
	require.Contains(t, output, "Invalid webhook transformation value")
	require.Zero(t, writes.Load())
	runtime.writeConfig(t, configuration("Webhook", "HEAD"))
	output, err = runtime.run(t.Context(), "destroy", "-auto-approve", "-input=false", "-no-color")
	require.Error(t, err, output)
	require.Contains(t, output, "Invalid webhook transformation value")
	require.Zero(t, deletes.Load())
	runtime.writeConfig(t, configuration("Webhook", "POST"))
	output, err = runtime.run(t.Context(), "destroy", "-auto-approve", "-input=false", "-no-color")
	require.NoError(t, err, output)
	require.Zero(t, writes.Load())
	require.EqualValues(t, 1, deletes.Load())
}
