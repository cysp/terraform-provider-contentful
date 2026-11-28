package provider_test

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

var errWebhookBasicPasswordLifecycle = errors.New("webhook Basic password lifecycle mismatch")

var rawWebhookWithoutBasicPassword = testJSON(map[string]any{
	"sys": map[string]any{
		"space":   map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Space", "id": "space"}},
		"type":    "WebhookDefinition",
		"id":      "raw-webhook",
		"version": 1,
	},
	"name":              "Basic webhook",
	"url":               "https://example.com/webhook",
	"topics":            []any{"Entry.publish"},
	"filters":           nil,
	"httpBasicUsername": "user",
	"headers":           []any{},
	"transformation":    nil,
	"active":            true,
})

func TestAccWebhookResourceBasicPasswordOmittedFromRawCMAResponse(t *testing.T) {
	t.Parallel()

	config := webhookBasicPasswordConfig("Basic webhook", "terraform_data.password.output", "") + `
resource "terraform_data" "password" {
  input = "basic-password-test-sentinel"
}
`
	testAccMockedResource(t, rawWebhookPasswordOmissionHandler{}, resource.TestCase{Steps: []resource.TestStep{
		{
			Config: config,
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectUnknownValue("contentful_webhook.test", tfjsonpath.New("http_basic_password")),
			}},
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("contentful_webhook.test", tfjsonpath.New("http_basic_password"), knownvalue.StringExact("basic-password-test-sentinel")),
			},
		},
		{
			Config: config,
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionNoop),
			}},
		},
	}})
}

type rawWebhookPasswordOmissionHandler struct{}

func (rawWebhookPasswordOmissionHandler) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
	responseWriter.Header().Set("Content-Type", "application/json")

	switch {
	case request.Method == http.MethodPost && request.URL.Path == "/spaces/space/webhook_definitions":
		responseWriter.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(responseWriter, rawWebhookWithoutBasicPassword)
	case request.Method == http.MethodGet && request.URL.Path == "/spaces/space/webhook_definitions/raw-webhook":
		_, _ = io.WriteString(responseWriter, rawWebhookWithoutBasicPassword)
	case request.Method == http.MethodDelete && request.URL.Path == "/spaces/space/webhook_definitions/raw-webhook":
		responseWriter.WriteHeader(http.StatusNoContent)
	default:
		http.NotFound(responseWriter, request)
	}
}

func TestAccWebhookResourceBasicPasswordLifecycle(t *testing.T) {
	t.Parallel()

	server := newTestCMAServer(t)
	server.RegisterSpaceEnvironment("space", "master")

	passwordPath := tfjsonpath.New("http_basic_password")
	tests := []struct {
		webhookName string
		password    string
		action      plancheck.ResourceActionType
		stateValue  knownvalue.Check
		storedValue cm.OptNilString
	}{
		{"Created webhook", `"password-one"`, plancheck.ResourceActionCreate, knownvalue.StringExact("password-one"), cm.NewOptNilString("password-one")},
		{"Created webhook", `"password-one"`, plancheck.ResourceActionNoop, knownvalue.StringExact("password-one"), cm.NewOptNilString("password-one")},
		{"Updated webhook", `"password-one"`, plancheck.ResourceActionUpdate, knownvalue.StringExact("password-one"), cm.NewOptNilString("password-one")},
		{"Updated webhook", `"password-one"`, plancheck.ResourceActionNoop, knownvalue.StringExact("password-one"), cm.NewOptNilString("password-one")},
		{"Updated webhook", `"password-two"`, plancheck.ResourceActionUpdate, knownvalue.StringExact("password-two"), cm.NewOptNilString("password-two")},
		{"Updated webhook", `"password-two"`, plancheck.ResourceActionNoop, knownvalue.StringExact("password-two"), cm.NewOptNilString("password-two")},
		{"Updated webhook", "null", plancheck.ResourceActionUpdate, knownvalue.Null(), cm.NewOptNilStringNull()},
		{"Updated webhook", "null", plancheck.ResourceActionNoop, knownvalue.Null(), cm.NewOptNilStringNull()},
	}

	steps := make([]resource.TestStep, 0, len(tests))
	for _, test := range tests {
		steps = append(steps, resource.TestStep{
			Config: webhookBasicPasswordConfig(test.webhookName, test.password, ""),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_webhook.test", test.action),
			}},
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("contentful_webhook.test", passwordPath, test.stateValue),
			},
			Check: checkStoredWebhookPassword(server, test.storedValue),
		})
	}

	testAccMockedResource(t, server, resource.TestCase{Steps: steps})
}

func TestAccWebhookResourceBasicPasswordDelegationAfterImport(t *testing.T) {
	t.Parallel()

	server := newTestCMAServer(t)
	server.RegisterSpaceEnvironment("space", "master")
	server.SetWebhookDefinition("space", "imported-webhook", cm.WebhookDefinitionData{
		Name:              "Imported webhook",
		Active:            cm.NewOptBool(true),
		URL:               "https://example.com/webhook",
		Topics:            []string{"Entry.publish"},
		HttpBasicUsername: cm.NewOptNilString("user"),
		HttpBasicPassword: cm.NewOptNilString("remote-password"),
	})

	usernameOnly := webhookBasicAuthConfig("Imported webhook", `"user"`, "null", "")
	config := webhookBasicPasswordConfig("Imported webhook", `"configured-after-import"`, "")
	recorder := &taxonomyRequestBodyRecorder{next: server}
	testAccMockedResource(t, recorder, resource.TestCase{Steps: []resource.TestStep{
		{
			Config:             usernameOnly,
			ResourceName:       "contentful_webhook.test",
			ImportState:        true,
			ImportStateId:      "space/imported-webhook",
			ImportStatePersist: true,
			ImportStateCheck: func(states []*terraform.InstanceState) error {
				if len(states) != 1 {
					return fmt.Errorf("%w: imported %d webhook states, want 1", errWebhookBasicPasswordLifecycle, len(states))
				}

				if _, exists := states[0].Attributes["http_basic_password"]; exists {
					return fmt.Errorf("%w: import manufactured http_basic_password in state", errWebhookBasicPasswordLifecycle)
				}

				return nil
			},
		},
		{
			Config: usernameOnly,
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionNoop),
			}},
		},
		{
			Config: webhookBasicAuthConfig("Renamed webhook", `"user"`, "null", ""),
			Check:  checkStoredWebhookPassword(server, cm.NewOptNilString("remote-password")),
		},
		{
			Config: webhookBasicPasswordConfig("Ignored password", `"ignored-password"`, "lifecycle { ignore_changes = [http_basic_password] }"),
			Check:  checkStoredWebhookPassword(server, cm.NewOptNilString("remote-password")),
		},
		{
			Config:      webhookBasicAuthConfig("Ignored password", `"changed-user"`, "null", ""),
			ExpectError: regexp.MustCompile("Missing webhook Basic password"),
		},
		{
			Config: config,
			Check:  checkStoredWebhookPassword(server, cm.NewOptNilString("configured-after-import")),
		},
		{
			Config: usernameOnly,
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("contentful_webhook.test", tfjsonpath.New("http_basic_password"), knownvalue.Null()),
			},
			Check: checkStoredWebhookPassword(server, cm.NewOptNilString("configured-after-import")),
		},
		{
			Config: webhookBasicPasswordConfig("Imported webhook", "null", ""),
			Check:  checkStoredWebhookPassword(server, cm.NewOptNilStringNull()),
		},
	}})

	updates := recorder.matchingRequests(http.MethodPut, "")
	require.Len(t, updates, 4)

	body := map[string]any{
		"url": "https://example.com/webhook", "topics": []string{"Entry.publish"},
		"active": true, "filters": nil, "transformation": nil,
	}
	checkPUT := func(index int, version, name string) {
		t.Helper()

		body["name"] = name

		require.Equal(t, "/spaces/space/webhook_definitions/imported-webhook", updates[index].path)
		require.Equal(t, version, updates[index].version)
		require.JSONEq(t, testJSON(body), string(updates[index].body))
	}
	checkPUT(0, "0", "Renamed webhook")
	checkPUT(1, "1", "Ignored password")

	body["httpBasicUsername"], body["httpBasicPassword"] = "user", "configured-after-import"

	checkPUT(2, "2", "Imported webhook")

	body["httpBasicUsername"], body["httpBasicPassword"] = nil, nil

	checkPUT(3, "3", "Imported webhook")
}

func TestAccWebhookResourceImportBlockClearsOmittedBasicCredentials(t *testing.T) {
	t.Parallel()

	server := newTestCMAServer(t)
	server.RegisterSpaceEnvironment("space", "master")
	recorder := &taxonomyRequestBodyRecorder{next: server}
	server.SetWebhookDefinition("space", "existing", cm.WebhookDefinitionData{
		Name: "Existing", URL: "https://example.com/webhook", Topics: []string{"Entry.publish"}, Active: cm.NewOptBool(false),
		HttpBasicUsername: cm.NewOptNilString("user"), HttpBasicPassword: cm.NewOptNilString("imported-password"),
	})

	config := webhookBasicAuthConfig("Existing", "null", "null", "active = false") + `
import {
  to = contentful_webhook.test
  id = "space/existing"
}
`
	testAccMockedResource(t, recorder, resource.TestCase{Steps: []resource.TestStep{{
		Config: config, Check: checkStoredWebhookPassword(server, cm.NewOptNilStringNull()),
	}}})

	puts := recorder.matchingRequests(http.MethodPut, "/spaces/space/webhook_definitions/existing")
	require.Len(t, puts, 1)
	require.Equal(t, "0", puts[0].version)
	require.JSONEq(t, `{"name":"Existing","url":"https://example.com/webhook","topics":["Entry.publish"],"active":false,"filters":null,"transformation":null,"httpBasicUsername":null,"httpBasicPassword":null}`, string(puts[0].body))
}

func TestAccWebhookResourceBasicPasswordIgnoreChangesUsesEffectivePlan(t *testing.T) {
	t.Parallel()

	server := newTestCMAServer(t)
	server.RegisterSpaceEnvironment("space", "master")

	testAccMockedResource(t, server, resource.TestCase{Steps: []resource.TestStep{
		{Config: webhookBasicPasswordConfig("Initial webhook", `"effective-plan-password"`, "")},
		{
			Config: webhookBasicAuthConfig("Updated webhook", `"user"`, "null", "lifecycle { ignore_changes = [http_basic_password] }"),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionUpdate),
				plancheck.ExpectKnownValue("contentful_webhook.test", tfjsonpath.New("http_basic_password"), knownvalue.StringExact("effective-plan-password")),
			}},
			Check: checkStoredWebhookPassword(server, cm.NewOptNilString("effective-plan-password")),
		},
		{
			Config: webhookBasicPasswordConfig("Updated webhook", `""`, "lifecycle { ignore_changes = [http_basic_password] }"),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_webhook.test", plancheck.ResourceActionNoop),
			}},
		},
	}})
}

func TestAccWebhookResourceIncompleteBasicCredentialsOnCreate(t *testing.T) {
	t.Parallel()

	server := newTestCMAServer(t)
	server.RegisterSpaceEnvironment("space", "master")
	counter := &webhookCreateCounter{next: server}
	testAccMockedResource(t, counter, resource.TestCase{Steps: []resource.TestStep{
		{
			Config:      webhookBasicAuthConfig("Webhook", `"user"`, "null", ""),
			ExpectError: regexp.MustCompile("Missing webhook Basic password"),
		},
		{
			Config:      webhookBasicAuthConfig("Webhook", "null", `"password"`, ""),
			ExpectError: regexp.MustCompile("Missing webhook Basic username"),
		},
	}})
	require.Zero(t, counter.creates.Load())
}

func webhookBasicPasswordConfig(name, password, lifecycle string) string {
	username := `"user"`
	if password == "null" {
		username = "null"
	}

	return webhookBasicAuthConfig(name, username, password, lifecycle)
}

func webhookBasicAuthConfig(name, username, password, lifecycle string) string {
	return fmt.Sprintf(`
resource "contentful_webhook" "test" {
  space_id            = "space"
  name                = %q
  url                 = "https://example.com/webhook"
  topics              = ["Entry.publish"]
  http_basic_username = %s
  http_basic_password = %s
  %s
}
`, name, username, password, lifecycle)
}

func checkStoredWebhookPassword(server *cmt.Server, expected cm.OptNilString) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState := state.RootModule().Resources["contentful_webhook.test"]
		stored, ok := server.StoredWebhookDefinition("space", resourceState.Primary.Attributes["webhook_id"])

		if !ok {
			return fmt.Errorf("%w: stored webhook not found", errWebhookBasicPasswordLifecycle)
		}

		if stored.HttpBasicPassword != expected {
			return fmt.Errorf("%w: stored password %#v, want %#v", errWebhookBasicPasswordLifecycle, stored.HttpBasicPassword, expected)
		}

		return nil
	}
}

func newWebhookBasicTestServer(t *testing.T) (*cmt.Server, *taxonomyRequestBodyRecorder, *mutationJSONResponseAdapter) {
	t.Helper()

	server := newTestCMAServer(t)
	server.RegisterSpaceEnvironment("space", "master")
	recorder := &taxonomyRequestBodyRecorder{next: server}

	return server, recorder, &mutationJSONResponseAdapter{delegate: recorder}
}

const webhookBasicAddress = "contentful_webhook.test"

func importExistingBasicWebhook(config string) resource.TestStep {
	return resource.TestStep{Config: config, ResourceName: webhookBasicAddress, ImportState: true, ImportStateId: "space/existing", ImportStatePersist: true}
}

func TestAccWebhookResourceDelegateLongImportedUsername(t *testing.T) {
	t.Parallel()
	server, recorder, adapter := newWebhookBasicTestServer(t)
	username := strings.Repeat("u", 51)
	server.SetWebhookDefinition("space", "existing", cm.WebhookDefinitionData{
		Name: "Existing", URL: "https://example.com/webhook", Topics: []string{"Entry.publish"},
		HttpBasicUsername: cm.NewOptNilString(username), HttpBasicPassword: cm.NewOptNilString("imported-password"),
	})
	config := webhookBasicAuthConfig("Renamed", strconv.Quote(username), "null", "")
	testAccMockedResource(t, adapter, resource.TestCase{Steps: []resource.TestStep{
		importExistingBasicWebhook(config),
		{Config: config, Check: checkStoredWebhookPassword(server, cm.NewOptNilString("imported-password"))},
	}})

	puts := recorder.matchingRequests(http.MethodPut, "/spaces/space/webhook_definitions/existing")
	require.Len(t, puts, 1)
	require.Equal(t, "0", puts[0].version)
	body := testJSONObject(string(puts[0].body))
	require.NotContains(t, body, "httpBasicUsername")
	require.NotContains(t, body, "httpBasicPassword")
}

func TestAccWebhookResourceDelegateImportedUnicodeUsername(t *testing.T) {
	t.Parallel()
	server, recorder, adapter := newWebhookBasicTestServer(t)

	const (
		decomposed = "e\u0301"
		composed   = "\u00e9"
		path       = "/spaces/space/webhook_definitions/existing"
	)
	server.SetWebhookDefinition("space", "existing", cm.WebhookDefinitionData{
		Name: "Imported", URL: "https://example.com/webhook", Topics: []string{"Entry.publish"}, Active: cm.NewOptBool(true),
		HttpBasicUsername: cm.NewOptNilString(decomposed), HttpBasicPassword: cm.NewOptNilString(rand.Text()),
	})

	config := func(name, extra string) string {
		return webhookBasicAuthConfig(name, "null", "null", "lifecycle { ignore_changes = [http_basic_username] }\n"+extra)
	}
	delegated := config("Renamed", "")
	rotate := config("Renamed", `http_basic_password_wo = "first-value"`)
	returnDecomposed := func() {
		adapter.mutateNext(http.MethodPut, func(doc map[string]any) { doc["httpBasicUsername"] = decomposed })
	}
	observations := &webhookProtocolObservations{}
	testObservedWebhookResource(t, adapter, observations, []resource.TestStep{
		importExistingBasicWebhook(delegated),
		{Config: delegated, PostApplyFunc: func() {
			puts := recorder.matchingRequests(http.MethodPut, path)
			require.Len(t, puts, 1)
			require.Equal(t, "0", puts[0].version)
			body := testJSONObject(string(puts[0].body))
			require.NotContains(t, body, "httpBasicUsername")
			require.NotContains(t, body, "httpBasicPassword")

			stored, exists := server.StoredWebhookDefinition("space", "existing")
			require.True(t, exists)
			require.Equal(t, decomposed, stored.HttpBasicUsername.Or(""))
		}},
		// Delegation preserves the raw spelling, but a password write must acknowledge the submitted bytes.
		{Config: rotate, PreConfig: returnDecomposed, ExpectError: regexp.MustCompile("(?s)Error running apply:.*Contentful returned a different webhook Basic username")},
		{Config: rotate, PreConfig: func() {
			puts := recorder.matchingRequests(http.MethodPut, path)
			require.Len(t, puts, 2)
			require.Equal(t, "1", puts[1].version)
			require.Equal(t, composed, testJSONObject(string(puts[1].body))["httpBasicUsername"])

			stored, exists := server.StoredWebhookDefinition("space", "existing")
			require.True(t, exists)
			require.Equal(t, composed, stored.HttpBasicUsername.Or(""))
			require.Nil(t, observations.lastApply().verifiers(t).Basic)
		}, PostApplyFunc: func() {
			require.NotNil(t, observations.lastApply().verifiers(t).Basic)
		}},
		{Config: config("Delegated", ""), PreConfig: returnDecomposed, PostApplyFunc: func() {
			require.Nil(t, observations.lastApply().verifiers(t).Basic)
		}},
	})

	puts := recorder.matchingRequests(http.MethodPut, path)
	require.Len(t, puts, 4)
	require.Equal(t, "2", puts[2].version)
	// Reasserting the committed value is a CMA no-op.
	require.Equal(t, "2", puts[3].version)
	require.NotContains(t, testJSONObject(string(puts[3].body)), "httpBasicPassword")
}

func TestAccWebhookResourceUnicodeWriteOnlyCredentials(t *testing.T) {
	t.Parallel()
	fixture := newSeededWebhookFixture(t)
	password, header := rand.Text(), rand.Text()
	config := webhookRecoveryConfig(true, "")
	decomposed := webhookRecoveryStep(config, password+"e\u0301", header+"e\u0301")
	composed := webhookRecoveryStep(config, password+"é", header+"é")
	composed.PlanOnly = true
	composed.PostApplyFunc = func() {
		fixture.requirePuts(t,
			webhookExpectedRecoveryPut("0", webhookPriorPassword, webhookPriorHeader, true),
			webhookExpectedRecoveryPut("1", password+"é", header+"é", true),
		)
		fixture.requireStored(t, password+"é", header+"é", true)
	}
	testAccMockedResource(t, fixture, resource.TestCase{Steps: []resource.TestStep{
		importedWebhookRecoveryStep(t, fixture), decomposed, composed,
	}})
}
