package provider_test

import (
	"maps"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccLocaleResource(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	server.RegisterSpaceEnvironment("space-id", "master")
	server.SetLocale("space-id", "master", "en-US", cm.LocaleData{
		Name:                 "English (United States)",
		Code:                 "en-US",
		FallbackCode:         cm.NewNilStringNull(),
		ContentDeliveryApi:   true,
		ContentManagementApi: true,
		Optional:             false,
	}, true)

	configVariables := config.Variables{
		"space_id":       config.StringVariable("space-id"),
		"environment_id": config.StringVariable("master"),
		"name":           config.StringVariable("German"),
		"code":           config.StringVariable("de-DE"),
		"fallback_code":  config.StringVariable("en-US"),
	}

	configVariables1 := maps.Clone(configVariables)

	configVariables2 := maps.Clone(configVariables)
	configVariables2["name"] = config.StringVariable("German (Germany)")
	configVariables2["optional"] = config.BoolVariable(true)

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("contentful_locale.test", "id", "space-id/master/de-DE"),
					resource.TestCheckResourceAttr("contentful_locale.test", "locale_id", "de-DE"),
					resource.TestCheckResourceAttr("contentful_locale.test", "name", "German"),
					resource.TestCheckResourceAttr("contentful_locale.test", "code", "de-DE"),
					resource.TestCheckResourceAttr("contentful_locale.test", "fallback_code", "en-US"),
					resource.TestCheckResourceAttr("contentful_locale.test", "content_delivery_api", "true"),
					resource.TestCheckResourceAttr("contentful_locale.test", "content_management_api", "true"),
					resource.TestCheckResourceAttr("contentful_locale.test", "optional", "false"),
					resource.TestCheckResourceAttr("contentful_locale.test", "default", "false"),
					resource.TestCheckResourceAttr("contentful_locale.test", "internal_code", "de-DE"),
				),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("contentful_locale.test", "name", "German (Germany)"),
					resource.TestCheckResourceAttr("contentful_locale.test", "optional", "true"),
				),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables2,
				ImportState:     true,
				ImportStateId:   "space-id/master/de-DE",
				ResourceName:    "contentful_locale.test",
			},
		},
	})
}

func TestAccLocaleResourceImport(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	server.SetLocale("space-id", "master", "fr-FR", cm.LocaleData{
		Name:                 "French",
		Code:                 "fr-FR",
		FallbackCode:         cm.NewNilString("en-US"),
		ContentDeliveryApi:   true,
		ContentManagementApi: true,
		Optional:             true,
	}, false)

	configVariables := config.Variables{
		"space_id":       config.StringVariable("space-id"),
		"environment_id": config.StringVariable("master"),
		"name":           config.StringVariable("French"),
		"code":           config.StringVariable("fr-FR"),
		"fallback_code":  config.StringVariable("en-US"),
		"optional":       config.BoolVariable(true),
	}

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ImportState:     true,
				ImportStateId:   "space-id/master/fr-FR",
				ResourceName:    "contentful_locale.test",
			},
		},
	})
}
