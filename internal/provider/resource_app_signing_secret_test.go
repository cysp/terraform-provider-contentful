package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestAccAppSigningSecretResource(t *testing.T) {
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
					"organization_id":      config.StringVariable("organization-id"),
					"app_definition_id":    config.StringVariable("app-definition-id"),
					"signing_secret_value": config.StringVariable("secret"),
				},
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"organization_id":      config.StringVariable("organization-id"),
					"app_definition_id":    config.StringVariable("app-definition-id"),
					"signing_secret_value": config.StringVariable("updated-secret"),
				},
			},
		},
	})
}

func TestAccAppSigningSecretResourceWriteOnly(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()
	server.SetAppDefinition("organization-id", "app-definition-id", cm.AppDefinitionData{
		Name: "Test App",
	})

	const signingSecretConfig = `
resource "contentful_app_signing_secret" "test" {
  organization_id   = "organization-id"
  app_definition_id = "app-definition-id"
  value_wo          = "write-only-secret"
}
`

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: signingSecretConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("contentful_app_signing_secret.test", "value"),
					resource.TestCheckNoResourceAttr("contentful_app_signing_secret.test", "value_wo"),
				),
			},
			{
				Config: signingSecretConfig,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_app_signing_secret.test", plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccAppSigningSecretImport(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	configVariables := config.Variables{
		"organization_id":   config.StringVariable("organization-id"),
		"app_definition_id": config.StringVariable("app-definition-id"),
	}

	server.SetAppDefinition("organization-id", "app-definition-id", cm.AppDefinitionData{
		Name: "Test App",
	})

	server.SetAppSigningSecret("organization-id", "app-definition-id", cm.AppSigningSecretRequestData{
		Value: "secret",
	})

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("contentful_app_signing_secret.test", "id", "organization-id/app-definition-id"),
				),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ImportState:     true,
				ResourceName:    "contentful_app_signing_secret.test",
			},
		},
	})
}
