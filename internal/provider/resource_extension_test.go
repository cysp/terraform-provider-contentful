package provider_test

import (
	"testing"

	cmts "github.com/cysp/terraform-provider-contentful/internal/contentful-management-testserver"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccExtensionResource(t *testing.T) {
	t.Parallel()

	testserver := cmts.NewContentfulManagementTestServer()
	defer testserver.Server().Close()

	extensionID := "acctest_" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	configVariables := config.Variables{
		"space_id":          config.StringVariable("0p38pssr0fi3"),
		"environment_id":    config.StringVariable("test"),
		"test_extension_id": config.StringVariable(extensionID),
	}

	ContentfulProviderMockableResourceTest(t, testserver.Server(), resource.TestCase{
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
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
				ImportState:     true,
				ResourceName:    "contentful_extension.test",
			},
		},
	})
}
