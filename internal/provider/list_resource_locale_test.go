package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/querycheck/queryfilter"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestAccLocaleListResource(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	server.SetLocale("space-id", "master", "en-US", cm.LocaleData{
		Name:                 "English (United States)",
		Code:                 "en-US",
		FallbackCode:         cm.NewNilStringNull(),
		ContentDeliveryApi:   true,
		ContentManagementApi: true,
		Optional:             false,
	}, true)

	server.SetLocale("space-id", "master", "de-DE", cm.LocaleData{
		Name:                 "German",
		Code:                 "de-DE",
		FallbackCode:         cm.NewNilString("en-US"),
		ContentDeliveryApi:   true,
		ContentManagementApi: true,
		Optional:             true,
	}, false)

	configVariables := config.Variables{
		"space_id":       config.StringVariable("space-id"),
		"environment_id": config.StringVariable("master"),
	}

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		Steps: []resource.TestStep{
			{
				Config: `
				provider "contentful" {}

				variable "space_id" {
					type = string
				}

				variable "environment_id" {
					type = string
				}

				list "contentful_locale" "locales" {
					provider = contentful

					config {
						space_id       = var.space_id
						environment_id = var.environment_id
					}

					include_resource = true
				}
				`,
				ConfigVariables: configVariables,
				Query:           true,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLengthAtLeast("contentful_locale.locales", 2),
					querycheck.ExpectIdentity("contentful_locale.locales", map[string]knownvalue.Check{
						"space_id":       knownvalue.StringExact("space-id"),
						"environment_id": knownvalue.StringExact("master"),
						"locale_id":      knownvalue.StringExact("en-US"),
					}),
					querycheck.ExpectResourceKnownValues("contentful_locale.locales", queryfilter.ByResourceIdentity(map[string]knownvalue.Check{
						"space_id":       knownvalue.StringExact("space-id"),
						"environment_id": knownvalue.StringExact("master"),
						"locale_id":      knownvalue.StringExact("de-DE"),
					}), []querycheck.KnownValueCheck{
						{
							Path:       tfjsonpath.New("id"),
							KnownValue: knownvalue.StringExact("space-id/master/de-DE"),
						},
						{
							Path:       tfjsonpath.New("name"),
							KnownValue: knownvalue.StringExact("German"),
						},
						{
							Path:       tfjsonpath.New("fallback_code"),
							KnownValue: knownvalue.StringExact("en-US"),
						},
						{
							Path:       tfjsonpath.New("optional"),
							KnownValue: knownvalue.Bool(true),
						},
						{
							Path:       tfjsonpath.New("default"),
							KnownValue: knownvalue.Bool(false),
						},
					}),
				},
			},
		},
	})
}
