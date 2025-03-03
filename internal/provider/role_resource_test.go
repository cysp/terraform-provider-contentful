package provider_test

import (
	"regexp"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmts "github.com/cysp/terraform-provider-contentful/internal/contentful-management-testserver"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

//nolint:paralleltest
func TestAccRoleResourceImport(t *testing.T) {
	testserver := cmts.NewContentfulManagementTestServer()
	defer testserver.Server().Close()

	configVariables := config.Variables{
		"space_id": config.StringVariable("0p38pssr0fi3"),
	}

	testserver.SetRole("0p38pssr0fi3", &cm.Role{
		Sys: cm.RoleSys{
			Type: cm.RoleSysTypeRole,
			ID:   "2EZrF9oDqi4AnsLNy21n6z",
		},
		Name: "author",
	})

	ContentfulProviderMockableResourceTest(t, testserver.Server(), resource.TestCase{
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
	testserver := cmts.NewContentfulManagementTestServer()
	defer testserver.Server().Close()

	configVariables := config.Variables{
		"space_id": config.StringVariable("0p38pssr0fi3"),
	}

	ContentfulProviderMockableResourceTest(t, testserver.Server(), resource.TestCase{
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
				ResourceName:    "contentful_role.admin",
				ImportState:     true,
				ImportStateId:   "0p38pssr0fi3/admin",
				ExpectError:     regexp.MustCompile(`Cannot import non-existent remote object`),
			},
		},
	})
}

//nolint:paralleltest
func TestAccRoleResourceCreateUpdateDelete(t *testing.T) {
	testserver := cmts.NewContentfulManagementTestServer()
	defer testserver.Server().Close()

	ContentfulProviderMockedResourceTest(t, testserver.Server(), resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: `
					resource "contentful_role" "test" {
						space_id = "0p38pssr0fi3"
						name = "Test"
						permissions = {}
						policies = []
					}
					`,
			},
			{
				Config: `
					resource "contentful_role" "test" {
						space_id = "0p38pssr0fi3"
						name = "Test"
						permissions = { foo = ["bar"] }
						policies = []
					}
					`,
			},
		},
	})
}
