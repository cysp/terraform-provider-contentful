package provider_test

import (
	"regexp"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmts "github.com/cysp/terraform-provider-contentful/internal/contentful-management-testserver"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

//nolint:paralleltest
func TestAccEditorInterfaceResourceImport(t *testing.T) {
	testserver := cmts.NewContentfulManagementTestServer()
	defer testserver.Server().Close()

	configVariables := config.Variables{
		"space_id":        config.StringVariable("0p38pssr0fi3"),
		"environment_id":  config.StringVariable("test"),
		"content_type_id": config.StringVariable("author"),
	}

	testserver.SetContentType("0p38pssr0fi3", "test", "author", cm.ContentTypeRequestFields{
		Name: "Author",
	})

	testserver.SetEditorInterface("0p38pssr0fi3", "test", "author", cm.EditorInterfaceFields{})

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
	testserver := cmts.NewContentfulManagementTestServer()
	defer testserver.Server().Close()

	configVariables := config.Variables{
		"space_id":        config.StringVariable("0p38pssr0fi3"),
		"environment_id":  config.StringVariable("test"),
		"content_type_id": config.StringVariable("nonexistent"),
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
	testserver := cmts.NewContentfulManagementTestServer()
	defer testserver.Server().Close()

	configVariables := config.Variables{
		"space_id":        config.StringVariable("0p38pssr0fi3"),
		"environment_id":  config.StringVariable("nonexistent"),
		"content_type_id": config.StringVariable("nonexistent"),
	}

	ContentfulProviderMockableResourceTest(t, testserver.Server(), resource.TestCase{
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
	testserver := cmts.NewContentfulManagementTestServer()
	defer testserver.Server().Close()

	configVariables := config.Variables{
		"space_id":        config.StringVariable("0p38pssr0fi3"),
		"environment_id":  config.StringVariable("test"),
		"content_type_id": config.StringVariable("nonexistent"),
	}

	ContentfulProviderMockableResourceTest(t, testserver.Server(), resource.TestCase{
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
func TestAccEditorInterfaceResourceUpdate(t *testing.T) {
	testserver := cmts.NewContentfulManagementTestServer()
	defer testserver.Server().Close()

	configVariables := config.Variables{
		"space_id":        config.StringVariable("0p38pssr0fi3"),
		"environment_id":  config.StringVariable("test"),
		"content_type_id": config.StringVariable("author"),
	}

	testserver.SetContentType("0p38pssr0fi3", "test", "author", cm.ContentTypeRequestFields{
		Name: "Author",
	})

	testserver.SetEditorInterface("0p38pssr0fi3", "test", "author", cm.EditorInterfaceFields{})

	ContentfulProviderMockableResourceTest(t, testserver.Server(), resource.TestCase{
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
