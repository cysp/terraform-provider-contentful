package provider_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type previewEnvironmentPlanCheckFunc func(context.Context, plancheck.CheckPlanRequest, *plancheck.CheckPlanResponse)

func (f previewEnvironmentPlanCheckFunc) CheckPlan(ctx context.Context, req plancheck.CheckPlanRequest, resp *plancheck.CheckPlanResponse) {
	f(ctx, req, resp)
}

func previewEnvironmentResourceConfig(name, previewEnvironmentID string, contentTypeIDs ...string) string {
	var configurations strings.Builder
	for _, contentTypeID := range contentTypeIDs {
		fmt.Fprintf(&configurations, `
    %q = {
      url = %q
    }`, contentTypeID, "https://preview.example.invalid/"+contentTypeID+"/{entry.sys.id}")
	}

	selectedID := ""
	if previewEnvironmentID != "" {
		selectedID = fmt.Sprintf("  preview_environment_id = %q\n", previewEnvironmentID)
	}

	return fmt.Sprintf(`
resource "contentful_preview_environment" "test" {
  space_id = "0p38pssr0fi3"
  name     = %q
%s
  content_type_configurations = {%s
  }
}
`, name, selectedID, configurations.String())
}

func capturePreviewEnvironmentID(ids *[]string) resource.TestCheckFunc {
	return resource.TestCheckResourceAttrWith(
		"contentful_preview_environment.test",
		"preview_environment_id",
		func(value string) error {
			*ids = append(*ids, value)

			return nil
		},
	)
}

func registerLivePreviewEnvironmentCleanup(t *testing.T, ids *[]string) {
	t.Helper()

	if os.Getenv("TF_ACC") == "" || os.Getenv("TF_ACC_MOCKED") != "" {
		return
	}

	t.Cleanup(func() {
		if len(*ids) == 0 {
			return
		}

		client := livePreviewEnvironmentClient(t)

		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), time.Minute)
		defer cancel()

		seen := make(map[string]struct{}, len(*ids))
		for _, previewEnvironmentID := range *ids {
			if _, ok := seen[previewEnvironmentID]; ok {
				continue
			}

			seen[previewEnvironmentID] = struct{}{}

			deleteResponse, err := client.DeletePreviewEnvironment(ctx, cm.DeletePreviewEnvironmentParams{
				SpaceID:              "0p38pssr0fi3",
				PreviewEnvironmentID: previewEnvironmentID,
			})
			require.NoError(t, err)

			if _, ok := deleteResponse.(*cm.NoContent); !ok {
				statusResponse, statusOK := deleteResponse.(cm.StatusCodeResponse)
				require.True(t, statusOK && statusResponse.GetStatusCode() == http.StatusNotFound, "unexpected cleanup delete response: %T", deleteResponse)
			}

			require.EventuallyWithT(t, func(collect *assert.CollectT) {
				getResponse, err := client.GetPreviewEnvironment(ctx, cm.GetPreviewEnvironmentParams{
					SpaceID:              "0p38pssr0fi3",
					PreviewEnvironmentID: previewEnvironmentID,
				})
				assert.NoError(collect, err)

				statusResponse, ok := getResponse.(cm.StatusCodeResponse)
				assert.True(collect, ok && statusResponse.GetStatusCode() == http.StatusNotFound)
			}, time.Minute, time.Second, "content preview platform %q still exists after cleanup", previewEnvironmentID)
		}
	})
}

//nolint:paralleltest
func TestAccPreviewEnvironmentResourceLifecycle(t *testing.T) {
	parallelWhenMocked(t)

	server, _ := cmt.NewContentfulManagementServer()
	server.RegisterSpaceEnvironment("0p38pssr0fi3", "master")

	name := "acctest_preview_" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")
	configVariables := config.Variables{
		"space_id": config.StringVariable("0p38pssr0fi3"),
		"name":     config.StringVariable(name),
	}

	var cleanupIDs []string
	registerLivePreviewEnvironmentCleanup(t, &cleanupIDs)

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_preview_environment.test", plancheck.ResourceActionCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					capturePreviewEnvironmentID(&cleanupIDs),
					resource.TestCheckResourceAttrSet("contentful_preview_environment.test", "preview_environment_id"),
					resource.TestCheckResourceAttr("contentful_preview_environment.test", "description", ""),
					resource.TestCheckResourceAttr("contentful_preview_environment.test", "content_type_configurations.%", "1"),
					resource.TestCheckResourceAttr(
						"contentful_preview_environment.test",
						"content_type_configurations.page.url",
						"https://preview.example.invalid/{env_id}/pages/{entry.sys.id}",
					),
				),
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_preview_environment.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					capturePreviewEnvironmentID(&cleanupIDs),
					resource.TestCheckResourceAttr("contentful_preview_environment.test", "name", name+" updated"),
					resource.TestCheckResourceAttr("contentful_preview_environment.test", "description", "updated description"),
					resource.TestCheckResourceAttr("contentful_preview_environment.test", "content_type_configurations.%", "1"),
					resource.TestCheckNoResourceAttr("contentful_preview_environment.test", "content_type_configurations.page"),
					resource.TestCheckResourceAttr(
						"contentful_preview_environment.test",
						"content_type_configurations.author.url",
						"https://preview.example.invalid/{env_id}/authors/{entry.sys.id}",
					),
				),
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_preview_environment.test", plancheck.ResourceActionNoop),
					},
				},
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_preview_environment.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					capturePreviewEnvironmentID(&cleanupIDs),
					resource.TestCheckResourceAttr("contentful_preview_environment.test", "content_type_configurations.%", "2"),
					resource.TestCheckResourceAttr(
						"contentful_preview_environment.test",
						"content_type_configurations.page.url",
						"https://preview.example.invalid/{env_id}/pages/{entry.sys.id}?replacement=true",
					),
				),
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_preview_environment.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.TestCheckResourceAttr(
					"contentful_preview_environment.test",
					"content_type_configurations.%",
					"0",
				),
			},
			{
				ConfigDirectory:   config.TestStepDirectory(),
				ConfigVariables:   configVariables,
				ResourceName:      "contentful_preview_environment.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

//nolint:paralleltest
func TestAccPreviewEnvironmentResourceSelectedID(t *testing.T) {
	parallelWhenMocked(t)

	server, _ := cmt.NewContentfulManagementServer()
	server.RegisterSpaceEnvironment("0p38pssr0fi3", "master")

	previewEnvironmentID := "acctest-preview-" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")
	configVariables := config.Variables{
		"space_id":               config.StringVariable("0p38pssr0fi3"),
		"name":                   config.StringVariable(previewEnvironmentID),
		"preview_environment_id": config.StringVariable(previewEnvironmentID),
		"include_page":           config.BoolVariable(true),
	}
	cleanupIDs := []string{previewEnvironmentID}
	registerLivePreviewEnvironmentCleanup(t, &cleanupIDs)

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("contentful_preview_environment.test", "preview_environment_id", previewEnvironmentID),
					resource.TestCheckResourceAttr("contentful_preview_environment.test", "content_type_configurations.%", "1"),
				),
			},
			{
				PreConfig: func() {
					configVariables["name"] = config.StringVariable(previewEnvironmentID + " updated")
					configVariables["include_page"] = config.BoolVariable(false)
				},
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_preview_environment.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.TestCheckResourceAttr(
					"contentful_preview_environment.test",
					"content_type_configurations.%",
					"0",
				),
			},
			{
				ConfigDirectory:   config.TestNameDirectory(),
				ConfigVariables:   configVariables,
				ResourceName:      "contentful_preview_environment.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

//nolint:paralleltest
func TestAccPreviewEnvironmentResourceRejectsEmptyContentTypeID(t *testing.T) {
	parallelWhenMocked(t)

	server, _ := cmt.NewContentfulManagementServer()
	server.RegisterSpaceEnvironment("0p38pssr0fi3", "master")

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      previewEnvironmentResourceConfig("Preview", "", ""),
				ExpectError: regexp.MustCompile(`content_type_configurations\[""\].*length must be at least 1`),
			},
		},
	})
}

func TestAccPreviewEnvironmentResourceMapOrderIsIgnored(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()
	server.RegisterSpaceEnvironment("0p38pssr0fi3", "master")

	resourceName := "acctest_preview_" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")
	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: previewEnvironmentResourceConfig(resourceName, "", "page", "author"),
			},
			{
				Config: previewEnvironmentResourceConfig(resourceName, "", "author", "page"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_preview_environment.test", plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

//nolint:paralleltest
func TestAccPreviewEnvironmentResourceOutOfBandDeletionRecreates(t *testing.T) {
	parallelWhenMocked(t)

	server, _ := cmt.NewContentfulManagementServer()
	server.RegisterSpaceEnvironment("0p38pssr0fi3", "master")

	resourceName := "acctest_preview_" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")
	testConfig := previewEnvironmentResourceConfig(resourceName, "", "page")

	var previewEnvironmentID string

	var cleanupIDs []string
	registerLivePreviewEnvironmentCleanup(t, &cleanupIDs)

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testConfig,
				Check: resource.ComposeTestCheckFunc(
					capturePreviewEnvironmentID(&cleanupIDs),
					resource.TestCheckResourceAttrWith(
						"contentful_preview_environment.test",
						"preview_environment_id",
						func(value string) error {
							previewEnvironmentID = value

							return nil
						},
					),
				),
			},
			{
				PreConfig: func() {
					deletePreviewEnvironmentOutOfBand(t, server, "0p38pssr0fi3", previewEnvironmentID)
				},
				Config: testConfig,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_preview_environment.test", plancheck.ResourceActionCreate),
					},
				},
				Check: capturePreviewEnvironmentID(&cleanupIDs),
			},
		},
	})
}

//nolint:paralleltest
func TestAccPreviewEnvironmentResourceStaleVersionConflict(t *testing.T) {
	parallelWhenMocked(t)

	server, _ := cmt.NewContentfulManagementServer()
	server.RegisterSpaceEnvironment("0p38pssr0fi3", "master")

	previewEnvironmentID := "acctest-preview-" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")
	cleanupIDs := []string{previewEnvironmentID}
	registerLivePreviewEnvironmentCleanup(t, &cleanupIDs)
	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{Config: previewEnvironmentResourceConfig(previewEnvironmentID, previewEnvironmentID, "page")},
			{
				Config: previewEnvironmentResourceConfig(previewEnvironmentID+" updated", previewEnvironmentID, "page"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_preview_environment.test", plancheck.ResourceActionUpdate),
						previewEnvironmentPlanCheckFunc(func(ctx context.Context, _ plancheck.CheckPlanRequest, _ *plancheck.CheckPlanResponse) {
							incrementPreviewEnvironmentVersionOutOfBand(ctx, t, server, "0p38pssr0fi3", previewEnvironmentID)
						}),
					},
				},
				ExpectError: regexp.MustCompile("Failed to update content preview platform"),
			},
		},
	})
}

func deletePreviewEnvironmentOutOfBand(t *testing.T, server *cmt.Server, spaceID, previewEnvironmentID string) {
	t.Helper()

	if os.Getenv("TF_ACC_MOCKED") != "" {
		server.DeletePreviewEnvironment(spaceID, previewEnvironmentID)

		return
	}

	client := livePreviewEnvironmentClient(t)
	response, err := client.DeletePreviewEnvironment(t.Context(), cm.DeletePreviewEnvironmentParams{
		SpaceID:              spaceID,
		PreviewEnvironmentID: previewEnvironmentID,
	})
	require.NoError(t, err)

	_, ok := response.(*cm.NoContent)
	require.True(t, ok, "unexpected out-of-band delete response: %T", response)
}

func incrementPreviewEnvironmentVersionOutOfBand(ctx context.Context, t *testing.T, server *cmt.Server, spaceID, previewEnvironmentID string) {
	t.Helper()

	if os.Getenv("TF_ACC_MOCKED") != "" {
		server.IncrementPreviewEnvironmentVersion(spaceID, previewEnvironmentID)

		return
	}

	client := livePreviewEnvironmentClient(t)
	getResponse, err := client.GetPreviewEnvironment(ctx, cm.GetPreviewEnvironmentParams{
		SpaceID:              spaceID,
		PreviewEnvironmentID: previewEnvironmentID,
	})
	require.NoError(t, err)

	previewEnvironment, ok := getResponse.(*cm.PreviewEnvironment)
	require.True(t, ok, "unexpected out-of-band read response: %T", getResponse)

	configurations := make([]cm.PreviewEnvironmentConfigurationData, 0, len(previewEnvironment.Configurations))
	for _, configuration := range previewEnvironment.Configurations {
		configurations = append(configurations, cm.PreviewEnvironmentConfigurationData{
			URL:        configuration.URL,
			EntityType: configuration.EntityType.Or("ContentType"),
			EntityId:   configuration.EntityId.Or(configuration.ContentType.Or("")),
			Enabled:    configuration.Enabled,
		})
	}

	updateResponse, err := client.PutPreviewEnvironment(ctx, &cm.PreviewEnvironmentData{
		Name:           previewEnvironment.Name + " out-of-band",
		Description:    previewEnvironment.Description,
		Configurations: configurations,
	}, cm.PutPreviewEnvironmentParams{
		SpaceID:              spaceID,
		PreviewEnvironmentID: previewEnvironmentID,
		XContentfulVersion:   previewEnvironment.Sys.Version,
	})
	require.NoError(t, err)

	_, ok = updateResponse.(*cm.PreviewEnvironment)
	require.True(t, ok, "unexpected out-of-band update response: %T", updateResponse)
}

func livePreviewEnvironmentClient(t *testing.T) *cm.Client {
	t.Helper()

	accessToken := os.Getenv("CONTENTFUL_MANAGEMENT_ACCESS_TOKEN")
	require.NotEmpty(t, accessToken, "CONTENTFUL_MANAGEMENT_ACCESS_TOKEN must be set for live acceptance tests")

	client, err := cm.NewClient(
		cm.DefaultServerURL,
		cm.NewAccessTokenSecuritySource(accessToken),
		cm.WithClient(cm.NewTransportClient(http.DefaultClient, "terraform-provider-contentful/acceptance-test")),
	)
	require.NoError(t, err)

	return client
}
