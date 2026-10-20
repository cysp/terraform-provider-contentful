package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccAppInstallationResourceTimeoutOnlyPreservesRemoteRotation(t *testing.T) {
	t.Parallel()
	fixture := newAppInstallationSecretFixture(t)
	intended := map[string]any{"secret": "alpha", "visible": "one"}
	external := map[string]any{"secret": "rotated", "visible": "one"}
	ignored := map[string]any{"secret": "ignored", "visible": "one"}
	steps := make([]resource.TestStep, 0, 5)

	steps = append(steps, resource.TestStep{Config: appInstallationSecretConfig(intended, "")})
	for _, change := range []struct {
		settings string
		timeouts knownvalue.Check
	}{
		{settings: `timeouts = { update = "3m" }`, timeouts: knownvalue.ObjectExact(map[string]knownvalue.Check{"create": knownvalue.Null(), "read": knownvalue.Null(), "update": knownvalue.StringExact("3m"), "delete": knownvalue.Null()})},
		{settings: `timeouts = { update = "4m" }`, timeouts: knownvalue.ObjectExact(map[string]knownvalue.Check{"create": knownvalue.Null(), "read": knownvalue.Null(), "update": knownvalue.StringExact("4m"), "delete": knownvalue.Null()})},
		{settings: "", timeouts: knownvalue.Null()},
	} {
		steps = append(steps, resource.TestStep{
			PreConfig: func() { fixture.setStored(external) },
			Config: appInstallationSecretConfig(ignored, change.settings+`
lifecycle { ignore_changes = [parameters] }`),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(appInstallationTestAddress, plancheck.ResourceActionUpdate)}},
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(appInstallationTestAddress, tfjsonpath.New("parameters"), knownvalue.StringExact(testJSON(intended))),
				statecheck.ExpectKnownValue(appInstallationTestAddress, tfjsonpath.New("timeouts"), change.timeouts),
				statecheck.ExpectIdentityValueMatchesState(appInstallationTestAddress, tfjsonpath.New("space_id")),
				statecheck.ExpectIdentityValueMatchesState(appInstallationTestAddress, tfjsonpath.New("environment_id")),
				statecheck.ExpectIdentityValueMatchesState(appInstallationTestAddress, tfjsonpath.New("app_definition_id")),
			},
			Check: fixture.check(t, external, []appInstallationSecretRequest{appInstallationSecretPut(intended, "")}),
		})
	}

	steps = append(steps, resource.TestStep{
		Config: appInstallationSecretConfig(ignored, `marketplace = ["terms"]
lifecycle { ignore_changes = [parameters] }`),
		ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(appInstallationTestAddress, plancheck.ResourceActionUpdate)}},
		Check:            fixture.check(t, intended, []appInstallationSecretRequest{appInstallationSecretPut(intended, ""), appInstallationSecretPut(intended, "terms")}),
	})
	testAccMockedResource(t, fixture, resource.TestCase{CheckDestroy: fixture.checkDestroyed(t), Steps: steps})
}

func TestAccAppInstallationResourceImportedTimeoutOnlyPreservesRemoteRotation(t *testing.T) {
	t.Parallel()
	fixture := newAppInstallationSecretFixture(t)
	original := map[string]any{"secret": "remote", "visible": "one"}
	external := map[string]any{"secret": "externally changed", "visible": "one"}

	fixture.setStored(original)

	ignored := map[string]any{"secret": "ignored", "visible": "one"}
	testAccMockedResource(t, fixture, resource.TestCase{CheckDestroy: fixture.checkDestroyed(t), Steps: []resource.TestStep{
		{
			Config:       appInstallationSecretConfig(ignored, `lifecycle { ignore_changes = [parameters] }`),
			ResourceName: appInstallationTestAddress, ImportState: true, ImportStateId: "space/master/app", ImportStatePersist: true,
			ImportStateCheck: func(states []*terraform.InstanceState) error {
				require.Len(t, states, 1)
				assert.JSONEq(t, testJSON(map[string]any{"secret": "******", "visible": "one"}), states[0].Attributes["parameters"], "import retains original mask")

				return nil
			},
		},
		{
			PreConfig: func() { fixture.setStored(external) },
			Config: appInstallationSecretConfig(ignored, `timeouts = { update = "3m" }
lifecycle { ignore_changes = [parameters] }`),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(appInstallationTestAddress, plancheck.ResourceActionUpdate)}},
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(appInstallationTestAddress, tfjsonpath.New("parameters"), knownvalue.StringExact(testJSON(map[string]any{"secret": "******************", "visible": "one"}))),
				statecheck.ExpectKnownValue(appInstallationTestAddress, tfjsonpath.New("timeouts"), knownvalue.ObjectExact(map[string]knownvalue.Check{"create": knownvalue.Null(), "read": knownvalue.Null(), "update": knownvalue.StringExact("3m"), "delete": knownvalue.Null()})),
				statecheck.ExpectIdentityValueMatchesState(appInstallationTestAddress, tfjsonpath.New("space_id")),
				statecheck.ExpectIdentityValueMatchesState(appInstallationTestAddress, tfjsonpath.New("environment_id")),
				statecheck.ExpectIdentityValueMatchesState(appInstallationTestAddress, tfjsonpath.New("app_definition_id")),
			},
			Check: fixture.check(t, external, nil),
		},
	}})
}
