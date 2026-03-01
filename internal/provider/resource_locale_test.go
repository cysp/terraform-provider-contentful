package provider_test

import (
	"maps"
	"regexp"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

//nolint:paralleltest
func TestAccLocaleResourceImport(t *testing.T) {
	server, _ := cmt.NewContentfulManagementServer()

	server.SetLocale("0p38pssr0fi3", "en-US", cm.LocaleRequest{
		Name:                 "English (US)",
		Code:                 "en-US",
		Optional:             cm.NewOptBool(true),
		ContentDeliveryAPI:   cm.NewOptBool(true),
		ContentManagementAPI: cm.NewOptBool(true),
		FallbackCode:         cm.NewOptNilStringNull(),
	})

	configVariables := config.Variables{
		"space_id":  config.StringVariable("0p38pssr0fi3"),
		"locale_id": config.StringVariable("en-US"),
		"name":      config.StringVariable("English (US)"),
		"code":      config.StringVariable("en-US"),
	}

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory:    config.TestNameDirectory(),
				ConfigVariables:    configVariables,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				ConfigDirectory:    config.TestNameDirectory(),
				ConfigVariables:    configVariables,
				ResourceName:       "contentful_locale.test",
				ImportState:        true,
				ImportStateId:      "0p38pssr0fi3/en-US",
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

//nolint:paralleltest
func TestAccLocaleResourceImportNotFound(t *testing.T) {
	server, _ := cmt.NewContentfulManagementServer()

	configVariables := config.Variables{
		"space_id":  config.StringVariable("0p38pssr0fi3"),
		"locale_id": config.StringVariable("nonexistent"),
		"name":      config.StringVariable("English (US)"),
		"code":      config.StringVariable("en-US"),
	}

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory:    config.TestNameDirectory(),
				ConfigVariables:    configVariables,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ResourceName:    "contentful_locale.test",
				ImportState:     true,
				ImportStateId:   "0p38pssr0fi3/nonexistent",
				ExpectError:     regexp.MustCompile(`Cannot import non-existent remote object`),
			},
		},
	})
}

//nolint:paralleltest
func TestAccLocaleResourceCreateUpdate(t *testing.T) {
	server, _ := cmt.NewContentfulManagementServer()

	localeID := "acctest-" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	configVariables := config.Variables{
		"space_id":  config.StringVariable("0p38pssr0fi3"),
		"locale_id": config.StringVariable(localeID),
		"name":      config.StringVariable("Locale " + localeID),
		"code":      config.StringVariable(localeID),
	}

	configVariables1 := maps.Clone(configVariables)

	configVariables2 := maps.Clone(configVariables)
	configVariables2["name"] = config.StringVariable("Locale " + localeID + " Updated")

	configVariables3 := maps.Clone(configVariables2)
	configVariables3["fallback_code"] = config.StringVariable("en-US")
	configVariables3["optional"] = config.BoolVariable(true)

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables1,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_locale.test", plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables2,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_locale.test", plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables3,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_locale.test", plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}
