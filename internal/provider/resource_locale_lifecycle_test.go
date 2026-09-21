package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

func TestAccLocaleResourceLifecycle(t *testing.T) {
	t.Parallel()

	const (
		resourceAddress = "contentful_locale.test"
		localeID        = "imported-locale"
	)

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")
	server.SetLocale(cmt.NewLocaleFromData("space", "environment", "default-locale", cm.LocaleData{
		Name: "English (United States)", Code: "en-US", FallbackCode: cm.NewNilStringNull(),
		ContentDeliveryApi: true, ContentManagementApi: true,
	}, true))

	server.SetLocale(cmt.NewLocaleFromData("space", "environment", localeID, cm.LocaleData{
		Name: "German", Code: "de-DE", FallbackCode: cm.NewNilString("en-US"),
		ContentDeliveryApi: true, ContentManagementApi: true,
	}, false))

	initial := config.Variables{"locale": config.ObjectVariable(map[string]config.Variable{
		"code": config.StringVariable("de-DE"), "fallback_code": config.StringVariable("en-US"),
	})}
	updated := config.Variables{"locale": config.ObjectVariable(map[string]config.Variable{
		"code": config.StringVariable("de-AT"), "content_management_api": config.BoolVariable(false), "optional": config.BoolVariable(true),
	})}
	restored := config.Variables{"locale": config.ObjectVariable(map[string]config.Variable{
		"code": config.StringVariable("de-AT"),
	})}
	localeIDCheck := statecheck.CompareValue(compare.ValuesSame())
	recreatedLocaleIDCheck := statecheck.CompareValue(compare.ValuesDiffer())

	resourceIdentity := statecheck.ExpectIdentity(resourceAddress, map[string]knownvalue.Check{
		"space_id":       knownvalue.StringExact("space"),
		"environment_id": knownvalue.StringExact("environment"),
		"locale_id":      knownvalue.StringExact(localeID),
	})

	defaults := []statecheck.StateCheck{
		statecheck.ExpectKnownValue(resourceAddress, tfjsonpath.New("content_delivery_api"), knownvalue.Bool(true)),
		statecheck.ExpectKnownValue(resourceAddress, tfjsonpath.New("content_management_api"), knownvalue.Bool(true)),
		statecheck.ExpectKnownValue(resourceAddress, tfjsonpath.New("optional"), knownvalue.Bool(false)),
		statecheck.ExpectKnownValue(resourceAddress, tfjsonpath.New("default"), knownvalue.Bool(false)),
	}

	testAccMockedResource(t, server, resource.TestCase{Steps: []resource.TestStep{
		{
			ConfigDirectory: config.TestNameDirectory(),
			ConfigVariables: initial,
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction(resourceAddress, plancheck.ResourceActionNoop),
			}},
			ConfigStateChecks: append([]statecheck.StateCheck{
				localeIDCheck.AddStateValue(resourceAddress, tfjsonpath.New("locale_id")),
				resourceIdentity,
				statecheck.ExpectKnownValue(resourceAddress, tfjsonpath.New("id"), knownvalue.StringExact("space/environment/imported-locale")),
				statecheck.ExpectKnownValue(resourceAddress, tfjsonpath.New("locale_id"), knownvalue.StringExact(localeID)),
				statecheck.ExpectKnownValue(resourceAddress, tfjsonpath.New("fallback_code"), knownvalue.StringExact("en-US")),
			}, defaults...),
		},
		{
			ConfigFile: config.TestNameFile("main.tf"), ConfigVariables: initial,
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectEmptyPlan(),
			}},
		},
		{
			ConfigFile: config.TestNameFile("main.tf"), ConfigVariables: updated,
			ConfigStateChecks: []statecheck.StateCheck{
				resourceIdentity,
				localeIDCheck.AddStateValue(resourceAddress, tfjsonpath.New("locale_id")),
			},
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction(resourceAddress, plancheck.ResourceActionUpdate),
			}},
		},
		{
			ConfigFile: config.TestNameFile("main.tf"), ConfigVariables: updated, ResourceName: resourceAddress, ImportState: true, ImportStateVerify: true,
		},
		{
			ConfigFile: config.TestNameFile("main.tf"),
			ConfigVariables: config.Variables{"locale": config.ObjectVariable(map[string]config.Variable{
				"code": config.StringVariable("de-AT"), "content_delivery_api": config.BoolVariable(false),
			})},
			ConfigStateChecks: []statecheck.StateCheck{
				localeIDCheck.AddStateValue(resourceAddress, tfjsonpath.New("locale_id")),
				statecheck.ExpectKnownValue(resourceAddress, tfjsonpath.New("content_management_api"), knownvalue.Bool(true)),
			},
		},
		{
			ConfigFile: config.TestNameFile("main.tf"), ConfigVariables: restored,
			ConfigStateChecks: append([]statecheck.StateCheck{
				localeIDCheck.AddStateValue(resourceAddress, tfjsonpath.New("locale_id")),
				recreatedLocaleIDCheck.AddStateValue(resourceAddress, tfjsonpath.New("locale_id")),
			}, defaults...),
		},
		{
			PreConfig: func() {
				response, err := server.Handler().DeleteLocale(t.Context(), cm.DeleteLocaleParams{SpaceID: "space", EnvironmentID: "environment", LocaleID: localeID})
				require.NoError(t, err)
				require.IsType(t, &cm.NoContent{}, response)
			},
			ConfigFile: config.TestNameFile("main.tf"),
			ConfigVariables: config.Variables{"locale": config.ObjectVariable(map[string]config.Variable{
				"code": config.StringVariable("de-AT"), "fallback_code": config.StringVariable("en-US"),
			})},
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction(resourceAddress, plancheck.ResourceActionCreate),
			}},
			ConfigStateChecks: append([]statecheck.StateCheck{
				recreatedLocaleIDCheck.AddStateValue(resourceAddress, tfjsonpath.New("locale_id")),
				statecheck.ExpectKnownValue(resourceAddress, tfjsonpath.New("fallback_code"), knownvalue.StringExact("en-US")),
			}, defaults...),
		},
	}})
}
