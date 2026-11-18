package provider_test

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAccWebhookResourceWriteOnlyLifecycle(t *testing.T) {
	t.Parallel()
	runtime, fixture := newWebhookRecoveryFixture(t)

	const (
		initialPassword = "synthetic-password-uq3Lny8k"
		nextPassword    = "synthetic-password-2-jrn14X38"
		token           = "synthetic-token-p9Jx2u8W"
		nextToken       = "synthetic-token-2-d7L3Bm28"
		sibling         = "synthetic-sibling-zx7xrpG9"
	)

	checkSecrets := func(data string) {
		t.Helper()

		for _, secret := range []string{initialPassword, nextPassword, token, nextToken, sibling} {
			require.NotContains(t, data, secret)
		}
	}
	write := func(name, color, key, passwordArg string, ordinaryToken bool) {
		t.Helper()

		value := ""
		writeOnly := `, "X-Token" = var.token`

		if ordinaryToken {
			value = fmt.Sprintf(", value = sensitive(%q)", nextToken)
			writeOnly = ""
		}

		runtime.writeConfig(t, fmt.Sprintf(`
variable "password" {
 type = string
 sensitive = true
 ephemeral = true
}
variable "token" {
 type = string
 sensitive = true
 ephemeral = true
}
variable "sibling" {
 type = string
 sensitive = true
 ephemeral = true
}
resource "contentful_webhook" "test" {
 space_id = "space"
 url = "https://example.invalid/webhook"
 name = %q
 topics = ["Entry.publish"]
 http_basic_username = "user"
 %s
 headers = {
   "X-Color" = { value = %q }
   %q = { secret = true }
   "X-Token" = { secret = true %s }
 }
 header_values_wo = { %q = var.sibling %s }
 lifecycle { ignore_changes = [headers["X-Color"].value] }
}
`, name, passwordArg, color, key, value, key, writeOnly))
	}
	cli := func(command, password, header string, args ...string) string {
		t.Helper()

		flags := []string{command, "-input=false", "-no-color", "-var=password=" + password, "-var=token=" + header, "-var=sibling=" + sibling}
		output, err := runtime.run(t.Context(), append(flags, args...)...)
		checkSecrets(output)
		require.NoError(t, err, output)

		return output
	}
	state := func() []byte {
		t.Helper()

		b, err := os.ReadFile(filepath.Join(runtime.workingDirectory, "terraform.tfstate"))
		require.NoError(t, err)

		return b
	}

	const passwordArgument = "http_basic_password_wo = var.password" // #nosec G101 -- HCL variable reference.
	write("Created", "red", "X-Sibling", passwordArgument, false)
	cli("apply", initialPassword, token, "-auto-approve")
	checkSecrets(string(state()))

	instance := readWebhookRecoveryInstance(t, runtime)
	webhookID, ok := instance.Attributes["webhook_id"].(string)
	require.True(t, ok)
	require.NotEmpty(t, webhookID)

	const collection = "/spaces/space/webhook_definitions"

	color := map[string]any{"key": "X-Color", "value": "red", "secret": false}
	secretHeader := func(k string, value ...string) map[string]any {
		h := map[string]any{"key": k, "secret": true}
		if len(value) > 0 {
			h["value"] = value[0]
		}

		return h
	}
	body := func(name, password string, headers ...map[string]any) string {
		return testJSON(map[string]any{"name": name, "url": "https://example.invalid/webhook", "active": true, "topics": []string{"Entry.publish"}, "httpBasicUsername": "user", "httpBasicPassword": password, "filters": nil, "transformation": nil, "headers": headers})
	}
	want := make([]webhookMutation, 0, 4)
	want = append(want, webhookMutation{method: http.MethodPost, path: collection, body: body("Created", initialPassword, color, secretHeader("X-Sibling", sibling), secretHeader("X-Token", token))})
	require.Equal(t, want, fixture.mutations())
	// The rotation is supplied only at apply time. The ignored ordinary
	// value comes from the effective Plan even when a sibling changes.
	write("Updated", "ignored-blue", "X-Sibling", passwordArgument, false)
	cli("plan", initialPassword, token, "-out=update.tfplan")

	out, err := runtime.run(t.Context(), "show", "-json", "update.tfplan")
	require.NoError(t, err, out)
	checkSecrets(out)

	archive, err := zip.OpenReader(filepath.Join(runtime.workingDirectory, "update.tfplan"))
	require.NoError(t, err)

	for _, member := range archive.File {
		reader, err := member.Open()
		require.NoError(t, err)
		content, err := io.ReadAll(reader)
		require.NoError(t, err)
		require.NoError(t, reader.Close())
		checkSecrets(string(content))
	}

	require.NoError(t, archive.Close())
	cli("apply", nextPassword, nextToken, "update.tfplan")
	want = append(want, webhookMutation{method: http.MethodPut, path: collection + "/" + webhookID, version: "1", body: body("Updated", nextPassword, color, secretHeader("X-Sibling"), secretHeader("X-Token", nextToken))})
	require.Equal(t, want, fixture.mutations())
	checkSecrets(string(state()))
	cli("plan", nextPassword, nextToken, "-detailed-exitcode", "-out=noop.tfplan")
	cli("apply", initialPassword, token, "noop.tfplan")
	require.Equal(t, want, fixture.mutations())

	// A rename removes the old key and sends the same bytes at the new key.
	write("Updated", "ignored-blue", "X-Other", passwordArgument, false)
	cli("apply", nextPassword, nextToken, "-auto-approve")
	want = append(want, webhookMutation{method: http.MethodPut, path: collection + "/" + webhookID, version: "2", body: body("Updated", nextPassword, color, secretHeader("X-Other", sibling), secretHeader("X-Token"))})
	require.Equal(t, want, fixture.mutations())
	checkSecrets(string(state()))

	logs, err := os.ReadFile(runtime.logPath)
	require.NoError(t, err)
	checkSecrets(string(logs))

	// Migrate two fields in both directions with an independent write-only
	// sibling. Only ordinary storage contains their acknowledged plaintext.
	write("Updated", "ignored-blue", "X-Other", fmt.Sprintf("http_basic_password = %q", nextPassword), true)
	cli("apply", nextPassword, nextToken, "-auto-approve")
	require.Equal(t, want, fixture.mutations())

	public := string(state())
	require.Contains(t, public, nextPassword)
	require.Contains(t, public, nextToken)
	require.NotContains(t, public, sibling)
	write("Updated", "ignored-blue", "X-Other", passwordArgument, false)
	cli("apply", nextPassword, nextToken, "-auto-approve")
	checkSecrets(string(state()))
	cli("plan", nextPassword, nextToken, "-detailed-exitcode")
	require.Equal(t, want, fixture.mutations())

	// Omitting both Basic credentials and explicitly emptying headers clears
	// the remote definition and removes the last path-bound verifiers.
	runtime.writeConfig(t, `
resource "contentful_webhook" "test" {
 space_id = "space"
 name = "Updated"
 url = "https://example.invalid/webhook"
 topics = ["Entry.publish"]
 headers = {}
}
`)
	output, err := runtime.run(t.Context(), "apply", "-no-color", "-input=false", "-auto-approve")
	require.NoError(t, err, output)
	checkSecrets(output)

	cleared := testJSONObject(body("Updated", ""))
	cleared["httpBasicPassword"] = nil
	cleared["httpBasicUsername"] = nil
	cleared["headers"] = []any{}
	want = append(want, webhookMutation{method: http.MethodPut, path: collection + "/" + webhookID, version: "3", body: testJSON(cleared)})
	require.Equal(t, want, fixture.mutations())
	require.Empty(t, readWebhookRecoveryInstance(t, runtime).verifiers)
}
