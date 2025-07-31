package provider_test

import (
	"regexp"
	"testing"

	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPersonalAccessTokenResource(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	personalAccessTokenID := acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	configVariables := config.Variables{
		"personal_access_token_id": config.StringVariable(personalAccessTokenID),
	}

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("contentful_personal_access_token.test", "token"),
				),
			},
			{
				ConfigDirectory:         config.TestNameDirectory(),
				ConfigVariables:         configVariables,
				ResourceName:            "contentful_personal_access_token.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"expires_in", "token"},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("contentful_personal_access_token.test", "token"),
				),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("contentful_personal_access_token.test", "token"),
				),
			},
		},
	})
}

func TestAccPersonalAccessTokenResourceInvalidScopes(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	personalAccessTokenID := acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	configVariables := config.Variables{
		"personal_access_token_id": config.StringVariable(personalAccessTokenID),
	}

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ExpectError:     regexp.MustCompile(`Failed to create personal access token`),
			},
		},
	})
}

func TestAccPersonalAccessTokenResourceImportNotFound(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	personalAccessTokenID := acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	configVariables := config.Variables{
		"personal_access_token_id": config.StringVariable(personalAccessTokenID),
	}

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ExpectError:     regexp.MustCompile(`Cannot import non-existent remote object`),
			},
		},
	})
}
