package provider_test

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAccSigningSecretWriteOnlySavedPlan(t *testing.T) {
	t.Parallel()

	for _, kind := range []string{"app", "webhook"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()

			fixture := &signingSecretFixture{kind: kind, timestamps: true}
			server := httptest.NewServer(fixture)
			t.Cleanup(server.Close)
			runtime := newTerraformTestRuntime(t)
			runtime.providerURL = server.URL
			runtime.logPath = filepath.Join(runtime.workingDirectory, "provider.log")
			config := func(extra string) {
				runtime.writeConfig(t, `variable "secret" {
 type=string
 sensitive=true
 ephemeral=true
}
`+signingSecretConfig(kind, "value_wo=var.secret\n"+extra))
			}
			initial, changed, third := strings.Repeat("a", 64), strings.Repeat("b", 64), strings.Repeat("c", 64)
			run := func(args ...string) string {
				output, err := runtime.run(t.Context(), args...)
				require.NoError(t, err, output)

				for _, secret := range []string{initial, changed, third} {
					require.NotContains(t, output, secret)
				}

				return output
			}

			config("")
			run("apply", "-auto-approve", "-input=false", "-no-color", "-var=secret="+initial)
			fixture.requireValues(t, initial)
			config(`timeouts={read="1m"}`)
			run("plan", "-input=false", "-no-color", "-var=secret="+initial, "-out=change.tfplan")
			// A saved update must use the actual apply-time ephemeral value.
			run("apply", "-input=false", "-no-color", "-var=secret="+changed, "change.tfplan")
			fixture.requireValues(t, initial, changed)
			run("plan", "-input=false", "-no-color", "-var=secret="+changed, "-out=noop.tfplan")
			// Terraform does not invoke Update when applying this saved no-op plan.
			run("apply", "-input=false", "-no-color", "-var=secret="+third, "noop.tfplan")
			fixture.requireValues(t, initial, changed)
			run("apply", "-auto-approve", "-input=false", "-no-color", "-var=secret="+third)
			fixture.requireValues(t, initial, changed, third)

			logs, err := os.ReadFile(runtime.logPath)
			require.NoError(t, err)
			state, err := os.ReadFile(filepath.Join(runtime.workingDirectory, "terraform.tfstate"))
			require.NoError(t, err)

			var decoded struct {
				Resources []struct {
					Instances []struct {
						Private []byte `json:"private"`
					} `json:"instances"`
				} `json:"resources"`
			}
			require.NoError(t, json.Unmarshal(state, &decoded))

			for _, secret := range []string{initial, changed, third} {
				require.NotContains(t, string(logs), secret)
				require.NotContains(t, string(state), secret)

				for _, resource := range decoded.Resources {
					for _, instance := range resource.Instances {
						private := map[string][]byte{}
						require.NoError(t, json.Unmarshal(instance.Private, &private))

						for _, value := range private {
							require.NotContains(t, string(value), secret)
						}
					}
				}
			}

			output, err := runtime.run(t.Context(), "show", "-json", "change.tfplan")
			require.NoError(t, err, output)

			for _, secret := range []string{initial, changed, third} {
				require.NotContains(t, output, secret)
			}

			require.Contains(t, output, fmt.Sprintf(`"type":"contentful_%s_signing_secret"`, kind))
		})
	}
}
