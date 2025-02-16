package provider_test

import (
	"regexp"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmts "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testserver"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

//nolint:paralleltest
func TestAccRoleResourceImport(t *testing.T) {
	ContentfulProviderMockableResourceTest(t, nil, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: `
				resource "contentful_role" "author" {
					space_id = "0p38pssr0fi3"
					name = "Author"
					permissions = {}
					policies = []
				}
				`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
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
	ContentfulProviderMockableResourceTest(t, nil, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: `
				resource "contentful_role" "admin" {
					space_id = "0p38pssr0fi3"
					name = "Admin"
					permissions = {}
					policies = []
				}
				`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				ResourceName:  "contentful_role.admin",
				ImportState:   true,
				ImportStateId: "0p38pssr0fi3/admin",
				ExpectError:   regexp.MustCompile(`Cannot import non-existent remote object`),
			},
		},
	})
}

//nolint:paralleltest
func TestAccRoleResourceCreateUpdateDelete(t *testing.T) {
	testserver := cmts.NewContentfulManagementTestServer()
	defer testserver.Server.Close()

	role := cm.Role{
		Sys: cm.RoleSys{
			Type: cm.RoleSysTypeRole,
			ID:   "abcdef",
		},
		Name:        "Test",
		Permissions: cm.RolePermissions{},
		Policies:    []cm.RolePoliciesItem{},
	}

	testserver.HandleRoleCreation("0p38pssr0fi3", &role)
	testserver.HandleRole("0p38pssr0fi3", &role)

	ContentfulProviderMockedResourceTest(t, testserver.Server, resource.TestCase{
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
				PreConfig: func() {
					role.Permissions["foo"] = cm.RolePermissionsItem{Type: cm.StringArrayRolePermissionsItem, StringArray: []string{"bar"}}
				},
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
