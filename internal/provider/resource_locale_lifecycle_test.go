package provider_test

import (
	"fmt"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

func TestAccLocaleResourceLifecycle(t *testing.T) {
	t.Parallel()

	const resourceAddress = "contentful_locale.test"

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")
	server.SetLocale(cmt.NewLocaleFromData("space", "environment", "default-locale", cm.LocaleData{
		Name: "English (United States)", Code: "en-US", FallbackCode: cm.NewNilStringNull(),
		ContentDeliveryApi: true, ContentManagementApi: true,
	}, true))

	configuration := func(code, attributes string) string {
		return fmt.Sprintf(`resource "contentful_locale" "test" {
  space_id = "space"
  environment_id = "environment"
  name = "German"
  code = %[1]q
  %[2]s
}`, code, attributes)
	}
	initial := configuration("de-DE", `fallback_code = "en-US"`)
	updated := configuration("de-AT", `content_management_api = false
 optional = true`)
	identity := statecheck.CompareValue(compare.ValuesSame())
	recreatedIdentity := statecheck.CompareValue(compare.ValuesDiffer())

	var localeID string

	defaults := []statecheck.StateCheck{
		statecheck.ExpectKnownValue(resourceAddress, tfjsonpath.New("content_delivery_api"), knownvalue.Bool(true)),
		statecheck.ExpectKnownValue(resourceAddress, tfjsonpath.New("content_management_api"), knownvalue.Bool(true)),
		statecheck.ExpectKnownValue(resourceAddress, tfjsonpath.New("optional"), knownvalue.Bool(false)),
		statecheck.ExpectKnownValue(resourceAddress, tfjsonpath.New("default"), knownvalue.Bool(false)),
	}

	testAccMockedResource(t, server, resource.TestCase{Steps: []resource.TestStep{
		{
			Config: initial,
			ConfigStateChecks: append([]statecheck.StateCheck{
				identity.AddStateValue(resourceAddress, tfjsonpath.New("locale_id")),
			}, defaults...),
		},
		{
			Config: updated,
			ConfigStateChecks: []statecheck.StateCheck{
				identity.AddStateValue(resourceAddress, tfjsonpath.New("locale_id")),
			},
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction(resourceAddress, plancheck.ResourceActionUpdate),
			}},
		},
		{
			Config: updated, ResourceName: resourceAddress, ImportState: true, ImportStateVerify: true,
		},
		{
			Config: updated, ResourceName: resourceAddress, ImportState: true,
			ImportStateKind: resource.ImportBlockWithResourceIdentity,
		},
		{
			Config: configuration("de-AT", `content_delivery_api = false`),
			ConfigStateChecks: []statecheck.StateCheck{
				identity.AddStateValue(resourceAddress, tfjsonpath.New("locale_id")),
				statecheck.ExpectKnownValue(resourceAddress, tfjsonpath.New("content_management_api"), knownvalue.Bool(true)),
			},
		},
		{
			Config: configuration("de-AT", ""),
			ConfigStateChecks: append([]statecheck.StateCheck{
				identity.AddStateValue(resourceAddress, tfjsonpath.New("locale_id")),
				recreatedIdentity.AddStateValue(resourceAddress, tfjsonpath.New("locale_id")),
				statecheck.ExpectKnownValue(resourceAddress, tfjsonpath.New("locale_id"), knownvalue.StringFunc(func(value string) error {
					localeID = value

					return nil
				})),
			}, defaults...),
		},
		{
			PreConfig: func() {
				response, err := server.Handler().DeleteLocale(t.Context(), cm.DeleteLocaleParams{SpaceID: "space", EnvironmentID: "environment", LocaleID: localeID})
				require.NoError(t, err)
				require.IsType(t, &cm.NoContent{}, response)
			},
			Config: configuration("de-AT", ""),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction(resourceAddress, plancheck.ResourceActionCreate),
			}},
			ConfigStateChecks: []statecheck.StateCheck{
				recreatedIdentity.AddStateValue(resourceAddress, tfjsonpath.New("locale_id")),
			},
		},
	}})
}
