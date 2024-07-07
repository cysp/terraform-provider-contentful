package provider_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

//nolint:paralleltest
func TestAccWebhookResourceImport(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "contentful_webhook" "test" {
					space_id = "0p38pssr0fi3"
					name = "test"
					url = "https://example.org/webhook"
				}
				`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				ResourceName:       "contentful_webhook.test",
				ImportState:        true,
				ImportStateId:      "0p38pssr0fi3/4HJlhYqVjoWwxFmvOj5r1Q",
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

//nolint:paralleltest
func TestAccWebhookResourceImportNotFound(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				resource "contentful_webhook" "test" {
					space_id = "0p38pssr0fi3"
					name = "test"
					url = "https://example.org/webhook"
				}
				`,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				ResourceName:  "contentful_webhook.test",
				ImportState:   true,
				ImportStateId: "0p38pssr0fi3/nonexistent",
				ExpectError:   regexp.MustCompile(`Cannot import non-existent remote object`),
			},
		},
	})
}

func TestAccWebhookResourceCreate(t *testing.T) {
	t.Parallel()

	contentTypeID := "acctest_" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "contentful_webhook" %[1]q {
					space_id = "0p38pssr0fi3"
					name = "%[1]s"
					url = "https://example.org/webhook"
					topics = ["Entry.publish", "Entry.unpublish"]
				}
				`, contentTypeID),
			},
		},
	})
}

func TestAccWebhookResourceUpdate(t *testing.T) {
	t.Parallel()

	contentTypeID := "acctest_" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "contentful_webhook" %[1]q {
					space_id = "0p38pssr0fi3"
					name = "%[1]s"
					url = "https://example.org/webhook"
					topics = ["Entry.publish", "Entry.unpublish"]
				}
				`, contentTypeID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_webhook."+contentTypeID, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: fmt.Sprintf(`
				resource "contentful_webhook" %[1]q {
					space_id = "0p38pssr0fi3"
					name = "%[1]s"
					url = "https://example.org/webhook"
					topics = ["Entry.save", "Entry.publish", "Entry.unpublish"]
				}
				`, contentTypeID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_webhook."+contentTypeID, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: fmt.Sprintf(`
				resource "contentful_webhook" %[1]q {
					space_id = "0p38pssr0fi3"
					name = "%[1]s"
					url = "https://example.org/webhook"
					topics = ["Entry.save", "Entry.publish", "Entry.unpublish"]
				}
				`, contentTypeID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_webhook."+contentTypeID, plancheck.ResourceActionNoop),
					},
				},
			},
			{
				Config: fmt.Sprintf(`
				resource "contentful_webhook" %[1]q {
					space_id = "0p38pssr0fi3"
					name = "%[1]s"
					url = "https://example.org/webhook"
					topics = ["Entry.save", "Entry.publish", "Entry.unpublish"]
					filters = [
						{
							equals = {
								doc   = "sys.environment.sys.id"
								value = "master"
							}
						},
					]
					transformation = {
						method                 = "POST"
						content_type           = "application/vnd.contentful.management.v1+json"
						include_content_length = true
					}
				}
				`, contentTypeID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_webhook."+contentTypeID, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}
