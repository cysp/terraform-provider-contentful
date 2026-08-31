package provider_test

import (
	"maps"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/stretchr/testify/require"
)

type editorInterfaceRequestCountingHandler struct {
	next http.Handler

	gets atomic.Int64
	puts atomic.Int64

	mu       sync.Mutex
	requests []string
	versions []string
}

func (h *editorInterfaceRequestCountingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/editor_interface") {
		h.mu.Lock()

		h.requests = append(h.requests, r.Method)
		if r.Method == http.MethodPut {
			h.versions = append(h.versions, r.Header.Get("X-Contentful-Version"))
		}
		h.mu.Unlock()

		switch r.Method {
		case http.MethodGet:
			h.gets.Add(1)
		case http.MethodPut:
			h.puts.Add(1)
		}
	}

	h.next.ServeHTTP(w, r)
}

func (h *editorInterfaceRequestCountingHandler) Requests() []string {
	h.mu.Lock()
	defer h.mu.Unlock()

	return append([]string(nil), h.requests...)
}

func (h *editorInterfaceRequestCountingHandler) PutVersions() []string {
	h.mu.Lock()
	defer h.mu.Unlock()

	return append([]string(nil), h.versions...)
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
				ExpectError:     regexp.MustCompile(`Resource Import Passthrough Multipart ID Mismatch`),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ResourceName:    "contentful_editor_interface.test",
				ImportState:     true,
				ImportStateId:   "a/b",
				ExpectError:     regexp.MustCompile(`Resource Import Passthrough Multipart ID Mismatch`),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ResourceName:    "contentful_editor_interface.test",
				ImportState:     true,
				ImportStateId:   "a/b/c/d",
				ExpectError:     regexp.MustCompile(`Resource Import Passthrough Multipart ID Mismatch`),
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

func TestAccEditorInterfaceResourceCreateRequiresImportForModifiedInterface(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)

	server.SetContentType("0p38pssr0fi3", "test", "author", cm.ContentTypeRequestData{Name: "Author"})
	server.SetEditorInterface("0p38pssr0fi3", "test", "author", cm.EditorInterfaceData{})

	response, err := server.Handler().PutEditorInterface(t.Context(), &cm.EditorInterfaceData{}, cm.PutEditorInterfaceParams{
		SpaceID:            "0p38pssr0fi3",
		EnvironmentID:      "test",
		ContentTypeID:      "author",
		XContentfulVersion: 1,
	})
	require.NoError(t, err)

	responseStatus, ok := response.(*cm.EditorInterfaceStatusCode)
	require.True(t, ok)
	require.Equal(t, 2, responseStatus.Response.Sys.Version)

	handler := &editorInterfaceRequestCountingHandler{next: server}
	configVariables := config.Variables{
		"space_id":        config.StringVariable("0p38pssr0fi3"),
		"environment_id":  config.StringVariable("test"),
		"content_type_id": config.StringVariable("author"),
	}

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ExpectError:     regexp.MustCompile(`Editor interface must be imported`),
			},
		},
	})

	require.Equal(t, int64(0), handler.gets.Load())
	require.Equal(t, int64(1), handler.puts.Load())
	require.Equal(t, []string{http.MethodPut}, handler.Requests())
	require.Equal(t, []string{"1"}, handler.PutVersions())
}

func TestAccEditorInterfaceResourceCreateManagesInitialInterface(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)

	server.SetContentType("0p38pssr0fi3", "test", "author", cm.ContentTypeRequestData{Name: "Author"})
	server.SetEditorInterface("0p38pssr0fi3", "test", "author", cm.EditorInterfaceData{})

	handler := &editorInterfaceRequestCountingHandler{next: server}
	configVariables := config.Variables{
		"space_id":        config.StringVariable("0p38pssr0fi3"),
		"environment_id":  config.StringVariable("test"),
		"content_type_id": config.StringVariable("author"),
	}

	ContentfulProviderMockedResourceTest(t, handler, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
			},
		},
	})

	requests := handler.Requests()
	require.NotEmpty(t, requests)
	require.Equal(t, http.MethodPut, requests[0])
	require.Equal(t, int64(1), handler.puts.Load())
	require.Equal(t, []string{"1"}, handler.PutVersions())
}

func TestAccEditorInterfaceResourceCreateAfterContentTypeUpdateUsesObservedVersion(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("0p38pssr0fi3", "test")

	handler := &editorInterfaceRequestCountingHandler{next: server}
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

	requests := handler.Requests()
	require.NotEmpty(t, requests)
	require.Equal(t, http.MethodPut, requests[0])
	require.Equal(t, int64(1), handler.puts.Load())
	require.Equal(t, []string{"2"}, handler.PutVersions())
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
