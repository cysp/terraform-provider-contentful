package provider_test

import (
	"regexp"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

//nolint:paralleltest
func TestAccLocaleDataSource(t *testing.T) {
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
	}

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.contentful_locale.test", "id", "0p38pssr0fi3/en-US"),
					resource.TestCheckResourceAttr("data.contentful_locale.test", "name", "English (US)"),
					resource.TestCheckResourceAttr("data.contentful_locale.test", "code", "en-US"),
				),
			},
		},
	})
}

func TestAccLocaleDataSourceNotFound(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	configVariables := config.Variables{
		"space_id":  config.StringVariable("0p38pssr0fi3"),
		"locale_id": config.StringVariable("nonexistent"),
	}

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ExpectError:     regexp.MustCompile(`Provider produced null object`),
			},
		},
	})
}
