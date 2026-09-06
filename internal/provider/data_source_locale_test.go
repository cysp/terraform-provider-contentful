package provider_test

import (
	"regexp"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccLocaleDataSource(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	server.SetLocale("0p38pssr0fi3", "master", "2EElC09UknSbiccBgPK9ib", cm.LocaleData{
		Name:                 "English (Australia)",
		Code:                 "en-AU",
		FallbackCode:         cm.NewNilStringNull(),
		ContentDeliveryApi:   true,
		ContentManagementApi: true,
		Optional:             false,
	}, true)

	configVariables := config.Variables{
		"space_id":       config.StringVariable("0p38pssr0fi3"),
		"environment_id": config.StringVariable("master"),
		"locale_id":      config.StringVariable("2EElC09UknSbiccBgPK9ib"),
	}

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.contentful_locale.test", "id", "0p38pssr0fi3/master/2EElC09UknSbiccBgPK9ib"),
					resource.TestCheckResourceAttr("data.contentful_locale.test", "locale_id", "2EElC09UknSbiccBgPK9ib"),
					resource.TestCheckResourceAttrSet("data.contentful_locale.test", "name"),
					resource.TestCheckResourceAttr("data.contentful_locale.test", "code", "en-AU"),
					resource.TestCheckResourceAttr("data.contentful_locale.test", "default", "true"),
				),
			},
		},
	})
}

func TestAccLocaleDataSourceNotFound(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	configVariables := config.Variables{
		"space_id":       config.StringVariable("space-id"),
		"environment_id": config.StringVariable("master"),
		"locale_id":      config.StringVariable("nonexistent"),
	}

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ExpectError:     regexp.MustCompile(`Failed to read locale`),
			},
		},
	})
}
