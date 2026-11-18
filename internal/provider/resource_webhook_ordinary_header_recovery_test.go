package provider_test

import (
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
