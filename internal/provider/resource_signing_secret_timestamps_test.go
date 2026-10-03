package provider_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
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
						{Config: signingSecretConfig(kind, fmt.Sprintf("value=%q", initial)), ConfigStateChecks: checks("2026-09-01T00:00:00Z", "2026-09-01T01:00:00Z")},
						{PreConfig: func() { fixture.requireValues(t, initial) }, Config: signingSecretConfig(kind, fmt.Sprintf("value=%q\ntimeouts={read=\"1m\"}", initial)), ConfigStateChecks: checks("2026-09-01T00:00:00Z", "2026-09-01T01:00:00Z")},
						{PreConfig: func() { fixture.requireValues(t, initial) }, Config: signingSecretConfig(kind, fmt.Sprintf("value=%q", changed)), ConfigStateChecks: checks("2026-09-02T00:00:00Z", "2026-09-02T01:00:00Z")},
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
