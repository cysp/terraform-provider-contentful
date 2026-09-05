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

	server.SetLocale("0p38pssr0fi3", "master", "2EElC09UknSbiccBgPK9ib", cm.LocaleData{
		Name:                 "English (Australia)",
		Code:                 "en-AU",
		FallbackCode:         cm.NewNilStringNull(),
		ContentDeliveryApi:   true,
		ContentManagementApi: true,
		Optional:             false,
	}, true)

	server.SetLocale("0p38pssr0fi3", "master", "5bEdTq1ah9h5lYOdHICIES", cm.LocaleData{
		Name:                 "English (Canada)",
		Code:                 "en-CA",
		FallbackCode:         cm.NewNilString("en-AU"),
		ContentDeliveryApi:   true,
		ContentManagementApi: true,
		Optional:             true,
	}, false)

	configVariables := config.Variables{
		"space_id":       config.StringVariable("0p38pssr0fi3"),
		"environment_id": config.StringVariable("master"),
	}

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
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
					querycheck.ExpectLengthAtLeast("contentful_locale.locales", 1),
					querycheck.ExpectIdentity("contentful_locale.locales", map[string]knownvalue.Check{
						"space_id":       knownvalue.StringExact("0p38pssr0fi3"),
						"environment_id": knownvalue.StringExact("master"),
						"locale_id":      knownvalue.StringExact("2EElC09UknSbiccBgPK9ib"),
					}),
					querycheck.ExpectResourceKnownValues("contentful_locale.locales", queryfilter.ByResourceIdentity(map[string]knownvalue.Check{
						"space_id":       knownvalue.StringExact("0p38pssr0fi3"),
						"environment_id": knownvalue.StringExact("master"),
						"locale_id":      knownvalue.StringExact("2EElC09UknSbiccBgPK9ib"),
					}), []querycheck.KnownValueCheck{
						{
							Path:       tfjsonpath.New("id"),
							KnownValue: knownvalue.StringExact("0p38pssr0fi3/master/2EElC09UknSbiccBgPK9ib"),
						},
						{
							Path:       tfjsonpath.New("code"),
							KnownValue: knownvalue.StringExact("en-AU"),
						},
						{
							Path:       tfjsonpath.New("default"),
							KnownValue: knownvalue.Bool(true),
						},
						{
							Path:       tfjsonpath.New("timeouts"),
							KnownValue: knownvalue.Null(),
						},
					}),
				},
			},
		},
	})
}
