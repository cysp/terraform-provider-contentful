package provider_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccSigningSecretWriteOnlyLifecycle(t *testing.T) {
	t.Parallel()

	for _, kind := range []string{"app", "webhook"} {
		for _, timestamps := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/timestamps=%t", kind, timestamps), func(t *testing.T) {
				t.Parallel()

				fixture := &signingSecretFixture{kind: kind, timestamps: timestamps}
				initial, changed, third := strings.Repeat("a", 64), strings.Repeat("b", 64), strings.Repeat("c", 64)
				address := "contentful_" + kind + "_signing_secret.test"
				step := func(attribute, value, extra string, priorRequests []string) resource.TestStep {
					var expected knownvalue.Check = knownvalue.Null()
					if attribute == "value" {
						expected = knownvalue.StringExact(value)
					}

					return resource.TestStep{
						PreConfig: func() { fixture.requireValues(t, priorRequests...) },
						Config:    signingSecretConfig(kind, fmt.Sprintf("%s=%q\n%s", attribute, value, extra)),
						ConfigStateChecks: []statecheck.StateCheck{
							statecheck.ExpectKnownValue(address, tfjsonpath.New("value"), expected),
							statecheck.ExpectKnownValue(address, tfjsonpath.New("value_wo"), knownvalue.Null()),
						},
					}
				}
				unchanged := step("value_wo", initial, "", []string{initial})
				unchanged.ConfigPlanChecks = resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(address, plancheck.ResourceActionNoop)}}
				recreated := step("value_wo", initial, "", []string{initial, changed, third, initial})
				recreated.PreConfig = func() {
					fixture.requireValues(t, initial, changed, third, initial)
					fixture.mu.Lock()
					defer fixture.mu.Unlock()

					fixture.exists = false
				}
				recreated.ConfigPlanChecks = resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(address, plancheck.ResourceActionCreate)}}
				testAccMockedResource(t, fixture, resource.TestCase{
					Steps: []resource.TestStep{
						step("value_wo", initial, "", nil),
						unchanged,
						step("value_wo", changed, "", []string{initial}),
						step("value", changed, "", []string{initial, changed}),
						step("value", third, "", []string{initial, changed}),
						step("value_wo", third, "", []string{initial, changed, third}),
						step("value_wo", initial, "", []string{initial, changed, third}),
						step("value_wo", initial, `timeouts={read="1m"}`, []string{initial, changed, third, initial}),
						recreated,
					},
					CheckDestroy: func(_ *terraform.State) error {
						fixture.requireValues(t, initial, changed, third, initial, initial)

						return nil
					},
				})
			})
		}
	}
}

func TestAccSigningSecretWriteOnlyValidation(t *testing.T) {
	t.Parallel()

	for _, kind := range []string{"app", "webhook"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()

			fixture := &signingSecretFixture{kind: kind}
			value := strings.Repeat("a", 64)
			testAccMockedResource(t, fixture, resource.TestCase{
				Steps: []resource.TestStep{
					{Config: signingSecretConfig(kind, ""), ExpectError: regexp.MustCompile(`Invalid Attribute Combination`)},
					{Config: signingSecretConfig(kind, fmt.Sprintf("value=%q\nvalue_wo=%q", value, value)), ExpectError: regexp.MustCompile(`Invalid Attribute Combination`)},
					{Config: signingSecretConfig(kind, `value_wo="invalid"`), ExpectError: regexp.MustCompile(`Invalid .* signing secret value`)},
				},
			})
			fixture.requireValues(t)
		})
	}
}

func TestAccSigningSecretWriteOnlyImport(t *testing.T) {
	t.Parallel()

	for _, kind := range []string{"app", "webhook"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()

			fixture := &signingSecretFixture{kind: kind, exists: true, timestamps: true}
			value := strings.Repeat("a", 64)
			config := signingSecretConfig(kind, fmt.Sprintf("value_wo=%q", value))

			importID := "organization/app"
			if kind == "webhook" {
				importID = "space"
			}

			testAccMockedResource(t, fixture, resource.TestCase{Steps: []resource.TestStep{
				{Config: config, ResourceName: "contentful_" + kind + "_signing_secret.test", ImportState: true, ImportStateId: importID, ImportStatePersist: true},
				{PreConfig: func() { fixture.requireValues(t) }, Config: config},
			}})
			fixture.requireValues(t, value)
		})
	}
}

// Unknown write-only inputs must resolve at apply before deciding whether the
// secret differs from the acknowledged value.
func TestAccSigningSecretUnknownWriteOnlyValue(t *testing.T) {
	t.Parallel()

	for _, kind := range []string{"app", "webhook"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()

			fixture := &signingSecretFixture{kind: kind, timestamps: true}
			initial, changed := strings.Repeat("a", 64), strings.Repeat("b", 64)
			config := func(value string, nonce int) string {
				return fmt.Sprintf(`resource "terraform_data" "secret" {
 input={value=%q, nonce=%d}
}
`, value, nonce) + signingSecretConfig(kind, "value_wo=terraform_data.secret.output.value")
			}
			testAccMockedResource(t, fixture, resource.TestCase{Steps: []resource.TestStep{
				{Config: config(initial, 1)},
				{PreConfig: func() { fixture.requireValues(t, initial) }, Config: config(initial, 2)},
				{PreConfig: func() { fixture.requireValues(t, initial) }, Config: config(changed, 3)},
			}})
			fixture.requireValues(t, initial, changed)
		})
	}
}
