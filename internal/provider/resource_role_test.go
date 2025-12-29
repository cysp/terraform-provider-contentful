package provider_test

import (
	"regexp"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

//nolint:paralleltest
func TestAccRoleResourceImport(t *testing.T) {
	server, _ := cmt.NewContentfulManagementServer()

	configVariables := config.Variables{
		"space_id": config.StringVariable("0p38pssr0fi3"),
	}

	server.SetRole("0p38pssr0fi3", "2EZrF9oDqi4AnsLNy21n6z", cm.RoleData{
		Name: "author",
	})

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory:    config.TestNameDirectory(),
				ConfigVariables:    configVariables,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				ConfigDirectory:    config.TestNameDirectory(),
				ConfigVariables:    configVariables,
				ResourceName:       "contentful_role.author",
				ImportState:        true,
				ImportStateId:      "0p38pssr0fi3/2EZrF9oDqi4AnsLNy21n6z",
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

//nolint:paralleltest
func TestAccRoleResourceImportNotFound(t *testing.T) {
	server, _ := cmt.NewContentfulManagementServer()

	configVariables := config.Variables{
		"space_id": config.StringVariable("0p38pssr0fi3"),
	}

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory:    config.TestNameDirectory(),
				ConfigVariables:    configVariables,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ResourceName:    "contentful_role.test",
				ImportState:     true,
				ImportStateId:   "0p38pssr0fi3/nonexistent",
				ExpectError:     regexp.MustCompile(`Cannot import non-existent remote object`),
			},
		},
	})
}

//nolint:paralleltest
func TestAccRoleResourceCreateUpdateDelete(t *testing.T) {
	server, _ := cmt.NewContentfulManagementServer()

	server.RegisterSpaceEnvironment("0p38pssr0fi3", "master")

	configVariables := config.Variables{
		"space_id": config.StringVariable("0p38pssr0fi3"),
	}

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
			},
		},
	})
}

//nolint:paralleltest
func TestAccRoleResourceDeleted(t *testing.T) {
	server, _ := cmt.NewContentfulManagementServer()

	server.RegisterSpaceEnvironment("0p38pssr0fi3", "master")

	server.SetRole("0p38pssr0fi3", "test", cm.RoleData{
		Name: "Test",
	})

	configVariables := config.Variables{
		"space_id": config.StringVariable("0p38pssr0fi3"),
	}

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
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
						plancheck.ExpectResourceAction("contentful_role.test", plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}
