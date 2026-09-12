package provider_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
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

			waitForPreviewEnvironmentDeletion(ctx, t, client, "0p38pssr0fi3", previewEnvironmentID)
		}
	})
}

func waitForPreviewEnvironmentDeletion(ctx context.Context, t *testing.T, client *cm.Client, spaceID, previewEnvironmentID string) {
	t.Helper()

	require.EventuallyWithT(t, func(collect *assert.CollectT) {
		response, err := client.GetPreviewEnvironment(ctx, cm.GetPreviewEnvironmentParams{
			SpaceID:              spaceID,
			PreviewEnvironmentID: previewEnvironmentID,
		})
		assert.NoError(collect, err)

		statusResponse, ok := response.(cm.StatusCodeResponse)
		assert.True(collect, ok && statusResponse.GetStatusCode() == http.StatusNotFound)
	}, time.Minute, time.Second, "content preview platform %q still exists after deletion", previewEnvironmentID)
}

//nolint:paralleltest
func TestAccPreviewEnvironmentResourceLifecycle(t *testing.T) {
	parallelWhenMocked(t)

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("0p38pssr0fi3", "master")

	name := "acctest_preview_" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")
	configVariables := config.Variables{
		"space_id": config.StringVariable("0p38pssr0fi3"),
		"name":     config.StringVariable(name),
	}

	identity := statecheck.CompareValue(compare.ValuesSame())

	var cleanupIDs []string
	registerLivePreviewEnvironmentCleanup(t, &cleanupIDs)

	testAccMockableResource(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_preview_environment.test", plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity.AddStateValue("contentful_preview_environment.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("contentful_preview_environment.test", tfjsonpath.New("preview_environment_id"), knownvalue.StringRegexp(regexp.MustCompile(`(?s)^.+$`))),
					statecheck.ExpectKnownValue("contentful_preview_environment.test", tfjsonpath.New("description"), knownvalue.StringExact("")),
					statecheck.ExpectKnownValue("contentful_preview_environment.test", tfjsonpath.New("content_type_configurations"), knownvalue.MapExact(map[string]knownvalue.Check{
						"page": knownvalue.ObjectExact(map[string]knownvalue.Check{
							"url": knownvalue.StringExact("https://preview.example.invalid/{env_id}/pages/{entry.sys.id}"),
						}),
					})),
				},
				Check: capturePreviewEnvironmentID(&cleanupIDs),
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_preview_environment.test", plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity.AddStateValue("contentful_preview_environment.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("contentful_preview_environment.test", tfjsonpath.New("name"), knownvalue.StringExact(name+" updated")),
					statecheck.ExpectKnownValue("contentful_preview_environment.test", tfjsonpath.New("description"), knownvalue.StringExact("updated description")),
					statecheck.ExpectKnownValue("contentful_preview_environment.test", tfjsonpath.New("content_type_configurations"), knownvalue.MapExact(map[string]knownvalue.Check{
						"author": knownvalue.ObjectExact(map[string]knownvalue.Check{
							"url": knownvalue.StringExact("https://preview.example.invalid/{env_id}/authors/{entry.sys.id}"),
						}),
					})),
				},
				Check: capturePreviewEnvironmentID(&cleanupIDs),
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
				ConfigStateChecks: []statecheck.StateCheck{
					identity.AddStateValue("contentful_preview_environment.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("contentful_preview_environment.test", tfjsonpath.New("content_type_configurations"), knownvalue.MapSizeExact(2)),
					statecheck.ExpectKnownValue("contentful_preview_environment.test", tfjsonpath.New("content_type_configurations").AtMapKey("page").AtMapKey("url"), knownvalue.StringExact("https://preview.example.invalid/{env_id}/pages/{entry.sys.id}?replacement=true")),
				},
				Check: capturePreviewEnvironmentID(&cleanupIDs),
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_preview_environment.test", plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity.AddStateValue("contentful_preview_environment.test", tfjsonpath.New("id")),
					statecheck.ExpectKnownValue("contentful_preview_environment.test", tfjsonpath.New("content_type_configurations"), knownvalue.MapSizeExact(0)),
				},
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

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("0p38pssr0fi3", "master")

	previewEnvironmentID := "acctest-preview-" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")
	cleanupIDs := []string{previewEnvironmentID}
	registerLivePreviewEnvironmentCleanup(t, &cleanupIDs)

	testAccMockableResource(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"space_id":               config.StringVariable("0p38pssr0fi3"),
					"name":                   config.StringVariable(previewEnvironmentID),
					"preview_environment_id": config.StringVariable(previewEnvironmentID),
					"include_page":           config.BoolVariable(true),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("contentful_preview_environment.test", tfjsonpath.New("preview_environment_id"), knownvalue.StringExact(previewEnvironmentID)),
					statecheck.ExpectKnownValue("contentful_preview_environment.test", tfjsonpath.New("content_type_configurations"), knownvalue.MapSizeExact(1)),
				},
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"space_id":               config.StringVariable("0p38pssr0fi3"),
					"name":                   config.StringVariable(previewEnvironmentID + " updated"),
					"preview_environment_id": config.StringVariable(previewEnvironmentID),
					"include_page":           config.BoolVariable(false),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_preview_environment.test", plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("contentful_preview_environment.test", tfjsonpath.New("content_type_configurations"), knownvalue.MapSizeExact(0)),
				},
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"space_id":               config.StringVariable("0p38pssr0fi3"),
					"name":                   config.StringVariable(previewEnvironmentID + " updated"),
					"preview_environment_id": config.StringVariable(previewEnvironmentID),
					"include_page":           config.BoolVariable(false),
				},
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

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("0p38pssr0fi3", "master")

	testAccMockableResource(t, server, resource.TestCase{
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

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("0p38pssr0fi3", "master")

	resourceName := "acctest_preview_" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")
	testAccMockedResource(t, server, resource.TestCase{
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

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("0p38pssr0fi3", "master")

	resourceName := "acctest_preview_" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")
	testConfig := previewEnvironmentResourceConfig(resourceName, "", "page")

	var previewEnvironmentID string

	var cleanupIDs []string
	registerLivePreviewEnvironmentCleanup(t, &cleanupIDs)

	testAccMockableResource(t, server, resource.TestCase{
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

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("0p38pssr0fi3", "master")

	previewEnvironmentID := "acctest-preview-" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")
	cleanupIDs := []string{previewEnvironmentID}
	registerLivePreviewEnvironmentCleanup(t, &cleanupIDs)
	testAccMockableResource(t, server, resource.TestCase{
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

	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()

	client := livePreviewEnvironmentClient(t)
	response, err := client.DeletePreviewEnvironment(ctx, cm.DeletePreviewEnvironmentParams{
		SpaceID:              spaceID,
		PreviewEnvironmentID: previewEnvironmentID,
	})
	require.NoError(t, err)

	_, ok := response.(*cm.NoContent)
	require.True(t, ok, "unexpected out-of-band delete response: %T", response)

	waitForPreviewEnvironmentDeletion(ctx, t, client, spaceID, previewEnvironmentID)
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

func TestAccPreviewEnvironmentResourceGeneratedIDReplacement(t *testing.T) {
	t.Parallel()

	for _, createBeforeDestroy := range []bool{false, true} {
		t.Run(fmt.Sprintf("create_before_destroy=%t", createBeforeDestroy), func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
			require.NoError(t, err)
			server.RegisterSpaceEnvironment("0p38pssr0fi3", "master")

			configuration := fmt.Sprintf(`
resource "contentful_preview_environment" "test" {
  space_id = "0p38pssr0fi3"
  name = "Preview"
  content_type_configurations = {
    page = { url = "https://preview.invalid/page" }
  }
  lifecycle {
    create_before_destroy = %t
  }
}
`, createBeforeDestroy)

			var ids []string
			testAccMockedResource(t, server, resource.TestCase{
				Steps: []resource.TestStep{
					{Config: configuration, Check: capturePreviewEnvironmentID(&ids)},
					{
						Config: configuration,
						Taint:  []string{"contentful_preview_environment.test"},
						Check:  capturePreviewEnvironmentID(&ids),
					},
				},
			})
			require.Len(t, ids, 2)
			assert.NotEqual(t, ids[0], ids[1])
		})
	}
}

func TestAccPreviewEnvironmentResourceVersionedConfigurationDelta(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("0p38pssr0fi3", "master")

	var (
		requestMutex sync.Mutex
		mutations    []string
		bodies       []string
		handlerErr   error
	)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			body, readErr := io.ReadAll(r.Body)

			requestMutex.Lock()
			if readErr != nil {
				handlerErr = readErr
			}

			mutations = append(mutations, r.Method+" "+r.URL.Path+" version="+r.Header.Get("X-Contentful-Version"))
			bodies = append(bodies, string(body))
			requestMutex.Unlock()

			r.Body = io.NopCloser(bytes.NewReader(body))
		}

		server.ServeHTTP(w, r)
	})

	createdConfig := previewEnvironmentResourceConfig("Preview", "preview", "page")
	renamedConfig := previewEnvironmentResourceConfig("Renamed", "preview", "page")
	changedURLConfig := strings.ReplaceAll(renamedConfig, "https://preview.example.invalid/page/{entry.sys.id}", "https://preview.invalid/changed")

	testAccMockedResource(t, handler, resource.TestCase{
		Steps: []resource.TestStep{
			{Config: createdConfig},
			{
				Config: renamedConfig,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("contentful_preview_environment.test", tfjsonpath.New("name"), knownvalue.StringExact("Renamed")),
				},
			},
			{
				Config: changedURLConfig,
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("contentful_preview_environment.test", tfjsonpath.New("content_type_configurations"), knownvalue.MapExact(map[string]knownvalue.Check{
						"page": knownvalue.ObjectExact(map[string]knownvalue.Check{"url": knownvalue.StringExact("https://preview.invalid/changed")}),
					})),
				},
			},
			{
				Config: previewEnvironmentResourceConfig("Renamed", "preview"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("contentful_preview_environment.test", tfjsonpath.New("content_type_configurations"), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
		},
	})

	requestMutex.Lock()
	defer requestMutex.Unlock()

	require.NoError(t, handlerErr)
	require.Equal(t, []string{
		"PUT /spaces/0p38pssr0fi3/preview_environments/preview version=0",
		"PUT /spaces/0p38pssr0fi3/preview_environments/preview version=0",
		"PUT /spaces/0p38pssr0fi3/preview_environments/preview version=1",
		"PUT /spaces/0p38pssr0fi3/preview_environments/preview version=1",
		"DELETE /spaces/0p38pssr0fi3/preview_environments/preview version=",
	}, mutations)
	require.Len(t, bodies, 5)
	assert.JSONEq(t, `{"name":"Preview","description":"","configurations":[{"entityType":"ContentType","entityId":"page","url":"https://preview.example.invalid/page/{entry.sys.id}","enabled":true}]}`, bodies[0])
	assert.JSONEq(t, `{"name":"Renamed","description":"","configurations":[]}`, bodies[1])
	assert.JSONEq(t, `{"name":"Renamed","description":"","configurations":[{"entityType":"ContentType","entityId":"page","url":"https://preview.invalid/changed","enabled":true}]}`, bodies[2])
	assert.JSONEq(t, `{"name":"Renamed","description":"","configurations":[{"entityType":"ContentType","entityId":"page","url":"https://preview.invalid/changed","enabled":false}]}`, bodies[3])
	assert.Empty(t, bodies[4])
}
