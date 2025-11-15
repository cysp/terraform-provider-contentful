package provider_test

import (
	"regexp"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

//nolint:paralleltest
func TestAccWebhookResourceImport(t *testing.T) {
	server, _ := cmt.NewContentfulManagementServer()

	configVariables := config.Variables{
		"space_id": config.StringVariable("0p38pssr0fi3"),
	}

	server.SetWebhookDefinition("0p38pssr0fi3", "6umfVRwmSpcSRdc1jSW6qQ", cm.WebhookDefinitionData{
		Name: "test",
		URL:  "https://example.com",
		Headers: []cm.WebhookDefinitionHeader{
			{
				Key:   "X-Contentful-Test",
				Value: cm.NewOptString("test"),
			},
			{
				Key:    "X-Contentful-Secret",
				Value:  cm.NewOptString("secret"),
				Secret: cm.NewOptBool(true),
			},
		},
		Topics: []string{},
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
				ConfigDirectory:    config.TestNameDirectory(),
				ConfigVariables:    configVariables,
				ResourceName:       "contentful_webhook.test",
				ImportState:        true,
				ImportStateId:      "0p38pssr0fi3/6umfVRwmSpcSRdc1jSW6qQ",
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

//nolint:paralleltest
func TestAccWebhookResourceImportNotFound(t *testing.T) {
	server, _ := cmt.NewContentfulManagementServer()

	configVariables := config.Variables{
		"space_id": config.StringVariable("0p38pssr0fi3"),
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
				ResourceName:    "contentful_webhook.test",
				ImportState:     true,
				ImportStateId:   "0p38pssr0fi3/nonexistent",
				ExpectError:     regexp.MustCompile(`Cannot import non-existent remote object`),
			},
		},
	})
}

func TestAccWebhookResourceCreate(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	webhookID := "acctest_" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	configVariables := config.Variables{
		"space_id":   config.StringVariable("0p38pssr0fi3"),
		"webhook_id": config.StringVariable(webhookID),
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

func TestAccWebhookResourceUpdate(t *testing.T) {
	t.Parallel()

	server, _ := cmt.NewContentfulManagementServer()

	webhookID := "acctest_" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	configVariables := config.Variables{
		"space_id":   config.StringVariable("0p38pssr0fi3"),
		"webhook_id": config.StringVariable(webhookID),
	}

	ContentfulProviderMockableResourceTest(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionNoop),
					},
				},
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}
