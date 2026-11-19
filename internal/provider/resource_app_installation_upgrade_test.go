package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

//nolint:tparallel // The parent sets the namespace shared by both fixture runs.
func TestAccAppInstallationResourceParameterStateUpgrade(t *testing.T) {
	// v0.1.0 imported this fixture. Importing the same installation with v0.0.74
	// produced identical state apart from Terraform's randomly assigned lineage.
	t.Setenv(resource.EnvTfAccProviderNamespace, "cysp")

	directory := config.StaticDirectory("testdata/app-installation-parameter-upgrade")
	parameters := tfjsonpath.New("parameters")

	t.Run("without refresh", func(t *testing.T) {
		t.Parallel()

		fixture := newAppInstallationSecretFixture(t)
		fixture.setStored(map[string]any{"retained": "external", "visible": "external"})

		expected := installationParameterCheck(map[string]any{"retained": "original", "visible": "one"})

		testAccMockedResource(t, fixture, resource.TestCase{
			AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
			Steps: []resource.TestStep{{
				ConfigDirectory: directory,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(appInstallationTestAddress, plancheck.ResourceActionNoop),
					plancheck.ExpectKnownValue(appInstallationTestAddress, parameters, expected),
					plancheck.ExpectKnownValue(appInstallationTestAddress, tfjsonpath.New("marketplace"), knownvalue.SetExact(nil)),
				}},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(appInstallationTestAddress, parameters, expected),
					statecheck.ExpectKnownValue(appInstallationTestAddress, tfjsonpath.New("marketplace"), knownvalue.SetExact(nil)),
					statecheck.ExpectKnownValue(appInstallationTestAddress, tfjsonpath.New("id"), knownvalue.StringExact("space/master/app")),
					statecheck.ExpectIdentityValueMatchesState(appInstallationTestAddress, tfjsonpath.New("space_id")),
					statecheck.ExpectIdentityValueMatchesState(appInstallationTestAddress, tfjsonpath.New("environment_id")),
					statecheck.ExpectIdentityValueMatchesState(appInstallationTestAddress, tfjsonpath.New("app_definition_id")),
				},
				PostApplyFunc: func() {
					fixture.mu.Lock()
					defer fixture.mu.Unlock()

					require.Empty(t, fixture.methods)
				},
			}},
			CheckDestroy: fixture.checkDestroyed(t),
		})
	})

	t.Run("with ignored drift and a managed edit", func(t *testing.T) {
		t.Parallel()

		fixture := newAppInstallationSecretFixture(t)
		fixture.setStored(map[string]any{"retained": "external", "visible": "one"})

		edited := map[string]any{"retained": "external", "visible": "two"}
		retained := map[string]any{"retained": "external"}
		firstPut := appInstallationSecretPut(edited, "")

		testAccMockedResource(t, fixture, resource.TestCase{Steps: []resource.TestStep{
			{
				ConfigDirectory: directory,
				ConfigVariables: config.Variables{"parameters": config.MapVariable(map[string]config.Variable{"visible": config.StringVariable("two")})},
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(appInstallationTestAddress, plancheck.ResourceActionUpdate),
					plancheck.ExpectKnownValue(appInstallationTestAddress, parameters, installationParameterCheck(edited)),
				}},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(appInstallationTestAddress, parameters, installationParameterCheck(edited)),
				},
				Check: fixture.check(t, edited, []appInstallationSecretRequest{firstPut}),
			},
			{
				Config: appInstallationSecretConfig(nil, `lifecycle { ignore_changes = [parameters["retained"]] }`),
				Check:  fixture.check(t, retained, []appInstallationSecretRequest{firstPut, appInstallationSecretPut(retained, "")}),
			},
		}, CheckDestroy: fixture.checkDestroyed(t)})
	})
}
