package provider_test

import (
	"maps"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/stretchr/testify/require"
)

type editorInterfaceRequestRecorder struct {
	next http.Handler

	mu       sync.Mutex
	requests []string
}

func (h *editorInterfaceRequestRecorder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/editor_interface") {
		request := r.Method
		if r.Method == http.MethodPut {
			request += ":" + r.Header.Get("X-Contentful-Version")
		}

		h.mu.Lock()
		h.requests = append(h.requests, request)
		h.mu.Unlock()
	}

	h.next.ServeHTTP(w, r)
}

const editorInterfaceInitialConfig = `
resource "contentful_editor_interface" "test" {
  space_id        = "0p38pssr0fi3"
  environment_id  = "test"
  content_type_id = "author"
  controls        = []
}
`

func (h *editorInterfaceRequestRecorder) Requests() []string {
	h.mu.Lock()
	defer h.mu.Unlock()

	return append([]string(nil), h.requests...)
}

//nolint:paralleltest
func TestAccEditorInterfaceResourceImport(t *testing.T) {
	parallelWhenMocked(t)

	server, _ := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))

	configVariables := config.Variables{
		"space_id":        config.StringVariable("0p38pssr0fi3"),
		"environment_id":  config.StringVariable("test"),
		"content_type_id": config.StringVariable("author"),
	}

	server.SetContentType("0p38pssr0fi3", "test", "author", cm.ContentTypeRequestData{
		Name: "Author",
	})

	server.SetEditorInterface("0p38pssr0fi3", "test", "author", cm.EditorInterfaceData{})

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory:    config.TestNameDirectory(),
				ConfigVariables:    configVariables,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ResourceName:    "contentful_editor_interface.test",
				ImportState:     true,
				ImportStateId:   "a",
				ExpectError:     regexp.MustCompile(`Invalid import ID`),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ResourceName:    "contentful_editor_interface.test",
				ImportState:     true,
				ImportStateId:   "a/b",
				ExpectError:     regexp.MustCompile(`Invalid import ID`),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ResourceName:    "contentful_editor_interface.test",
				ImportState:     true,
				ImportStateId:   "a/b/c/d",
				ExpectError:     regexp.MustCompile(`Invalid import ID`),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ResourceName:    "contentful_editor_interface.test",
				ImportState:     true,
				ImportStateId:   "0p38pssr0fi3/test/author",
				// ImportStateVerify: true,
			},
		},
	})
}

//nolint:paralleltest
func TestAccEditorInterfaceResourceImportNotFound(t *testing.T) {
	parallelWhenMocked(t)

	server, _ := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))

	configVariables := config.Variables{
		"space_id":        config.StringVariable("0p38pssr0fi3"),
		"environment_id":  config.StringVariable("test"),
		"content_type_id": config.StringVariable("nonexistent"),
	}

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory:    config.TestNameDirectory(),
				ConfigVariables:    configVariables,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ResourceName:    "contentful_editor_interface.test",
				ImportState:     true,
				ImportStateId:   "0p38pssr0fi3/test/nonexistent",
				ExpectError:     regexp.MustCompile(`Cannot import non-existent remote object`),
			},
		},
	})
}

//nolint:paralleltest
func TestAccEditorInterfaceResourceCreateNotFoundEnvironment(t *testing.T) {
	parallelWhenMocked(t)

	server, _ := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))

	configVariables := config.Variables{
		"space_id":        config.StringVariable("0p38pssr0fi3"),
		"environment_id":  config.StringVariable("nonexistent"),
		"content_type_id": config.StringVariable("nonexistent"),
	}

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ExpectError:     regexp.MustCompile(`Failed to create editor interface`),
			},
		},
	})
}

//nolint:paralleltest
func TestAccEditorInterfaceResourceCreateNotFoundContentType(t *testing.T) {
	parallelWhenMocked(t)

	server, _ := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))

	configVariables := config.Variables{
		"space_id":        config.StringVariable("0p38pssr0fi3"),
		"environment_id":  config.StringVariable("test"),
		"content_type_id": config.StringVariable("nonexistent"),
	}

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ExpectError:     regexp.MustCompile(`Failed to create editor interface`),
			},
		},
	})
}

func TestAccEditorInterfaceResourceCreateManagesInitialInterface(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)

	server.SetContentType("0p38pssr0fi3", "test", "author", cm.ContentTypeRequestData{Name: "Author"})
	server.SetEditorInterface("0p38pssr0fi3", "test", "author", cm.EditorInterfaceData{})

	handler := &editorInterfaceRequestRecorder{next: server}
	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: editorInterfaceInitialConfig,
			},
		},
	})

	require.Equal(t, []string{"PUT:1", "GET"}, handler.Requests())
}

func TestAccEditorInterfaceResourceCreateAfterContentTypeUpdateUsesObservedVersion(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("0p38pssr0fi3", "test")

	handler := &editorInterfaceRequestRecorder{next: server}
	configVariables := config.Variables{
		"space_id":        config.StringVariable("0p38pssr0fi3"),
		"environment_id":  config.StringVariable("test"),
		"content_type_id": config.StringVariable("author"),
	}

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{
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

	require.Equal(t, []string{"PUT:2", "GET"}, handler.Requests())
}

//nolint:paralleltest
func TestAccEditorInterfaceResourceUpdate(t *testing.T) {
	parallelWhenMocked(t)

	server, _ := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))

	configVariables := config.Variables{
		"space_id":        config.StringVariable("0p38pssr0fi3"),
		"environment_id":  config.StringVariable("test"),
		"content_type_id": config.StringVariable("author"),
	}

	server.SetContentType("0p38pssr0fi3", "test", "author", cm.ContentTypeRequestData{
		Name: "Author",
	})

	server.SetEditorInterface("0p38pssr0fi3", "test", "author", cm.EditorInterfaceData{})

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory:    config.TestStepDirectory(),
				ConfigVariables:    configVariables,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				ConfigDirectory:    config.TestStepDirectory(),
				ConfigVariables:    configVariables,
				ResourceName:       "contentful_editor_interface.test",
				ImportState:        true,
				ImportStateId:      "0p38pssr0fi3/test/author",
				ImportStatePersist: true,
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_editor_interface.test", plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_editor_interface.test", plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_editor_interface.test", plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

//nolint:paralleltest
func TestAccEditorInterfaceResourceUpdateWithContentType(t *testing.T) {
	parallelWhenMocked(t)

	server, _ := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))

	server.RegisterSpaceEnvironment("0p38pssr0fi3", "test")

	contentTypeID := "acctest_" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	configVariables := config.Variables{
		"space_id":        config.StringVariable("0p38pssr0fi3"),
		"environment_id":  config.StringVariable("test"),
		"content_type_id": config.StringVariable(contentTypeID),
	}

	configVariables1 := maps.Clone(configVariables)

	configVariables2 := maps.Clone(configVariables)
	configVariables2["content_type_additional_fields"] = config.ListVariable(
		config.StringVariable("a"),
	)

	configVariables3 := maps.Clone(configVariables)
	configVariables3["content_type_additional_fields"] = config.ListVariable(
		config.StringVariable("a"),
		config.StringVariable("b"),
	)

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables1,
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables2,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionUpdate),
						plancheck.ExpectResourceAction("contentful_editor_interface.test", plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables3,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionUpdate),
						plancheck.ExpectResourceAction("contentful_editor_interface.test", plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccEditorInterfaceResourceUpdateWithContentTypeMultipleSpaceEnvironments(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))

	server.RegisterSpaceEnvironment("space-a", "environment-a-a")
	server.RegisterSpaceEnvironment("space-a", "environment-a-b")
	server.RegisterSpaceEnvironment("space-b", "environment-b-a")

	contentTypeID := "acctest_" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	configVariables := config.Variables{
		"content_type_id": config.StringVariable(contentTypeID),
	}

	configVariables1 := maps.Clone(configVariables)

	configVariables2 := maps.Clone(configVariables)
	configVariables2["content_type_additional_fields"] = config.ListVariable(
		config.StringVariable("a"),
	)

	configVariables3 := maps.Clone(configVariables)
	configVariables3["content_type_additional_fields"] = config.ListVariable(
		config.StringVariable("a"),
		config.StringVariable("b"),
	)

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables1,
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables2,
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables3,
			},
		},
	})
}
