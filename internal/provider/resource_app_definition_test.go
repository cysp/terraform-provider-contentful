package provider_test

import (
	"net/http"
	"regexp"
	"sync/atomic"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

//nolint:paralleltest
func TestAccAppDefinitionResource(t *testing.T) {
	parallelWhenMocked(t)

	server, _ := cmt.NewContentfulManagementServer()

	configVariables := config.Variables{
		"organization_id": config.StringVariable("2zuSjSO4A0e6GKBrhJRe2m"),
	}

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
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

func TestAccAppDefinitionResourceRejectsEmptySourceBeforeContentful(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)

	var requestCount atomic.Int64

	handler := http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		requestCount.Add(1)
		server.ServeHTTP(responseWriter, request)
	})

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: `
resource "contentful_app_definition" "test" {
  organization_id = "organization"
  name            = "Empty source"
  src             = ""
  locations       = []
}
`,
				ExpectError: regexp.MustCompile(`at least 1`),
			},
		},
	})

	require.Zero(t, requestCount.Load())
}

func TestAccAppDefinitionResourceImport(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	configVariables := config.Variables{
		"organization_id": config.StringVariable("2zuSjSO4A0e6GKBrhJRe2m"),
	}

	server.SetAppDefinition("2zuSjSO4A0e6GKBrhJRe2m", "app-definition-id", cm.AppDefinitionData{
		Name:   "Test App",
		Bundle: cm.NewOptAppBundleLink(cm.NewAppBundleLink("app-bundle-id")),
	})

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
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
	parallelWhenMocked(t)

	server, _ := cmt.NewContentfulManagementServer()

	configVariables := config.Variables{
		"organization_id": config.StringVariable("2zuSjSO4A0e6GKBrhJRe2m"),
	}

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
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
