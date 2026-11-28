package provider_test

import (
	"fmt"
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
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/require"
)

func TestAccWebhookResourceMixedEmptyBasicCredentials(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")
	recorder := &taxonomyRequestBodyRecorder{next: server}
	testAccMockedResource(t, recorder, resource.TestCase{Steps: []resource.TestStep{
		{
			Config:   webhookBasicAuthConfig("Webhook", `""`, `"password"`, ""),
			PlanOnly: true, ExpectError: regexp.MustCompile("Incomplete webhook Basic credentials"),
		},
		{
			Config:   webhookBasicAuthConfig("Webhook", `"user"`, "null", `http_basic_password_wo = ""`),
			PlanOnly: true, ExpectError: regexp.MustCompile("Incomplete webhook Basic credentials"),
		},
		{
			Config: webhookBasicAuthConfig("Webhook", `"user"`, "terraform_data.password.output", "") + `
resource "terraform_data" "password" {
  input = ""
}
`,
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectUnknownValue("contentful_webhook.test", tfjsonpath.New("http_basic_password")),
			}},
			ExpectError: regexp.MustCompile("(?s)Error running apply:.*Incomplete webhook Basic credentials"),
		},
		{
			Config: webhookBasicAuthConfig("Webhook", `""`, `""`, ""),
			Check:  checkStoredWebhookPassword(server, cm.NewOptNilString("")),
		},
	}})
	require.Len(t, recorder.matchingRequests(http.MethodPost, ""), 1)
}

func TestAccWebhookResourceReplacementRequiresBasicPassword(t *testing.T) {
	t.Parallel()
	server, recorder, adapter := newWebhookBasicTestServer(t)
	server.SetWebhookDefinition("space", "existing", cm.WebhookDefinitionData{
		Name: "Existing", URL: "https://example.com/webhook", Topics: []string{"Entry.publish"},
		HttpBasicUsername: cm.NewOptNilString("user"), HttpBasicPassword: cm.NewOptNilString("imported-password"),
	})

	config := webhookBasicAuthConfig("Existing", `"user"`, "null", webhookReplacementLifecycle)

	testAccMockedResource(t, adapter, resource.TestCase{Steps: []resource.TestStep{
		{Config: config + webhookTriggerBefore + `
import {
  to = contentful_webhook.test
  id = "space/existing"
}
`},
		{Config: config + webhookTriggerAfter, PlanOnly: true, ExpectError: regexp.MustCompile("Missing webhook Basic password")},
		{Config: config + webhookTriggerBefore, PlanOnly: true, PreConfig: func() {
			require.Empty(t, recorder.matchingRequests(http.MethodDelete, ""))
			require.Empty(t, recorder.matchingRequests(http.MethodPost, ""))
			require.Empty(t, recorder.matchingRequests(http.MethodPut, ""))

			prior, exists := server.StoredWebhookDefinition("space", "existing")
			require.True(t, exists)
			require.Equal(t, cm.NewOptNilString("imported-password"), prior.HttpBasicPassword)
		}},
	}})

	_, exists := server.StoredWebhookDefinition("space", "existing")
	require.False(t, exists)
}

func TestAccWebhookResourceRotatePasswordWithIgnoredUsername(t *testing.T) {
	t.Parallel()
	server, recorder, adapter := newWebhookBasicTestServer(t)
	server.SetWebhookDefinition("space", "existing", cm.WebhookDefinitionData{
		Name: "Existing", URL: "https://example.com/webhook", Topics: []string{"Entry.publish"},
		HttpBasicUsername: cm.NewOptNilString("external-user"), HttpBasicPassword: cm.NewOptNilString("imported-password"),
	})

	const lifecycle = `lifecycle {
  ignore_changes = [http_basic_username]
  replace_triggered_by = [terraform_data.trigger]
 }`

	config := func(user, password, extra string) string {
		return webhookBasicAuthConfig("Existing", user, password, lifecycle+"\n"+extra)
	}
	ordinary := config("null", `"ordinary-rotation"`, "")
	writeOnly := config("null", "null", `http_basic_password_wo = "write-only-rotation"`)
	testAccMockedResource(t, adapter, resource.TestCase{Steps: []resource.TestStep{
		importExistingBasicWebhook(config("null", "null", "") + webhookTriggerBefore),
		{Config: ordinary + webhookTriggerBefore},
		{Config: writeOnly + webhookTriggerBefore},
		{Config: writeOnly + webhookTriggerAfter, PlanOnly: true, ExpectError: regexp.MustCompile("Missing webhook Basic username")},
		{Config: webhookBasicPasswordConfig("Existing", "null", "") + webhookTriggerBefore, PreConfig: func() {
			require.Empty(t, recorder.matchingRequests(http.MethodDelete, ""))
			puts := recorder.matchingRequests(http.MethodPut, "/spaces/space/webhook_definitions/existing")
			require.Len(t, puts, 2)

			for i, password := range []string{"ordinary-rotation", "write-only-rotation"} {
				require.Equal(t, strconv.Itoa(i), puts[i].version)
				require.JSONEq(t, testJSON(map[string]any{
					"name": "Existing", "url": "https://example.com/webhook", "topics": []string{"Entry.publish"},
					"active": true, "filters": nil, "transformation": nil,
					"httpBasicUsername": "external-user", "httpBasicPassword": password,
				}), string(puts[i].body))
			}
		}},
		// The preliminary update retains a null username. Replacement uses both configured values.
		{Config: config(`"new-user"`, `"new-password"`, "") + webhookTriggerAfter, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
			plancheck.ExpectResourceAction(webhookBasicAddress, plancheck.ResourceActionDestroyBeforeCreate),
		}}},
	}})

	posts := recorder.matchingRequests(http.MethodPost, "/spaces/space/webhook_definitions")
	require.Len(t, posts, 1)
	created := testJSONObject(string(posts[0].body))
	require.Equal(t, "new-user", created["httpBasicUsername"])
	require.Equal(t, "new-password", created["httpBasicPassword"])
}

func TestAccWebhookResourceReplacementBasicLengthLimits(t *testing.T) {
	t.Parallel()
	_, recorder, adapter := newWebhookBasicTestServer(t)

	prior := webhookBasicAuthConfig("Webhook", "null", "null", webhookReplacementLifecycle) + webhookTriggerBefore
	cases := []struct{ username, ordinary, extra, dependency, attr string }{
		{username: strconv.Quote(strings.Repeat("😀", 26)), ordinary: `"password"`, attr: "http_basic_username"},
		{username: `"user"`, ordinary: strconv.Quote(strings.Repeat("😀", 128)), attr: "http_basic_password"},
		{username: `"user"`, ordinary: "null", extra: "http_basic_password_wo = " + strconv.Quote(strings.Repeat("x", 256)), attr: "http_basic_password_wo"},
		{username: strconv.Quote(strings.Repeat("u", 51)), ordinary: "terraform_data.prerequisite.output", dependency: `resource "terraform_data" "prerequisite" { input = "pending" }`, attr: "http_basic_username"},
	}

	steps := make([]resource.TestStep, 0, len(cases)+2)
	steps = append(steps, resource.TestStep{Config: prior})

	for _, test := range cases {
		config := webhookBasicAuthConfig("Webhook", test.username, test.ordinary, webhookReplacementLifecycle+"\n"+test.extra) + webhookTriggerAfter + "\n" + test.dependency
		steps = append(steps, resource.TestStep{Config: config, PlanOnly: true, ExpectError: regexp.MustCompile("(?s)Webhook Basic credential too long.*" + test.attr)})
	}

	steps = append(steps, resource.TestStep{Config: prior, PlanOnly: true, PreConfig: func() {
		require.Empty(t, recorder.matchingRequests(http.MethodDelete, ""))
		require.Len(t, recorder.matchingRequests(http.MethodPost, ""), 1)
	}})
	testAccMockedResource(t, adapter, resource.TestCase{Steps: steps})
}

func TestAccWebhookResourceReplaceIgnoredEmptyBasicCredential(t *testing.T) {
	t.Parallel()

	for _, attr := range []string{"http_basic_username", "http_basic_password"} {
		t.Run(attr, func(t *testing.T) {
			t.Parallel()
			_, recorder, adapter := newWebhookBasicTestServer(t)

			lifecycle := fmt.Sprintf(`lifecycle {
  ignore_changes = [%s]
  replace_triggered_by = [terraform_data.trigger]
 }`, attr)
			update := webhookBasicAuthConfig("Webhook", `"new-user"`, `"new-password"`, lifecycle)
			testAccMockedResource(t, adapter, resource.TestCase{Steps: []resource.TestStep{
				{Config: webhookBasicAuthConfig("Webhook", `""`, `""`, "") + webhookTriggerBefore},
				{Config: update + webhookTriggerBefore, ExpectError: regexp.MustCompile("(?s)Error running apply:.*Incomplete webhook Basic credentials")},
				{Config: update + webhookTriggerAfter, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(webhookBasicAddress, plancheck.ResourceActionDestroyBeforeCreate),
				}}},
			}})
			require.Empty(t, recorder.matchingRequests(http.MethodPut, ""))
			posts := recorder.matchingRequests(http.MethodPost, "")
			require.Len(t, posts, 2)
			body := testJSONObject(string(posts[1].body))
			require.Equal(t, "new-user", body["httpBasicUsername"])
			require.Equal(t, "new-password", body["httpBasicPassword"])
		})
	}
}

func TestAccWebhookResourceReplacementApplyTimePassword(t *testing.T) {
	// One ephemeral variable changes between the framework's saved plan and its apply.
	const variable = "TF_VAR_replacement_password"

	atLimit := strings.Repeat("😀", 127) + "x"
	t.Setenv(variable, atLimit)
	_, recorder, adapter := newWebhookBasicTestServer(t)
	config := webhookBasicAuthConfig("Webhook", strconv.Quote(strings.Repeat("😀", 25)), "null", `http_basic_password_wo = var.replacement_password`) + `
variable "replacement_password" {
  type = string
  sensitive = true
  ephemeral = true
}
`

	var rejected bool

	testAccMockedResource(t, adapter, resource.TestCase{ErrorCheck: func(err error) error {
		message := err.Error()
		if strings.Contains(message, atLimit) {
			t.Fatal("replacement diagnostic exposed the ephemeral password")
		}

		if !regexp.MustCompile("(?s)Error running apply:.*Webhook Basic credential too long.*http_basic_password_wo").MatchString(message) {
			return err
		}

		rejected = true

		return nil
	}, Steps: []resource.TestStep{
		{Config: config},
		{
			Config: config, Taint: []string{webhookBasicAddress},
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction(webhookBasicAddress, plancheck.ResourceActionDestroyBeforeCreate),
				afterPlan(func() { t.Setenv(variable, atLimit+"x") }),
			}},
		},
	}})
	require.True(t, rejected, "apply accepted an oversized ephemeral replacement password")
	require.Len(t, recorder.matchingRequests(http.MethodDelete, ""), 1)
	require.Len(t, recorder.matchingRequests(http.MethodPost, ""), 1)
}

func TestAccWebhookResourceUnknownBasicPasswordSelection(t *testing.T) {
	t.Parallel()

	const conditional = `terraform_data.choice.output ? "selected" : null`
	for _, test := range []struct {
		name, ordinary, writeOnly string
		wantConflict              bool
	}{
		{name: "ordinary", ordinary: `"selected"`, writeOnly: conditional},
		{name: "write-only", ordinary: conditional, writeOnly: `"selected"`},
		{name: "conflicting", ordinary: `"selected"`, writeOnly: conditional, wantConflict: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
			require.NoError(t, err)
			server.RegisterSpaceEnvironment("space", "master")
			recorder := &taxonomyRequestBodyRecorder{next: server}

			step := resource.TestStep{Config: webhookBasicAuthConfig("Webhook", `"user"`, test.ordinary, "http_basic_password_wo = "+test.writeOnly) +
				fmt.Sprintf("\nresource \"terraform_data\" \"choice\" { input = %t }\n", test.wantConflict)}
			if test.wantConflict {
				step.ExpectError = regexp.MustCompile("(?s)Error running apply:.*Conflicting webhook Basic passwords")
			}

			testAccMockedResource(t, recorder, resource.TestCase{Steps: []resource.TestStep{step}})

			posts := recorder.matchingRequests(http.MethodPost, "")
			if test.wantConflict {
				require.Empty(t, posts)

				return
			}

			require.Len(t, posts, 1)
			require.Equal(t, "selected", testJSONObject(string(posts[0].body))["httpBasicPassword"])
		})
	}
}

func TestAccWebhookResourceUnknownHeaderValueSelection(t *testing.T) {
	t.Parallel()

	const conditional = `terraform_data.choice.output ? "selected" : null`
	for _, test := range []struct {
		name, ordinary, writeOnly string
		wantConflict              bool
	}{
		{name: "ordinary", ordinary: `"selected"`, writeOnly: conditional},
		{name: "write-only", ordinary: conditional, writeOnly: `"selected"`},
		{name: "conflicting", ordinary: `"selected"`, writeOnly: conditional, wantConflict: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
			require.NoError(t, err)
			server.RegisterSpaceEnvironment("space", "master")
			recorder := &taxonomyRequestBodyRecorder{next: server}

			step := resource.TestStep{Config: webhookBasicAuthConfig("Webhook", "null", "null", fmt.Sprintf("headers = { X = { secret = true, value = %s, value_wo = %s } }", test.ordinary, test.writeOnly)) +
				fmt.Sprintf("\nresource \"terraform_data\" \"choice\" { input = %t }\n", test.wantConflict)}
			if test.wantConflict {
				step.ExpectError = regexp.MustCompile("(?s)Error running apply:.*Conflicting webhook header values")
			}

			testAccMockedResource(t, recorder, resource.TestCase{Steps: []resource.TestStep{step}})

			posts := recorder.matchingRequests(http.MethodPost, "")
			if test.wantConflict {
				require.Empty(t, posts)

				return
			}

			require.Len(t, posts, 1)
			require.Equal(t, []any{map[string]any{"key": "X", "secret": true, "value": "selected"}}, testJSONObject(string(posts[0].body))["headers"])
			require.Empty(t, recorder.matchingRequests(http.MethodPut, ""))
		})
	}
}

func TestAccWebhookResourceKnownPasswordLengthWithUnknownAlternative(t *testing.T) {
	t.Parallel()
	_, recorder, adapter := newWebhookBasicTestServer(t)

	prior := webhookBasicAuthConfig("Webhook", `"user"`, `"initial"`, "") + webhookTriggerBefore
	replacement := webhookBasicAuthConfig("Webhook", `"user"`, strconv.Quote(strings.Repeat("x", 256)), `http_basic_password_wo = terraform_data.choice.output ? "selected" : null
 lifecycle { replace_triggered_by = [terraform_data.trigger] }`) + `
resource "terraform_data" "choice" { input = false }
resource "terraform_data" "trigger" { input = 1 }
`
	testAccMockedResource(t, adapter, resource.TestCase{Steps: []resource.TestStep{
		{Config: prior},
		{Config: replacement, PlanOnly: true, ExpectError: regexp.MustCompile("Webhook Basic credential too long")},
		{Config: prior, PlanOnly: true, PreConfig: func() { require.Empty(t, recorder.matchingRequests(http.MethodDelete, "")) }},
	}})
}

func TestAccWebhookResourcePlanOversizedBasicCredential(t *testing.T) {
	t.Parallel()

	longPassword := strconv.Quote(strings.Repeat("x", 256))
	longUsername := strconv.Quote(strings.Repeat("u", 51))

	const (
		user   = `"user"`
		header = `headers = { X = { value = "new" } }`
	)

	priorOrdinary := webhookBasicAuthConfig("Webhook", user, `"initial"`, "")
	priorWriteOnly := webhookBasicAuthConfig("Webhook", user, "null", `http_basic_password_wo = "initial"`)
	writeOnly := "http_basic_password_wo = " + longPassword
	unknownAlternative := webhookBasicAuthConfig("Webhook", user, longPassword, `http_basic_password_wo = terraform_data.choice.output ? "selected" : null`) + `
resource "terraform_data" "choice" { input = false }
`

	for _, test := range []struct {
		name, prior, config, diagnostic string
	}{
		{"rename with ordinary password", priorWriteOnly, webhookBasicAuthConfig("Renamed", user, longPassword, ""), "255 UTF-16"},
		{"rename with write-only password", priorWriteOnly, webhookBasicAuthConfig("Renamed", user, "null", writeOnly), "255 UTF-16"},
		{"write-only rotation", priorWriteOnly, webhookBasicAuthConfig("Webhook", user, "null", writeOnly), "255 UTF-16"},
		{"ordinary rotation", priorOrdinary, webhookBasicAuthConfig("Webhook", user, longPassword, ""), "255 UTF-16"},
		{"ordinary password and new header", priorWriteOnly, webhookBasicAuthConfig("Webhook", user, longPassword, header), "255 UTF-16"},
		{"write-only password and new header", priorWriteOnly, webhookBasicAuthConfig("Webhook", user, "null", writeOnly+"\n"+header), "255 UTF-16"},
		{"changed username", priorOrdinary, webhookBasicAuthConfig("Webhook", longUsername, `"initial"`, ""), "50 UTF-16"},
		{"ordinary rotation with unknown alternative", priorOrdinary, unknownAlternative, "255 UTF-16"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
			require.NoError(t, err)
			server.RegisterSpaceEnvironment("space", "master")
			testAccMockedResource(t, server, resource.TestCase{Steps: []resource.TestStep{
				{Config: test.prior},
				{
					Config:   test.config,
					PlanOnly: true, ExpectError: regexp.MustCompile(test.diagnostic),
				},
			}})
		})
	}
}

func TestAccWebhookResourceMigrateOversizedBasicPasswordWithoutPUT(t *testing.T) {
	t.Parallel()
	server, recorder, adapter := newWebhookBasicTestServer(t)
	password := strings.Repeat("x", 256)
	server.SetWebhookDefinition("space", "existing", cm.WebhookDefinitionData{
		Name: "Existing", URL: "https://example.com/webhook", Topics: []string{"Entry.publish"}, Active: cm.NewOptBool(true),
		HttpBasicUsername: cm.NewOptNilString("user"), HttpBasicPassword: cm.NewOptNilString(password),
	})
	// Model a legacy ordinary value that exceeds the current write limit.
	adapter.mutateNext(http.MethodGet, func(doc map[string]any) { doc["httpBasicPassword"] = password })

	writeOnly := webhookBasicAuthConfig("Existing", `"user"`, "null", "http_basic_password_wo = "+strconv.Quote(password))
	ordinary := func(extra string) string {
		return webhookBasicAuthConfig("Existing", `"user"`, strconv.Quote(password), extra)
	}
	observations := &webhookProtocolObservations{}
	testObservedWebhookResource(t, adapter, observations, []resource.TestStep{
		importExistingBasicWebhook(writeOnly),
		{Config: writeOnly, ConfigStateChecks: []statecheck.StateCheck{
			statecheck.ExpectKnownValue(webhookBasicAddress, tfjsonpath.New("http_basic_password"), knownvalue.Null()),
		}, PostApplyFunc: func() { require.NotNil(t, observations.lastApply().verifiers(t).Basic) }},
		{Config: ordinary(`headers = { X = { value = "new" } }`), PlanOnly: true, ExpectError: regexp.MustCompile("Webhook Basic credential too long")},
		{Config: ordinary(""), ConfigStateChecks: []statecheck.StateCheck{
			statecheck.ExpectKnownValue(webhookBasicAddress, tfjsonpath.New("http_basic_password"), knownvalue.StringExact(password)),
		}, PostApplyFunc: func() { require.Nil(t, observations.lastApply().verifiers(t).Basic) }},
	})
	require.Empty(t, recorder.matchingRequests(http.MethodPut, ""))
}

func TestAccWebhookResourceReplaceIgnoredOversizedBasicPassword(t *testing.T) {
	t.Parallel()
	server, recorder, adapter := newWebhookBasicTestServer(t)
	password := strings.Repeat("x", 256)
	server.SetWebhookDefinition("space", "existing", cm.WebhookDefinitionData{
		Name: "Imported", URL: "https://example.com/webhook", Topics: []string{"Entry.publish"}, Active: cm.NewOptBool(true),
		HttpBasicUsername: cm.NewOptNilString("user"), HttpBasicPassword: cm.NewOptNilString(password),
	})
	adapter.mutateNext(http.MethodGet, func(doc map[string]any) { doc["httpBasicPassword"] = password })

	const lifecycle = `lifecycle {
  ignore_changes = [http_basic_password]
  replace_triggered_by = [terraform_data.trigger]
 }`

	config := func(name string) string { return webhookBasicAuthConfig(name, `"user"`, `"valid"`, lifecycle) }
	testAccMockedResource(t, adapter, resource.TestCase{Steps: []resource.TestStep{
		importExistingBasicWebhook(config("Imported") + webhookTriggerBefore),
		{Config: config("Imported") + webhookTriggerBefore},
		{Config: config("Replacement") + webhookTriggerAfter, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
			plancheck.ExpectResourceAction(webhookBasicAddress, plancheck.ResourceActionDestroyBeforeCreate),
		}}},
	}})
	require.Empty(t, recorder.matchingRequests(http.MethodPut, ""))
	posts := recorder.matchingRequests(http.MethodPost, "")
	require.Len(t, posts, 1)
	require.Equal(t, "valid", testJSONObject(string(posts[0].body))["httpBasicPassword"])
}

func TestAccWebhookResourceRemovedUsernameWithKnownWriteOnlyPassword(t *testing.T) {
	t.Parallel()
	_, recorder, adapter := newWebhookBasicTestServer(t)
	config := webhookBasicAuthConfig("Webhook", "null", "null", `http_basic_password_wo = "rotation"
 depends_on = [terraform_data.prerequisite]`) + `
resource "terraform_data" "prerequisite" { input = "scheduled" }
`
	testAccMockedResource(t, adapter, resource.TestCase{Steps: []resource.TestStep{
		{Config: webhookBasicAuthConfig("Webhook", `"user"`, `"initial"`, "")},
		{Config: config, PlanOnly: true, ExpectError: regexp.MustCompile("Missing webhook Basic username")},
	}})
	require.Empty(t, recorder.matchingRequests(http.MethodPut, ""))
}

func TestAccWebhookResourcePlanChangedIncompleteBasicPair(t *testing.T) {
	t.Parallel()

	for _, test := range []struct{ username, password, diagnostic string }{
		{`"new"`, `""`, "Incomplete webhook Basic credentials"},
		{"null", `"changed"`, "Missing webhook Basic username"},
		{`"new"`, "null", "Missing webhook Basic password"},
	} {
		t.Run(test.diagnostic, func(t *testing.T) {
			t.Parallel()
			_, recorder, adapter := newWebhookBasicTestServer(t)
			config := webhookBasicAuthConfig("Changed", test.username, test.password, "depends_on = [terraform_data.prerequisite]") + `
resource "terraform_data" "prerequisite" { input = "ready" }
`
			testAccMockedResource(t, adapter, resource.TestCase{Steps: []resource.TestStep{
				{Config: webhookBasicAuthConfig("Initial", `"user"`, `"initial"`, "")},
				{Config: config, PlanOnly: true, ExpectError: regexp.MustCompile(test.diagnostic)},
			}})
			require.Empty(t, recorder.matchingRequests(http.MethodPut, ""))
		})
	}
}

func TestAccWebhookResourcePlanEmptyUsernameWithWriteOnlyPassword(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")
	recorder := &taxonomyRequestBodyRecorder{next: server}
	testAccMockedResource(t, recorder, resource.TestCase{Steps: []resource.TestStep{
		{Config: webhookBasicAuthConfig("Original", `"user"`, "null", `http_basic_password_wo = "initial"`)},
		{
			Config:   webhookBasicAuthConfig("Original", `""`, "null", `http_basic_password_wo = "next"`),
			PlanOnly: true, ExpectError: regexp.MustCompile("Incomplete webhook Basic credentials"),
		},
	}})
	require.Empty(t, recorder.matchingRequests(http.MethodPut, ""))
}
