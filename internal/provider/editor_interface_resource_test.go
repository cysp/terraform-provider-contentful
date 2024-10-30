package provider_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

//nolint:paralleltest
func TestAccEditorInterfaceResourceImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "contentful_editor_interface" "test" {
					space_id = "0p38pssr0fi3"
					environment_id = "test"
					content_type_id = "author"
				}
				`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				ResourceName:  "contentful_editor_interface.test",
				ImportState:   true,
				ImportStateId: "a",
				ExpectError:   regexp.MustCompile(`Resource Import Passthrough Multipart ID Mismatch`),
			},
			{
				ResourceName:  "contentful_editor_interface.test",
				ImportState:   true,
				ImportStateId: "a/b",
				ExpectError:   regexp.MustCompile(`Resource Import Passthrough Multipart ID Mismatch`),
			},
			{
				ResourceName:  "contentful_editor_interface.test",
				ImportState:   true,
				ImportStateId: "a/b/c/d",
				ExpectError:   regexp.MustCompile(`Resource Import Passthrough Multipart ID Mismatch`),
			},
			{
				ResourceName:  "contentful_editor_interface.test",
				ImportState:   true,
				ImportStateId: "0p38pssr0fi3/test/author",
				// ImportStateVerify: true,
			},
		},
	})
}

//nolint:paralleltest
func TestAccEditorInterfaceResourceImportNotFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "contentful_editor_interface" "test" {
					space_id = "0p38pssr0fi3"
					environment_id = "test"
					content_type_id = "nonexistent"
				}
				`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				ResourceName:  "contentful_editor_interface.test",
				ImportState:   true,
				ImportStateId: "0p38pssr0fi3/test/nonexistent",
				ExpectError:   regexp.MustCompile(`Cannot import non-existent remote object`),
			},
		},
	})
}

//nolint:paralleltest
func TestAccEditorInterfaceResourceCreateNotFoundEnvironment(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "contentful_editor_interface" "test" {
					space_id = "0p38pssr0fi3"
					environment_id = "nonexistent"
					content_type_id = "nonexistent"
				}
				`,
				ExpectError: regexp.MustCompile(`Failed to create editor interface`),
			},
		},
	})
}

//nolint:paralleltest
func TestAccEditorInterfaceResourceCreateNotFoundContentType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "contentful_editor_interface" "test" {
					space_id = "0p38pssr0fi3"
					environment_id = "test"
					content_type_id = "nonexistent"
				}
				`,
				ExpectError: regexp.MustCompile(`Failed to create editor interface`),
			},
		},
	})
}

//nolint:paralleltest
func TestAccEditorInterfaceResourceUpdate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "contentful_editor_interface" "test" {
					space_id = "0p38pssr0fi3"
					environment_id = "test"
					content_type_id = "author"
				}
				`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				ResourceName:       "contentful_editor_interface.test",
				ImportState:        true,
				ImportStateId:      "0p38pssr0fi3/test/author",
				ImportStatePersist: true,
			},
			{
				Config: `
				resource "contentful_editor_interface" "test" {
					space_id = "0p38pssr0fi3"
					environment_id = "test"
					content_type_id = "author"

					controls = [
						{
						field_id         = "name",
						widget_namespace = "builtin",
						widget_id        = "singleLine",
						},
						{
						field_id         = "avatar",
						widget_namespace = "builtin",
						widget_id        = "assetLinkEditor",
						},
						{
						field_id         = "blurb",
						widget_namespace = "builtin",
						widget_id        = "richTextEditor",
						}
					]

					sidebar = [{
						widget_namespace = "app"
						widget_id        = "1WkQ2J9LERPtbMTdUfSHka"
						settings = jsonencode({
							foo = "bar"
						})
					}]
				}
				`,
			},
			{
				Config: `
				resource "contentful_editor_interface" "test" {
					space_id = "0p38pssr0fi3"
					environment_id = "test"
					content_type_id = "author"

					controls = [
						{
						field_id         = "name",
						widget_namespace = "builtin",
						widget_id        = "singleLine",
						},
						{
						field_id         = "avatar",
						widget_namespace = "builtin",
						widget_id        = "assetLinkEditor",
						},
						{
						field_id         = "blurb",
						widget_namespace = "builtin",
						widget_id        = "richTextEditor",
						}
					]

					sidebar = [{
						widget_namespace = "app"
						widget_id        = "1WkQ2J9LERPtbMTdUfSHka"
						settings = jsonencode({
							bar = "baz"
						})
					}]
				}
				`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_editor_interface.test", plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: `
				resource "contentful_editor_interface" "test" {
					space_id = "0p38pssr0fi3"
					environment_id = "test"
					content_type_id = "author"

					controls = [
						{
						field_id         = "name",
						widget_namespace = "builtin",
						widget_id        = "singleLine",
						},
						{
						field_id         = "avatar",
						widget_namespace = "builtin",
						widget_id        = "assetLinkEditor",
						},
						{
						field_id         = "blurb",
						widget_namespace = "builtin",
						widget_id        = "richTextEditor",
						}
					]

					sidebar = [{
						widget_namespace = "app"
						widget_id        = "1WkQ2J9LERPtbMTdUfSHka"
					}]
				}
				`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_editor_interface.test", plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

// New integration test to increase test coverage
func TestAccEditorInterfaceResourceWithControlsAndSidebar(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "contentful_editor_interface" "test" {
					space_id = "0p38pssr0fi3"
					environment_id = "test"
					content_type_id = "author"

					controls = [
						{
						field_id         = "name",
						widget_namespace = "builtin",
						widget_id        = "singleLine",
						},
						{
						field_id         = "avatar",
						widget_namespace = "builtin",
						widget_id        = "assetLinkEditor",
						},
						{
						field_id         = "blurb",
						widget_namespace = "builtin",
						widget_id        = "richTextEditor",
						}
					]

					sidebar = [{
						widget_namespace = "app"
						widget_id        = "1WkQ2J9LERPtbMTdUfSHka"
						settings = jsonencode({
							foo = "bar"
						})
					}]
				}
				`,
			},
		},
	})
}
