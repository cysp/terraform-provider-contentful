package provider_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPreviewApiKeyDataSourceNotFound(t *testing.T) {
	t.Parallel()

	configVariables := config.Variables{
		"space_id":           config.StringVariable("0p38pssr0fi3"),
		"preview_api_key_id": config.StringVariable("nonexistent"),
	}

	ContentfulProviderMockableResourceTest(t, nil, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ExpectError:     regexp.MustCompile(`Provider produced null object`),
			},
		},
	})
}
