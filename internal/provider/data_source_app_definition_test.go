package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

//nolint:paralleltest
func TestAccAppDefinitionDataSourceRead(t *testing.T) {
	parallelWhenMocked(t)

	server := newTestCMAServer(t)

	configVariables := config.Variables{
		"organization_id":   config.StringVariable("2zuSjSO4A0e6GKBrhJRe2m"),
		"app_definition_id": config.StringVariable("2fxGxOcam8Fo5m1wC11fhn"),
	}

	server.SetAppDefinition("2zuSjSO4A0e6GKBrhJRe2m", "2fxGxOcam8Fo5m1wC11fhn", cm.AppDefinitionData{
		Name:   "Test App",
		Bundle: cm.NewOptAppBundleLink(cm.NewAppBundleLink("app-bundle-id")),
	})

	testAccMockableResource(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("data.contentful_app_definition.test", tfjsonpath.New("id"), knownvalue.StringExact("2zuSjSO4A0e6GKBrhJRe2m/2fxGxOcam8Fo5m1wC11fhn")),
				},
			},
		},
	})
}
