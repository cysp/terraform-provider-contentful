package provider_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/stretchr/testify/require"
)

func TestAccWebhookResourceWriteOnlyLostUpdate(t *testing.T) {
	t.Parallel()

	runtime, fixture, first := importWebhookRecoveryFixture(t)
	prior := readWebhookRecoveryInstance(t, runtime).verifiers
	// The fixture commits the request before returning an inconclusive 503.
	fixture.onNextResponse(http.MethodPut, func(w http.ResponseWriter, _ *httptest.ResponseRecorder) {
		http.Error(w, "fixture discarded acknowledged response", http.StatusServiceUnavailable)
	})

	output, err := runWebhookRecoveryCLI(t, runtime, webhookRotatedPassword, webhookRotatedHeader, "apply", "-auto-approve")
	require.Error(t, err, output)
	require.Contains(t, output, "Failed to update webhook")

	second := webhookExpectedRecoveryPut("1", webhookRotatedPassword, webhookRotatedHeader, true)
	fixture.requirePuts(t, first, second)
	fixture.requireStored(t, webhookRotatedPassword, webhookRotatedHeader, true)
	require.Equal(t, prior, readWebhookRecoveryInstance(t, runtime).verifiers)

	output, err = runWebhookRecoveryCLI(t, runtime, webhookPriorPassword, webhookPriorHeader, "plan", "-detailed-exitcode")
	require.NoError(t, err, output)
	fixture.requirePuts(t, first, second)
	fixture.requireStored(t, webhookRotatedPassword, webhookRotatedHeader, true)

	// Recovery of the header still needs a complete Basic pair on the full PUT.
	runtime.writeConfig(t, strings.Replace(webhookRecoveryConfig(true, ""), "http_basic_password_wo = var.password", "", 1))
	output, err = runWebhookRecoveryCLI(t, runtime, webhookPriorPassword, webhookRotatedHeader, "apply", "-auto-approve")
	require.Error(t, err, output)
	require.Contains(t, output, "Missing webhook Basic password")
	fixture.requirePuts(t, first, second)
	require.Equal(t, prior, readWebhookRecoveryInstance(t, runtime).verifiers)
	runtime.writeConfig(t, webhookRecoveryConfig(true, ""))

	output, err = runWebhookRecoveryCLI(t, runtime, webhookRotatedPassword, webhookRotatedHeader, "apply", "-auto-approve")
	require.NoError(t, err, output)

	third := webhookExpectedRecoveryPut("2", webhookRotatedPassword, webhookRotatedHeader, true)
	fixture.requirePuts(t, first, second, third)
	require.NotEqual(t, prior, readWebhookRecoveryInstance(t, runtime).verifiers)
	runtime.writeConfig(t, webhookRecoveryConfig(false, ""))
	output, err = runWebhookRecoveryCLI(t, runtime, webhookRotatedPassword, webhookRotatedHeader, "apply", "-auto-approve")
	require.NoError(t, err, output)
	fixture.requirePuts(t, first, second, third, webhookExpectedRecoveryPut("2", webhookRotatedPassword, "", false))
	fixture.requireStored(t, webhookRotatedPassword, webhookRotatedHeader, false)
}

func TestAccWebhookResourceWriteOnlyReadableResponse(t *testing.T) {
	t.Parallel()

	runtime, fixture, first := importWebhookRecoveryFixture(t)
	prior := readWebhookRecoveryInstance(t, runtime).verifiers
	puts := []webhookMutation{first}

	// Faults share one webhook. Only a change to the stored definition increments its remote version.
	for _, method := range []string{http.MethodGet, http.MethodPut} {
		for _, field := range []string{"password", "header", "case alias"} {
			t.Log(method, field)

			echo := func(w http.ResponseWriter, rec *httptest.ResponseRecorder) {
				body := mutateTestJSON(rec.Body.String(), func(document map[string]any) {
					if field == "password" {
						document["httpBasicPassword"] = webhookRotatedPassword
						delete(document, "httpBasicUsername")

						return
					}

					key := "X-Managed"
					if field == "case alias" {
						key = "x-managed"
					}

					document["headers"] = []map[string]any{
						{"key": key, "secret": false, "value": webhookRotatedHeader},
						{"key": "X-Sibling", "secret": true},
					}
				})
				_, _ = io.WriteString(w, body)
			}

			fixture.onNextResponse(method, echo)

			options := []string{"-auto-approve"}
			if method == http.MethodGet {
				options = append(options, "-refresh-only")
			}

			output, err := runWebhookRecoveryCLI(t, runtime, webhookRotatedPassword, webhookRotatedHeader, "apply", options...)
			require.Error(t, err, output)

			if method == http.MethodPut {
				puts = append(puts, webhookExpectedRecoveryPut(strconv.Itoa(min(len(puts), 2)), webhookRotatedPassword, webhookRotatedHeader, true))
				fixture.requireStored(t, webhookRotatedPassword, webhookRotatedHeader, true)
				// A rejected PUT can checkpoint safe recovery state. Its next Read
				// must still recognize the redacted path if the same echo recurs.
				fixture.onNextResponse(http.MethodGet, echo)

				output, err = runWebhookRecoveryCLI(t, runtime, webhookRotatedPassword, webhookRotatedHeader, "apply", "-auto-approve", "-refresh-only")
				require.Error(t, err, output)
			}

			fixture.requirePuts(t, puts...)
			require.Equal(t, prior, readWebhookRecoveryInstance(t, runtime).verifiers)
		}
	}

	logs, err := os.ReadFile(runtime.logPath)
	require.NoError(t, err)
	requireNoWebhookRecoverySecrets(t, string(logs))
}

func TestAccWebhookResourceWriteOnlyCreateRecovery(t *testing.T) {
	t.Parallel()

	runtime, fixture := newWebhookRecoveryFixture(t)
	runtime.writeConfig(t, webhookRecoveryConfig(false, ""))
	fixture.onNextResponse(http.MethodPost, func(w http.ResponseWriter, rec *httptest.ResponseRecorder) {
		body := mutateTestJSON(rec.Body.String(), func(document map[string]any) {
			document["name"] = "Returned name"
		})
		w.WriteHeader(rec.Code)
		_, _ = io.WriteString(w, body)
	})

	output, err := runWebhookRecoveryCLI(t, runtime, webhookPriorPassword, webhookPriorHeader, "apply", "-auto-approve")
	require.Error(t, err, output)
	require.Contains(t, output, "Contentful returned a different webhook name")

	created := webhookExpectedRecoveryPut("", webhookPriorPassword, webhookPriorHeader, false)
	created.method = http.MethodPost
	created.path = "/spaces/space/webhook_definitions"
	require.Equal(t, []webhookMutation{created}, fixture.mutations())

	instance := readWebhookRecoveryInstance(t, runtime)
	require.Equal(t, "tainted", instance.Status)
	require.Equal(t, "Returned name", instance.Attributes["name"])
	require.Empty(t, instance.verifiers)
	webhookID, ok := instance.Attributes["webhook_id"].(string)
	require.True(t, ok)
	require.Equal(t, "space/"+webhookID, instance.Attributes["id"])
	stored, exists := fixture.server.StoredWebhookDefinition("space", webhookID)
	require.True(t, exists)
	require.Equal(t, cm.NewOptNilString(webhookPriorPassword), stored.HttpBasicPassword)
}

// The effective ordinary value authorizes publication of those bytes. A different
// readable response at the former write-only path could contain the old secret.
func TestAccWebhookResourceWriteOnlyHeaderBecomesOrdinary(t *testing.T) {
	t.Parallel()

	runtime, fixture, _ := importWebhookRecoveryFixture(t)
	runtime.writeConfig(t, webhookRecoveryConfig(true, "X-Managed"))
	prior := readWebhookRecoveryInstance(t, runtime).verifiers

	fixture.onNextResponse(http.MethodPut, func(w http.ResponseWriter, rec *httptest.ResponseRecorder) {
		body := mutateTestJSON(rec.Body.String(), func(document map[string]any) {
			document["headers"] = []map[string]any{
				{"key": "X-Managed", "secret": false, "value": webhookPriorHeader},
				{"key": "X-Sibling", "secret": true},
			}
		})
		_, _ = io.WriteString(w, body)
	})

	output, err := runWebhookRecoveryCLI(t, runtime, webhookPriorPassword, webhookPriorHeader, "apply", "-auto-approve")
	require.Error(t, err, output)
	require.Equal(t, prior, readWebhookRecoveryInstance(t, runtime).verifiers)
}

// A case-only rename has a new Terraform key. Publishing an explicitly
// configured ordinary value at that spelling is safe when the response agrees.
func TestAccWebhookResourceWriteOnlyCaseRenameToOrdinary(t *testing.T) {
	t.Parallel()

	runtime, fixture, first := importWebhookRecoveryFixture(t)
	runtime.writeConfig(t, webhookRecoveryConfig(true, "x-managed"))

	output, err := runWebhookRecoveryCLI(t, runtime, webhookPriorPassword, webhookPriorHeader, "apply", "-auto-approve")
	require.NoError(t, err, output)

	update := webhookExpectedRecoveryPut("1", webhookPriorPassword, "", true)
	update.body = mutateTestJSON(update.body, func(document map[string]any) {
		document["headers"] = []map[string]any{
			{"key": "X-Sibling", "secret": true},
			{"key": "x-managed", "secret": false, "value": "public-replacement"},
		}
	})
	fixture.requirePuts(t, first, update)

	instance := readWebhookRecoveryInstance(t, runtime)
	require.Equal(t, map[string]any{
		"X-Sibling": map[string]any{"secret": true, "value": nil, "value_wo": nil},
		"x-managed": map[string]any{"secret": false, "value": "public-replacement", "value_wo": nil},
	}, instance.Attributes["headers"])
}

func TestAccWebhookResourceWriteOnlyHeaderRecreatedAfterRefresh(t *testing.T) {
	t.Parallel()

	runtime, fixture, first := importWebhookRecoveryFixture(t)

	sibling := cm.WebhookDefinitionHeader{Key: "X-Sibling", Secret: cm.NewOptBool(true), Value: cm.NewOptString(webhookSiblingHeader)}
	fixture.updateRemoteHeaders(t, 1, cm.WebhookDefinitionHeaders{sibling})

	output, err := runWebhookRecoveryCLI(t, runtime, webhookPriorPassword, webhookPriorHeader, "apply", "-refresh-only", "-auto-approve")
	require.NoError(t, err, output)
	require.Equal(t, map[string]any{"X-Sibling": map[string]any{"secret": true, "value": nil, "value_wo": nil}}, readWebhookRecoveryInstance(t, runtime).Attributes["headers"])
	fixture.requirePuts(t, first)

	fixture.updateRemoteHeaders(t, 2, cm.WebhookDefinitionHeaders{{Key: "X-Managed", Secret: cm.NewOptBool(true), Value: cm.NewOptString("externally-recreated-secret")}, sibling})

	output, err = runWebhookRecoveryCLI(t, runtime, webhookPriorPassword, webhookPriorHeader, "apply", "-auto-approve")
	require.NoError(t, err, output)
	fixture.requirePuts(t, first, webhookExpectedRecoveryPut("3", webhookPriorPassword, webhookPriorHeader, true))
	fixture.requireStored(t, webhookPriorPassword, webhookPriorHeader, true)
}

func (fixture *webhookSecretRecoveryFixture) updateRemoteHeaders(t *testing.T, version int, headers cm.WebhookDefinitionHeaders) {
	t.Helper()
	response, err := fixture.server.Handler().UpdateWebhookDefinition(t.Context(), &cm.WebhookDefinitionData{
		Name: "Imported", URL: "https://example.invalid/webhook", Topics: []string{"Entry.publish"}, Active: cm.NewOptBool(true),
		HttpBasicUsername: cm.NewOptNilString("user"), HttpBasicPassword: cm.NewOptNilString(webhookPriorPassword),
		Headers: headers,
	}, cm.UpdateWebhookDefinitionParams{SpaceID: "space", WebhookDefinitionID: "imported", XContentfulVersion: cm.NewOptInt(version)})
	require.NoError(t, err)
	require.IsType(t, &cm.WebhookDefinitionStatusCode{}, response)
}

func TestAccWebhookResourceDelegatesHeaderByName(t *testing.T) {
	t.Parallel()
	runtime, fixture, first := importWebhookRecoveryFixture(t)

	const outside = "synthetic-externally-rotated-sibling"
	fixture.updateRemoteHeaders(t, 1, cm.WebhookDefinitionHeaders{
		{Key: "X-Managed", Secret: cm.NewOptBool(true)},
		{Key: "X-Sibling", Secret: cm.NewOptBool(true), Value: cm.NewOptString(outside)},
	})

	runtime.writeConfig(t, webhookRecoveryHeaderConfig("", "", `lifecycle { ignore_changes = [headers["X-Sibling"]] }`))
	output, err := runWebhookRecoveryCLI(t, runtime, webhookPriorPassword, webhookRotatedHeader, "apply", "-auto-approve")
	require.NoError(t, err, output)
	fixture.requirePuts(t, first, webhookExpectedRecoveryPut("2", webhookPriorPassword, webhookRotatedHeader, true))
	stored, exists := fixture.server.StoredWebhookDefinition("space", "imported")
	require.True(t, exists)
	require.Equal(t, cm.WebhookDefinitionHeaders{
		{Key: "X-Managed", Secret: cm.NewOptBool(true), Value: cm.NewOptString(webhookRotatedHeader)},
		{Key: "X-Sibling", Secret: cm.NewOptBool(true), Value: cm.NewOptString(outside)},
	}, stored.Headers)
	output, err = runWebhookRecoveryCLI(t, runtime, webhookPriorPassword, webhookRotatedHeader, "plan", "-detailed-exitcode")
	require.NoError(t, err, output)

	// Ignoring the write-only leaf suppresses a rotation-only plan. An update
	// scheduled by a different argument still receives apply-time Config.
	runtime.writeConfig(t, webhookRecoveryHeaderConfig("", "", `lifecycle { ignore_changes = [headers["X-Sibling"], headers["X-Managed"].value_wo] }`))
	output, err = runWebhookRecoveryCLI(t, runtime, webhookPriorPassword, webhookPriorHeader, "plan", "-detailed-exitcode")
	require.NoError(t, err, output)
	output, err = runWebhookRecoveryCLI(t, runtime, webhookRotatedPassword, webhookPriorHeader, "apply", "-auto-approve")
	require.NoError(t, err, output)
	fixture.requirePuts(t, first, webhookExpectedRecoveryPut("2", webhookPriorPassword, webhookRotatedHeader, true), webhookExpectedRecoveryPut("3", webhookRotatedPassword, webhookPriorHeader, true))
}

func TestAccWebhookResourceWriteOnlyStalePlan(t *testing.T) {
	t.Parallel()
	runtime, fixture, first := importWebhookRecoveryFixture(t)
	before := readWebhookRecoveryInstance(t, runtime)
	out, err := runWebhookRecoveryCLI(t, runtime, webhookRotatedPassword, webhookRotatedHeader, "plan", "-out=rotate.tfplan")
	require.NoError(t, err, out)
	fixture.updateRemoteHeaders(t, 1, cm.WebhookDefinitionHeaders{
		{Key: "X-Managed", Secret: cm.NewOptBool(true)},
		{Key: "X-Sibling", Secret: cm.NewOptBool(true), Value: cm.NewOptString("externally-rotated")},
	})
	out, err = runWebhookRecoveryCLI(t, runtime, webhookRotatedPassword, webhookRotatedHeader, "apply", "rotate.tfplan")
	require.Error(t, err, out)
	require.Contains(t, out, "HTTP 409")
	fixture.requirePuts(t, first, webhookExpectedRecoveryPut("1", webhookRotatedPassword, webhookRotatedHeader, true))
	after := readWebhookRecoveryInstance(t, runtime)
	require.Equal(t, before.Attributes, after.Attributes)
	require.Equal(t, before.Private, after.Private)

	fields, found := fixture.server.StoredWebhookDefinition("space", "imported")
	require.True(t, found)
	require.Equal(t, 2, fields.Sys.Version)
	require.Equal(t, cm.NewOptNilString(webhookPriorPassword), fields.HttpBasicPassword)
	require.Equal(t, cm.WebhookDefinitionHeaders{
		{Key: "X-Managed", Secret: cm.NewOptBool(true), Value: cm.NewOptString(webhookPriorHeader)},
		{Key: "X-Sibling", Secret: cm.NewOptBool(true), Value: cm.NewOptString("externally-rotated")},
	}, fields.Headers)
}
