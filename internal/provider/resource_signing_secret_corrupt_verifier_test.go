package provider_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Terraform can plan an update before deciding to replace or destroy. Invalid
// comparison data must block Update without blocking those recovery operations.
func TestAccSigningSecretCorruptVerifierRecovery(t *testing.T) {
	t.Parallel()

	for _, kind := range []string{"app", "webhook"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()

			fixture := &signingSecretFixture{kind: kind, timestamps: true}
			server := httptest.NewServer(fixture)
			t.Cleanup(server.Close)
			runtime := newTerraformTestRuntime(t)
			runtime.providerURL = server.URL
			value := strings.Repeat("a", 64)
			runtime.writeConfig(t, signingSecretConfig(kind, fmt.Sprintf("value_wo=%q", value)))

			run := func(arguments ...string) string {
				t.Helper()
				output, err := runtime.run(t.Context(), append([]string{arguments[0], "-input=false", "-no-color"}, arguments[1:]...)...)
				require.NoError(t, err, output)

				return output
			}
			run("apply", "-auto-approve")
			fixture.requireValues(t, value)
			corruptSigningSecretVerifier(t, runtime)

			output := run("plan", "-out=update.tfplan")
			require.Contains(t, output, "Warning: Invalid write-only secret private state")
			require.Contains(t, strings.Join(strings.Fields(output), " "), "An in-place update will fail before writing the secret.")
			output, err := runtime.run(t.Context(), "apply", "-input=false", "-no-color", "update.tfplan")
			require.Error(t, err, output)
			require.Contains(t, output, "Error: Invalid write-only secret private state")
			fixture.requireValues(t, value)

			output = run("plan", "-replace=contentful_"+kind+"_signing_secret.test", "-out=replace.tfplan")
			require.Contains(t, output, "Warning: Invalid write-only secret private state")
			require.Contains(t, output, "will be replaced")
			run("apply", "replace.tfplan")
			fixture.requireValues(t, value, value)
			fixture.requireMutations(t, "PUT", "DELETE", "PUT")
			// The new verifier is usable: a fresh plan is a no-op without warnings.
			output = run("plan", "-detailed-exitcode")
			require.NotContains(t, output, "Invalid write-only secret private state")

			corruptSigningSecretVerifier(t, runtime)

			output = run("plan", "-destroy", "-out=destroy.tfplan")
			require.Contains(t, output, "will be destroyed")
			run("apply", "destroy.tfplan")
			fixture.requireMutations(t, "PUT", "DELETE", "PUT", "DELETE")
			fixture.mu.Lock()
			defer fixture.mu.Unlock()

			require.False(t, fixture.exists)
		})
	}
}

func corruptSigningSecretVerifier(t *testing.T, runtime *terraformTestRuntime) {
	t.Helper()

	filename := filepath.Join(runtime.workingDirectory, "terraform.tfstate")
	data, err := os.ReadFile(filename)
	require.NoError(t, err)

	document := testJSONObject(string(data))
	resources, ok := document["resources"].([]any)
	require.True(t, ok)
	require.Len(t, resources, 1)
	resource, ok := resources[0].(map[string]any)
	require.True(t, ok)
	instances, ok := resource["instances"].([]any)
	require.True(t, ok)
	require.Len(t, instances, 1)
	instance, ok := instances[0].(map[string]any)
	require.True(t, ok)
	privateText, ok := instance["private"].(string)
	require.True(t, ok)

	privateData, err := base64.StdEncoding.DecodeString(privateText)
	require.NoError(t, err)

	var private map[string][]byte
	require.NoError(t, json.Unmarshal(privateData, &private))
	private["write_only_secret_hashes"] = []byte(testJSON([]map[string]string{{"path": "value_wo", "hash": "invalid"}}))
	instance["private"] = base64.StdEncoding.EncodeToString([]byte(testJSON(private)))
	require.NoError(t, os.WriteFile(filename, []byte(testJSON(document)), 0o600))
}
