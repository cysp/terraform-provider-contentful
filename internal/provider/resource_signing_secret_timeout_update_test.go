package provider_test

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// An ignored write-only input is still available during apply. A timeout update
// compares those bytes normally, including when applying a saved plan.
func TestAccSigningSecretWriteOnlyTimeoutUpdate(t *testing.T) {
	t.Parallel()

	for _, kind := range []string{"app", "webhook"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()

			for _, timestamps := range []bool{true, false} {
				for _, changed := range []bool{false, true} {
					t.Run(fmt.Sprintf("timestamps=%t/changed=%t", timestamps, changed), func(t *testing.T) {
						fixture := &signingSecretFixture{kind: kind, timestamps: timestamps}
						server := httptest.NewServer(fixture)
						t.Cleanup(server.Close)
						runtime := newTerraformTestRuntime(t)
						runtime.providerURL = server.URL
						config := func(extra string) {
							runtime.writeConfig(t, `variable "secret" {
 type=string
 sensitive=true
 ephemeral=true
}
`+signingSecretConfig(kind, "value_wo=var.secret\n"+extra))
						}
						run := func(args ...string) {
							output, err := runtime.run(t.Context(), args...)
							require.NoError(t, err, output)
						}
						initial, different := strings.Repeat("a", 64), strings.Repeat("b", 64)

						config("")
						run("apply", "-auto-approve", "-input=false", "-no-color", "-var=secret="+initial)
						fixture.requireValues(t, initial)

						config("timeouts={read=\"1m\"}\nlifecycle {ignore_changes=[value_wo]}")
						run("plan", "-input=false", "-no-color", "-var=secret="+different, "-out=update.tfplan")

						value := initial
						writes := []string{initial}

						if changed {
							value = different
							writes = append(writes, value)
						}

						run("apply", "-input=false", "-no-color", "-var=secret="+value, "update.tfplan")
						fixture.requireValues(t, writes...)

						// Ignoring the input can still suppress the resource update itself.
						run("plan", "-input=false", "-no-color", "-var=secret="+different, "-out=noop.tfplan")
						run("apply", "-input=false", "-no-color", "-var=secret="+initial, "noop.tfplan")
						fixture.requireValues(t, writes...)
					})
				}
			}
		})
	}
}
