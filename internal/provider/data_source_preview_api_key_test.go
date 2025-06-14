package provider_test

import (
	"regexp"
	"testing"

	cmts "github.com/cysp/terraform-provider-contentful/internal/contentful-management-testserver"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPreviewApiKeyDataSourceNotFound(t *testing.T) {
	t.Parallel()

	testserver := cmts.NewContentfulManagementTestServer()
	defer testserver.Server().Close()

	configVariables := config.Variables{
		"space_id":           config.StringVariable("0p38pssr0fi3"),
		"preview_api_key_id": config.StringVariable("nonexistent"),
	}

	ContentfulProviderMockableResourceTest(t, testserver.Server(), resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ExpectError:     regexp.MustCompile(`Provider produced null object`),
			},
		},
	})
}
