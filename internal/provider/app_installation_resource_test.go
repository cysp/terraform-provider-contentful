package provider_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

//nolint:paralleltest
func TestAccAppInstallationResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "contentful_app_installation" "test" {
					space_id = "0p38pssr0fi3"
					environment_id = "master"
					app_definition_id = "1WkQ2J9LERPtbMTdUfSHka"
				}
				`,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"contentful_app_installation.test",
						tfjsonpath.New("space_id"),
						knownvalue.StringExact("0p38pssr0fi3"),
					),
					statecheck.ExpectKnownValue(
						"contentful_app_installation.test",
						tfjsonpath.New("environment_id"),
						knownvalue.StringExact("master"),
					),
					statecheck.ExpectKnownValue(
						"contentful_app_installation.test",
						tfjsonpath.New("app_definition_id"),
						knownvalue.StringExact("1WkQ2J9LERPtbMTdUfSHka"),
					),
				},
			},
		},
	})
}

//nolint:paralleltest
func TestAccAppInstallationResourceImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "contentful_app_installation" "test" {
					space_id = "0p38pssr0fi3"
					environment_id = "master"
					app_definition_id = "1WkQ2J9LERPtbMTdUfSHka"
				}
				`,
			},
			{
				ResourceName:  "contentful_app_installation.test",
				ImportState:   true,
				ImportStateId: "a",
				ExpectError:   regexp.MustCompile(`Resource Import Passthrough Multipart ID Mismatch`),
			},
			{
				ResourceName:  "contentful_app_installation.test",
				ImportState:   true,
				ImportStateId: "a/b",
				ExpectError:   regexp.MustCompile(`Resource Import Passthrough Multipart ID Mismatch`),
			},
			{
				ResourceName:  "contentful_app_installation.test",
				ImportState:   true,
				ImportStateId: "a/b/c/d",
				ExpectError:   regexp.MustCompile(`Resource Import Passthrough Multipart ID Mismatch`),
			},
			{
				ResourceName:  "contentful_app_installation.test",
				ImportState:   true,
				ImportStateId: "0p38pssr0fi3/master/1WkQ2J9LERPtbMTdUfSHka",
				// ImportStateCheck: func(is []*terraform.InstanceState) error {
				// 	if len(is) != 1 {
				// 		return fmt.Errorf("expected 1 state, got %d", len(is))
				// 	}

				// 	if is[0].Attributes["space_id"] != "0p38pssr0fi3" {
				// 		return fmt.Errorf("expected space_id to be 0p38pssr0fi3, got %s", is[0].Attributes["space_id"])
				// 	}

				// 	if is[0].Attributes["app_definition_id"] != "1WkQ2J9LERPtbMTdUfSHka" {
				// 		return fmt.Errorf("expected app_definition_id to be 1WkQ2J9LERPtbMTdUfSHka, got %s", is[0].Attributes["app_definition_id"])
				// 	}

				// 	return nil
				// },
			},
		},
	})
}

//nolint:paralleltest
func TestAccAppInstallationResourceCreateNotFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "contentful_app_installation" "test" {
					space_id = "0p38pssr0fi3"
					environment_id = "master"
					app_definition_id = "12345"
				}
				`,
				ExpectError: regexp.MustCompile(`Failed to create app installation`),
			},
		},
	})
}

//nolint:paralleltest
func TestAccAppInstallationResourceUpdate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "contentful_app_installation" "test" {
					space_id = "0p38pssr0fi3"
					environment_id = "master"
					app_definition_id = "1WkQ2J9LERPtbMTdUfSHka"
				}
				`,
			},
			{
				Config: `
				resource "contentful_app_installation" "test" {
					space_id = "0p38pssr0fi3"
					environment_id = "master"
					app_definition_id = "1WkQ2J9LERPtbMTdUfSHka"
					parameters = jsonencode({ foo = "bar" })
				}
				`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_app_installation.test", plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: `
				resource "contentful_app_installation" "test" {
					space_id = "0p38pssr0fi3"
					environment_id = "test"
					app_definition_id = "1WkQ2J9LERPtbMTdUfSHka"
				}
				`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_app_installation.test", plancheck.ResourceActionReplace),
					},
				},
			},
		},
	})
}

//nolint:paralleltest
func TestAccAppInstallationResourceDeleted(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "contentful_app_installation" "test" {
					space_id = "0p38pssr0fi3"
					environment_id = "master"
					app_definition_id = "1WkQ2J9LERPtbMTdUfSHka"
				}
				`,
			},
			{
				Config: `
				resource "contentful_app_installation" "test" {
					space_id = "0p38pssr0fi3"
					environment_id = "master"
					app_definition_id = "1WkQ2J9LERPtbMTdUfSHka"
				}

				import {
					id = "0p38pssr0fi3/master/1WkQ2J9LERPtbMTdUfSHka"
					to = contentful_app_installation.test_dup
				}

				resource "contentful_app_installation" "test_dup" {
					space_id = "0p38pssr0fi3"
					environment_id = "master"
					app_definition_id = "1WkQ2J9LERPtbMTdUfSHka"
				}
				`,
			},
			{
				Config: `
				resource "contentful_app_installation" "test" {
					space_id = "0p38pssr0fi3"
					environment_id = "master"
					app_definition_id = "1WkQ2J9LERPtbMTdUfSHka"
				}
				`,
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_app_installation.test", plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}
