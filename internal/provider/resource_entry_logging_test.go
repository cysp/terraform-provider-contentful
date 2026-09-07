package provider_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccEntryResourceLifecycleLogsExcludeContent(t *testing.T) {
	t.Parallel()

	for name, entryID := range map[string]string{"specified ID": `"log-entry"`, "generated ID": "null"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			runtime := newTerraformTestRuntime(t)
			server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
			require.NoError(t, err)
			server.RegisterSpaceEnvironment("log-space", "master")
			testServer := httptest.NewServer(server)
			t.Cleanup(testServer.Close)
			runtime.providerURL = testServer.URL
			runtime.logPath = filepath.Join(runtime.workingDirectory, "provider.log")

			for _, content := range []string{"LOG_ENTRY_CREATE_SENTINEL", "LOG_ENTRY_UPDATE_SENTINEL\n\"quoted\"\\escaped"} {
				runtime.writeConfig(t, entryLoggingTestConfig(entryID, content))
				output, applyErr := runtime.run(t.Context(), "apply", "-auto-approve", "-input=false", "-no-color")
				require.NoError(t, applyErr, output)
				assert.NotContains(t, output, "LOG_ENTRY_")
			}

			output, err := runtime.run(t.Context(), "destroy", "-auto-approve", "-input=false", "-no-color")
			require.NoError(t, err, output)
			assert.NotContains(t, output, "LOG_ENTRY_")

			logOutput, err := os.ReadFile(runtime.logPath)
			require.NoError(t, err)

			logs := string(logOutput)
			assert.NotContains(t, logs, "LOG_ENTRY_")

			for _, operation := range []string{"create", "read", "update", "publish", "unpublish", "delete"} {
				assert.Contains(t, logs, "entry."+operation)
			}

			assert.Contains(t, logs, "status_code=201")
			assert.Contains(t, logs, "status_code=204")
			assert.Contains(t, logs, "version=1")
			assert.Contains(t, logs, "log-space")

			for line := range strings.SplitSeq(logs, "\n") {
				if strings.Contains(line, ": entry.") {
					assert.NotContains(t, line, "request=")
					assert.NotContains(t, line, "response=")
				}
			}
		})
	}
}

func TestAccEntryResourceErrorDiagnosticsPreserveUpstreamMessages(t *testing.T) {
	t.Parallel()

	runtime := newTerraformTestRuntime(t)
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.contentful.management.v1+json")
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte(`{"sys":{"type":"Error","id":"ValidationFailed"},"message":"Contentful rejected UPSTREAM_ENTRY_SENTINEL","details":{"errors":[{"name":"invalid","path":["fields","secret","en-US"],"details":"Check the configured field value"}]}}`))
		assert.NoError(t, err)
	}))
	t.Cleanup(testServer.Close)
	runtime.providerURL = testServer.URL
	runtime.logPath = filepath.Join(runtime.workingDirectory, "provider.log")
	runtime.writeConfig(t, entryLoggingTestConfig(`"log-entry"`, "UPSTREAM_ENTRY_SENTINEL"))

	output, err := runtime.run(t.Context(), "apply", "-auto-approve", "-input=false", "-no-color")
	require.Error(t, err)
	assert.Contains(t, output, "Contentful rejected UPSTREAM_ENTRY_SENTINEL")
	assert.Contains(t, output, "Check the configured field value")

	logOutput, err := os.ReadFile(runtime.logPath)
	require.NoError(t, err)

	var events int

	for line := range strings.SplitSeq(string(logOutput), "\n") {
		if strings.Contains(line, ": entry.create:") {
			events++

			assert.Contains(t, line, "status_code=400")
			assert.NotContains(t, line, "UPSTREAM_ENTRY_SENTINEL")
			assert.NotContains(t, line, "response=")
		}
	}

	assert.Equal(t, 1, events)
}

func TestAccEntryResourceDecodeFailureLogsExcludeResponseBody(t *testing.T) {
	t.Parallel()

	runtime := newTerraformTestRuntime(t)
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.contentful.management.v1+json")
		w.WriteHeader(http.StatusCreated)
		_, err := w.Write([]byte(`{"fields":{"secret":{"en-US":"ENTRY_MALFORMED_RESPONSE_SENTINEL"}},"sys":`))
		assert.NoError(t, err)
	}))
	t.Cleanup(testServer.Close)
	runtime.providerURL = testServer.URL
	runtime.logPath = filepath.Join(runtime.workingDirectory, "provider.log")
	runtime.writeConfig(t, entryLoggingTestConfig(`"log-entry"`, "ENTRY_REQUEST_SENTINEL"))

	output, err := runtime.run(t.Context(), "apply", "-auto-approve", "-input=false", "-no-color")
	require.Error(t, err)
	assert.Contains(t, output, "Failed to create entry")

	logOutput, err := os.ReadFile(runtime.logPath)
	require.NoError(t, err)

	var events int

	for line := range strings.SplitSeq(string(logOutput), "\n") {
		if strings.Contains(line, ": entry.create:") {
			events++

			assert.Contains(t, line, "error_type=")
			assert.NotContains(t, line, "ENTRY_MALFORMED_RESPONSE_SENTINEL")
			assert.NotContains(t, line, "ENTRY_REQUEST_SENTINEL")
			assert.NotContains(t, line, "err=")
			assert.NotContains(t, line, "response=")
		}
	}

	assert.Equal(t, 1, events)
}

func entryLoggingTestConfig(entryID, content string) string {
	return fmt.Sprintf(`
variable "content" {
  type = string
  sensitive = true
  default = %q
}


resource "contentful_entry" "test" {
  space_id = "log-space"
  environment_id = "master"
  content_type_id = "log-type"
  entry_id = %s
  fields = { secret = jsonencode({ "en-US" = var.content }) }
}
`, content, entryID)
}
