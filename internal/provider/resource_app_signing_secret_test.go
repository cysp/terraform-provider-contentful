package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

//nolint:paralleltest
func TestAccAppSigningSecretResource(t *testing.T) {
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

//nolint:paralleltest
func TestAccAppSigningSecretImport(t *testing.T) {
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
