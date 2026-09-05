package provider_test

import (
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	extensionVisiblePlanSentinel       = "PLAN_EXTENSION_VISIBLE_SENTINEL_7f3d1b"
	extensionSensitivePlanSentinel     = "PLAN_EXTENSION_SENSITIVE_SENTINEL_0ae76c"
	webhookPasswordPlanSentinel        = "PLAN_WEBHOOK_BASIC_PASSWORD_SENTINEL_91ca42" //nolint:gosec // Harmless test sentinel, not a credential.
	webhookVisibleHeaderPlanSentinel   = "PLAN_WEBHOOK_VISIBLE_HEADER_SENTINEL_d802e5"
	webhookSensitiveHeaderPlanSentinel = "PLAN_WEBHOOK_SENSITIVE_HEADER_SENTINEL_39baf8"
)

func TestAccSecretCapableValuesFollowTerraformPlanSensitivity(t *testing.T) {
	t.Parallel()

	runtime := newTerraformTestRuntime(t)
	runtime.writeConfig(t, fmt.Sprintf(`
variable "webhook_password" {
  type    = string
  default = %q
}

variable "extension_sensitive_parameter" {
  type      = string
  sensitive = true
  default   = %q
}

variable "webhook_sensitive_header" {
  type      = string
  sensitive = true
  default   = %q
}

resource "contentful_extension" "visible" {
  space_id       = "plan-space"
  environment_id = "plan-environment"
  extension_id   = "plan-extension-visible"

  extension = {
    name        = "Plan Extension Visible"
    src         = "https://example.invalid/extension.html"
    field_types = []
  }

  parameters = jsonencode({
    api_key = %q
  })
}

resource "contentful_extension" "sensitive" {
  space_id       = "plan-space"
  environment_id = "plan-environment"
  extension_id   = "plan-extension-sensitive"

  extension = {
    name        = "Plan Extension Sensitive"
    src         = "https://example.invalid/extension.html"
    field_types = []
  }

  parameters = jsonencode({
    api_key = var.extension_sensitive_parameter
  })
}

resource "contentful_webhook" "plan" {
  space_id            = "plan-space"
  name                = "Plan Webhook"
  url                 = "https://example.invalid/webhook"
  topics              = ["Entry.publish"]
  http_basic_password = var.webhook_password

  headers = {
    Authorization = {
      value  = %q
      secret = true
    }
    SensitiveAuthorization = {
      value  = var.webhook_sensitive_header
      secret = true
    }
  }
}
`, webhookPasswordPlanSentinel, extensionSensitivePlanSentinel, webhookSensitiveHeaderPlanSentinel, extensionVisiblePlanSentinel, webhookVisibleHeaderPlanSentinel))

	output, err := runtime.run(t.Context(), "plan", "-input=false", "-no-color", "-lock=false")
	require.NoError(t, err, output)

	assert.Contains(t, output, extensionVisiblePlanSentinel)
	assert.Contains(t, output, webhookVisibleHeaderPlanSentinel)

	for _, sentinel := range []string{
		extensionSensitivePlanSentinel,
		webhookPasswordPlanSentinel,
		webhookSensitiveHeaderPlanSentinel,
	} {
		assert.NotContains(t, output, sentinel)
	}

	assert.Regexp(t, `(?m)^\s+\+ parameters\s+= \(sensitive value\)$`, output)
	assert.Regexp(t, `(?m)^\s+\+ http_basic_password\s+= \(sensitive value\)$`, output)
	assert.Regexp(t, fmt.Sprintf(`(?ms)^\s+\+ "Authorization"\s+= \{\s+\+ secret\s+= true\s+\+ value\s+= %q\s+\},$`, webhookVisibleHeaderPlanSentinel), output)
	assert.Regexp(t, `(?ms)^\s+\+ "SensitiveAuthorization"\s+= \{\s+\+ secret\s+= true\s+\+ value\s+= \(sensitive value\)\s+\},$`, output)
}

func TestAccExtensionLifecycleLogsExcludeParameters(t *testing.T) {
	t.Parallel()

	const (
		createSentinel = "LOG_EXTENSION_CREATE_API_KEY_SENTINEL_2f601a"
		updateSentinel = "LOG_EXTENSION_UPDATE_API_KEY_SENTINEL_644be9"
	)

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("log-space", "log-environment")

	testServer := httptest.NewServer(server)
	t.Cleanup(testServer.Close)

	runtime := newTerraformTestRuntime(t)
	runtime.providerURL = testServer.URL
	runtime.logPath = filepath.Join(runtime.workingDirectory, "provider.log")

	runtime.writeConfig(t, extensionLoggingTestConfig(createSentinel, "Created"))
	output, err := runtime.run(t.Context(), "apply", "-auto-approve", "-input=false", "-no-color")
	require.NoError(t, err, output)

	runtime.writeConfig(t, extensionLoggingTestConfig(updateSentinel, "Updated"))
	output, err = runtime.run(t.Context(), "apply", "-auto-approve", "-input=false", "-no-color")
	require.NoError(t, err, output)

	logOutput, err := os.ReadFile(runtime.logPath)
	require.NoError(t, err)

	logs := string(logOutput)

	assert.Contains(t, logs, "extension.create")
	assert.Contains(t, logs, "extension.read")
	assert.Contains(t, logs, "extension.update")

	for _, sentinel := range []string{createSentinel, updateSentinel} {
		assert.NotContains(t, logs, sentinel)
	}
}

func TestAccWebhookLifecycleLogsAndOutputExcludeBasicPassword(t *testing.T) {
	t.Parallel()

	sentinels := []string{
		"LOG_WEBHOOK_BASIC_PASSWORD_CREATE_SENTINEL_6c5fd1",
		"LOG_WEBHOOK_BASIC_PASSWORD_UPDATE_SENTINEL_b86312",
	}

	server, err := cmt.NewContentfulManagementServer()
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("log-space", "master")

	testServer := httptest.NewServer(server)
	t.Cleanup(testServer.Close)

	runtime := newTerraformTestRuntime(t)
	runtime.providerURL = testServer.URL
	runtime.logPath = filepath.Join(runtime.workingDirectory, "provider.log")

	for index, sentinel := range sentinels {
		runtime.writeConfig(t, webhookLoggingTestConfig(sentinel, fmt.Sprintf("Webhook %d", index)))
		output, applyErr := runtime.run(t.Context(), "apply", "-auto-approve", "-input=false", "-no-color")
		require.NoError(t, applyErr, output)
		assert.NotContains(t, output, sentinel)
	}

	logOutput, err := os.ReadFile(runtime.logPath)
	require.NoError(t, err)

	logs := string(logOutput)
	assert.Contains(t, logs, "webhook.create")
	assert.Contains(t, logs, "webhook.read")
	assert.Contains(t, logs, "webhook.update")

	for _, sentinel := range sentinels {
		assert.NotContains(t, logs, sentinel)
	}
}

type terraformTestRuntime struct {
	terraformPath    string
	workingDirectory string
	cliConfigPath    string
	providerURL      string
	logPath          string
}

func newTerraformTestRuntime(t *testing.T) *terraformTestRuntime {
	t.Helper()

	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC must be set for Terraform CLI tests")
	}

	terraformPath := os.Getenv("TF_ACC_TERRAFORM_PATH")
	if terraformPath == "" {
		var err error
		terraformPath, err = exec.LookPath("terraform")
		require.NoError(t, err, "Terraform CLI is required for acceptance tests")
	}

	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok)

	repositoryRoot := filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
	workingDirectory := t.TempDir()
	providerDirectory := filepath.Join(workingDirectory, "providers")
	require.NoError(t, os.Mkdir(providerDirectory, 0o700))

	providerPath := filepath.Join(providerDirectory, "terraform-provider-contentful")
	build := exec.CommandContext(t.Context(), "go", "build", "-o", providerPath, ".")
	build.Dir = repositoryRoot
	buildOutput, err := build.CombinedOutput()
	require.NoError(t, err, string(buildOutput))

	cliConfigPath := filepath.Join(workingDirectory, "terraform.rc")
	require.NoError(t, os.WriteFile(cliConfigPath, fmt.Appendf(nil, `
provider_installation {
  dev_overrides {
    "registry.terraform.io/cysp/contentful" = %q
  }
  direct {}
}
`, providerDirectory), 0o600))

	return &terraformTestRuntime{
		terraformPath:    terraformPath,
		workingDirectory: workingDirectory,
		cliConfigPath:    cliConfigPath,
		providerURL:      "https://contentful.invalid",
	}
}

func (runtime *terraformTestRuntime) writeConfig(t *testing.T, resources string) {
	t.Helper()

	configuration := fmt.Sprintf(`
terraform {
  required_providers {
    contentful = {
      source = "cysp/contentful"
    }
  }
}

provider "contentful" {
  access_token = %q
  url          = %q
}
%s
`, cmt.ValidAccessToken, runtime.providerURL, resources)
	require.NoError(t, os.WriteFile(filepath.Join(runtime.workingDirectory, "main.tf"), []byte(configuration), 0o600))
}

func (runtime *terraformTestRuntime) run(ctx context.Context, arguments ...string) (string, error) {
	//nolint:gosec // Test-only Terraform CLI path and arguments are fixed by the caller.
	command := exec.CommandContext(ctx, runtime.terraformPath, arguments...)
	command.Dir = runtime.workingDirectory

	command.Env = append(os.Environ(),
		"TF_CLI_CONFIG_FILE="+runtime.cliConfigPath,
		"TF_IN_AUTOMATION=1",
	)
	if runtime.logPath != "" {
		command.Env = append(command.Env,
			"TF_LOG_PROVIDER=INFO",
			"TF_LOG_PATH="+runtime.logPath,
		)
	}

	output, err := command.CombinedOutput()

	return string(output), err
}

func extensionLoggingTestConfig(sentinel, name string) string {
	return fmt.Sprintf(`
resource "contentful_extension" "logs" {
  space_id       = "log-space"
  environment_id = "log-environment"
  extension_id   = "log-extension"

  extension = {
    name        = %q
    src         = "https://example.invalid/extension.html"
    field_types = []
  }

  parameters = jsonencode({
    api_key = %q
  })
}
`, name, sentinel)
}

func webhookLoggingTestConfig(sentinel, name string) string {
	return fmt.Sprintf(`
resource "contentful_webhook" "logs" {
  space_id            = "log-space"
  name                = %q
  url                 = "https://example.invalid/webhook"
  topics              = ["Entry.publish"]
  http_basic_username = "user"
  http_basic_password = %q
}
`, name, sentinel)
}
