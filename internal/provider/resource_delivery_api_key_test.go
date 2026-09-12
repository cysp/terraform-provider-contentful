package provider_test

import (
	"maps"
	"regexp"
	"testing"

	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

func TestAccDeliveryAPIKeyResourceLifecycle(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)

	server.RegisterSpaceEnvironment("0p38pssr0fi3", "master")

	apiKeyName := "acctest_" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	configVariables := config.Variables{
		"space_id":                   config.StringVariable("0p38pssr0fi3"),
		"environment_id":             config.StringVariable("test"),
		"test_delivery_api_key_name": config.StringVariable(apiKeyName),
	}

	updatedVariables := maps.Clone(configVariables)
	updatedVariables["test_delivery_api_key_name"] = config.StringVariable(apiKeyName + " updated")

	identity := statecheck.CompareValue(compare.ValuesSame())
	deliveryToken := statecheck.CompareValue(compare.ValuesSame())
	previewToken := statecheck.CompareValue(compare.ValuesSame())

	testAccMockableResource(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_delivery_api_key.test", plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity.AddStateValue("contentful_delivery_api_key.test", tfjsonpath.New("id")),
					statecheck.ExpectSensitiveValue("contentful_delivery_api_key.test", tfjsonpath.New("access_token")),
					statecheck.ExpectSensitiveValue("data.contentful_preview_api_key.test", tfjsonpath.New("access_token")),
					statecheck.CompareValuePairs("contentful_delivery_api_key.test", tfjsonpath.New("preview_api_key_id"), "data.contentful_preview_api_key.test", tfjsonpath.New("preview_api_key_id"), compare.ValuesSame()),
					statecheck.ExpectKnownValue("contentful_delivery_api_key.test", tfjsonpath.New("access_token"), knownvalue.StringRegexp(regexp.MustCompile(`.+`))),
					statecheck.ExpectKnownValue("data.contentful_preview_api_key.test", tfjsonpath.New("access_token"), knownvalue.StringRegexp(regexp.MustCompile(`.+`))),
					deliveryToken.AddStateValue("contentful_delivery_api_key.test", tfjsonpath.New("access_token")),
					previewToken.AddStateValue("data.contentful_preview_api_key.test", tfjsonpath.New("access_token")),
				},
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
				ConfigVariables: updatedVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_delivery_api_key.test", plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity.AddStateValue("contentful_delivery_api_key.test", tfjsonpath.New("id")),
					statecheck.ExpectSensitiveValue("contentful_delivery_api_key.test", tfjsonpath.New("access_token")),
					statecheck.ExpectSensitiveValue("data.contentful_preview_api_key.test", tfjsonpath.New("access_token")),
					statecheck.CompareValuePairs("contentful_delivery_api_key.test", tfjsonpath.New("preview_api_key_id"), "data.contentful_preview_api_key.test", tfjsonpath.New("preview_api_key_id"), compare.ValuesSame()),
					statecheck.ExpectKnownValue("contentful_delivery_api_key.test", tfjsonpath.New("name"), knownvalue.StringExact(apiKeyName+" updated")),
					deliveryToken.AddStateValue("contentful_delivery_api_key.test", tfjsonpath.New("access_token")),
					previewToken.AddStateValue("data.contentful_preview_api_key.test", tfjsonpath.New("access_token")),
				},
			},
		},
	})
}

func TestAccDeliveryAPIKeyResourceImportNotFound(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)

	apiKeyName := "acctest_" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	configVariables := config.Variables{
		"space_id":                   config.StringVariable("0p38pssr0fi3"),
		"environment_id":             config.StringVariable("test"),
		"test_delivery_api_key_name": config.StringVariable(apiKeyName),
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
