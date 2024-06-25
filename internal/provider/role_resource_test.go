package provider_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

//nolint:paralleltest
func TestAccRoleResourceImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
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
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
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
