package provider_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

//nolint:paralleltest
func TestAccPersonalAccessTokenResourceLifecycle(t *testing.T) {
	parallelWhenMocked(t)

	server := newTestCMAServer(t)

	personalAccessTokenID := acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	configVariables := config.Variables{
		"personal_access_token_id": config.StringVariable(personalAccessTokenID),
	}

	var token string

	testAccMockableResource(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory:   config.TestNameDirectory(),
				ConfigVariables:   configVariables,
				ConfigStateChecks: []statecheck.StateCheck{statecheck.ExpectSensitiveValue("contentful_personal_access_token.test", tfjsonpath.New("token"))},
				Check: func(state *terraform.State) error {
					check := resource.TestCheckResourceAttrSet("contentful_personal_access_token.test", "token")

					err := check(state)
					if err != nil {
						return err
					}

					token = state.RootModule().Resources["contentful_personal_access_token.test"].Primary.Attributes["token"]

					return nil
				},
			},
			{
				ConfigDirectory:   config.TestNameDirectory(),
				ConfigVariables:   configVariables,
				ResourceName:      "contentful_personal_access_token.test",
				ImportState:       true,
				ImportStateVerify: true,
				// expires_in is a create-time duration; token is returned only on create.
				// Typed imported nulls are covered by MockImportedTimeoutUpdate.
				ImportStateVerifyIgnore: []string{"expires_in", "token"},
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					require.Len(t, states, 1)
					require.NotContains(t, states[0].Attributes, "token", "import must not manufacture a token")
					require.NotContains(t, states[0].Attributes, "expires_in", "import cannot recover the create-time duration")

					return nil
				},
			},
			{
				RefreshState: true,
				Check: func(state *terraform.State) error {
					return resource.TestCheckResourceAttr("contentful_personal_access_token.test", "token", token)(state)
				},
			},
		},
	})
}

//nolint:paralleltest
func TestAccPersonalAccessTokenResourceInvalidScopes(t *testing.T) {
	parallelWhenMocked(t)

	server := newTestCMAServer(t)

	personalAccessTokenID := acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	configVariables := config.Variables{
		"personal_access_token_id": config.StringVariable(personalAccessTokenID),
	}

	testAccMockableResource(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ExpectError:     regexp.MustCompile(`Failed to create personal access token`),
			},
		},
	})
}

//nolint:paralleltest
func TestAccPersonalAccessTokenResourceImportNotFound(t *testing.T) {
	parallelWhenMocked(t)

	server := newTestCMAServer(t)

	personalAccessTokenID := acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	configVariables := config.Variables{
		"personal_access_token_id": config.StringVariable(personalAccessTokenID),
	}

	testAccMockableResource(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ExpectError:     regexp.MustCompile(`Cannot import non-existent remote object`),
			},
		},
	})
}
