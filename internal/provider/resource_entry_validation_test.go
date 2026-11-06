package provider_test

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccEntryResourceRejectsMalformedJSONBeforeHTTP(t *testing.T) {
	t.Parallel()

	runtime := newTerraformTestRuntime(t)

	var requests atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		http.Error(w, "unexpected request", http.StatusBadRequest)
	}))
	t.Cleanup(server.Close)
	runtime.providerURL = server.URL
	runtime.writeConfig(t, `resource "contentful_entry" "test" {
  space_id        = "space"
  environment_id  = "master"
  content_type_id = "article"
  entry_id        = "entry"
  fields          = { title = "{" }
}`)

	validateOutput, validateErr := runtime.run(t.Context(), "validate", "-no-color")
	applyOutput, applyErr := runtime.run(t.Context(), "apply", "-auto-approve", "-input=false", "-no-color")
	assert.Zero(t, requests.Load())
	require.Error(t, validateErr, validateOutput)
	assert.Contains(t, validateOutput, "Invalid JSON String Value")
	require.Error(t, applyErr, applyOutput)
	assert.Contains(t, applyOutput, "Invalid JSON String Value")
	assert.NotContains(t, applyOutput, "Failed to create entry")
}
