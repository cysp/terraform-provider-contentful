package provider_test

import (
	"net/http"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/querycheck/queryfilter"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccLocaleListResource(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")

	server.SetLocale(cmt.NewLocaleFromData("space", "master", "locale-a", cm.LocaleData{
		Name:                 "English (Australia)",
		Code:                 "en-AU",
		FallbackCode:         cm.NewNilStringNull(),
		ContentDeliveryApi:   true,
		ContentManagementApi: true,
		Optional:             false,
	}, true))

	server.SetLocale(cmt.NewLocaleFromData("space", "master", "locale-b", cm.LocaleData{
		Name:                 "English (Canada)",
		Code:                 "en-CA",
		FallbackCode:         cm.NewNilString("en-AU"),
		ContentDeliveryApi:   true,
		ContentManagementApi: true,
		Optional:             true,
	}, false))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "sys.id", r.URL.Query().Get("order"))
		server.ServeHTTP(w, r)
	})

	testAccMockedResource(t, handler, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		Steps: []resource.TestStep{
			{
				Config: `
				provider "contentful" {}

				list "contentful_locale" "locales" {
					provider = contentful

					config {
						space_id       = "space"
						environment_id = "master"
					}

					include_resource = true
				}
				`,
				Query: true,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectLength("contentful_locale.locales", 2),
					querycheck.ExpectResourceKnownValues("contentful_locale.locales", queryfilter.ByResourceIdentity(map[string]knownvalue.Check{
						"space_id":       knownvalue.StringExact("space"),
						"environment_id": knownvalue.StringExact("master"),
						"locale_id":      knownvalue.StringExact("locale-a"),
					}), []querycheck.KnownValueCheck{
						{
							Path:       tfjsonpath.New("id"),
							KnownValue: knownvalue.StringExact("space/master/locale-a"),
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
