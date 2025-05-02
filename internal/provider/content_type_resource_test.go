package provider_test

import (
	"regexp"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmts "github.com/cysp/terraform-provider-contentful/internal/contentful-management-testserver"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestAccContentTypeResourceImport(t *testing.T) {
	t.Parallel()

	testserver := cmts.NewContentfulManagementTestServer()
	defer testserver.Server().Close()

	configVariables := config.Variables{
		"space_id":       config.StringVariable("0p38pssr0fi3"),
		"environment_id": config.StringVariable("test"),
	}

	testserver.SetContentType("0p38pssr0fi3", "test", "author", cm.ContentTypeRequestFields{
		Name: "Author",
	})

	ContentfulProviderMockableResourceTest(t, testserver.Server(), resource.TestCase{
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
				ResourceName:    "contentful_content_type.author",
				ImportState:     true,
				ImportStateId:   "a",
				ExpectError:     regexp.MustCompile(`Resource Import Passthrough Multipart ID Mismatch`),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ResourceName:    "contentful_content_type.author",
				ImportState:     true,
				ImportStateId:   "a/b",
				ExpectError:     regexp.MustCompile(`Resource Import Passthrough Multipart ID Mismatch`),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ResourceName:    "contentful_content_type.author",
				ImportState:     true,
				ImportStateId:   "a/b/c/d",
				ExpectError:     regexp.MustCompile(`Resource Import Passthrough Multipart ID Mismatch`),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ResourceName:    "contentful_content_type.author",
				ImportState:     true,
				ImportStateId:   "0p38pssr0fi3/test/author",
				// ImportStateVerify: true,
			},
		},
	})
}

func TestAccContentTypeResourceImportNotFound(t *testing.T) {
	t.Parallel()

	testserver := cmts.NewContentfulManagementTestServer()
	defer testserver.Server().Close()

	configVariables := config.Variables{
		"space_id":       config.StringVariable("0p38pssr0fi3"),
		"environment_id": config.StringVariable("test"),
	}

	ContentfulProviderMockableResourceTest(t, testserver.Server(), resource.TestCase{
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
				ResourceName:    "contentful_content_type.test",
				ImportState:     true,
				ImportStateId:   "0p38pssr0fi3/test/nonexistent",
				ExpectError:     regexp.MustCompile(`Cannot import non-existent remote object`),
			},
		},
	})
}

func TestAccContentTypeResourceCreateNotFoundEnvironment(t *testing.T) {
	t.Parallel()

	testserver := cmts.NewContentfulManagementTestServer()
	defer testserver.Server().Close()

	configVariables := config.Variables{
		"space_id":       config.StringVariable("0p38pssr0fi3"),
		"environment_id": config.StringVariable("nonexistent"),
	}

	ContentfulProviderMockableResourceTest(t, testserver.Server(), resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
				ExpectError:     regexp.MustCompile(`Failed to create content type`),
			},
		},
	})
}

func TestAccContentTypeResourceCreate(t *testing.T) {
	t.Parallel()

	testserver := cmts.NewContentfulManagementTestServer()
	defer testserver.Server().Close()

	contentTypeID := "acctest_" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	configVariables := config.Variables{
		"space_id":             config.StringVariable("0p38pssr0fi3"),
		"environment_id":       config.StringVariable("test"),
		"test_content_type_id": config.StringVariable(contentTypeID),
	}

	ContentfulProviderMockableResourceTest(t, testserver.Server(), resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
			},
		},
	})
}

func TestAccContentTypeResourceUpdate(t *testing.T) {
	t.Parallel()

	testserver := cmts.NewContentfulManagementTestServer()
	defer testserver.Server().Close()

	contentTypeID := "acctest_" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	configVariables := config.Variables{
		"space_id":             config.StringVariable("0p38pssr0fi3"),
		"environment_id":       config.StringVariable("test"),
		"test_content_type_id": config.StringVariable(contentTypeID),
	}

	ContentfulProviderMockableResourceTest(t, testserver.Server(), resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionCreate),
						plancheck.ExpectResourceAction("contentful_editor_interface.test", plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionUpdate),
						plancheck.ExpectResourceAction("contentful_editor_interface.test", plancheck.ResourceActionNoop),
					},
				},
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction("contentful_editor_interface.test", plancheck.ResourceActionNoop),
					},
				},
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionUpdate),
						plancheck.ExpectResourceAction("contentful_editor_interface.test", plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccContentTypeResourceUpdateResourceLinks(t *testing.T) {
	t.Parallel()

	testserver := cmts.NewContentfulManagementTestServer()
	defer testserver.Server().Close()

	contentTypeID := "acctest_" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	configVariables := config.Variables{
		"space_id":             config.StringVariable("0p38pssr0fi3"),
		"environment_id":       config.StringVariable("test"),
		"test_content_type_id": config.StringVariable(contentTypeID),
	}

	ContentfulProviderMockedResourceTest(t, testserver.Server(), resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionCreate),
						plancheck.ExpectResourceAction("contentful_editor_interface.test", plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionUpdate),
						plancheck.ExpectResourceAction("contentful_editor_interface.test", plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccContentTypeResourceDeleted(t *testing.T) {
	t.Parallel()

	testserver := cmts.NewContentfulManagementTestServer()
	defer testserver.Server().Close()

	contentTypeID := "acctest_" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	configVariables := config.Variables{
		"space_id":             config.StringVariable("0p38pssr0fi3"),
		"environment_id":       config.StringVariable("test"),
		"test_content_type_id": config.StringVariable(contentTypeID),
	}

	ContentfulProviderMockableResourceTest(t, testserver.Server(), resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
			},
			{
				ConfigDirectory:    config.TestStepDirectory(),
				ConfigVariables:    configVariables,
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}
