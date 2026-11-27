package provider_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/require"
)

func newWebhookBasicFixture(t *testing.T) (*terraformTestRuntime, *cmt.Server, *taxonomyRequestBodyRecorder, *mutationJSONResponseAdapter) {
	t.Helper()
	runtime := newTerraformTestRuntime(t)
	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")
	recorder := &taxonomyRequestBodyRecorder{next: server}
	adapter := &mutationJSONResponseAdapter{delegate: recorder}
	httpServer := httptest.NewServer(adapter)
	t.Cleanup(httpServer.Close)
	runtime.providerURL = httpServer.URL
	runtime.logPath = filepath.Join(runtime.workingDirectory, "provider.log")

	return runtime, server, recorder, adapter
}

func TestAccWebhookResourceWriteOnlyBasicPassword(t *testing.T) {
	t.Parallel()
	runtime, _, recorder, adapter := newWebhookBasicFixture(t)

	const (
		first     = "original-synthetic-basic"
		planned   = "planned-synthetic-basic"
		applied   = "applied-synthetic-basic"
		writeOnly = "http_basic_password_wo = var.password"
	)

	config := func(password, name string) {
		runtime.writeConfig(t, fmt.Sprintf(`
variable "password" {
 type=string
 sensitive=true
 ephemeral=true
}
resource "contentful_webhook" "test" {
 space_id="space"
 name=%q
 url="https://example.invalid/webhook"
 topics=["Entry.publish"]
 http_basic_username="user"
 %s
 headers={"X-Key"={secret=true,value="ordinary-header"}}
}`, name, password))
	}
	cli := func(command, password string, opts ...string) (string, error) {
		return runtime.run(t.Context(), append([]string{command, "-input=false", "-no-color", "-var=password=" + password}, opts...)...)
	}
	run := func(command, password string, opts ...string) {
		t.Helper()

		output, err := cli(command, password, opts...)
		require.NoError(t, err, output)
	}
	body := func(name, password string) string {
		return testJSON(map[string]any{
			"name": name, "url": "https://example.invalid/webhook", "active": true, "topics": []string{"Entry.publish"},
			"httpBasicUsername": "user", "httpBasicPassword": password, "filters": nil, "transformation": nil,
			"headers": []map[string]any{{"key": "X-Key", "secret": true, "value": "ordinary-header"}},
		})
	}
	state := func() webhookStateInstance {
		t.Helper()
		instances := readWebhookStateInstances(t, runtime)
		require.Len(t, instances, 1)

		return instances[0]
	}
	verifier := func(instance webhookStateInstance) []byte {
		t.Helper()

		var private map[string][]byte
		require.NoError(t, json.Unmarshal(instance.Private, &private))

		return private["write_only_secret_hashes"]
	}

	config(writeOnly, "Managed")
	run("apply", first, "-auto-approve")

	after := state()
	webhookID, ok := after.Attributes["webhook_id"].(string)
	require.True(t, ok)

	created := recorder.matchingRequests(http.MethodPost, "")
	require.Len(t, created, 1)
	require.Equal(t, "/spaces/space/webhook_definitions", created[0].path)
	require.Empty(t, created[0].version)
	require.JSONEq(t, body("Managed", first), string(created[0].body))
	require.Nil(t, after.Attributes["http_basic_password"])
	require.Nil(t, after.Attributes["http_basic_password_wo"])
	original := verifier(after)
	require.Contains(t, string(original), `"path":"http_basic_password_wo"`)
	require.NotContains(t, string(original), first)

	run("plan", first, "-out=noop.tfplan", "-detailed-exitcode")
	run("apply", planned, "noop.tfplan")
	require.Empty(t, recorder.matchingRequests(http.MethodPut, ""))
	run("plan", planned, "-out=rotate.tfplan")
	run("apply", applied, "rotate.tfplan")

	puts := recorder.matchingRequests(http.MethodPut, "")
	require.Len(t, puts, 1)
	require.Equal(t, "/spaces/space/webhook_definitions/"+webhookID, puts[0].path)
	require.Equal(t, "0", puts[0].version)
	require.JSONEq(t, body("Managed", applied), string(puts[0].body))

	rotated := verifier(state())
	require.NotEqual(t, original, rotated)
	run("plan", applied, "-detailed-exitcode")

	config(fmt.Sprintf("http_basic_password = %q", applied), "Managed")
	run("apply", planned, "-auto-approve")

	after = state()
	require.Equal(t, applied, after.Attributes["http_basic_password"])
	require.Empty(t, verifier(after))
	config(writeOnly, "Managed")
	run("apply", applied, "-auto-approve")
	require.Equal(t, puts, recorder.matchingRequests(http.MethodPut, ""))

	baseline := verifier(state())
	require.NotEmpty(t, baseline)

	config(writeOnly, "Changed")
	adapter.mutateNext(http.MethodPut, func(doc map[string]any) { doc["name"] = "Contradictory" })

	output, err := cli("apply", first, "-auto-approve")
	require.Error(t, err, output)
	require.Contains(t, output, "Contentful returned a different webhook name")
	require.NotContains(t, output, first)

	puts = recorder.matchingRequests(http.MethodPut, "")
	require.Len(t, puts, 2)
	require.Equal(t, "1", puts[1].version)
	require.Equal(t, puts[0].path, puts[1].path)
	require.JSONEq(t, body("Changed", first), string(puts[1].body))

	after = state()
	require.Equal(t, baseline, verifier(after))
	require.Nil(t, after.Attributes["http_basic_password"])
	require.Equal(t, "Contradictory", after.Attributes["name"])
	require.Equal(t, "2", testJSON(after.Attributes["version"]))
	require.Equal(t, map[string]any{"X-Key": map[string]any{"secret": true, "value": "ordinary-header", "value_wo": nil}}, after.Attributes["headers"])

	// A refreshed no-op compares against Terraform's stored acknowledgement;
	// only a fresh differing value deliberately writes the password again.
	run("plan", applied, "-detailed-exitcode")
	run("apply", planned, "-auto-approve")

	puts = recorder.matchingRequests(http.MethodPut, "")
	require.Len(t, puts, 3)
	require.Equal(t, "2", puts[2].version)
	require.JSONEq(t, body("Changed", planned), string(puts[2].body))
	run("plan", planned, "-detailed-exitcode")
}
