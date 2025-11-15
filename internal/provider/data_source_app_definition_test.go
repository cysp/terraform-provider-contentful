package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

//nolint:paralleltest
func TestAccAppDefinitionDataSource(t *testing.T) {
	server, _ := cmt.NewContentfulManagementServer()

	configVariables := config.Variables{
		"organization_id":   config.StringVariable("2zuSjSO4A0e6GKBrhJRe2m"),
		"app_definition_id": config.StringVariable("2fxGxOcam8Fo5m1wC11fhn"),
	}

	server.SetAppDefinition("2zuSjSO4A0e6GKBrhJRe2m", "2fxGxOcam8Fo5m1wC11fhn", cm.AppDefinitionData{
		Name: "Test App",
		Bundle: cm.NewOptAppBundleLink(cm.AppBundleLink{
			Sys: cm.AppBundleLinkSys{
				Type:     cm.AppBundleLinkSysTypeLink,
				LinkType: cm.AppBundleLinkSysLinkTypeAppBundle,
				ID:       "app-bundle-id",
			},
		}),
	})

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.contentful_app_definition.test", "id", "2zuSjSO4A0e6GKBrhJRe2m/2fxGxOcam8Fo5m1wC11fhn"),
				),
			},
		},
	})
}
