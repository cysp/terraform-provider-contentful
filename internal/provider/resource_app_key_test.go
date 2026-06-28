package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestAccAppKeyResourceJWK(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	server.SetAppDefinition("organization-id", "app-definition-id", cm.AppDefinitionData{
		Name: "Test App",
	})

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"organization_id":   config.StringVariable("organization-id"),
					"app_definition_id": config.StringVariable("app-definition-id"),
					"key_kid":           config.StringVariable("key-id"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("contentful_app_key.test", "id", "organization-id/app-definition-id/key-id"),
					resource.TestCheckResourceAttr("contentful_app_key.test", "key_kid", "key-id"),
					resource.TestCheckResourceAttr("contentful_app_key.test", "jwk.alg", "RS256"),
					resource.TestCheckResourceAttr("contentful_app_key.test", "jwk.kty", "RSA"),
					resource.TestCheckResourceAttr("contentful_app_key.test", "jwk.use", "sig"),
					resource.TestCheckResourceAttr("contentful_app_key.test", "jwk.kid", "key-id"),
					resource.TestCheckResourceAttr("contentful_app_key.test", "jwk.x5c.#", "1"),
					resource.TestCheckResourceAttr("contentful_app_key.test", "jwk.x5c.0", "certificate"),
					resource.TestCheckResourceAttr("contentful_app_key.test", "jwk.x5t", "key-id"),
					resource.TestCheckNoResourceAttr("contentful_app_key.test", "private_key"),
				),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"organization_id":   config.StringVariable("organization-id"),
					"app_definition_id": config.StringVariable("app-definition-id"),
					"key_kid":           config.StringVariable("replacement-key-id"),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_app_key.test", plancheck.ResourceActionReplace),
					},
				},
			},
		},
	})
}

func TestAccAppKeyResourceGenerate(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	server.SetAppDefinition("organization-id", "app-definition-id", cm.AppDefinitionData{
		Name: "Test App",
	})

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"organization_id":   config.StringVariable("organization-id"),
					"app_definition_id": config.StringVariable("app-definition-id"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("contentful_app_key.test", "id", "organization-id/app-definition-id/generated-key"),
					resource.TestCheckResourceAttr("contentful_app_key.test", "key_kid", "generated-key"),
					resource.TestCheckResourceAttr("contentful_app_key.test", "jwk.kid", "generated-key"),
					resource.TestCheckResourceAttr("contentful_app_key.test", "private_key", "generated-private-key"),
				),
			},
			{
				RefreshState: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("contentful_app_key.test", "private_key", "generated-private-key"),
				),
			},
		},
	})
}

func TestAccAppKeyResourceNullJWK(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	server.SetAppDefinition("organization-id", "app-definition-id", cm.AppDefinitionData{
		Name: "Test App",
	})

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("contentful_app_key.test", "id", "organization-id/app-definition-id/generated-key"),
					resource.TestCheckResourceAttr("contentful_app_key.test", "key_kid", "generated-key"),
					resource.TestCheckResourceAttr("contentful_app_key.test", "jwk.kid", "generated-key"),
					resource.TestCheckResourceAttr("contentful_app_key.test", "private_key", "generated-private-key"),
				),
			},
		},
	})
}

func TestAccAppKeyResourceImport(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	configVariables := config.Variables{
		"organization_id":   config.StringVariable("organization-id"),
		"app_definition_id": config.StringVariable("app-definition-id"),
		"key_kid":           config.StringVariable("key-id"),
	}

	server.SetAppDefinition("organization-id", "app-definition-id", cm.AppDefinitionData{
		Name: "Test App",
	})

	server.SetAppKey("organization-id", "app-definition-id", cm.AppKeyRequestData{
		Jwk: cm.NewOptAppKeyJWK(testAppKeyJWK("key-id")),
	})

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("contentful_app_key.test", "id", "organization-id/app-definition-id/key-id"),
				),
			},
			{
				ConfigDirectory:   config.TestNameDirectory(),
				ConfigVariables:   configVariables,
				ImportState:       true,
				ImportStateVerify: true,
				ResourceName:      "contentful_app_key.test",
			},
		},
	})
}

func testAppKeyJWK(keyID string) cm.AppKeyJWK {
	return cm.AppKeyJWK{
		Alg: cm.AppKeyJWKAlgRS256,
		Kty: cm.AppKeyJWKKtyRSA,
		Use: cm.AppKeyJWKUseSig,
		X5c: []string{"certificate"},
		Kid: keyID,
		X5t: keyID,
	}
}
