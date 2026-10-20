package provider_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const appInstallationTestAddress = "contentful_app_installation.test"

func TestAccAppInstallationResourceSecretsLifecycle(t *testing.T) {
	t.Parallel()

	initial := map[string]any{"secret": "alpha", "visible": "one", "remove": "present"}
	// #nosec G101 -- synthetic Secret parameter used only against the local fixture.
	rotated := map[string]any{"secret": "é🌸bravo", "visible": "one", "remove": "present"}
	// #nosec G101 -- synthetic Secret parameter used only against the local fixture.
	changed := map[string]any{"secret": "é🌸bravo", "visible": "two"}
	noSecret := map[string]any{"visible": "two"}

	type transition struct {
		parameters map[string]any
		action     plancheck.ResourceActionType
	}

	for _, test := range []struct {
		name        string
		mask        string
		transitions []transition
	}{
		{name: "observed responses", transitions: []transition{
			{initial, plancheck.ResourceActionCreate},
			{rotated, plancheck.ResourceActionUpdate},
			{changed, plancheck.ResourceActionUpdate},
			{noSecret, plancheck.ResourceActionUpdate},
			{nil, plancheck.ResourceActionUpdate},
		}},
		// Masked Create and Update responses exercise both mutation call sites.
		{name: "different length mask", mask: "*", transitions: []transition{
			{initial, plancheck.ResourceActionCreate},
			{rotated, plancheck.ResourceActionUpdate},
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			fixture := newAppInstallationSecretFixture(t)
			fixture.maskOverride = test.mask
			identity := statecheck.CompareValue(compare.ValuesSame())

			var requests []appInstallationSecretRequest

			steps := make([]resource.TestStep, 0, len(test.transitions))

			for _, transition := range test.transitions {
				requests = append(requests, appInstallationSecretPut(transition.parameters, ""))

				var expected knownvalue.Check = knownvalue.Null()
				if transition.parameters != nil {
					expected = knownvalue.StringExact(testJSON(transition.parameters))
				}

				steps = append(steps, resource.TestStep{
					Config:            appInstallationSecretConfig(transition.parameters, ""),
					ConfigPlanChecks:  resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(appInstallationTestAddress, transition.action)}},
					ConfigStateChecks: []statecheck.StateCheck{identity.AddStateValue(appInstallationTestAddress, tfjsonpath.New("id")), statecheck.ExpectKnownValue(appInstallationTestAddress, tfjsonpath.New("parameters"), expected)},
					Check:             fixture.check(t, transition.parameters, append([]appInstallationSecretRequest(nil), requests...)),
				})
			}

			testAccMockedResource(t, fixture, resource.TestCase{Steps: steps, CheckDestroy: fixture.checkDestroyed(t)})
		})
	}
}

func TestAccAppInstallationResourceSecretsImportAndEffectivePlan(t *testing.T) {
	t.Parallel()

	fixture := newAppInstallationSecretFixture(t)
	stored := map[string]any{"secret": "remote", "visible": "one"}
	imported := map[string]any{"secret": "******", "visible": "one"}
	configured := map[string]any{"secret": "owned", "visible": "one"}

	fixture.setStored(stored)
	testAccMockedResource(t, fixture, resource.TestCase{CheckDestroy: fixture.checkDestroyed(t), Steps: []resource.TestStep{
		{
			Config: appInstallationSecretConfig(configured, ""), ResourceName: appInstallationTestAddress, ImportState: true, ImportStateId: "space/master/app", ImportStatePersist: true,
			ImportStateCheck: func(states []*terraform.InstanceState) error {
				require.Len(t, states, 1)
				assert.JSONEq(t, testJSON(imported), states[0].Attributes["parameters"], "import retains returned mask")

				return nil
			},
		},
		{
			Config: appInstallationSecretConfig(configured, `marketplace = ["terms"]
lifecycle { ignore_changes = [parameters] }`),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(appInstallationTestAddress, plancheck.ResourceActionUpdate), plancheck.ExpectKnownValue(appInstallationTestAddress, tfjsonpath.New("parameters"), knownvalue.StringExact(testJSON(imported)))}},
			Check:            fixture.check(t, stored, []appInstallationSecretRequest{appInstallationSecretPut(imported, "terms")}),
		},
		{
			Config:            appInstallationSecretConfig(configured, `marketplace = ["terms"]`),
			ConfigPlanChecks:  resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(appInstallationTestAddress, plancheck.ResourceActionUpdate)}},
			ConfigStateChecks: []statecheck.StateCheck{statecheck.ExpectKnownValue(appInstallationTestAddress, tfjsonpath.New("parameters"), knownvalue.StringExact(testJSON(configured)))},
			Check:             fixture.check(t, configured, []appInstallationSecretRequest{appInstallationSecretPut(imported, "terms"), appInstallationSecretPut(configured, "terms")}),
		},
		{
			Config: appInstallationSecretConfig(map[string]any{"secret": "ignored-change", "visible": "ignored-change"}, `marketplace = ["privacy", "terms"]
lifecycle { ignore_changes = [parameters] }`),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(appInstallationTestAddress, plancheck.ResourceActionUpdate), plancheck.ExpectKnownValue(appInstallationTestAddress, tfjsonpath.New("parameters"), knownvalue.StringExact(testJSON(configured)))}},
			Check:            fixture.check(t, configured, []appInstallationSecretRequest{appInstallationSecretPut(imported, "terms"), appInstallationSecretPut(configured, "terms"), appInstallationSecretPut(configured, "privacy,terms")}),
		},
	}})
}

func TestAccAppInstallationResourceSecretsExternalRotation(t *testing.T) {
	t.Parallel()

	fixture := newAppInstallationSecretFixture(t)
	initial := map[string]any{"secret": "alpha", "visible": "one"}
	// #nosec G101 -- synthetic external rotation used only against the local fixture.
	external := map[string]any{"secret": "externally-rotated-longer", "visible": "one"}
	changed := map[string]any{"secret": "alpha", "visible": "two"}
	testAccMockedResource(t, fixture, resource.TestCase{CheckDestroy: fixture.checkDestroyed(t), Steps: []resource.TestStep{
		{Config: appInstallationSecretConfig(initial, "")},
		{
			PreConfig: func() { fixture.setStored(external) }, Config: appInstallationSecretConfig(initial, ""),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(appInstallationTestAddress, plancheck.ResourceActionNoop)}},
			Check:            fixture.check(t, external, []appInstallationSecretRequest{appInstallationSecretPut(initial, "")}),
		},
		{
			Config:           appInstallationSecretConfig(changed, ""),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(appInstallationTestAddress, plancheck.ResourceActionUpdate)}},
			Check:            fixture.check(t, changed, []appInstallationSecretRequest{appInstallationSecretPut(initial, ""), appInstallationSecretPut(changed, "")}),
		},
	}})
}

func TestAccAppInstallationResourceSecretsMutationContradictions(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name     string
		response map[string]any
	}{
		{"readable secret", map[string]any{"secret": "different", "visible": "one"}},
		{"readable ordinary", map[string]any{"secret": "*****", "visible": "different"}},
		{"missing ordinary key", map[string]any{"secret": "*****"}},
		{"missing secret key", map[string]any{"visible": "one"}},
		{"added key", map[string]any{"secret": "*****", "visible": "one", "added": true}},
	} {
		for _, operation := range []string{"create", "update"} {
			t.Run(test.name+"/"+operation, func(t *testing.T) {
				t.Parallel()
				fixture := newAppInstallationSecretFixture(t)
				parameters := map[string]any{"secret": "alpha", "visible": "one"}
				initial := map[string]any{"secret": "before", "visible": "before"}

				var (
					steps          []resource.TestStep
					requests       []appInstallationSecretRequest
					storedAfterPut []map[string]any
				)

				if operation == "update" {
					steps = append(steps, resource.TestStep{Config: appInstallationSecretConfig(initial, ""), Check: fixture.check(t, initial, []appInstallationSecretRequest{appInstallationSecretPut(initial, "")})})
					requests = append(requests, appInstallationSecretPut(initial, ""))
					storedAfterPut = append(storedAfterPut, initial)
				}

				requests = append(requests, appInstallationSecretPut(parameters, ""))
				storedAfterPut = append(storedAfterPut, parameters)
				steps = append(steps, resource.TestStep{
					PreConfig: func() {
						fixture.mu.Lock()
						defer fixture.mu.Unlock()

						fixture.putResponseOverride = test.response
					},
					Config:      appInstallationSecretConfig(parameters, ""),
					ExpectError: regexp.MustCompile(`Provider produced inconsistent result after apply`),
				})
				testAccMockedResource(t, fixture, resource.TestCase{Steps: steps, CheckDestroy: fixture.checkDestroyed(t)})
				fixture.mu.Lock()
				defer fixture.mu.Unlock()

				assert.Equal(t, requests, fixture.requests, "exact mutation requests")
				assert.Equal(t, storedAfterPut, fixture.storedAfterPut, "backing values after each PUT")
			})
		}
	}
}

func TestAccAppInstallationResourceSecretsUnsuccessfulWrite(t *testing.T) {
	t.Parallel()

	fixture := newAppInstallationSecretFixture(t)
	fixture.failPut = true
	parameters := map[string]any{"secret": "alpha", "visible": "one"}
	testAccMockedResource(t, fixture, resource.TestCase{Steps: []resource.TestStep{{Config: appInstallationSecretConfig(parameters, ""), ExpectError: regexp.MustCompile(`Failed to create app installation`)}}})
	// Failed creation has no resource state to destroy. Its stored remote write
	// does not become success merely because a later GET could return a mask.
	require.NoError(t, fixture.check(t, parameters, []appInstallationSecretRequest{appInstallationSecretPut(parameters, "")})(nil))
	fixture.mu.Lock()
	defer fixture.mu.Unlock()

	require.Equal(t, []string{"PUT"}, fixture.methods)
}

func TestAccAppInstallationResourceSecretsReadableDrift(t *testing.T) {
	t.Parallel()

	fixture := newAppInstallationSecretFixture(t)
	configured := map[string]any{"secret": "alpha", "visible": "one", "retained": "present"}
	externallyChanged := map[string]any{"secret": "alpha", "visible": "external", "retained": "present"}
	externallyRemoved := map[string]any{"secret": "alpha", "visible": "one"}
	steps := make([]resource.TestStep, 0, 3)
	steps = append(steps, resource.TestStep{Config: appInstallationSecretConfig(configured, "")})

	for index, drift := range []map[string]any{externallyChanged, externallyRemoved} {
		requests := make([]appInstallationSecretRequest, index+2)
		for i := range requests {
			requests[i] = appInstallationSecretPut(configured, "")
		}

		steps = append(steps, resource.TestStep{
			PreConfig: func() { fixture.setStored(drift) }, Config: appInstallationSecretConfig(configured, ""),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(appInstallationTestAddress, plancheck.ResourceActionUpdate)}},
			Check:            fixture.check(t, configured, requests),
		})
	}

	testAccMockedResource(t, fixture, resource.TestCase{Steps: steps, CheckDestroy: fixture.checkDestroyed(t)})
}
