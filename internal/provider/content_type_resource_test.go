package provider_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestAccContentTypeResourceImport(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "contentful_content_type" "test" {
					space_id = "0p38pssr0fi3"
					environment_id = "test"
					content_type_id = "author"

					name = "Author"
					description = "An author"

					display_field = "name"

					fields = []
				}
				`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				ResourceName:  "contentful_content_type.test",
				ImportState:   true,
				ImportStateId: "a",
				ExpectError:   regexp.MustCompile(`Resource Import Passthrough Multipart ID Mismatch`),
			},
			{
				ResourceName:  "contentful_content_type.test",
				ImportState:   true,
				ImportStateId: "a/b",
				ExpectError:   regexp.MustCompile(`Resource Import Passthrough Multipart ID Mismatch`),
			},
			{
				ResourceName:  "contentful_content_type.test",
				ImportState:   true,
				ImportStateId: "a/b/c/d",
				ExpectError:   regexp.MustCompile(`Resource Import Passthrough Multipart ID Mismatch`),
			},
			{
				ResourceName:  "contentful_content_type.test",
				ImportState:   true,
				ImportStateId: "0p38pssr0fi3/test/author",
				// ImportStateVerify: true,
			},
		},
	})
}

func TestAccContentTypeResourceImportNotFound(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "contentful_content_type" "test" {
					space_id = "0p38pssr0fi3"
					environment_id = "test"
					content_type_id = "nonexistent"

					name = ""
					description = ""

					display_field = ""

					fields = []
				}
				`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				ResourceName:  "contentful_content_type.test",
				ImportState:   true,
				ImportStateId: "0p38pssr0fi3/test/nonexistent",
				ExpectError:   regexp.MustCompile(`Cannot import non-existent remote object`),
			},
		},
	})
}

func TestAccContentTypeResourceCreateNotFoundEnvironment(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "contentful_content_type" "test" {
					space_id = "0p38pssr0fi3"
					environment_id = "nonexistent"
					content_type_id = "nonexistent"

					name = ""
					description = ""

					display_field = ""

					fields = []
				}
				`,
				ExpectError: regexp.MustCompile(`Failed to create content type`),
			},
		},
	})
}

func TestAccContentTypeResourceCreate(t *testing.T) {
	t.Parallel()

	contentTypeID := "acctest_" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "contentful_content_type" %[1]q {
					space_id = "0p38pssr0fi3"
					environment_id = "test"
					content_type_id = %[1]q

					name = "Test"
					description = "Test content type (%[1]s)"

					display_field = "name"

					fields = [
						{
							id  	  = "name"
							name      = "Name"
							type      = "Symbol"
							required  = true
							localized = false
						}
					]
				}
				`, contentTypeID),
			},
		},
	})
}

func TestAccContentTypeResourceUpdate(t *testing.T) {
	t.Parallel()

	contentTypeID := "acctest_" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "contentful_content_type" %[1]q {
					space_id = "0p38pssr0fi3"
					environment_id = "test"
					content_type_id = %[1]q

					name = "Test"
					description = "Test content type (%[1]s)"

					display_field = "name"

					fields = [
						{
							id  	  = "name"
							name      = "Name"
							type      = "Symbol"
							required  = true
							localized = false
						}
					]
				}

				resource "contentful_editor_interface" %[1]q {
					space_id = "0p38pssr0fi3"
					environment_id = "test"
					content_type_id = %[1]q

					depends_on = [contentful_content_type.%[1]s]
				}
				`, contentTypeID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type."+contentTypeID, plancheck.ResourceActionCreate),
						plancheck.ExpectResourceAction("contentful_editor_interface."+contentTypeID, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: fmt.Sprintf(`
				resource "contentful_content_type" %[1]q {
					space_id = "0p38pssr0fi3"
					environment_id = "test"
					content_type_id = %[1]q

					name = "Test"
					description = "Test content type (%[1]s)"

					display_field = "name"

					fields = [
						{
							id  	  = "name"
							name      = "Name"
							type      = "Symbol"
							required  = true
							localized = false
						},
						{
							id  	  = "slug"
							name      = "Slug"
							type      = "Symbol"
							required  = true
							localized = false
						}
					]
				}

				resource "contentful_editor_interface" %[1]q {
					space_id = "0p38pssr0fi3"
					environment_id = "test"
					content_type_id = %[1]q
				}
				`, contentTypeID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type."+contentTypeID, plancheck.ResourceActionUpdate),
						plancheck.ExpectResourceAction("contentful_editor_interface."+contentTypeID, plancheck.ResourceActionNoop),
					},
				},
			},
			{
				Config: fmt.Sprintf(`
				resource "contentful_content_type" %[1]q {
					space_id = "0p38pssr0fi3"
					environment_id = "test"
					content_type_id = %[1]q

					name = "Test"
					description = "Test content type (%[1]s)"

					display_field = "name"

					fields = [
						{
							id  	  = "name"
							name      = "Name"
							type      = "Symbol"
							required  = true
							localized = false
						},
						{
							id  	  = "slug"
							name      = "Slug"
							type      = "Symbol"
							required  = true
							localized = false
						}
					]
				}

				resource "contentful_editor_interface" %[1]q {
					space_id = "0p38pssr0fi3"
					environment_id = "test"
					content_type_id = %[1]q
				}
				`, contentTypeID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type."+contentTypeID, plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction("contentful_editor_interface."+contentTypeID, plancheck.ResourceActionNoop),
					},
				},
			},
			{
				Config: fmt.Sprintf(`
				resource "contentful_content_type" %[1]q {
					space_id = "0p38pssr0fi3"
					environment_id = "test"
					content_type_id = %[1]q

					name = "Test"
					description = "Test content type (%[1]s)"

					display_field = "name"

					fields = [
						{
							id  	  = "name"
							name      = "Name"
							type      = "Symbol"
							required  = true
							localized = false
						},
						{
							id  	  = "slug"
							name      = "Slug"
							type      = "Symbol"
							required  = true
							localized = false
							validations = [
								jsonencode({
									regexp = {
										pattern = "^[a-z0-9]+(?:-[a-z0-9]+)*$"
										flags = null
									}
								}),
							]
						},
						{
							id  	  = "flags"
							name      = "Flags"
							type      = "Array"
							items = {
								type = "Symbol"
								validations = [
									jsonencode({
										in = ["abc", "def", "ghi"]
									}),
								]
							}
							default_value = jsonencode({
								"en-AU": ["def"],
							})
							required  = true
							localized = false
						}
					]
				}

				resource "contentful_editor_interface" %[1]q {
					space_id = "0p38pssr0fi3"
					environment_id = "test"
					content_type_id = %[1]q
				}
				`, contentTypeID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type."+contentTypeID, plancheck.ResourceActionUpdate),
						plancheck.ExpectResourceAction("contentful_editor_interface."+contentTypeID, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccContentTypeResourceDeleted(t *testing.T) {
	t.Parallel()

	contentTypeID := "acctest_" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "contentful_content_type" "test" {
					space_id = "0p38pssr0fi3"
					environment_id = "test"
					content_type_id = %[1]q

					name = "Test"
					description = "Test content type (%[1]s)"

					display_field = "name"

					fields = [
						{
							id  	  = "name"
							name      = "Name"
							type      = "Symbol"
							required  = true
							localized = false
						}
					]
				}
				`, contentTypeID),
			},
			{
				Config: fmt.Sprintf(`
				resource "contentful_content_type" "test" {
					space_id = "0p38pssr0fi3"
					environment_id = "test"
					content_type_id = %[1]q

					name = "Test"
					description = "Test content type (%[1]s)"

					display_field = "name"

					fields = [
						{
							id  	  = "name"
							name      = "Name"
							type      = "Symbol"
							required  = true
							localized = false
						}
					]
				}

				import {
					id = "0p38pssr0fi3/test/%[1]s"
					to = contentful_content_type.test_dup
				}

				resource "contentful_content_type" "test_dup" {
					space_id = "0p38pssr0fi3"
					environment_id = "test"
					content_type_id = %[1]q

					name = "Test"
					description = "Test content type (%[1]s)"

					display_field = "name"

					fields = [
						{
							id  	  = "name"
							name      = "Name"
							type      = "Symbol"
							required  = true
							localized = false
						}
					]
				}
				`, contentTypeID),
			},
			{
				Config: fmt.Sprintf(`
				resource "contentful_content_type" "test" {
					space_id = "0p38pssr0fi3"
					environment_id = "test"
					content_type_id = %[1]q

					name = "Test"
					description = "Test content type (%[1]s)"

					display_field = "name"

					fields = [
						{
							id  	  = "name"
							name      = "Name"
							type      = "Symbol"
							required  = true
							localized = false
						}
					]
				}
				`, contentTypeID),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}
