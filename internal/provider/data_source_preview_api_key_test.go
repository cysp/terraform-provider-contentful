package provider_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

//nolint:paralleltest
func TestAccPreviewAPIKeyDataSourceNotFound(t *testing.T) {
	parallelWhenMocked(t)

	server := newTestCMAServer(t)

	configVariables := config.Variables{
		"space_id":           config.StringVariable("0p38pssr0fi3"),
		"preview_api_key_id": config.StringVariable("nonexistent"),
	}

	testAccMockableResource(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ExpectError:     regexp.MustCompile(`Failed to read preview API key`),
			},
		},
	})
}
