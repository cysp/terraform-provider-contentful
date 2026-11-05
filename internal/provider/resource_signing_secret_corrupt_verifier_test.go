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

func TestAccSigningSecretOrdinaryValueDiscardsObsoleteVerifier(t *testing.T) {
	t.Parallel()

	initial, changed := strings.Repeat("a", 64), strings.Repeat("b", 64)
	for _, kind := range []string{"app", "webhook"} {
		for _, test := range []struct {
			name, attribute, value string
			wantValues             []string
		}{
			{name: "ordinary unchanged", attribute: "value", value: initial, wantValues: []string{initial}},
			{name: "ordinary changed", attribute: "value", value: changed, wantValues: []string{initial, changed}},
			{name: "write-only unchanged", attribute: "value_wo", value: initial, wantValues: []string{initial}},
			{name: "write-only changed", attribute: "value_wo", value: changed, wantValues: []string{initial, changed}},
		} {
			t.Run(kind+"/"+test.name, func(t *testing.T) {
				t.Parallel()

				fixture := &signingSecretFixture{kind: kind, timestamps: true}
				server := httptest.NewServer(fixture)
				t.Cleanup(server.Close)
				runtime := newTerraformTestRuntime(t)
				runtime.providerURL = server.URL
				run := func(arguments ...string) string {
					t.Helper()
					output, err := runtime.run(t.Context(), append([]string{arguments[0], "-input=false", "-no-color"}, arguments[1:]...)...)
					require.NoError(t, err, output)
					require.NotContains(t, output, "Invalid write-only secret private state")

					return output
				}

				runtime.writeConfig(t, signingSecretConfig(kind, fmt.Sprintf("value=%q", initial)))
				run("apply", "-auto-approve")
				corruptSigningSecretVerifier(t, runtime)

				runtime.writeConfig(t, signingSecretConfig(kind, fmt.Sprintf("%s=%q\ntimeouts={read=\"1m\"}", test.attribute, test.value)))
				run("plan", "-out=update.tfplan")
				run("apply", "update.tfplan")
				fixture.requireValues(t, test.wantValues...)

				data, err := os.ReadFile(filepath.Join(runtime.workingDirectory, "terraform.tfstate"))
				require.NoError(t, err)

				var state struct {
					Resources []struct {
						Instances []struct {
							Private    []byte         `json:"private"`
							Attributes map[string]any `json:"attributes"`
						} `json:"instances"`
					} `json:"resources"`
				}
				require.NoError(t, json.Unmarshal(data, &state))
				require.Len(t, state.Resources, 1)
				require.Len(t, state.Resources[0].Instances, 1)
				instance := state.Resources[0].Instances[0]

				private := map[string][]byte{}
				if len(instance.Private) != 0 {
					require.NoError(t, json.Unmarshal(instance.Private, &private))
				}

				require.Nil(t, instance.Attributes["value_wo"])

				if test.attribute == "value" {
					require.Equal(t, test.value, instance.Attributes["value"])
					require.Empty(t, private["write_only_secret_hashes"])
				} else {
					require.Nil(t, instance.Attributes["value"])
					require.NotEmpty(t, private["write_only_secret_hashes"])
					require.NotContains(t, string(data), test.value)
				}
				// The retained ordinary value or new verifier makes a fresh plan a no-op.
				run("plan", "-detailed-exitcode")
				fixture.requireValues(t, test.wantValues...)

				// Comparison still detects the next changed write-only value.
				next := strings.Repeat("c", 64)
				runtime.writeConfig(t, signingSecretConfig(kind, fmt.Sprintf("value_wo=%q\ntimeouts={read=\"1m\"}", next)))
				run("apply", "-auto-approve")
				fixture.requireValues(t, append(test.wantValues, next)...)
				run("plan", "-detailed-exitcode")
			})
		}
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

	private := map[string][]byte{}

	if encoded, exists := instance["private"]; exists {
		privateText, ok := encoded.(string)
		require.True(t, ok)

		privateData, err := base64.StdEncoding.DecodeString(privateText)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(privateData, &private))
	}

	private["write_only_secret_hashes"] = []byte(testJSON([]map[string]string{{"path": "value_wo", "hash": "invalid"}}))
	instance["private"] = base64.StdEncoding.EncodeToString([]byte(testJSON(private)))
	require.NoError(t, os.WriteFile(filename, []byte(testJSON(document)), 0o600))
}
