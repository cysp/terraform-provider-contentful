package provider_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPersonalAccessTokenResource(t *testing.T) {
	t.Parallel()

	personalAccessTokenID := acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "contentful_personal_access_token" "test" {
					name = "terraform-provider-contentful-acctest-%[1]s"
					scopes = ["content_management_read"]
					expires_in = 5 * 60
				}
				`, personalAccessTokenID),
			},
			{
				ResourceName:            "contentful_personal_access_token.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"expires_in", "token"},
			},
		},
	})
}

func TestAccPersonalAccessTokenResourceInvalidScopes(t *testing.T) {
	t.Parallel()

	personalAccessTokenID := acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				resource "contentful_personal_access_token" "test" {
					name = "terraform-provider-contentful-acctest-%[1]s"
					scopes = ["content_management_invalid"]
					expires_in = 5 * 60
				}
				`, personalAccessTokenID),
				ExpectError: regexp.MustCompile(`Failed to create personal access token`),
			},
		},
	})
}

func TestAccPersonalAccessTokenResourceImportNotFound(t *testing.T) {
	t.Parallel()

	personalAccessTokenID := acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
				import {
					id = "%[1]s"
					to = contentful_personal_access_token.test
				}

				resource "contentful_personal_access_token" "test" {
					name = "terraform-provider-contentful-acctest-%[1]s"
					scopes = ["content_management_read"]
				}
				`, personalAccessTokenID),
				ExpectError: regexp.MustCompile(`Cannot import non-existent remote object`),
			},
		},
	})
}
