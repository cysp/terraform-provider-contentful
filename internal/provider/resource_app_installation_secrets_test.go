package provider_test

import (
	"encoding/json"
	"net/http/httptest"
	"regexp"
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
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
					expected = installationParameterCheck(transition.parameters)
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
				assert.Equal(t, `"******"`, states[0].Attributes["parameters.secret"], "import retains returned mask")
				assert.Equal(t, `"one"`, states[0].Attributes["parameters.visible"])

				return nil
			},
		},
		{
			Config: appInstallationSecretConfig(configured, `marketplace = ["terms"]
lifecycle { ignore_changes = [parameters] }`),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(appInstallationTestAddress, plancheck.ResourceActionUpdate), plancheck.ExpectKnownValue(appInstallationTestAddress, tfjsonpath.New("parameters"), installationParameterCheck(imported))}},
			Check:            fixture.check(t, stored, []appInstallationSecretRequest{appInstallationSecretPut(imported, "terms")}),
		},
		{
			Config:            appInstallationSecretConfig(configured, `marketplace = ["terms"]`),
			ConfigPlanChecks:  resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(appInstallationTestAddress, plancheck.ResourceActionUpdate)}},
			ConfigStateChecks: []statecheck.StateCheck{statecheck.ExpectKnownValue(appInstallationTestAddress, tfjsonpath.New("parameters"), installationParameterCheck(configured))},
			Check:             fixture.check(t, configured, []appInstallationSecretRequest{appInstallationSecretPut(imported, "terms"), appInstallationSecretPut(configured, "terms")}),
		},
		{
			Config: appInstallationSecretConfig(map[string]any{"secret": "ignored-change", "visible": "ignored-change"}, `marketplace = ["privacy", "terms"]
lifecycle { ignore_changes = [parameters] }`),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(appInstallationTestAddress, plancheck.ResourceActionUpdate), plancheck.ExpectKnownValue(appInstallationTestAddress, tfjsonpath.New("parameters"), installationParameterCheck(configured))}},
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

func TestAccAppInstallationResourceEmptyMarketplacePreservesRemoteRotation(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, before, after string
	}{
		{name: "omitted to empty", after: "marketplace = []"},
		{name: "empty to omitted", before: "marketplace = []"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			fixture := newAppInstallationSecretFixture(t)
			initial := map[string]any{"secret": "alpha", "visible": "one"}
			external := map[string]any{"secret": "rotated", "visible": "one"}
			testAccMockedResource(t, fixture, resource.TestCase{CheckDestroy: fixture.checkDestroyed(t), Steps: []resource.TestStep{
				{Config: appInstallationSecretConfig(initial, test.before)},
				{
					PreConfig: func() { fixture.setStored(external) },
					Config:    appInstallationSecretConfig(initial, test.after),
					ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(appInstallationTestAddress, plancheck.ResourceActionNoop),
					}},
					ConfigStateChecks: []statecheck.StateCheck{
						statecheck.ExpectKnownValue(appInstallationTestAddress, tfjsonpath.New("marketplace"), knownvalue.SetExact(nil)),
					},
					Check: fixture.check(t, external, []appInstallationSecretRequest{appInstallationSecretPut(initial, "")}),
				},
			}})
		})
	}
}

func TestAccAppInstallationResourceContradictoryUpdate(t *testing.T) {
	t.Parallel()

	fixture := newAppInstallationSecretFixture(t)
	server := httptest.NewServer(fixture)
	t.Cleanup(server.Close)

	runtime := newTerraformTestRuntime(t)
	runtime.providerURL = server.URL
	initial := map[string]any{"secret": "alpha", "visible": "one"}
	updated := map[string]any{"secret": "alpha", "visible": "two"}

	runtime.writeConfig(t, appInstallationSecretConfig(initial, ""))
	output, err := runtime.run(t.Context(), "apply", "-json", "-auto-approve", "-input=false")
	require.NoError(t, err, output)

	firstPut := appInstallationSecretPut(initial, "")
	require.NoError(t, fixture.check(t, initial, []appInstallationSecretRequest{firstPut})(nil))

	fixture.mu.Lock()
	fixture.putResponseOverride = []byte(testJSON(map[string]any{"secret": "*****"}))
	fixture.mu.Unlock()

	runtime.writeConfig(t, appInstallationSecretConfig(updated, ""))
	output, err = runtime.run(t.Context(), "apply", "-json", "-auto-approve", "-input=false")
	require.Error(t, err)
	assert.Contains(t, output, "Contentful returned different App Installation parameters")
	require.NoError(t, fixture.check(t, updated, []appInstallationSecretRequest{firstPut, appInstallationSecretPut(updated, "")})(nil))
}

func TestAccAppInstallationResourceReturnedParametersInRecoveryState(t *testing.T) {
	t.Parallel()

	const returned = "synthetic-response-only-credential"

	fixture := newAppInstallationSecretFixture(t)
	fixture.putResponseOverride = []byte(testJSON(map[string]string{"secret": returned}))
	server := httptest.NewServer(fixture)
	t.Cleanup(server.Close)

	runtime := newTerraformTestRuntime(t)
	runtime.providerURL = server.URL
	runtime.writeConfig(t, `
resource "contentful_app_installation" "test" {
  space_id          = "space"
  environment_id    = "master"
  app_definition_id = "app"
}`)

	output, err := runtime.run(t.Context(), "apply", "-json", "-auto-approve", "-input=false")
	require.Error(t, err)
	assert.Contains(t, output, "Contentful returned different App Installation parameters")
	assert.NotContains(t, output, returned)
	require.NoError(t, fixture.check(t, nil, []appInstallationSecretRequest{appInstallationSecretPut(nil, "")})(nil))

	output, err = runtime.run(t.Context(), "show", "-json")
	require.NoError(t, err, output)

	var state tfjson.State
	require.NoError(t, json.Unmarshal([]byte(output), &state))
	require.NotNil(t, state.Values)
	require.NotNil(t, state.Values.RootModule)
	require.Len(t, state.Values.RootModule.Resources, 1)
	instance := state.Values.RootModule.Resources[0]
	assert.True(t, instance.Tainted)
	assert.Equal(t, "space/master/app", instance.AttributeValues["id"])
	assert.Equal(t, map[string]any{"space_id": "space", "environment_id": "master", "app_definition_id": "app"}, instance.IdentityValues)
	assert.Equal(t, map[string]any{"secret": testJSON(returned)}, instance.AttributeValues["parameters"])
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

func TestAccAppInstallationResourceInvalidParameterDiagnostic(t *testing.T) {
	t.Parallel()
	runtime := newTerraformTestRuntime(t)

	const sentinel = "private-sentinel-malformed-json"

	runtime.writeConfig(t, `
variable "credential" {
  type = string
  sensitive = true
}
resource "contentful_app_installation" "invalid" {
  space_id = "space"
  environment_id = "master"
  app_definition_id = "app"
  parameters = { apiKey = var.credential }
}
`)

	for _, format := range []string{"-no-color", "-json"} {
		output, err := runtime.run(t.Context(), "plan", format, "-input=false", "-var=credential="+sentinel)
		require.Error(t, err)
		assert.Contains(t, output, "Invalid installation parameter JSON")
		assert.NotContains(t, output, sentinel)
	}
}

func TestAccAppInstallationResourceFormattingAndMarketplace(t *testing.T) {
	t.Parallel()
	fixture := newAppInstallationSecretFixture(t)
	remote := map[string]any{"visible": "one"}
	initialPut := appInstallationSecretPut(remote, "")
	puts := []appInstallationSecretRequest{initialPut}

	const formatted = `
resource "contentful_app_installation" "test" {
  space_id = "space"
  environment_id = "master"
  app_definition_id = "app"
  parameters = { visible = "  \"one\" " }
}
`
	testAccMockedResource(t, fixture, resource.TestCase{Steps: []resource.TestStep{
		{Config: appInstallationSecretConfig(remote, ""), Check: fixture.check(t, remote, puts)},
		{
			Config: formatted, Check: fixture.check(t, remote, puts),
			ConfigStateChecks: []statecheck.StateCheck{statecheck.ExpectKnownValue(appInstallationTestAddress, tfjsonpath.New("parameters").AtMapKey("visible"), knownvalue.StringExact(`  "one" `))},
		},
		{Config: formatted, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(appInstallationTestAddress, plancheck.ResourceActionNoop)}}},
		{Config: appInstallationSecretConfig(remote, `marketplace = ["terms"]`), Check: fixture.check(t, remote, []appInstallationSecretRequest{initialPut, appInstallationSecretPut(remote, "terms")})},
	}})
}
