package provider_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccLocaleResourceMissingVersionRefreshPreservesPriorState(t *testing.T) {
	t.Parallel()

	cli := newTerraformTestRuntime(t)

	var missingVersion atomic.Bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name, versionJSON := "Original", `,"version":7`
		if missingVersion.Load() {
			name, versionJSON = "Remote drift", ""
		}

		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
		}

		_, _ = fmt.Fprintf(w, `{
   "name":%q,"code":"en-AU","fallbackCode":null,
   "contentDeliveryApi":true,"contentManagementApi":true,"optional":false,"default":false,
   "sys":{"type":"Locale","id":"locale",
    "space":{"sys":{"type":"Link","linkType":"Space","id":"space"}},
    "environment":{"sys":{"type":"Link","linkType":"Environment","id":"environment"}}%s}
  }`, name, versionJSON)
	}))
	t.Cleanup(server.Close)

	cli.providerURL = server.URL
	cli.writeConfig(t, `resource "contentful_locale" "test" {
  space_id       = "space"
  environment_id = "environment"
  name           = "Original"
  code           = "en-AU"
 }`)

	output, err := cli.run(t.Context(), "apply", "-auto-approve", "-input=false", "-no-color")
	require.NoError(t, err, output)

	statePath := filepath.Join(cli.workingDirectory, "terraform.tfstate")
	before, err := os.ReadFile(statePath)
	require.NoError(t, err)

	missingVersion.Store(true)

	output, err = cli.run(t.Context(), "apply", "-refresh-only", "-auto-approve", "-input=false", "-no-color")
	require.Error(t, err)
	assert.Contains(t, output, "Missing locale version")
	assert.Contains(t, strings.Join(strings.Fields(output), " "), "Contentful returned a locale without sys.version. Refresh the resource successfully before updating it.")

	// Unlike mutation errors, Read errors discard the proposed state and private data.
	after, err := os.ReadFile(statePath)
	require.NoError(t, err)
	assert.JSONEq(t, string(before), string(after))
}
