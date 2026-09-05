package provider_test

import (
	"fmt"
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
func TestAccAppDefinitionResourceLifecycle(t *testing.T) {
	parallelWhenMocked(t)

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)

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

func TestAccAppDefinitionResourceRejectsEmptySourceBeforeContentfulMutation(t *testing.T) {
	t.Parallel()

	config := func(name, src string) string {
		return fmt.Sprintf(`
resource "contentful_app_definition" "test" {
  organization_id = "organization"
  name            = %q
  src             = %q
  locations       = []
}
`, name, src)
	}

	for name, test := range map[string]struct {
		mutationMethod string
		steps          []resource.TestStep
	}{
		"create": {
			mutationMethod: http.MethodPost,
			steps: []resource.TestStep{{
				Config:      config("Empty source", ""),
				ExpectError: regexp.MustCompile(`at least 1`),
			}},
		},
		"update": {
			mutationMethod: http.MethodPut,
			steps: []resource.TestStep{
				{Config: config("Valid source", "https://example.com/app.js")},
				{
					Config:      config("Empty source update", ""),
					ExpectError: regexp.MustCompile(`at least 1`),
				},
				{
					Config:  config("Valid source", "https://example.com/app.js"),
					Destroy: true,
				},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer()
			require.NoError(t, err)

			var mutationCount atomic.Int64

			handler := http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
				if request.Method == test.mutationMethod {
					mutationCount.Add(1)
				}

				server.ServeHTTP(responseWriter, request)
			})

			if name == "update" {
				test.steps[1].PreConfig = func() { mutationCount.Store(0) }
			}

			ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{Steps: test.steps})

			require.Zero(t, mutationCount.Load())
		})
	}
}

func TestAccAppDefinitionResourceImport(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)

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

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)

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
