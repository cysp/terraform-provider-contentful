package provider_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"sync/atomic"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

var (
	errUnexpectedExtensionResponse = errors.New("unexpected extension response")
	errExtensionSourceMismatch     = errors.New("extension source mismatch")
	errResolvedSrcdocOmitted       = errors.New("resolved srcdoc was omitted from Contentful request")
)

func TestAccExtensionResource(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	server.RegisterSpaceEnvironment("0p38pssr0fi3", "test")

	extensionID := "acctest_" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	configVariables := config.Variables{
		"space_id":          config.StringVariable("0p38pssr0fi3"),
		"environment_id":    config.StringVariable("test"),
		"test_extension_id": config.StringVariable(extensionID),
	}

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
				Check: testAccExtensionSources(
					server,
					"0p38pssr0fi3",
					"test",
					extensionID,
					cm.OptString{},
					cm.NewOptString("<!DOCTYPE html>"),
				),
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
				Check: testAccExtensionSources(
					server,
					"0p38pssr0fi3",
					"test",
					extensionID,
					cm.NewOptString("http://localhost:3000/entry-field.js"),
					cm.OptString{},
				),
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
				ImportState:     true,
				ResourceName:    "contentful_extension.test",
			},
		},
	})
}

func testAccExtensionSources(
	server *cmt.Server,
	spaceID string,
	environmentID string,
	extensionID string,
	expectedSrc cm.OptString,
	expectedSrcdoc cm.OptString,
) resource.TestCheckFunc {
	if os.Getenv("TF_ACC_MOCKED") == "" {
		checks := []resource.TestCheckFunc{}
		if src, ok := expectedSrc.Get(); ok {
			checks = append(checks, resource.TestCheckResourceAttr("contentful_extension.test", "extension.src", src))
		}

		if srcdoc, ok := expectedSrcdoc.Get(); ok {
			checks = append(checks, resource.TestCheckResourceAttr("contentful_extension.test", "extension.srcdoc", srcdoc))
		}

		retryClient := retryablehttp.NewClient()
		retryClient.Logger = nil
		retryClient.RetryMax = 5

		client, err := cm.NewClient(
			cm.DefaultServerURL,
			cm.NewAccessTokenSecuritySource(os.Getenv("CONTENTFUL_MANAGEMENT_ACCESS_TOKEN")),
			cm.WithClient(cm.NewTransportClient(retryClient.StandardClient(), cm.DefaultUserAgent)),
		)
		if err != nil {
			return func(*terraform.State) error {
				return fmt.Errorf("create Contentful client for Extension check: %w", err)
			}
		}

		checks = append(checks, testContentfulExtensionSources(
			func(ctx context.Context, params cm.GetExtensionParams) (cm.GetExtensionRes, error) {
				return client.GetExtension(ctx, params)
			},
			spaceID,
			environmentID,
			extensionID,
			expectedSrc,
			expectedSrcdoc,
		))

		return resource.ComposeTestCheckFunc(checks...)
	}

	return testContentfulExtensionSources(server.Handler().GetExtension, spaceID, environmentID, extensionID, expectedSrc, expectedSrcdoc)
}

type extensionGetter func(context.Context, cm.GetExtensionParams) (cm.GetExtensionRes, error)

func testContentfulExtensionSources(
	getExtension extensionGetter,
	spaceID string,
	environmentID string,
	extensionID string,
	expectedSrc cm.OptString,
	expectedSrcdoc cm.OptString,
) resource.TestCheckFunc {
	return func(*terraform.State) error {
		response, err := getExtension(context.Background(), cm.GetExtensionParams{
			SpaceID:       spaceID,
			EnvironmentID: environmentID,
			ExtensionID:   extensionID,
		})
		if err != nil {
			return fmt.Errorf("get extension from Contentful: %w", err)
		}

		extension, ok := response.(*cm.Extension)
		if !ok {
			return fmt.Errorf("%w: %T", errUnexpectedExtensionResponse, response)
		}

		if extension.Extension.Src != expectedSrc {
			return fmt.Errorf("%w: Contentful received src %#v, want %#v", errExtensionSourceMismatch, extension.Extension.Src, expectedSrc)
		}

		if extension.Extension.Srcdoc != expectedSrcdoc {
			return fmt.Errorf("%w: Contentful received srcdoc %#v, want %#v", errExtensionSourceMismatch, extension.Extension.Srcdoc, expectedSrcdoc)
		}

		return nil
	}
}

func TestAccExtensionResourceExplicitEmptySrcdocReachesContentful(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()
	server.RegisterSpaceEnvironment("space", "environment")

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: `
resource "contentful_extension" "test" {
  space_id       = "space"
  environment_id = "environment"
  extension_id   = "empty-srcdoc"

  extension = {
    name        = "Empty srcdoc"
    srcdoc      = ""
    field_types = []
  }
}
`,
				Check: testContentfulExtensionSources(
					server.Handler().GetExtension,
					"space",
					"environment",
					"empty-srcdoc",
					cm.OptString{},
					cm.NewOptString(""),
				),
			},
		},
	})
}

func TestAccExtensionResourceRejectsInvalidSourcesBeforeContentful(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		sources       string
		expectedError string
	}{
		"neither source": {
			expectedError: "No attribute specified when one (and only one)",
		},
		"empty src": {
			sources:       `src = ""`,
			expectedError: "at least 1",
		},
		"both sources": {
			sources: `
    src    = "https://example.com/extension.js"
    srcdoc = "<!doctype html>"
`,
			expectedError: "2 attributes specified when one (and only one)",
		},
	} {
		t.Run(name, func(t *testing.T) {
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
						Config: fmt.Sprintf(`
resource "contentful_extension" "test" {
  space_id       = "space"
  environment_id = "environment"
  extension_id   = "invalid-source"

  extension = {
    name        = "Invalid source"
    field_types = []
    %s
  }
}
`, test.sources),
						ExpectError: regexp.MustCompile(regexp.QuoteMeta(test.expectedError)),
					},
				},
			})

			require.Zero(t, requestCount.Load())
		})
	}
}

func TestAccExtensionResourceResolvedDependencyValueReachesContentful(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()
	server.RegisterSpaceEnvironment("space", "environment")

	const resolvedSrcdoc = "<!doctype html><title>resolved during apply</title>"

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: `
resource "terraform_data" "source" {
  input = "<!doctype html><title>resolved during apply</title>"
}

resource "contentful_extension" "test" {
  space_id       = "space"
  environment_id = "environment"
  extension_id   = "resolved-dependency"

  extension = {
    name        = "Resolved dependency"
    srcdoc      = terraform_data.source.output
    field_types = []
  }
}
`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectUnknownValue(
							"contentful_extension.test",
							tfjsonpath.New("extension").AtMapKey("srcdoc"),
						),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("contentful_extension.test", "extension.srcdoc", resolvedSrcdoc),
					func(*terraform.State) error {
						response, err := server.Handler().GetExtension(t.Context(), cm.GetExtensionParams{
							SpaceID:       "space",
							EnvironmentID: "environment",
							ExtensionID:   "resolved-dependency",
						})
						if err != nil {
							return fmt.Errorf("get extension from mock Contentful: %w", err)
						}

						extension, ok := response.(*cm.Extension)
						if !ok {
							return fmt.Errorf("%w: %T", errUnexpectedExtensionResponse, response)
						}

						actual, ok := extension.Extension.Srcdoc.Get()
						if !ok {
							return errResolvedSrcdocOmitted
						}

						if actual != resolvedSrcdoc {
							return fmt.Errorf("%w: Contentful received srcdoc %q, want %q", errExtensionSourceMismatch, actual, resolvedSrcdoc)
						}

						return nil
					},
				),
			},
		},
	})
}
