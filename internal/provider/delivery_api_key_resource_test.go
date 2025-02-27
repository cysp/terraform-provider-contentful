package provider_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDeliveryApiKeyResource(t *testing.T) {
	t.Parallel()

	apiKeyName := "acctest_" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	configVariables := config.Variables{
		"space_id":                   config.StringVariable("0p38pssr0fi3"),
		"environment_id":             config.StringVariable("test"),
		"test_delivery_api_key_name": config.StringVariable(apiKeyName),
	}

	ContentfulProviderMockableResourceTest(t, nil, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("contentful_delivery_api_key.test", "access_token"),
					resource.TestCheckResourceAttrSet("data.contentful_preview_api_key.test", "access_token"),
				),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ResourceName:    "contentful_delivery_api_key.test",
				ImportState:     true,
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
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				PreConfig: func() {
					configVariables["test_delivery_api_key_name"] = config.StringVariable(apiKeyName + " updated")
				},
			},
		},
	})
}

func TestAccDeliveryApiKeyResourceImportNotFound(t *testing.T) {
	t.Parallel()

	apiKeyName := "acctest_" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	configVariables := config.Variables{
		"space_id":                   config.StringVariable("0p38pssr0fi3"),
		"environment_id":             config.StringVariable("test"),
		"test_delivery_api_key_name": config.StringVariable(apiKeyName),
	}

	ContentfulProviderMockableResourceTest(t, nil, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ExpectError:     regexp.MustCompile(`Cannot import non-existent remote object`),
			},
		},
	})
}
