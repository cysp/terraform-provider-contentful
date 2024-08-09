package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDeliveryApiKeyResource(t *testing.T) {
	t.Parallel()

	apiKeyID := "acctest_" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "contentful_delivery_api_key" "test" {
					space_id = "0p38pssr0fi3"

					name = %[1]q
					description = "key: %[1]s"

					environments = ["test"]
				}

				data "contentful_preview_api_key" "test"{
					space_id = "0p38pssr0fi3"

					preview_api_key_id = contentful_delivery_api_key.test.preview_api_key_id
				}
				`, apiKeyID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("contentful_delivery_api_key.test", "access_token"),
					resource.TestCheckResourceAttrSet("data.contentful_preview_api_key.test", "access_token"),
				),
			},
			{
				ResourceName: "contentful_delivery_api_key.test",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					apiKey := s.RootModule().Resources["contentful_delivery_api_key.test"]
					if apiKey == nil {
						//nolint:err113,perfsprint
						return "", fmt.Errorf("resource not found")
					}

					return apiKey.Primary.Attributes["space_id"] + "/" + apiKey.Primary.Attributes["api_key_id"], nil
				},
				ImportStateVerifyIdentifierAttribute: "api_key_id",
				ImportStateVerify:                    true,
			},
			{
				Config: fmt.Sprintf(`
				resource "contentful_delivery_api_key" "test" {
					space_id = "0p38pssr0fi3"

					name = "%[1]s updated"
					description = "key: %[1]s updated"

					environments = ["test"]
				}

				data "contentful_preview_api_key" "test"{
					space_id = "0p38pssr0fi3"

					preview_api_key_id = contentful_delivery_api_key.test.preview_api_key_id
				}
				`, apiKeyID),
			},
		},
	})
}
