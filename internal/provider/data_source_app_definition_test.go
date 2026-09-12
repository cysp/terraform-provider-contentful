package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

//nolint:paralleltest
func TestAccAppDefinitionDataSourceRead(t *testing.T) {
	parallelWhenMocked(t)

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)

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
