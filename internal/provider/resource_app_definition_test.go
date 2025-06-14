package provider_test

import (
	"regexp"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmts "github.com/cysp/terraform-provider-contentful/internal/contentful-management-testserver"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

//nolint:paralleltest
func TestAccAppDefinitionResource(t *testing.T) {
	testserver := cmts.NewContentfulManagementTestServer()
	defer testserver.Server().Close()

	configVariables := config.Variables{
		"organization_id": config.StringVariable("2zuSjSO4A0e6GKBrhJRe2m"),
	}

	ContentfulProviderMockableResourceTest(t, testserver.Server(), resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
			},
		},
	})
}

//nolint:paralleltest
func TestAccAppDefinitionResourceImport(t *testing.T) {
	testserver := cmts.NewContentfulManagementTestServer()
	defer testserver.Server().Close()

	configVariables := config.Variables{
		"organization_id": config.StringVariable("2zuSjSO4A0e6GKBrhJRe2m"),
	}

	testserver.SetAppDefinition("2zuSjSO4A0e6GKBrhJRe2m", "app-definition-id", cm.AppDefinitionFields{
		Name: "Test App",
		Bundle: cm.NewOptAppBundleLink(cm.AppBundleLink{
			Sys: cm.AppBundleLinkSys{
				Type:     cm.AppBundleLinkSysTypeLink,
				LinkType: cm.AppBundleLinkSysLinkTypeAppBundle,
				ID:       "app-bundle-id",
			},
		}),
	})

	ContentfulProviderMockedResourceTest(t, testserver.Server(), resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ResourceName:    "contentful_app_definition.test",
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ResourceName:    "contentful_app_definition.test",
				ImportState:     true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("contentful_app_definition.test", "id", "2zuSjSO4A0e6GKBrhJRe2m/app-definition-id"),
					resource.TestCheckResourceAttr("contentful_app_definition.test", "bundle_id", "app-bundle-id"),
				),
			},
		},
	})
}

//nolint:paralleltest
func TestAccAppDefinitionResourceImportNotFound(t *testing.T) {
	testserver := cmts.NewContentfulManagementTestServer()
	defer testserver.Server().Close()

	configVariables := config.Variables{
		"organization_id": config.StringVariable("2zuSjSO4A0e6GKBrhJRe2m"),
	}

	ContentfulProviderMockableResourceTest(t, testserver.Server(), resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ResourceName:    "contentful_app_definition.test",
				ImportState:     true,
				ImportStateId:   "2zuSjSO4A0e6GKBrhJRe2m/nonexistent",
				ExpectError:     regexp.MustCompile(`Cannot import non-existent remote object`),
			},
		},
	})
}
