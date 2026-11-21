package provider_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sync"
	"sync/atomic"
	"testing"

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
	"github.com/stretchr/testify/require"
)

func TestAccWebhookResourceTopicsValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		topics          string
		expectError     *regexp.Regexp
		expectedCreates int64
	}{
		{
			name:        "omitted",
			expectError: regexp.MustCompile(`(?s)Missing required argument.*topics`),
		},
		{
			name:        "empty",
			topics:      "topics = []",
			expectError: regexp.MustCompile(`list must contain at least 1 elements`),
		},
		{
			name:            "one topic",
			topics:          `topics = ["Entry.publish"]`,
			expectedCreates: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
			require.NoError(t, err)
			server.RegisterSpaceEnvironment("space", "master")

			counter := &webhookCreateCounter{next: server}
			configuration := fmt.Sprintf(`
resource "contentful_webhook" "test" {
  space_id = "space"
  name     = "Topics validation"
  url      = "https://example.com/webhook"
  %s
}
`, test.topics)

			testAccMockedResource(t, counter, resource.TestCase{Steps: []resource.TestStep{{
				Config:      configuration,
				ExpectError: test.expectError,
			}}})

			require.Equal(t, test.expectedCreates, counter.creates.Load())
		})
	}
}

type webhookCreateCounter struct {
	next    http.Handler
	creates atomic.Int64
}

func (c *webhookCreateCounter) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodPost && request.URL.Path == "/spaces/space/webhook_definitions" {
		c.creates.Add(1)
	}

	c.next.ServeHTTP(responseWriter, request)
}

//nolint:paralleltest
func TestAccWebhookResourceImport(t *testing.T) {
	parallelWhenMocked(t)

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)

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
		Topics: []string{"Entry.publish"},
	})

	testAccMockableResource(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory:    config.TestNameDirectory(),
				ConfigVariables:    configVariables,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				ConfigDirectory:  config.TestNameDirectory(),
				ConfigVariables:  configVariables,
				ResourceName:     "contentful_webhook.test",
				ImportState:      true,
				ImportStateId:    "0p38pssr0fi3/6umfVRwmSpcSRdc1jSW6qQ",
				ImportStateCheck: testAccImportAttributes(map[string]string{"id": "0p38pssr0fi3/6umfVRwmSpcSRdc1jSW6qQ", "space_id": "0p38pssr0fi3", "webhook_id": "6umfVRwmSpcSRdc1jSW6qQ"}),
			},
		},
	})
}

func TestAccWebhookResourcePreservesImportedSecretHeadersOnUpdate(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")
	server.SetWebhookDefinition("space", "imported-webhook", cm.WebhookDefinitionData{
		Name: "Imported webhook",
		URL:  "https://example.com/webhook",
		Headers: cm.WebhookDefinitionHeaders{
			{Key: "X-Ordinary", Value: cm.NewOptString("ordinary"), Secret: cm.NewOptBool(false)},
			{Key: "X-Secret", Value: cm.NewOptString("existing-secret"), Secret: cm.NewOptBool(true)},
		},
		Topics: []string{"Entry.publish"},
	})

	recorder := &webhookWriteRecorder{next: server}

	testAccMockedResource(t, recorder, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:             webhookHeadersConfig("Imported webhook", ""),
				ResourceName:       "contentful_webhook.test",
				ImportState:        true,
				ImportStateId:      "space/imported-webhook",
				ImportStatePersist: true,
			},
			{
				Config: webhookHeadersConfig("Updated webhook", ""),
				PostApplyFunc: func() {
					recorder.requirePut(t, 1, "1", "Updated webhook", []map[string]any{
						{"key": "X-Ordinary", "value": "ordinary", "secret": false},
						{"key": "X-Secret", "secret": true},
					})

					stored, ok := server.StoredWebhookDefinition("space", "imported-webhook")
					require.True(t, ok)
					require.Contains(t, stored.Headers, cm.WebhookDefinitionHeader{Key: "X-Secret", Value: cm.NewOptString("existing-secret"), Secret: cm.NewOptBool(true)})
				},
			},
			{
				Config: webhookHeadersConfig("Cleared webhook", "headers = {}"),
				PostApplyFunc: func() {
					recorder.requirePut(t, 2, "2", "Cleared webhook", []map[string]any{})

					stored, ok := server.StoredWebhookDefinition("space", "imported-webhook")
					require.True(t, ok)
					require.Empty(t, stored.Headers)
				},
			},
		},
	})
}

func TestAccWebhookResourceRetainsSecretHeaderWhileUpdatingSibling(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")

	imported := cm.WebhookDefinitionData{
		Name: "Imported webhook", URL: "https://example.com/webhook", Topics: []string{"Entry.publish"}, Active: cm.NewOptBool(true),
		Headers: cm.WebhookDefinitionHeaders{{Key: "X-Secret", Secret: cm.NewOptBool(true), Value: cm.NewOptString("imported-value")}},
	}
	server.SetWebhookDefinition("space", "imported-webhook", imported)
	recorder := &webhookWriteRecorder{next: server}

	const address = "contentful_webhook.test"

	testAccMockedResource(t, recorder, resource.TestCase{Steps: []resource.TestStep{
		{Config: webhookHeadersConfig(imported.Name, ""), ResourceName: address, ImportState: true, ImportStateId: "space/imported-webhook", ImportStatePersist: true},
		{
			Config: webhookHeadersConfig(imported.Name, `headers = { X-Managed = { value = "first" } }
  lifecycle { ignore_changes = [headers["X-Secret"]] }`),
			PreConfig: func() {
				imported.Headers[0].Value = cm.NewOptString("externally-rotated")
				response, err := server.Handler().UpdateWebhookDefinition(t.Context(), &imported, cm.UpdateWebhookDefinitionParams{
					SpaceID: "space", WebhookDefinitionID: "imported-webhook", XContentfulVersion: 1,
				})
				require.NoError(t, err)
				require.IsType(t, &cm.WebhookDefinitionStatusCode{}, response)
			},
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(address, tfjsonpath.New("headers"), knownvalue.MapExact(map[string]knownvalue.Check{
					"X-Secret":  knownvalue.ObjectExact(map[string]knownvalue.Check{"secret": knownvalue.Bool(true), "value": knownvalue.Null()}),
					"X-Managed": knownvalue.ObjectExact(map[string]knownvalue.Check{"secret": knownvalue.Bool(false), "value": knownvalue.StringExact("first")}),
				})),
			},
			PostApplyFunc: func() {
				recorder.requirePut(t, 1, "2", imported.Name, []map[string]any{
					{"key": "X-Managed", "secret": false, "value": "first"}, {"key": "X-Secret", "secret": true},
				})

				stored, ok := server.StoredWebhookDefinition("space", "imported-webhook")
				require.True(t, ok)
				require.Equal(t, cm.WebhookDefinitionHeaders{
					{Key: "X-Managed", Secret: cm.NewOptBool(false), Value: cm.NewOptString("first")},
					{Key: "X-Secret", Secret: cm.NewOptBool(true), Value: cm.NewOptString("externally-rotated")},
				}, stored.Headers)
			},
		},
	}})
	recorder.mu.Lock()
	defer recorder.mu.Unlock()

	require.Len(t, recorder.writes, 2)
	require.Equal(t, http.MethodDelete, recorder.writes[1].method)
}

func webhookHeadersConfig(name, headers string) string {
	return fmt.Sprintf(`
resource "contentful_webhook" "test" {
  space_id = "space"
  name     = %[1]q
  url      = "https://example.com/webhook"
  topics   = ["Entry.publish"]
  %[2]s
}
`, name, headers)
}

type webhookWriteRecorder struct {
	next   http.Handler
	mu     sync.Mutex
	writes []webhookWrite
}

type webhookWrite struct {
	method, path, version, body string
}

func (r *webhookWriteRecorder) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(w, "reading webhook request", http.StatusBadRequest)

			return
		}

		req.Body = io.NopCloser(bytes.NewReader(body))

		r.mu.Lock()
		r.writes = append(r.writes, webhookWrite{method: req.Method, path: req.URL.Path, version: req.Header.Get("X-Contentful-Version"), body: string(body)})
		r.mu.Unlock()
	}

	r.next.ServeHTTP(w, req)
}

func (r *webhookWriteRecorder) requirePut(t *testing.T, count int, version, name string, headers []map[string]any) {
	t.Helper()
	r.mu.Lock()
	defer r.mu.Unlock()

	require.Len(t, r.writes, count)
	write := r.writes[count-1]
	require.Equal(t, http.MethodPut, write.method)
	require.Equal(t, "/spaces/space/webhook_definitions/imported-webhook", write.path)
	require.Equal(t, version, write.version)
	require.JSONEq(t, testJSON(map[string]any{
		"name": name, "url": "https://example.com/webhook", "topics": []string{"Entry.publish"}, "active": true,
		"headers": headers, "filters": nil, "transformation": nil, "httpBasicUsername": nil, "httpBasicPassword": nil,
	}), write.body)
}

//nolint:paralleltest
func TestAccWebhookResourceImportNotFound(t *testing.T) {
	parallelWhenMocked(t)

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)

	configVariables := config.Variables{
		"space_id": config.StringVariable("0p38pssr0fi3"),
	}

	testAccMockableResource(t, server, resource.TestCase{
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

func TestAccWebhookResourceLifecycle(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)

	server.RegisterSpaceEnvironment("0p38pssr0fi3", "master")

	webhookID := "acctest_" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	configVariables := config.Variables{
		"space_id":   config.StringVariable("0p38pssr0fi3"),
		"webhook_id": config.StringVariable(webhookID),
	}

	identity := statecheck.CompareValue(compare.ValuesSame())

	testAccMockableResource(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
				ConfigStateChecks: []statecheck.StateCheck{
					identity.AddStateValue("contentful_webhook.test", tfjsonpath.New("id")),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
				ConfigStateChecks: []statecheck.StateCheck{
					identity.AddStateValue("contentful_webhook.test", tfjsonpath.New("id")),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/TestAccWebhookResourceLifecycle/2"),
				ConfigVariables: configVariables,
				ConfigStateChecks: []statecheck.StateCheck{
					identity.AddStateValue("contentful_webhook.test", tfjsonpath.New("id")),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionNoop),
					},
				},
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
				ConfigStateChecks: []statecheck.StateCheck{
					identity.AddStateValue("contentful_webhook.test", tfjsonpath.New("id")),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}
