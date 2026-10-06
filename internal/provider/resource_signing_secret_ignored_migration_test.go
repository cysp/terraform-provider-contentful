package provider_test

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

func TestAccSigningSecretIgnoredMigration(t *testing.T) {
	t.Parallel()

	for _, kind := range []string{"app", "webhook"} {
		for _, changed := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/changed=%t", kind, changed), func(t *testing.T) {
				t.Parallel()

				initial, candidate := strings.Repeat("a", 64), strings.Repeat("a", 64)
				expectedWrites := []string{initial}

				if changed {
					candidate = strings.Repeat("b", 64)
					expectedWrites = append(expectedWrites, candidate)
				}

				fixture := &signingSecretFixture{kind: kind, timestamps: true}
				address := "contentful_" + kind + "_signing_secret.test"
				nullValues := []statecheck.StateCheck{
					statecheck.ExpectKnownValue(address, tfjsonpath.New("value"), knownvalue.Null()),
					statecheck.ExpectKnownValue(address, tfjsonpath.New("value_wo"), knownvalue.Null()),
				}
				testAccMockedResource(t, fixture, resource.TestCase{Steps: []resource.TestStep{
					{Config: signingSecretConfig(kind, fmt.Sprintf("value=%q", initial))},
					{
						// Removing ordinary value schedules Update even when planning Config
						// omits the ignored write-only input. Apply must compare the actual bytes.
						PreConfig:         func() { fixture.requireValues(t, initial) },
						Config:            signingSecretConfig(kind, fmt.Sprintf("value_wo=%q\nlifecycle {ignore_changes=[value_wo]}", candidate)),
						ConfigPlanChecks:  resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(address, plancheck.ResourceActionUpdate)}},
						ConfigStateChecks: nullValues,
					},
					{
						// The inverse migration with ordinary value ignored preserves its null
						// effective plan; it must not restore plaintext or write a new secret.
						PreConfig:         func() { fixture.requireValues(t, expectedWrites...) },
						Config:            signingSecretConfig(kind, fmt.Sprintf("value=%q\nlifecycle {ignore_changes=[value]}", initial)),
						ConfigPlanChecks:  resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(address, plancheck.ResourceActionNoop)}},
						ConfigStateChecks: nullValues,
					},
					{
						PreConfig:         func() { fixture.requireValues(t, expectedWrites...) },
						Config:            signingSecretConfig(kind, fmt.Sprintf("value=%q", candidate)),
						ConfigStateChecks: []statecheck.StateCheck{statecheck.ExpectKnownValue(address, tfjsonpath.New("value"), knownvalue.StringExact(candidate))},
					},
				}})
				fixture.requireValues(t, expectedWrites...)
				fixture.mu.Lock()
				defer fixture.mu.Unlock()

				require.False(t, fixture.exists)
			})
		}
	}
}

func TestAccSigningSecretIgnoredMigrationReplacement(t *testing.T) {
	t.Parallel()

	for _, kind := range []string{"app", "webhook"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()

			fixture := &signingSecretFixture{kind: kind, timestamps: true}
			server := httptest.NewServer(fixture)
			t.Cleanup(server.Close)
			runtime := newTerraformTestRuntime(t)
			runtime.providerURL = server.URL
			initial, candidate := strings.Repeat("a", 64), strings.Repeat("b", 64)
			runtime.writeConfig(t, signingSecretConfig(kind, fmt.Sprintf("value=%q", initial)))
			output, err := runtime.run(t.Context(), "apply", "-auto-approve", "-input=false", "-no-color")
			require.NoError(t, err, output)
			fixture.requireValues(t, initial)

			runtime.writeConfig(t, signingSecretConfig(kind, fmt.Sprintf("value_wo=%q\nlifecycle {ignore_changes=[value]}", candidate)))
			output, err = runtime.run(t.Context(), "plan", "-input=false", "-no-color", "-out=conflict.tfplan")
			require.NoError(t, err, output)
			output, err = runtime.run(t.Context(), "apply", "-input=false", "-no-color", "conflict.tfplan")
			require.Error(t, err, output)
			require.Contains(t, output, "Conflicting signing secret values")
			fixture.requireMutations(t, "PUT")

			output, err = runtime.run(t.Context(), "apply", "-auto-approve", "-input=false", "-no-color", "-replace=contentful_"+kind+"_signing_secret.test")
			require.NoError(t, err, output)
			fixture.requireValues(t, initial, candidate)
			fixture.requireMutations(t, "PUT", "DELETE", "PUT")
			output, err = runtime.run(t.Context(), "show", "-json")
			require.NoError(t, err, output)
			require.NotContains(t, output, initial)
			require.NotContains(t, output, candidate)
			output, err = runtime.run(t.Context(), "destroy", "-auto-approve", "-input=false", "-no-color")
			require.NoError(t, err, output)
			fixture.requireMutations(t, "PUT", "DELETE", "PUT", "DELETE")
		})
	}
}
