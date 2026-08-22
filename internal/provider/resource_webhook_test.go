package provider_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sync"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"
)

var errWebhookUpdateMismatch = errors.New("webhook update mismatch")

//nolint:paralleltest
func TestAccWebhookResourceImport(t *testing.T) {
	parallelWhenMocked(t)

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

func TestAccWebhookResourcePreservesImportedSecretHeadersOnUpdate(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(100))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")
	server.SetWebhookDefinition("space", "imported-webhook", cm.WebhookDefinitionData{
		Name: "Imported webhook",
		URL:  "https://example.com/webhook",
		Headers: cm.WebhookDefinitionHeaders{
			{Key: "X-Ordinary", Value: cm.NewOptString("ordinary"), Secret: cm.NewOptBool(false)},
			{Key: "X-Secret", Value: cm.NewOptString("existing-secret"), Secret: cm.NewOptBool(true)},
		},
		Topics: []string{},
	})

	recorder := &webhookUpdateRecorder{next: server}
	config := func(name, headers string) string {
		return fmt.Sprintf(`
resource "contentful_webhook" "test" {
  space_id = "space"
  name     = %q
  url      = "https://example.com/webhook"
  topics   = []
  %s
}
`, name, headers)
	}

	ContentfulProviderMockedResourceTest(t, recorder, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:             config("Imported webhook", ""),
				ResourceName:       "contentful_webhook.test",
				ImportState:        true,
				ImportStateId:      "space/imported-webhook",
				ImportStatePersist: true,
			},
			{
				Config: config("Updated webhook", ""),
				Check: resource.ComposeTestCheckFunc(
					recorder.checkPreservedHeadersRequest(),
					func(*terraform.State) error {
						stored, ok := server.StoredWebhookDefinition("space", "imported-webhook")
						if !ok {
							return fmt.Errorf("%w: updated webhook was not stored", errWebhookUpdateMismatch)
						}

						for _, header := range stored.Headers {
							if header.Key == "X-Secret" && header.Value == cm.NewOptString("existing-secret") {
								return nil
							}
						}

						return fmt.Errorf("%w: stored secret header was not preserved: %#v", errWebhookUpdateMismatch, stored.Headers)
					},
				),
			},
			{
				Config: config("Cleared webhook", "headers = {}"),
				Check: resource.ComposeTestCheckFunc(
					recorder.checkEmptyHeadersRequest(),
					func(*terraform.State) error {
						response, getErr := server.Handler().GetWebhookDefinition(t.Context(), cm.GetWebhookDefinitionParams{
							SpaceID:             "space",
							WebhookDefinitionID: "imported-webhook",
						})
						if getErr != nil {
							return fmt.Errorf("get webhook after clearing headers: %w", getErr)
						}

						webhook, ok := response.(*cm.WebhookDefinition)
						if !ok {
							return fmt.Errorf("%w: unexpected response %T", errWebhookUpdateMismatch, response)
						}

						if len(webhook.Headers) != 0 {
							return fmt.Errorf("%w: headers were not cleared: %#v", errWebhookUpdateMismatch, webhook.Headers)
						}

						return nil
					},
				),
			},
		},
	})
}

type webhookUpdateRecorder struct {
	next   http.Handler
	mu     sync.Mutex
	bodies []map[string]json.RawMessage
}

func (r *webhookUpdateRecorder) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodPut && request.URL.Path == "/spaces/space/webhook_definitions/imported-webhook" {
		body, err := io.ReadAll(request.Body)
		if err == nil {
			request.Body = io.NopCloser(bytes.NewReader(body))

			var decoded map[string]json.RawMessage
			if json.Unmarshal(body, &decoded) == nil {
				r.mu.Lock()
				r.bodies = append(r.bodies, decoded)
				r.mu.Unlock()
			}
		}
	}

	r.next.ServeHTTP(responseWriter, request)
}

func (r *webhookUpdateRecorder) checkPreservedHeadersRequest() resource.TestCheckFunc {
	return func(*terraform.State) error {
		r.mu.Lock()
		defer r.mu.Unlock()

		if len(r.bodies) != 1 {
			return fmt.Errorf("%w: recorded %d updates, want exactly 1", errWebhookUpdateMismatch, len(r.bodies))
		}

		var headers []map[string]json.RawMessage

		err := json.Unmarshal(r.bodies[0]["headers"], &headers)
		if err != nil {
			return fmt.Errorf("decode webhook headers: %w", err)
		}

		if len(headers) != 2 {
			return fmt.Errorf("%w: sent %d headers, want 2", errWebhookUpdateMismatch, len(headers))
		}

		byKey := make(map[string]map[string]json.RawMessage, len(headers))
		for _, header := range headers {
			var key string

			err = json.Unmarshal(header["key"], &key)
			if err != nil {
				return fmt.Errorf("decode webhook header key: %w", err)
			}

			byKey[key] = header
		}

		if string(byKey["X-Ordinary"]["value"]) != `"ordinary"` {
			return fmt.Errorf("%w: ordinary header value was %s", errWebhookUpdateMismatch, byKey["X-Ordinary"]["value"])
		}

		secret := byKey["X-Secret"]
		if string(secret["secret"]) != "true" {
			return fmt.Errorf("%w: secret header marker was %s", errWebhookUpdateMismatch, secret["secret"])
		}

		if _, ok := secret["value"]; ok {
			return fmt.Errorf("%w: secret header request included value %s; want the member omitted", errWebhookUpdateMismatch, secret["value"])
		}

		return nil
	}
}

func (r *webhookUpdateRecorder) checkEmptyHeadersRequest() resource.TestCheckFunc {
	return func(*terraform.State) error {
		r.mu.Lock()
		defer r.mu.Unlock()

		if len(r.bodies) != 2 {
			return fmt.Errorf("%w: recorded %d updates, want exactly 2", errWebhookUpdateMismatch, len(r.bodies))
		}

		if body, ok := r.bodies[1]["headers"]; !ok || string(body) != "[]" {
			return fmt.Errorf("%w: explicit empty headers did not produce an empty array", errWebhookUpdateMismatch)
		}

		return nil
	}
}

//nolint:paralleltest
func TestAccWebhookResourceImportNotFound(t *testing.T) {
	parallelWhenMocked(t)

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

	server.RegisterSpaceEnvironment("0p38pssr0fi3", "master")

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

	server.RegisterSpaceEnvironment("0p38pssr0fi3", "master")

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
