package provider_test

import (
	"regexp"
	"testing"

	cmts "github.com/cysp/terraform-provider-contentful/internal/contentful-management-testserver"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDeliveryApiKeyResource(t *testing.T) {
	t.Parallel()

	testserver := cmts.NewContentfulManagementTestServer()
	defer testserver.Server().Close()

	apiKeyName := "acctest_" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	configVariables := config.Variables{
		"space_id":                   config.StringVariable("0p38pssr0fi3"),
		"environment_id":             config.StringVariable("test"),
		"test_delivery_api_key_name": config.StringVariable(apiKeyName),
	}

	ContentfulProviderMockableResourceTest(t, testserver.Server(), resource.TestCase{
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
				ConfigDirectory:   config.TestNameDirectory(),
				ConfigVariables:   configVariables,
				ResourceName:      "contentful_delivery_api_key.test",
				ImportState:       true,
				ImportStateVerify: true,
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

	testserver := cmts.NewContentfulManagementTestServer()
	defer testserver.Server().Close()

	apiKeyName := "acctest_" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	configVariables := config.Variables{
		"space_id":                   config.StringVariable("0p38pssr0fi3"),
		"environment_id":             config.StringVariable("test"),
		"test_delivery_api_key_name": config.StringVariable(apiKeyName),
	}

	ContentfulProviderMockableResourceTest(t, testserver.Server(), resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ExpectError:     regexp.MustCompile(`Cannot import non-existent remote object`),
			},
		},
	})
}
