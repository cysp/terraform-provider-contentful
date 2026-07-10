package provider_test

import (
	"context"
	"regexp"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

func TestAccContentTypeResourceImport(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	configVariables := config.Variables{
		"space_id":       config.StringVariable("0p38pssr0fi3"),
		"environment_id": config.StringVariable("test"),
	}

	server.SetContentType("0p38pssr0fi3", "test", "author", cm.ContentTypeRequestData{
		Name:   "Author",
		Fields: []cm.ContentTypeRequestDataFieldsItem{},
	})

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

func TestAccContentTypeResourceImportWithTaxonomy(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	server.SetContentType("0p38pssr0fi3", "test", "author", cm.ContentTypeRequestData{
		Name:         "Author",
		Description:  cm.NewOptNilString("An author"),
		DisplayField: "name",
		Fields:       []cm.ContentTypeRequestDataFieldsItem{},
		Metadata: cm.NewOptContentTypeMetadata(cm.ContentTypeMetadata{
			Taxonomy: []cm.ContentTypeMetadataTaxonomyItem{
				{
					Sys: cm.ContentTypeMetadataTaxonomyItemSys{
						Type:     cm.ContentTypeMetadataTaxonomyItemSysTypeLink,
						ID:       "furniture",
						LinkType: cm.ContentTypeMetadataTaxonomyItemSysLinkTypeTaxonomyConceptScheme,
					},
				},
				{
					Sys: cm.ContentTypeMetadataTaxonomyItemSys{
						Type:     cm.ContentTypeMetadataTaxonomyItemSysTypeLink,
						ID:       "livingRoomFurniture",
						LinkType: cm.ContentTypeMetadataTaxonomyItemSysLinkTypeTaxonomyConcept,
					},
				},
			},
		}),
	})

	configVariables := config.Variables{
		"space_id":       config.StringVariable("0p38pssr0fi3"),
		"environment_id": config.StringVariable("test"),
	}

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory:    config.StaticDirectory("testdata/TestAccContentTypeResourceImport"),
				ConfigVariables:    configVariables,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				ConfigDirectory:    config.StaticDirectory("testdata/TestAccContentTypeResourceImport"),
				ConfigVariables:    configVariables,
				ResourceName:       "contentful_content_type.author",
				ImportState:        true,
				ImportStateId:      "0p38pssr0fi3/test/author",
				ImportStatePersist: true,
				ConfigStateChecks:  contentTypeTaxonomyStateChecks("contentful_content_type.author"),
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceImport"),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.author", plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: contentTypeTaxonomyStateChecks("contentful_content_type.author"),
			},
		},
	})
}

func TestAccContentTypeResourceImportNotFound(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	configVariables := config.Variables{
		"space_id":       config.StringVariable("0p38pssr0fi3"),
		"environment_id": config.StringVariable("test"),
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

	server, _ := cmt.NewContentfulManagementServer()

	configVariables := config.Variables{
		"space_id":       config.StringVariable("0p38pssr0fi3"),
		"environment_id": config.StringVariable("nonexistent"),
	}

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
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

	server, _ := cmt.NewContentfulManagementServer()

	server.RegisterSpaceEnvironment("0p38pssr0fi3", "test")

	contentTypeID := "acctest_" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	configVariables := config.Variables{
		"space_id":             config.StringVariable("0p38pssr0fi3"),
		"environment_id":       config.StringVariable("test"),
		"test_content_type_id": config.StringVariable(contentTypeID),
	}

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
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

	server, _ := cmt.NewContentfulManagementServer()

	server.RegisterSpaceEnvironment("0p38pssr0fi3", "test")

	contentTypeID := "acctest_" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	configVariables := config.Variables{
		"space_id":             config.StringVariable("0p38pssr0fi3"),
		"environment_id":       config.StringVariable("test"),
		"test_content_type_id": config.StringVariable(contentTypeID),
	}

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
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

//nolint:paralleltest
func TestAccContentTypeResourceUpdateMetadata(t *testing.T) {
	parallelWhenMocked(t)

	server, _ := cmt.NewContentfulManagementServer()

	server.RegisterSpaceEnvironment("0p38pssr0fi3", "test")

	contentTypeID := "acctest_" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	configVariables := config.Variables{
		"space_id":             config.StringVariable("0p38pssr0fi3"),
		"environment_id":       config.StringVariable("test"),
		"test_content_type_id": config.StringVariable(contentTypeID),
	}

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdateMetadata/annotations_only"),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: contentTypeTaxonomyStateChecks("contentful_content_type.test"),
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdateMetadata/1"),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: contentTypeMetadataWithoutAnnotationsStateChecks(),
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdateMetadata/1"),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: contentTypeMetadataWithoutAnnotationsStateChecks(),
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdateMetadata/3"),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdateMetadata/3"),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func contentTypeTaxonomyStateChecks(resourceAddress string) []statecheck.StateCheck {
	return []statecheck.StateCheck{
		statecheck.ExpectKnownValue(
			resourceAddress,
			tfjsonpath.New("metadata").AtMapKey("taxonomy").AtSliceIndex(0).AtMapKey("taxonomy_concept_scheme").AtMapKey("id"),
			knownvalue.StringExact("furniture"),
		),
		statecheck.ExpectKnownValue(
			resourceAddress,
			tfjsonpath.New("metadata").AtMapKey("taxonomy").AtSliceIndex(1).AtMapKey("taxonomy_concept").AtMapKey("id"),
			knownvalue.StringExact("livingRoomFurniture"),
		),
	}
}

func contentTypeMetadataWithoutAnnotationsStateChecks() []statecheck.StateCheck {
	return append(contentTypeTaxonomyStateChecks("contentful_content_type.test"), statecheck.ExpectKnownValue(
		"contentful_content_type.test",
		tfjsonpath.New("metadata").AtMapKey("annotations"),
		knownvalue.Null(),
	))
}

func TestAccContentTypeResourceTaxonomyDrift(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()
	server.RegisterSpaceEnvironment("0p38pssr0fi3", "test")

	contentTypeID := "acctest_" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")
	configVariables := config.Variables{
		"space_id":             config.StringVariable("0p38pssr0fi3"),
		"environment_id":       config.StringVariable("test"),
		"test_content_type_id": config.StringVariable(contentTypeID),
	}

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdateMetadata/1"),
				ConfigVariables: configVariables,
			},
			{
				PreConfig: func() {
					setContentTypeTaxonomyDrift(t, server, contentTypeID, []cm.ContentTypeMetadataTaxonomyItem{
						contentTypeTaxonomyConceptScheme("furniture"),
					})
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdateMetadata/1"),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"contentful_content_type.test",
						tfjsonpath.New("metadata").AtMapKey("taxonomy"),
						knownvalue.ListSizeExact(1),
					),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdateMetadata/2"),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: contentTypeTaxonomyStateChecks("contentful_content_type.test"),
			},
			{
				PreConfig: func() {
					setContentTypeTaxonomyDrift(t, server, contentTypeID, []cm.ContentTypeMetadataTaxonomyItem{})
				},
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdateMetadata/2"),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: contentTypeTaxonomyStateChecks("contentful_content_type.test"),
			},
		},
	})
}

func setContentTypeTaxonomyDrift(
	t *testing.T,
	server *cmt.Server,
	contentTypeID string,
	taxonomy []cm.ContentTypeMetadataTaxonomyItem,
) {
	t.Helper()

	response, err := server.Handler().GetContentType(context.Background(), cm.GetContentTypeParams{
		SpaceID: "0p38pssr0fi3", EnvironmentID: "test", ContentTypeID: contentTypeID,
	})
	require.NoError(t, err)

	contentType, ok := response.(*cm.ContentType)
	require.True(t, ok)

	metadata := contentType.Metadata.Or(cm.ContentTypeMetadata{})
	metadata.Taxonomy = taxonomy
	contentType.Metadata.SetTo(metadata)
	contentType.Sys.Version++
}

func contentTypeTaxonomyConceptScheme(id string) cm.ContentTypeMetadataTaxonomyItem {
	return cm.ContentTypeMetadataTaxonomyItem{
		Sys: cm.ContentTypeMetadataTaxonomyItemSys{
			Type:     cm.ContentTypeMetadataTaxonomyItemSysTypeLink,
			ID:       id,
			LinkType: cm.ContentTypeMetadataTaxonomyItemSysLinkTypeTaxonomyConceptScheme,
		},
	}
}

//nolint:paralleltest
func TestAccContentTypeResourceRemoveAnnotationsMetadata(t *testing.T) {
	parallelWhenMocked(t)

	server, _ := cmt.NewContentfulManagementServer()

	server.RegisterSpaceEnvironment("0p38pssr0fi3", "test")

	contentTypeID := "acctest_" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	configVariables := config.Variables{
		"space_id":             config.StringVariable("0p38pssr0fi3"),
		"environment_id":       config.StringVariable("test"),
		"test_content_type_id": config.StringVariable(contentTypeID),
	}

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdateMetadata/annotations_only"),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdateMetadata/1"),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/TestAccContentTypeResourceUpdateMetadata/1"),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_content_type.test", plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccContentTypeResourceUpdateResourceLinks(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	server.RegisterSpaceEnvironment("0p38pssr0fi3", "test")

	contentTypeID := "acctest_" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	configVariables := config.Variables{
		"space_id":             config.StringVariable("0p38pssr0fi3"),
		"environment_id":       config.StringVariable("test"),
		"test_content_type_id": config.StringVariable(contentTypeID),
	}

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
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

	server, _ := cmt.NewContentfulManagementServer()

	server.RegisterSpaceEnvironment("0p38pssr0fi3", "test")

	contentTypeID := "acctest_" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	configVariables := config.Variables{
		"space_id":             config.StringVariable("0p38pssr0fi3"),
		"environment_id":       config.StringVariable("test"),
		"test_content_type_id": config.StringVariable(contentTypeID),
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
