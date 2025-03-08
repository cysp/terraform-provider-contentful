package provider_test

import (
	"regexp"
	"testing"

	cmts "github.com/cysp/terraform-provider-contentful/internal/contentful-management-testserver"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

//nolint:paralleltest
func TestAccAppInstallationResource(t *testing.T) {
	testserver := cmts.NewContentfulManagementTestServer()
	defer testserver.Server().Close()

	configVariables := config.Variables{
		"space_id":          config.StringVariable("0p38pssr0fi3"),
		"environment_id":    config.StringVariable("master"),
		"app_definition_id": config.StringVariable("1WkQ2J9LERPtbMTdUfSHka"),
	}

	testserver.AddAppDefinitionID("1WkQ2J9LERPtbMTdUfSHka")

	ContentfulProviderMockableResourceTest(t, testserver.Server(), resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
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
	testserver := cmts.NewContentfulManagementTestServer()
	defer testserver.Server().Close()

	configVariables := config.Variables{
		"space_id":          config.StringVariable("0p38pssr0fi3"),
		"environment_id":    config.StringVariable("master"),
		"app_definition_id": config.StringVariable("1WkQ2J9LERPtbMTdUfSHka"),
	}

	ContentfulProviderMockableResourceTest(t, testserver.Server(), resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ResourceName:    "contentful_app_installation.test",
				ImportState:     true,
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ResourceName:    "contentful_app_installation.test",
				ImportState:     true,
				ImportStateId:   "a",
				ExpectError:     regexp.MustCompile(`Resource Import Passthrough Multipart ID Mismatch`),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ResourceName:    "contentful_app_installation.test",
				ImportState:     true,
				ImportStateId:   "a/b",
				ExpectError:     regexp.MustCompile(`Resource Import Passthrough Multipart ID Mismatch`),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ResourceName:    "contentful_app_installation.test",
				ImportState:     true,
				ImportStateId:   "a/b/c/d",
				ExpectError:     regexp.MustCompile(`Resource Import Passthrough Multipart ID Mismatch`),
			},
		},
	})
}

//nolint:paralleltest
func TestAccAppInstallationResourceImportNotFound(t *testing.T) {
	testserver := cmts.NewContentfulManagementTestServer()
	defer testserver.Server().Close()

	configVariables := config.Variables{
		"space_id":          config.StringVariable("0p38pssr0fi3"),
		"environment_id":    config.StringVariable("test"),
		"app_definition_id": config.StringVariable("nonexistent"),
	}

	ContentfulProviderMockableResourceTest(t, testserver.Server(), resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ResourceName:    "contentful_app_installation.test",
				ImportState:     true,
				ImportStateId:   "0p38pssr0fi3/test/nonexistent",
				ExpectError:     regexp.MustCompile(`Cannot import non-existent remote object`),
			},
		},
	})
}

//nolint:paralleltest
func TestAccAppInstallationResourceCreateNotFound(t *testing.T) {
	testserver := cmts.NewContentfulManagementTestServer()
	defer testserver.Server().Close()

	configVariables := config.Variables{
		"space_id":          config.StringVariable("0p38pssr0fi3"),
		"environment_id":    config.StringVariable("master"),
		"app_definition_id": config.StringVariable("nonexistent"),
	}

	ContentfulProviderMockableResourceTest(t, testserver.Server(), resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
				ExpectError:     regexp.MustCompile(`Failed to create app installation`),
			},
		},
	})
}

//nolint:paralleltest
func TestAccAppInstallationResourceUpdate(t *testing.T) {
	testserver := cmts.NewContentfulManagementTestServer()
	defer testserver.Server().Close()

	configVariables := config.Variables{
		"space_id":          config.StringVariable("0p38pssr0fi3"),
		"environment_id":    config.StringVariable("master"),
		"app_definition_id": config.StringVariable("1WkQ2J9LERPtbMTdUfSHka"),
	}

	ContentfulProviderMockableResourceTest(t, testserver.Server(), resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_app_installation.test", plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_app_installation.test", plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

//nolint:paralleltest
func TestAccAppInstallationResourceDeleted(t *testing.T) {
	testserver := cmts.NewContentfulManagementTestServer()
	defer testserver.Server().Close()

	configVariables := config.Variables{
		"space_id":          config.StringVariable("0p38pssr0fi3"),
		"environment_id":    config.StringVariable("master"),
		"app_definition_id": config.StringVariable("1WkQ2J9LERPtbMTdUfSHka"),
	}

	ContentfulProviderMockableResourceTest(t, testserver.Server(), resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
			},
			{
				ConfigDirectory:    config.TestStepDirectory(),
				ConfigVariables:    configVariables,
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
