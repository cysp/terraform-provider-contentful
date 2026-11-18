package provider_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/require"
)

const (
	webhookPriorPassword   = "synthetic-password-for-import-2ZB6"
	webhookPriorHeader     = "synthetic-managed-header-for-import-V5py"
	webhookRotatedPassword = "synthetic-new-password-after-import-B2gG"
	webhookRotatedHeader   = "synthetic-new-managed-header-after-import-7GbN"
	webhookSiblingHeader   = "synthetic-imported-unreadable-sibling-Aw6f"
)

type webhookMutation struct {
	method  string
	path    string
	version string
	body    string
}

type webhookSecretRecoveryFixture struct {
	server         *cmt.Server
	mu             sync.Mutex
	requests       []webhookMutation
	overrideMethod string
	override       func(http.ResponseWriter, *httptest.ResponseRecorder)
}

func (f *webhookSecretRecoveryFixture) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if r.Method == http.MethodPut || r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "fixture could not read PUT", http.StatusBadRequest)

			return
		}

		r.Body = io.NopCloser(bytes.NewReader(body))
		f.requests = append(f.requests, webhookMutation{method: r.Method, path: r.URL.Path, version: r.Header.Get("X-Contentful-Version"), body: testJSON(testJSONObject(string(body)))})
	}

	rec := httptest.NewRecorder()
	f.server.ServeHTTP(rec, r)
	maps.Copy(w.Header(), rec.Header())
	w.Header().Del("Content-Length")

	if f.overrideMethod == r.Method && (rec.Code == http.StatusOK || rec.Code == http.StatusCreated) {
		f.overrideMethod = ""
		f.override(w, rec)

		return
	}

	w.WriteHeader(rec.Code)
	_, _ = w.Write(rec.Body.Bytes())
}

func (f *webhookSecretRecoveryFixture) mutations() []webhookMutation {
	f.mu.Lock()
	defer f.mu.Unlock()

	return slices.Clone(f.requests)
}

func (f *webhookSecretRecoveryFixture) onNextResponse(method string, override func(http.ResponseWriter, *httptest.ResponseRecorder)) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.overrideMethod, f.override = method, override
}

func (f *webhookSecretRecoveryFixture) requirePuts(t *testing.T, expected ...webhookMutation) {
	t.Helper()
	f.mu.Lock()
	defer f.mu.Unlock()

	require.Empty(t, f.overrideMethod, "fixture fault was never exercised")
	require.Equal(t, expected, f.requests)
}

func webhookExpectedRecoveryPut(version, password, header string, includeSibling bool) webhookMutation {
	managed := map[string]any{"key": "X-Managed", "secret": true}
	if header != "" {
		managed["value"] = header
	}

	headers := []map[string]any{managed}
	if includeSibling {
		headers = append(headers, map[string]any{"key": "X-Sibling", "secret": true})
	}

	return webhookMutation{method: http.MethodPut, path: "/spaces/space/webhook_definitions/imported", version: version, body: testJSON(map[string]any{
		"name": "Imported", "url": "https://example.invalid/webhook", "topics": []string{"Entry.publish"},
		"active": true, "filters": nil, "transformation": nil,
		"httpBasicUsername": "user", "httpBasicPassword": password, "headers": headers,
	})}
}

func (f *webhookSecretRecoveryFixture) requireStored(t *testing.T, password, header string, includeSibling bool) {
	t.Helper()

	stored, ok := f.server.StoredWebhookDefinition("space", "imported")
	require.True(t, ok)
	require.Equal(t, cm.NewOptNilString(password), stored.HttpBasicPassword)

	expected := cm.WebhookDefinitionHeaders{{Key: "X-Managed", Secret: cm.NewOptBool(true), Value: cm.NewOptString(header)}}
	if includeSibling {
		expected = append(expected, cm.WebhookDefinitionHeader{Key: "X-Sibling", Secret: cm.NewOptBool(true), Value: cm.NewOptString(webhookSiblingHeader)})
	}

	require.Equal(t, expected, stored.Headers)
}

func webhookRecoveryConfig(includeSibling bool, ordinaryKey string) string {
	sibling := ""
	if includeSibling {
		sibling = `"X-Sibling" = { secret=true }`
	}

	managed := `"X-Managed" = { secret=true }`
	writeOnly := `header_values_wo = { "X-Managed" = var.header }`

	if ordinaryKey != "" {
		managed = fmt.Sprintf(`%q = { secret=false, value="public-replacement" }`, ordinaryKey)
		writeOnly = ""
	}

	return fmt.Sprintf(`
variable "password" {
  type      = string
  sensitive = true
  ephemeral = true
}
variable "header" {
  type      = string
  sensitive = true
  ephemeral = true
}
resource "contentful_webhook" "test" {
  space_id = "space"
  name     = "Imported"
  url      = "https://example.invalid/webhook"
  topics   = ["Entry.publish"]

  http_basic_username    = "user"
  http_basic_password_wo = var.password
  headers = {
    %s
    %s
  }
  %s
}
`, managed, sibling, writeOnly)
}

func runWebhookRecoveryCLI(t *testing.T, runtime *terraformTestRuntime, password, header, command string, args ...string) (string, error) {
	t.Helper()

	flags := []string{command, "-no-color", "-input=false", "-var=password=" + password, "-var=header=" + header}
	output, err := runtime.run(t.Context(), append(flags, args...)...)
	requireNoWebhookRecoverySecrets(t, output)

	return output, err
}

func requireNoWebhookRecoverySecrets(t *testing.T, text string) {
	t.Helper()

	for _, secret := range []string{webhookPriorPassword, webhookPriorHeader, webhookRotatedPassword, webhookRotatedHeader, webhookSiblingHeader} {
		require.NotContains(t, text, secret)
	}
}

type webhookRecoveryInstance struct {
	Attributes map[string]any `json:"attributes"`
	Status     string         `json:"status"`
	Private    []byte         `json:"private"`
	verifiers  []byte
}

func readWebhookRecoveryInstance(t *testing.T, runtime *terraformTestRuntime) webhookRecoveryInstance {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(runtime.workingDirectory, "terraform.tfstate"))
	require.NoError(t, err)
	requireNoWebhookRecoverySecrets(t, string(data))

	var state struct {
		Resources []struct {
			Instances []webhookRecoveryInstance `json:"instances"`
		} `json:"resources"`
	}
	require.NoError(t, json.Unmarshal(data, &state))
	require.Len(t, state.Resources, 1)
	require.Len(t, state.Resources[0].Instances, 1)

	instance := state.Resources[0].Instances[0]

	var private map[string][]byte
	if len(instance.Private) > 0 {
		require.NoError(t, json.Unmarshal(instance.Private, &private))
	}

	for _, field := range private {
		requireNoWebhookRecoverySecrets(t, string(field))
	}

	instance.verifiers = private["write_only_secret_hashes"]

	return instance
}

func newWebhookRecoveryFixture(t *testing.T) (*terraformTestRuntime, *webhookSecretRecoveryFixture) {
	t.Helper()
	runtime := newTerraformTestRuntime(t)
	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")
	fixture := &webhookSecretRecoveryFixture{server: server}
	httpServer := httptest.NewServer(fixture)
	t.Cleanup(httpServer.Close)
	runtime.providerURL = httpServer.URL
	runtime.logPath = filepath.Join(runtime.workingDirectory, "provider.log")

	return runtime, fixture
}

// Import supplies neither plaintext nor a verifier. Even supplying the remote bytes
// must issue a real PUT, while an independently managed redacted sibling survives.
func importWebhookRecoveryFixture(t *testing.T) (*terraformTestRuntime, *webhookSecretRecoveryFixture, webhookMutation) {
	t.Helper()

	runtime, fixture := newWebhookRecoveryFixture(t)
	fixture.server.SetWebhookDefinition("space", "imported", cm.WebhookDefinitionData{
		Name: "Imported", URL: "https://example.invalid/webhook", Topics: []string{"Entry.publish"}, Active: cm.NewOptBool(true),
		HttpBasicUsername: cm.NewOptNilString("user"), HttpBasicPassword: cm.NewOptNilString(webhookPriorPassword),
		Headers: cm.WebhookDefinitionHeaders{
			{Key: "X-Managed", Secret: cm.NewOptBool(true), Value: cm.NewOptString(webhookPriorHeader)},
			{Key: "X-Sibling", Secret: cm.NewOptBool(true), Value: cm.NewOptString(webhookSiblingHeader)},
		},
	})
	runtime.writeConfig(t, webhookRecoveryConfig(true, ""))
	output, err := runWebhookRecoveryCLI(t, runtime, webhookPriorPassword, webhookPriorHeader, "import", "contentful_webhook.test", "space/imported")
	require.NoError(t, err, output)
	require.Empty(t, readWebhookRecoveryInstance(t, runtime).verifiers)
	fixture.requirePuts(t)
	output, err = runWebhookRecoveryCLI(t, runtime, webhookPriorPassword, webhookPriorHeader, "apply", "-auto-approve")
	require.NoError(t, err, output)

	first := webhookExpectedRecoveryPut("1", webhookPriorPassword, webhookPriorHeader, true)
	fixture.requirePuts(t, first)
	fixture.requireStored(t, webhookPriorPassword, webhookPriorHeader, true)
	require.NotEmpty(t, readWebhookRecoveryInstance(t, runtime).verifiers)

	return runtime, fixture, first
}
