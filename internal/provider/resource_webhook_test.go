package provider_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
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
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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

func TestAccWebhookResourceOmittedAndEmptyHeaders(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")

	imported := cm.WebhookDefinitionHeader{Key: "X-Imported", Secret: cm.NewOptBool(true), Value: cm.NewOptString("imported-secret")}
	fields := cm.WebhookDefinitionData{
		Name: "Original", URL: "https://example.com/webhook", Topics: []string{"Entry.publish"}, Active: cm.NewOptBool(true),
		Headers: cm.WebhookDefinitionHeaders{imported},
	}
	server.SetWebhookDefinition("space", "imported-webhook", fields)
	recorder := &webhookWriteRecorder{next: server}

	testAccMockedResource(t, recorder, resource.TestCase{Steps: []resource.TestStep{
		{
			Config: webhookHeadersConfig(fields.Name, ""), ResourceName: "contentful_webhook.test",
			ImportState: true, ImportStateId: "space/imported-webhook", ImportStatePersist: true,
			ImportStateCheck: testAccImportAttributes(map[string]string{"version": "0", "headers.X-Imported.secret": "true"}),
		},
		{Config: webhookHeadersConfig(fields.Name, `headers = {
  X-Formerly-Plain = { value = "plain" }
  X-Imported = { secret = true }
  X-Managed = { secret = true, value = "managed" }
}`)},
		{
			Config: webhookHeadersConfig("Renamed", ""),
			PreConfig: func() {
				fields.Headers = cm.WebhookDefinitionHeaders{
					{Key: "X-Empty", Secret: cm.NewOptBool(false), Value: cm.NewOptString("")},
					{Key: "X-Format", Secret: cm.NewOptBool(false), Value: cm.NewOptString("json")},
					{Key: "X-Formerly-Plain", Secret: cm.NewOptBool(true), Value: cm.NewOptString("external-first")},
					imported,
					{Key: "X-Managed", Secret: cm.NewOptBool(true), Value: cm.NewOptString("external-second")},
				}
				response, err := server.Handler().UpdateWebhookDefinition(t.Context(), &fields, cm.UpdateWebhookDefinitionParams{
					SpaceID: "space", WebhookDefinitionID: "imported-webhook", XContentfulVersion: cm.NewOptInt(1),
				})
				require.NoError(t, err)
				require.IsType(t, &cm.WebhookDefinitionStatusCode{}, response)
			},
			PostApplyFunc: func() {
				recorder.requirePut(t, 2, "2", "Renamed", []map[string]any{
					{"key": "X-Empty", "secret": false, "value": ""},
					{"key": "X-Format", "secret": false, "value": "json"},
					{"key": "X-Formerly-Plain", "secret": true},
					{"key": "X-Imported", "secret": true},
					{"key": "X-Managed", "secret": true},
				})

				stored, ok := server.StoredWebhookDefinition("space", "imported-webhook")
				require.True(t, ok)
				require.Equal(t, fields.Headers, stored.Headers)
			},
		},
		{
			Config: webhookHeadersConfig("Renamed", "headers = {}"),
			PostApplyFunc: func() {
				recorder.requirePut(t, 3, "3", "Renamed", []map[string]any{})

				stored, ok := server.StoredWebhookDefinition("space", "imported-webhook")
				require.True(t, ok)
				require.Empty(t, stored.Headers)
			},
		},
	}})
	require.Len(t, recorder.writes, 4)
	require.Equal(t, http.MethodDelete, recorder.writes[3].method)
}

func TestAccWebhookResourceConfiguredSecretHeaders(t *testing.T) {
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

	step := func(managed, headers string, writes int, requestVersion string, responseVersion json.Number) resource.TestStep {
		return resource.TestStep{
			Config: webhookHeadersConfig(imported.Name, headers),
			ConfigPlanChecks: resource.ConfigPlanChecks{PostApplyPreRefresh: []plancheck.PlanCheck{
				expectResponseStateInPlan{address: address, valuePath: tfjsonpath.New("version"), value: responseVersion},
			}},
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(address, tfjsonpath.New("headers"), knownvalue.MapExact(map[string]knownvalue.Check{
					"X-Secret":  knownvalue.ObjectExact(map[string]knownvalue.Check{"secret": knownvalue.Bool(true), "value": knownvalue.Null()}),
					"X-Managed": knownvalue.ObjectExact(map[string]knownvalue.Check{"secret": knownvalue.Bool(false), "value": knownvalue.StringExact(managed)}),
				})),
			},
			PostApplyFunc: func() {
				recorder.requirePut(t, writes, requestVersion, imported.Name, []map[string]any{
					{"key": "X-Managed", "secret": false, "value": managed}, {"key": "X-Secret", "secret": true},
				})

				stored, ok := server.StoredWebhookDefinition("space", "imported-webhook")
				require.True(t, ok)
				require.Equal(t, cm.WebhookDefinitionHeaders{
					{Key: "X-Managed", Secret: cm.NewOptBool(false), Value: cm.NewOptString(managed)},
					{Key: "X-Secret", Secret: cm.NewOptBool(true), Value: cm.NewOptString("externally-rotated")},
				}, stored.Headers)
			},
		}
	}
	ignored := step("first", `headers = { X-Managed = { value = "first" } }
  lifecycle { ignore_changes = [headers["X-Secret"]] }`, 1, "1", "2")
	ignored.PreConfig = func() {
		imported.Headers[0].Value = cm.NewOptString("externally-rotated")
		response, err := server.Handler().UpdateWebhookDefinition(t.Context(), &imported, cm.UpdateWebhookDefinitionParams{
			SpaceID: "space", WebhookDefinitionID: "imported-webhook", XContentfulVersion: cm.NewOptInt(0),
		})
		require.NoError(t, err)
		require.IsType(t, &cm.WebhookDefinitionStatusCode{}, response)
	}
	explicit := step("second", `headers = { X-Managed = terraform_data.header.output.managed, X-Secret = { secret = terraform_data.header.output.secret } }`, 2, "2", "3")
	explicit.Config += `
resource "terraform_data" "header" {
  input = { managed = { value = "second" }, secret = true }
}
`
	unknownSecret := webhookHeadersConfig(imported.Name, `headers = { X-Secret = { secret = terraform_data.flag.output } }`) + `
resource "terraform_data" "flag" {
  input = true
}
`
	testAccMockedResource(t, recorder, resource.TestCase{Steps: []resource.TestStep{
		{Config: webhookHeadersConfig(imported.Name, ""), ResourceName: address, ImportState: true, ImportStateId: "space/imported-webhook", ImportStatePersist: true},
		ignored, explicit,
		{
			Config:      webhookHeadersConfig(imported.Name, `headers = { x-secret = { secret = true, value = null } }`),
			ExpectError: regexp.MustCompile("Failed to update webhook"),
		},
		{
			Config:      webhookHeadersConfig(imported.Name, `headers = { X-Managed = { secret = true }, X-Secret = { secret = true } }`),
			ExpectError: regexp.MustCompile("Contentful returned different webhook headers"),
		},
		{Config: strings.Replace(explicit.Config, `space_id = "space"`, `space_id = "another-space"`, 1), PlanOnly: true, ExpectError: regexp.MustCompile("Missing initial webhook header value")},
		{Config: unknownSecret, Taint: []string{address}, PlanOnly: true, ExpectError: regexp.MustCompile("Missing initial webhook header value")},
	}})
	recorder.requirePut(t, 3, "3", imported.Name, []map[string]any{{"key": "x-secret", "secret": true}})
	recorder.requirePut(t, 4, "3", imported.Name, []map[string]any{
		{"key": "X-Managed", "secret": true}, {"key": "X-Secret", "secret": true},
	})
	require.Len(t, recorder.writes, 5)
	require.Equal(t, http.MethodDelete, recorder.writes[4].method)
}

func TestAccWebhookResourceCreateWithNullHeaderValue(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")
	recorder := &webhookWriteRecorder{next: server}
	testAccMockedResource(t, recorder, resource.TestCase{Steps: []resource.TestStep{
		{
			Config:      webhookHeadersConfig("New webhook", `headers = { X-Header = { secret = true, value = null } }`),
			PlanOnly:    true,
			ExpectError: regexp.MustCompile("Missing initial webhook header value"),
		},
		{
			Config:      webhookHeadersConfig("New webhook", `headers = { X-Header = { value = null } }`),
			PlanOnly:    true,
			ExpectError: regexp.MustCompile("Missing initial webhook header value"),
		},
		{
			Config: webhookHeadersConfig("New webhook", `headers = { X-Header = { secret = true, value = terraform_data.token.output } }`) + `
resource "terraform_data" "token" { input = "test-token" }
`,
			PlanOnly: true, ExpectNonEmptyPlan: true,
		},
	}})
	recorder.mu.Lock()
	defer recorder.mu.Unlock()

	require.Empty(t, recorder.writes)
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

func (r *webhookWriteRecorder) requirePut(t *testing.T, writeNumber int, version, name string, headers []map[string]any) {
	t.Helper()
	r.mu.Lock()
	defer r.mu.Unlock()

	require.GreaterOrEqual(t, len(r.writes), writeNumber)
	write := r.writes[writeNumber-1]
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

//nolint:paralleltest // Live acceptance serializes access to the shared account.
func TestAccWebhookResourceLifecycle(t *testing.T) {
	parallelWhenMocked(t)

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)

	server.RegisterSpaceEnvironment("0p38pssr0fi3", "master")

	webhookID := "acctest_" + acctest.RandStringFromCharSet(8, "abcdefghijklmnopqrstuvwxyz")

	configVariables := config.Variables{
		"space_id":   config.StringVariable("0p38pssr0fi3"),
		"webhook_id": config.StringVariable(webhookID),
	}

	checkVersion := checkWebhookVersionMatchesContentful(t, server)
	identity := statecheck.CompareValue(compare.ValuesSame())
	stableVersion := statecheck.CompareValue(compare.ValuesSame())
	version := tfjsonpath.New("version")

	testAccMockableResource(t, server, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
				Check:           checkVersion,
				ConfigStateChecks: []statecheck.StateCheck{
					identity.AddStateValue("contentful_webhook.test", tfjsonpath.New("id")),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionCreate),
						plancheck.ExpectUnknownValue("contentful_webhook.test", version),
					},
				},
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: configVariables,
				Check:           checkVersion,
				ConfigStateChecks: []statecheck.StateCheck{
					identity.AddStateValue("contentful_webhook.test", tfjsonpath.New("id")),
					stableVersion.AddStateValue("contentful_webhook.test", version),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionUpdate),
						plancheck.ExpectUnknownValue("contentful_webhook.test", version),
					},
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/TestAccWebhookResourceLifecycle/2"),
				ConfigVariables: configVariables,
				Check:           checkVersion,
				ConfigStateChecks: []statecheck.StateCheck{
					identity.AddStateValue("contentful_webhook.test", tfjsonpath.New("id")),
					stableVersion.AddStateValue("contentful_webhook.test", version),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionNoop),
						plancheck.ExpectKnownValue("contentful_webhook.test", version, knownvalue.NotNull()),
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

func checkWebhookVersionMatchesContentful(t *testing.T, server *cmt.Server) resource.TestCheckFunc {
	t.Helper()

	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC must be set for acceptance tests")
	}

	endpoint, token := os.Getenv("CONTENTFUL_URL"), os.Getenv("CONTENTFUL_MANAGEMENT_ACCESS_TOKEN")
	if endpoint == "" {
		endpoint = cm.DefaultServerURL
	}

	if os.Getenv("TF_ACC_MOCKED") != "" {
		inspectionServer := httptest.NewServer(server)
		t.Cleanup(inspectionServer.Close)
		endpoint, token = inspectionServer.URL, cmt.ValidAccessToken
	}

	client, err := cm.NewClient(endpoint, cm.NewAccessTokenSecuritySource(token), cm.WithClient(cm.NewTransportClient(&http.Client{Timeout: time.Minute}, cm.DefaultUserAgent)))
	require.NoError(t, err)

	return func(state *terraform.State) error {
		const address = "contentful_webhook.test"

		instance := state.RootModule().Resources[address]
		require.NotNil(t, instance, "missing %s in state", address)
		require.NotNil(t, instance.Primary)

		attributes := instance.Primary.Attributes

		response, err := client.GetWebhookDefinition(t.Context(), cm.GetWebhookDefinitionParams{
			SpaceID: attributes["space_id"], WebhookDefinitionID: attributes["webhook_id"],
		})
		if err != nil {
			return fmt.Errorf("read webhook version: %w", err)
		}

		webhook, ok := response.(*cm.WebhookDefinition)
		require.Truef(t, ok, "unexpected GetWebhookDefinition response %T", response)

		return resource.TestCheckResourceAttr(address, "version", strconv.Itoa(webhook.Sys.Version))(state)
	}
}
