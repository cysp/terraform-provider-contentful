package provider_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/stretchr/testify/require"
)

// A forbidden rune makes ExpectError check the entire rendered error, including
// any HCL source Terraform attaches to an attribute diagnostic.
const diagnosticCredential = "§synthetic-webhook-credential"

func webhookErrorWithoutCredential(detail string) *regexp.Regexp {
	words := strings.Fields(regexp.QuoteMeta(detail))

	return regexp.MustCompile("^[^§]*" + strings.Join(words, `\s+`) + "[^§]*$")
}

func TestAccWebhookResourceCredentialDiagnosticSource(t *testing.T) {
	t.Parallel()

	literal := strconv.Quote(diagnosticCredential)

	long := strconv.Quote(strings.Repeat(diagnosticCredential, 12))
	for _, test := range []struct{ name, ordinary, extra, summary string }{
		{"ordinary password length", long, "", "Webhook Basic credential too long"},
		{"write-only password length", "null", "http_basic_password_wo = " + long, "Webhook Basic credential too long"},
		{"conflicting passwords", literal, "http_basic_password_wo = " + literal, "Conflicting webhook Basic passwords"},
		{"write-only readable header", "null", `headers = { "X-Managed" = { secret = false, value_wo = ` + literal + ` } }`, "Write-only webhook header must be secret"},
		{"null header beside managed header", "null", `headers = { Missing = null, Managed = { secret = true, value_wo = ` + literal + ` } }`, "Null webhook header"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			testAccMockedResource(t, newSeededWebhookFixture(t), resource.TestCase{Steps: []resource.TestStep{{
				Config:   webhookBasicAuthConfig("Webhook", `"user"`, test.ordinary, test.extra),
				PlanOnly: true, ExpectError: webhookErrorWithoutCredential(test.summary),
			}}})
		})
	}
}

func TestAccWebhookResourcePasswordReconciliationDiagnostic(t *testing.T) {
	t.Parallel()
	fixture := newSeededWebhookFixture(t)
	literal := strconv.Quote(diagnosticCredential)
	testAccMockedResource(t, fixture, resource.TestCase{Steps: []resource.TestStep{
		{Config: webhookBasicAuthConfig("Original", `"user"`, literal, "")},
		{
			Config: webhookBasicAuthConfig("Renamed", `"user"`, literal, ""),
			PreConfig: func() {
				webhookReplyFields(fixture, http.MethodPut, func(doc map[string]any) { doc["httpBasicPassword"] = nil })
			},
			ExpectError: webhookErrorWithoutCredential("Contentful returned a different webhook Basic password"),
		},
	}})
}

func TestAccWebhookResourceOperationErrorDiagnostics(t *testing.T) {
	t.Parallel()

	const (
		unauthorized = "HTTP 401. Contentful rejected authentication. Check the provider's Contentful management access token."
		rejected     = "HTTP 422. Contentful rejected the webhook definition. Review the configuration before applying again."
		ambiguous    = "Contentful did not acknowledge the webhook write. The operation may have reached Contentful. Reads cannot confirm secret bytes. Coordinate recovery with the webhook receiver before applying again."
		conflict     = "HTTP 409. Contentful reported a conflict on the versioned webhook update. Create a fresh Terraform plan and review any remote changes before applying again."
	)
	for _, test := range []struct {
		name, method, id        string
		status                  int
		credential, importState bool
		detail                  string
	}{
		{name: "credential-free create", method: http.MethodPost, status: 401, id: "AccessTokenInvalid", detail: unauthorized},
		{name: "credential-free read", method: http.MethodGet, status: 401, id: "AccessTokenInvalid", detail: unauthorized},
		{name: "credential-free update", method: http.MethodPut, status: 401, id: "AccessTokenInvalid", detail: unauthorized},
		{name: "credential-free delete", method: http.MethodDelete, status: 401, id: "AccessTokenInvalid", detail: unauthorized},
		{name: "rejected create", method: http.MethodPost, status: 422, id: "ValidationFailed", credential: true, detail: rejected},
		{name: "rejected update", method: http.MethodPut, status: 422, id: "ValidationFailed", credential: true, detail: rejected},
		{name: "versioned update conflict", method: http.MethodPut, status: 409, id: "Conflict", credential: true, detail: conflict},
		{name: "create conflict", method: http.MethodPost, status: 409, id: "Conflict", credential: true, detail: "HTTP 409. " + ambiguous},
		{name: "other update conflict", method: http.MethodPut, status: 409, id: "OtherConflict", credential: true, detail: "HTTP 409. " + ambiguous},
		{name: "other conflict status", method: http.MethodPut, status: 422, id: "Conflict", credential: true, detail: "HTTP 422. " + ambiguous},
		{name: "malformed update", method: http.MethodPut, status: 422, credential: true, detail: ambiguous},
		{name: "unauthorized create", method: http.MethodPost, status: 401, id: "AccessTokenInvalid", credential: true, detail: unauthorized},
		{name: "malformed unauthorized create", method: http.MethodPost, status: 401, credential: true, detail: ambiguous},
		{name: "unauthorized import", method: http.MethodGet, status: 401, id: "AccessTokenInvalid", credential: true, importState: true, detail: unauthorized},
		{name: "malformed unauthorized import", method: http.MethodGet, status: 401, credential: true, importState: true, detail: "Contentful did not confirm the webhook operation."},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			fixture := newSeededWebhookFixture(t)

			username, password := "null", "null"
			if test.credential {
				username, password = `"user"`, strconv.Quote(diagnosticCredential)
			}

			original := webhookBasicAuthConfig("Original", username, password, "")

			var steps []resource.TestStep
			if test.method != http.MethodPost && !test.importState {
				steps = append(steps, resource.TestStep{Config: original})
			}

			step := resource.TestStep{Config: original, ExpectError: webhookErrorWithoutCredential(test.detail)}
			switch {
			case test.importState:
				step.ImportState = true
				step.ImportStateId = "space/imported"
				step.ResourceName = "contentful_webhook.test"
			case test.method == http.MethodGet:
				step.PlanOnly = true
			case test.method == http.MethodPut:
				step.Config = webhookBasicAuthConfig("Renamed", username, password, "")
			case test.method == http.MethodDelete:
				step.Destroy = true
			}

			step.PreConfig = func() {
				fixture.onNextResponse(test.method, func(w http.ResponseWriter, _ *httptest.ResponseRecorder) {
					body := "{" + diagnosticCredential
					if test.id != "" {
						body = testJSON(map[string]any{
							"sys":     map[string]any{"type": "Error", "id": test.id},
							"message": diagnosticCredential, "details": map[string]any{"errors": []any{map[string]any{"name": "httpBasicPassword", "path": []any{}, "details": diagnosticCredential}}},
						})
					}

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(test.status)
					_, _ = io.WriteString(w, body)
				})
			}

			var observed webhookProtocolObservations
			testObservedWebhookResource(t, fixture, &observed, append(steps, step))

			calls := observed.applies
			if test.method == http.MethodGet {
				calls = observed.reads
			}

			requireWebhookOperationDiagnostic(t, calls, test.detail)
		})
	}
}

func requireWebhookOperationDiagnostic(t *testing.T, calls []webhookProtocolResult, detail string) {
	t.Helper()

	var sawError bool

	for _, call := range calls {
		if strings.Contains(call.logs, diagnosticCredential) {
			t.Fatal("provider logged credential bytes")
		}

		for _, d := range call.diagnostics {
			if d.Detail == detail {
				require.Nil(t, d.Attribute, "operation diagnostic must omit the source expression")

				sawError = true
			}
		}
	}

	require.True(t, sawError, "provider did not return the expected operation detail")
}

func TestAccWebhookResourceWriteOnlyStalePlan(t *testing.T) {
	t.Parallel()
	fixture := newSeededWebhookFixture(t)

	var (
		observed webhookProtocolObservations
		before   webhookProtocolResult
	)

	rotate := webhookRecoveryStep(webhookRecoveryConfig(true, ""), webhookRotatedPassword, webhookRotatedHeader)
	rotate.ConfigPlanChecks = resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
		plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionUpdate),
		afterPlan(func() {
			before = observed.lastRead()

			fixture.updateRemoteHeaders(t, 1, cm.WebhookDefinitionRequestHeaders{
				{Key: "X-Managed", Secret: cm.NewOptBool(true)},
				{Key: "X-Sibling", Secret: cm.NewOptBool(true), Value: cm.NewOptNilString("externally-rotated")},
			})
		}),
	}}
	rotate.ExpectError = webhookErrorWithoutCredential("HTTP 409. Contentful reported a conflict on the versioned webhook update. Create a fresh Terraform plan and review any remote changes before applying again.")
	// The final destroy follows the failed PUT. Inspect the checkpoint before the framework cleans up.
	next := webhookRecoveryStep(webhookRecoveryConfig(true, ""), webhookPriorPassword, webhookPriorHeader)
	next.PlanOnly = true
	next.PreConfig = func() {
		after := observed.lastApply()
		require.Equal(t, before.model(t), after.model(t))
		require.Equal(t, before.privateFields(t), after.privateFields(t))
		fixture.requirePuts(t, webhookExpectedRecoveryPut("0", webhookPriorPassword, webhookPriorHeader, true), webhookExpectedRecoveryPut("1", webhookRotatedPassword, webhookRotatedHeader, true))
		fields, ok := fixture.server.StoredWebhookDefinition("space", "imported")
		require.True(t, ok)
		require.Equal(t, 2, fields.Sys.Version)
		require.Equal(t, cm.NewOptNilString(webhookPriorPassword), fields.HttpBasicPassword)
		require.Equal(t, cm.WebhookDefinitionHeaders{
			{Key: "X-Managed", Secret: cm.NewOptBool(true), Value: cm.NewOptString(webhookPriorHeader)},
			{Key: "X-Sibling", Secret: cm.NewOptBool(true), Value: cm.NewOptString("externally-rotated")},
		}, fields.Headers)
	}
	testObservedWebhookResource(t, fixture, &observed, []resource.TestStep{importedWebhookRecoveryStep(t, fixture), rotate, next})
}
