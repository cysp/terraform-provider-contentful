package provider_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/assert"
)

func TestAccSigningSecretUpdatedAt(t *testing.T) {
	t.Parallel()

	for _, kind := range []string{"app", "webhook"} {
		for _, withUpdatedAt := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/updated_at=%t", kind, withUpdatedAt), func(t *testing.T) {
				t.Parallel()

				fixture := &signingSecretFixture{kind: kind, withUpdatedAt: withUpdatedAt}
				initial, changed := strings.Repeat("a", 64), strings.Repeat("b", 64)
				address := "contentful_" + kind + "_signing_secret.test"
				config := func(value, extra string) string {
					return signingSecretConfig(kind, fmt.Sprintf("value=%q\n%s", value, extra)) + fmt.Sprintf(`
resource "terraform_data" "updated_at" {
 triggers_replace = [%s.updated_at]
}
`, address)
				}
				expectedUpdatedAt := func(value string) knownvalue.Check {
					if withUpdatedAt {
						return knownvalue.StringExact(value)
					}

					return knownvalue.Null()
				}
				updatedAt := tfjsonpath.New("updated_at")
				writeStep := func(value, timestamp string) resource.TestStep {
					expected := expectedUpdatedAt(timestamp)

					return resource.TestStep{
						Config: config(value, ""),
						ConfigPlanChecks: resource.ConfigPlanChecks{
							PreApply: []plancheck.PlanCheck{
								plancheck.ExpectUnknownValue(address, updatedAt),
							},
							PostApplyPreRefresh: []plancheck.PlanCheck{
								plancheck.ExpectKnownValue(address, updatedAt, expected),
							},
						},
						ConfigStateChecks: []statecheck.StateCheck{
							statecheck.ExpectKnownValue(address, updatedAt, expected),
						},
					}
				}
				testAccMockedResource(t, fixture, resource.TestCase{
					Steps: []resource.TestStep{
						writeStep(initial, "2026-09-01T01:00:00.456Z"),
						{
							PreConfig: func() { fixture.requireValues(t, initial) },
							Config:    config(initial, `timeouts={read="1m"}`),
							ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
								plancheck.ExpectResourceAction("terraform_data.updated_at", plancheck.ResourceActionNoop),
							}},
							ConfigStateChecks: []statecheck.StateCheck{
								statecheck.ExpectKnownValue(address, updatedAt, expectedUpdatedAt("2026-09-01T01:00:00.456Z")),
							},
						},
						writeStep(changed, "2026-09-01T01:00:00.987654321Z"),
					},
					CheckDestroy: func(_ *terraform.State) error {
						fixture.requireValues(t, initial, changed)
						fixture.mu.Lock()
						defer fixture.mu.Unlock()

						assert.False(t, fixture.exists)

						return nil
					},
				})
			})
		}
	}
}

// An unknown updated_at can replace a dependent even when apply skips the secret PUT.
func TestAccSigningSecretWriteOnlyTimeoutUpdateReplacesTimestampDependent(t *testing.T) {
	t.Parallel()

	for _, kind := range []string{"app", "webhook"} {
		for _, withUpdatedAt := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/updated_at=%t", kind, withUpdatedAt), func(t *testing.T) {
				t.Parallel()

				fixture := &signingSecretFixture{kind: kind, withUpdatedAt: withUpdatedAt}
				initial := strings.Repeat("a", 64)
				address := "contentful_" + kind + "_signing_secret.test"
				dependentID := statecheck.CompareValue(compare.ValuesDiffer())
				config := func(extra string) string {
					return signingSecretConfig(kind, fmt.Sprintf("value_wo=%q\n%s", initial, extra)) + fmt.Sprintf(`
resource "terraform_data" "updated_at" {
 triggers_replace = [%s.updated_at]
}
`, address)
				}

				var updatedAt knownvalue.Check = knownvalue.Null()
				if withUpdatedAt {
					updatedAt = knownvalue.StringExact("2026-09-01T01:00:00.456Z")
				}

				checks := func() []statecheck.StateCheck {
					return []statecheck.StateCheck{
						statecheck.ExpectKnownValue(address, tfjsonpath.New("updated_at"), updatedAt),
						statecheck.ExpectKnownValue("terraform_data.updated_at", tfjsonpath.New("id"), knownvalue.NotNull()),
						dependentID.AddStateValue("terraform_data.updated_at", tfjsonpath.New("id")),
					}
				}
				testAccMockedResource(t, fixture, resource.TestCase{
					Steps: []resource.TestStep{
						{Config: config(""), ConfigStateChecks: checks()},
						{
							PreConfig: func() { fixture.requireValues(t, initial) },
							Config:    config(`timeouts={read="1m"}`),
							ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
								plancheck.ExpectResourceAction(address, plancheck.ResourceActionUpdate),
								plancheck.ExpectUnknownValue(address, tfjsonpath.New("updated_at")),
								plancheck.ExpectResourceAction("terraform_data.updated_at", plancheck.ResourceActionDestroyBeforeCreate),
							}},
							ConfigStateChecks: checks(),
						},
					},
					CheckDestroy: func(_ *terraform.State) error {
						fixture.requireValues(t, initial)
						fixture.requireMutations(t, "PUT", "DELETE")

						return nil
					},
				})
			})
		}
	}
}
