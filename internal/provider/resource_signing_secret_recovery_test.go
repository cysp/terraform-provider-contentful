package provider_test

import (
	"encoding/json"
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

func TestAccSigningSecretLostAcknowledgementRecovery(t *testing.T) {
	t.Parallel()

	for _, kind := range []string{"app", "webhook"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()

			fixture := &signingSecretFixture{kind: kind, timestamps: true}

			var loseAcknowledgement atomic.Bool

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodPut && loseAcknowledgement.Swap(false) {
					// Commit the request, then close the connection before the provider can
					// receive the successful response. The fixture retains the new value.
					recorded := httptest.NewRecorder()
					fixture.ServeHTTP(recorded, r)
					assert.Equal(t, http.StatusOK, recorded.Code)

					connection, _, err := http.NewResponseController(w).Hijack()
					if assert.NoError(t, err) {
						assert.NoError(t, connection.Close())
					}

					return
				}

				fixture.ServeHTTP(w, r)
			}))
			t.Cleanup(server.Close)
			runtime := newTerraformTestRuntime(t)
			runtime.providerURL = server.URL
			runtime.writeConfig(t, `variable "secret" {
 type=string
 sensitive=true
 ephemeral=true
}
`+signingSecretConfig(kind, "value_wo=var.secret"))

			initial, changed := strings.Repeat("a", 64), strings.Repeat("b", 64)
			run := func(value string, arguments ...string) (string, error) {
				arguments = append(arguments, "-input=false", "-no-color", "-var=secret="+value)
				output, err := runtime.run(t.Context(), arguments...)
				require.NotContains(t, output, initial)
				require.NotContains(t, output, changed)

				return output, err
			}
			// Terraform 1.12.0-1.12.2 drop resource identity when persisting an apply error;
			// compare the provider's acknowledged attributes and private baseline.
			// Protocol tests separately assert the returned identity.
			// https://github.com/hashicorp/terraform/pull/37396
			type stateInstance struct {
				Attributes map[string]any `json:"attributes"`
				Private    []byte         `json:"private"`
			}

			readInstance := func() stateInstance {
				data, err := os.ReadFile(filepath.Join(runtime.workingDirectory, "terraform.tfstate"))
				require.NoError(t, err)
				require.NotContains(t, string(data), initial)
				require.NotContains(t, string(data), changed)

				var state struct {
					Resources []struct {
						Instances []stateInstance `json:"instances"`
					} `json:"resources"`
				}
				require.NoError(t, json.Unmarshal(data, &state))
				require.Len(t, state.Resources, 1)
				require.Len(t, state.Resources[0].Instances, 1)

				return state.Resources[0].Instances[0]
			}
			output, err := run(initial, "apply", "-auto-approve")
			require.NoError(t, err, output)
			fixture.requireValues(t, initial)

			prior := readInstance()

			loseAcknowledgement.Store(true)

			output, err = run(changed, "apply", "-auto-approve")
			require.Error(t, err, output)
			require.Contains(t, output, "Failed to write "+kind+" signing secret")
			require.False(t, loseAcknowledgement.Load())
			fixture.requireValues(t, initial, changed)
			require.Equal(t, prior, readInstance())

			// The old verifier cannot detect that the unacknowledged value is remote.
			// Reverting the input is a no-op, so it cannot be a recovery procedure.
			output, err = run(initial, "plan", "-detailed-exitcode")
			require.NoError(t, err, output)
			fixture.requireValues(t, initial, changed)

			// A fresh apply of bytes different from the acknowledged baseline performs
			// a deliberate write even when those bytes are already present remotely.
			output, err = run(changed, "apply", "-auto-approve")
			require.NoError(t, err, output)
			fixture.requireValues(t, initial, changed, changed)
			require.NotEqual(t, prior, readInstance())

			output, err = run(changed, "plan", "-detailed-exitcode")
			require.NoError(t, err, output)
			fixture.requireValues(t, initial, changed, changed)
		})
	}
}
