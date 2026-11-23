package provider_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// An ordinary value remains public even if Contentful's secret flag is true.
// Keep both refresh drift and a contradictory write independently observable.
func TestAccWebhookResourceOrdinaryHeaderRecovery(t *testing.T) {
	t.Parallel()
	runtime, fixture := newWebhookRecoveryFixture(t)
	runtime.writeConfig(t, `
resource "contentful_webhook" "test" {
 space_id = "space"
 name = "Ordinary header"
 url = "https://example.invalid/webhook"
 topics = ["Entry.publish"]
 headers = {"X-Ordinary" = {secret=true, value="configured-public-value"}}
}
`)

	apply := func(args ...string) (string, error) {
		return runtime.run(t.Context(), append([]string{"apply", "-input=false", "-auto-approve", "-no-color"}, args...)...)
	}
	returnedHeader := func(secret bool, value string) func(http.ResponseWriter, *httptest.ResponseRecorder) {
		return func(w http.ResponseWriter, rec *httptest.ResponseRecorder) {
			body := mutateTestJSON(rec.Body.String(), func(document map[string]any) {
				document["headers"] = []map[string]any{{"key": "X-Ordinary", "secret": secret, "value": value}}
			})
			_, _ = io.WriteString(w, body)
		}
	}
	output, err := apply()
	require.NoError(t, err, output)
	fixture.onNextResponse(http.MethodGet, returnedHeader(false, "changed-public-value"))

	output, err = apply("-refresh-only")
	require.NoError(t, err, output)
	instance := readWebhookRecoveryInstance(t, runtime)
	require.Equal(t, map[string]any{"X-Ordinary": map[string]any{"secret": false, "value": "changed-public-value", "value_wo": nil}}, instance.Attributes["headers"])
	require.Len(t, fixture.mutations(), 1)

	fixture.onNextResponse(http.MethodPut, returnedHeader(true, "contradictory-public-value"))

	output, err = apply("-refresh=false")
	require.Error(t, err, output)
	require.Contains(t, output, "Contentful returned different webhook headers")
	instance = readWebhookRecoveryInstance(t, runtime)
	require.Equal(t, map[string]any{"X-Ordinary": map[string]any{"secret": true, "value": "contradictory-public-value", "value_wo": nil}}, instance.Attributes["headers"])

	mutations := fixture.mutations()
	require.Len(t, mutations, 2)
	require.Equal(t, http.MethodPut, mutations[1].method)
}

// A contradictory name does not change the omission fallback for an ordinary
// credential. Exercise the same update with and without a write-only sibling.
func TestAccWebhookResourceOrdinaryCredentialRecovery(t *testing.T) {
	t.Parallel()

	for _, sibling := range []bool{false, true} {
		t.Run(fmt.Sprintf("write_only_sibling=%t", sibling), func(t *testing.T) {
			t.Parallel()
			runtime, fixture := newWebhookRecoveryFixture(t)

			declaration := ""
			if sibling {
				declaration = `"X-Sibling" = { secret = true, value_wo = var.sibling }`
			}

			configure := func(value string) {
				runtime.writeConfig(t, fmt.Sprintf(`
variable "sibling" {
 type = string
 sensitive = true
 ephemeral = true
}
resource "contentful_webhook" "test" {
 space_id = "space"
 name = "Configured"
 url = "https://example.invalid/webhook"
 topics = ["Entry.publish"]
 http_basic_username = "user"
 http_basic_password = %q
 headers = {
   "X-Managed" = {secret=true, value=%q}
   %s
 }
}
`, "password-"+value, "header-"+value, declaration))
			}
			cli := func(command string, options ...string) (string, error) {
				return runtime.run(t.Context(), append([]string{command, "-input=false", "-no-color", "-var=sibling=" + webhookSiblingHeader}, options...)...)
			}

			configure("before")

			out, err := cli("apply", "-auto-approve")
			require.NoError(t, err, out)
			before := readWebhookRecoveryInstance(t, runtime)
			webhookID, ok := before.Attributes["webhook_id"].(string)
			require.True(t, ok)
			configure("after")
			fixture.onNextResponse(http.MethodPut, func(w http.ResponseWriter, rec *httptest.ResponseRecorder) {
				body := mutateTestJSON(rec.Body.String(), func(doc map[string]any) { doc["name"] = "Contradictory" })
				_, _ = io.WriteString(w, body)
			})

			out, err = cli("apply", "-auto-approve")
			require.Error(t, err, out)
			require.Contains(t, out, "Contentful returned a different webhook name")

			writes := fixture.mutations()
			require.Len(t, writes, 2)

			headers := []map[string]any{{"key": "X-Managed", "secret": true, "value": "header-after"}}
			public := map[string]any{"X-Managed": map[string]any{"secret": true, "value": "header-after", "value_wo": nil}}

			if sibling {
				headers = append(headers, map[string]any{"key": "X-Sibling", "secret": true})
				public["X-Sibling"] = map[string]any{"secret": true, "value": nil, "value_wo": nil}

				require.NotEmpty(t, before.verifiers)
			}

			require.Equal(t, webhookMutation{method: http.MethodPut, path: "/spaces/space/webhook_definitions/" + webhookID, version: "0", body: testJSON(map[string]any{
				"active": true, "name": "Configured", "url": "https://example.invalid/webhook", "topics": []string{"Entry.publish"},
				"httpBasicUsername": "user", "httpBasicPassword": "password-after", "headers": headers, "filters": nil, "transformation": nil,
			})}, writes[1])
			after := readWebhookRecoveryInstance(t, runtime)
			require.Equal(t, webhookID, after.Attributes["webhook_id"])
			require.Equal(t, "1", testJSON(after.Attributes["version"]))
			require.Equal(t, "Contradictory", after.Attributes["name"])
			require.Equal(t, before.verifiers, after.verifiers)
			require.Equal(t, "password-after", after.Attributes["http_basic_password"])
			require.Equal(t, public, after.Attributes["headers"])

			out, err = cli("plan", "-detailed-exitcode")
			require.NoError(t, err, out)
			require.Equal(t, writes, fixture.mutations())
		})
	}
}
