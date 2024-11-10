package provider_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPreviewApiKeyResourceImportNotFound(t *testing.T) {
	t.Parallel()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				data "contentful_preview_api_key" "test" {
					space_id = "0p38pssr0fi3"
  					preview_api_key_id = "unknown"
				}
				`,
				ExpectError: regexp.MustCompile(`Provider produced null object`),
			},
		},
	})
}
