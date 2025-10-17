package provider_test

import (
	"regexp"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccEntryResourceImport(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	configVariables := config.Variables{
		"space_id":       config.StringVariable("0p38pssr0fi3"),
		"environment_id": config.StringVariable("test"),
		"entry_id":       config.StringVariable("entry"),
	}

	server.SetEntry("0p38pssr0fi3", "test", "contentType", "entry", cm.EntryRequest{
		Fields: cm.NewOptEntryFields(cm.EntryFields{}),
	})

	ContentfulProviderMockedResourceTest(t, server, resource.TestCase{
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
				ResourceName:    "contentful_entry.test",
				ImportState:     true,
				ImportStateId:   "a",
				ExpectError:     regexp.MustCompile(`Resource Import Passthrough Multipart ID Mismatch`),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ResourceName:    "contentful_entry.test",
				ImportState:     true,
				ImportStateId:   "a/b",
				ExpectError:     regexp.MustCompile(`Resource Import Passthrough Multipart ID Mismatch`),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ResourceName:    "contentful_entry.test",
				ImportState:     true,
				ImportStateId:   "a/b/c/d",
				ExpectError:     regexp.MustCompile(`Resource Import Passthrough Multipart ID Mismatch`),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
				ResourceName:    "contentful_entry.test",
				ImportState:     true,
				ImportStateId:   "0p38pssr0fi3/test/entry",
			},
		},
	})
}

func TestAccEntryResourceImportNotFound(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	configVariables := config.Variables{
		"space_id":       config.StringVariable("0p38pssr0fi3"),
		"environment_id": config.StringVariable("test"),
		"entry_id":       config.StringVariable("nonexistent"),
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
				ResourceName:    "contentful_entry.test",
				ImportState:     true,
				ImportStateId:   "0p38pssr0fi3/test/nonexistent",
				ExpectError:     regexp.MustCompile(`Cannot import non-existent remote object`),
			},
		},
	})
}

func TestAccEntryResourceCreate(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	configVariables := config.Variables{
		"space_id":        config.StringVariable("0p38pssr0fi3"),
		"environment_id":  config.StringVariable("test"),
		"content_type_id": config.StringVariable("author"),
		"fields": config.MapVariable(map[string]config.Variable{
			"name": config.StringVariable(`{"en-AU":"name"}`),
		}),
	}

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
			},
		},
	})
}

func TestAccEntryResourceCreateWithID(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	entryID := "acctest_" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	configVariables := config.Variables{
		"space_id":        config.StringVariable("0p38pssr0fi3"),
		"environment_id":  config.StringVariable("test"),
		"entry_id":        config.StringVariable(entryID),
		"content_type_id": config.StringVariable("author"),
		"fields": config.MapVariable(map[string]config.Variable{
			"name": config.StringVariable(`{"en-AU":"name"}`),
		}),
	}

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables,
			},
		},
	})
}

func TestAccEntryResourceUpdate(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	configVariables1 := config.Variables{
		"space_id":        config.StringVariable("0p38pssr0fi3"),
		"environment_id":  config.StringVariable("test"),
		"content_type_id": config.StringVariable("author"),
		"fields": config.MapVariable(map[string]config.Variable{
			"name": config.StringVariable(`{"en-AU":"name"}`),
		}),
	}

	configVariables2 := config.Variables{
		"space_id":        config.StringVariable("0p38pssr0fi3"),
		"environment_id":  config.StringVariable("test"),
		"content_type_id": config.StringVariable("author"),
		"fields": config.MapVariable(map[string]config.Variable{
			"name": config.StringVariable(`{"en-AU":"name (updated)"}`),
		}),
	}

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables1,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionCreate),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue("contentful_entry.test", tfjsonpath.New("entry_id"), knownvalue.NotNull()),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue("contentful_entry.test", tfjsonpath.New("entry_id"), knownvalue.NotNull()),
					},
				},
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: configVariables2,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue("contentful_entry.test", tfjsonpath.New("entry_id"), knownvalue.NotNull()),
					},
				},
			},
		},
	})
}

func TestAccEntryResourceDeleted(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	entryID := "acctest_" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	configVariables := config.Variables{
		"space_id":        config.StringVariable("0p38pssr0fi3"),
		"environment_id":  config.StringVariable("test"),
		"entry_id":        config.StringVariable(entryID),
		"content_type_id": config.StringVariable("author"),
		"fields": config.MapVariable(map[string]config.Variable{
			"name": config.StringVariable(`{"en-AU":"name"}`),
		}),
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
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_entry.test", plancheck.ResourceActionDestroy),
					},
				},
			},
		},
	})
}
