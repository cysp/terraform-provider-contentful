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

func TestAccSigningSecretTimestamps(t *testing.T) {
	t.Parallel()

	for _, kind := range []string{"app", "webhook"} {
		for _, timestamps := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/timestamps=%t", kind, timestamps), func(t *testing.T) {
				t.Parallel()

				fixture := &signingSecretFixture{kind: kind, timestamps: timestamps}
				initial, changed := strings.Repeat("a", 64), strings.Repeat("b", 64)
				address := "contentful_" + kind + "_signing_secret.test"
				config := func(value, extra string) string {
					return signingSecretConfig(kind, fmt.Sprintf("value=%q\n%s", value, extra)) + fmt.Sprintf(`
resource "terraform_data" "timestamps" {
 triggers_replace = [%[1]s.created_at, %[1]s.updated_at]
}
`, address)
				}
				checks := func(createdAt, updatedAt string) []statecheck.StateCheck {
					var expectedCreatedAt, expectedUpdatedAt knownvalue.Check = knownvalue.Null(), knownvalue.Null()
					if timestamps {
						expectedCreatedAt = knownvalue.StringExact(createdAt)
						expectedUpdatedAt = knownvalue.StringExact(updatedAt)
					}

					return []statecheck.StateCheck{
						statecheck.ExpectKnownValue(address, tfjsonpath.New("created_at"), expectedCreatedAt),
						statecheck.ExpectKnownValue(address, tfjsonpath.New("updated_at"), expectedUpdatedAt),
					}
				}
				testAccMockedResource(t, fixture, resource.TestCase{
					Steps: []resource.TestStep{
						{Config: config(initial, ""), ConfigStateChecks: checks("2026-09-01T00:00:00Z", "2026-09-01T01:00:00Z")},
						{
							PreConfig: func() { fixture.requireValues(t, initial) },
							Config:    config(initial, `timeouts={read="1m"}`),
							ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
								plancheck.ExpectResourceAction("terraform_data.timestamps", plancheck.ResourceActionNoop),
							}},
							ConfigStateChecks: checks("2026-09-01T00:00:00Z", "2026-09-01T01:00:00Z"),
						},
						{PreConfig: func() { fixture.requireValues(t, initial) }, Config: config(changed, ""), ConfigStateChecks: checks("2026-09-02T00:00:00Z", "2026-09-02T01:00:00Z")},
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

// Unknown timestamps can replace a dependent even when apply skips the secret PUT.
func TestAccSigningSecretWriteOnlyTimeoutUpdateReplacesTimestampDependent(t *testing.T) {
	t.Parallel()

	for _, kind := range []string{"app", "webhook"} {
		for _, timestamps := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/timestamps=%t", kind, timestamps), func(t *testing.T) {
				t.Parallel()

				fixture := &signingSecretFixture{kind: kind, timestamps: timestamps}
				initial := strings.Repeat("a", 64)
				address := "contentful_" + kind + "_signing_secret.test"
				dependentID := statecheck.CompareValue(compare.ValuesDiffer())
				config := func(extra string) string {
					return signingSecretConfig(kind, fmt.Sprintf("value_wo=%q\n%s", initial, extra)) + fmt.Sprintf(`
resource "terraform_data" "timestamps" {
 triggers_replace = [%[1]s.created_at, %[1]s.updated_at]
}
`, address)
				}

				var createdAt, updatedAt knownvalue.Check = knownvalue.Null(), knownvalue.Null()
				if timestamps {
					createdAt = knownvalue.StringExact("2026-09-01T00:00:00Z")
					updatedAt = knownvalue.StringExact("2026-09-01T01:00:00Z")
				}

				checks := func() []statecheck.StateCheck {
					return []statecheck.StateCheck{
						statecheck.ExpectKnownValue(address, tfjsonpath.New("created_at"), createdAt),
						statecheck.ExpectKnownValue(address, tfjsonpath.New("updated_at"), updatedAt),
						statecheck.ExpectKnownValue("terraform_data.timestamps", tfjsonpath.New("id"), knownvalue.NotNull()),
						dependentID.AddStateValue("terraform_data.timestamps", tfjsonpath.New("id")),
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
								plancheck.ExpectUnknownValue(address, tfjsonpath.New("created_at")),
								plancheck.ExpectUnknownValue(address, tfjsonpath.New("updated_at")),
								plancheck.ExpectResourceAction("terraform_data.timestamps", plancheck.ResourceActionDestroyBeforeCreate),
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
