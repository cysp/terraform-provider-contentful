package provider_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testWebhookSigningSecretValue        = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaAb09+/=_-"
	testWebhookSigningSecretUpdatedValue = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbZy87-/=_+"
)

// This is a Terraform address, not a credential.
//
//nolint:gosec
const testWebhookSigningSecretAddress = "contentful_webhook_signing_secret.test"

type webhookSigningSecretRequest struct {
	method, path, body string
}
type webhookSigningSecretRecorder struct {
	handler  http.Handler
	mu       sync.Mutex
	requests []webhookSigningSecretRequest
}

func (r *webhookSigningSecretRecorder) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "failed to read test request", http.StatusInternalServerError)

		return
	}

	req.Body = io.NopCloser(bytes.NewReader(body))

	r.mu.Lock()
	r.requests = append(r.requests, webhookSigningSecretRequest{method: req.Method, path: req.URL.Path, body: string(body)})
	r.mu.Unlock()
	r.handler.ServeHTTP(w, req)
}

func (r *webhookSigningSecretRecorder) requireMutations(t *testing.T, paths, bodies []string, deletes int) {
	t.Helper()
	r.mu.Lock()
	defer r.mu.Unlock()

	var puts []webhookSigningSecretRequest

	deleteCount := 0

	for _, req := range r.requests {
		if req.method == http.MethodPut {
			puts = append(puts, req)
		}

		if req.method == http.MethodDelete {
			deleteCount++

			assert.Empty(t, req.body)
		}
	}

	require.Len(t, puts, len(bodies))

	for i, body := range bodies {
		assert.Equal(t, paths[i], puts[i].path)
		assert.JSONEq(t, body, puts[i].body)
	}

	assert.Equal(t, deletes, deleteCount)
}

func webhookSigningSecretConfig(spaceID, value, extra string) string {
	return fmt.Sprintf(`
resource "contentful_webhook_signing_secret" "test" {
  space_id = %q
  value = %q
  %s
}
`, spaceID, value, extra)
}

func setTestWebhookSigningSecret(t *testing.T, server *cmt.Server, spaceID, value string) {
	t.Helper()
	response, err := server.Handler().PutWebhookSigningSecret(t.Context(), &cm.WebhookSigningSecretRequestData{Value: value}, cm.PutWebhookSigningSecretParams{SpaceID: spaceID})
	require.NoError(t, err)

	switch response.(type) {
	case *cm.PutWebhookSigningSecretOK, *cm.PutWebhookSigningSecretCreated:
	default:
		t.Fatalf("unexpected signing secret response type %T", response)
	}
}

func webhookSigningSecretStateChecks(value knownvalue.Check) []statecheck.StateCheck {
	return []statecheck.StateCheck{
		statecheck.ExpectKnownValue(testWebhookSigningSecretAddress, tfjsonpath.New("id"), knownvalue.StringExact("space")),
		statecheck.ExpectKnownValue(testWebhookSigningSecretAddress, tfjsonpath.New("value"), value),
		statecheck.ExpectSensitiveValue(testWebhookSigningSecretAddress, tfjsonpath.New("value")),
		statecheck.ExpectIdentityValueMatchesState(testWebhookSigningSecretAddress, tfjsonpath.New("space_id")),
	}
}

func TestAccWebhookSigningSecretResourceLifecycle(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")
	setTestWebhookSigningSecret(t, server, "space", strings.Repeat("z", 64))
	recorder := &webhookSigningSecretRecorder{handler: server}
	initial := webhookSigningSecretConfig("space", testWebhookSigningSecretValue, "")
	rotated := webhookSigningSecretConfig("space", testWebhookSigningSecretUpdatedValue, "")

	steps := make([]resource.TestStep, 0, 6)

	steps = append(steps, []resource.TestStep{
		{Config: initial, ConfigStateChecks: webhookSigningSecretStateChecks(knownvalue.StringExact(testWebhookSigningSecretValue))},
		{Config: initial, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}}},
		{Config: rotated, ConfigStateChecks: webhookSigningSecretStateChecks(knownvalue.StringExact(testWebhookSigningSecretUpdatedValue))},
	}...)
	for _, externalValue := range []string{strings.Repeat("c", 60) + "/=_+", strings.Repeat("d", 64)} {
		steps = append(steps, resource.TestStep{
			PreConfig:         func() { setTestWebhookSigningSecret(t, server, "space", externalValue) },
			Config:            rotated,
			ConfigPlanChecks:  resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}},
			ConfigStateChecks: webhookSigningSecretStateChecks(knownvalue.StringExact(testWebhookSigningSecretUpdatedValue)),
		})
	}

	steps = append(steps, resource.TestStep{
		Config: webhookSigningSecretConfig("space", testWebhookSigningSecretValue, `timeouts = { read = "1m" }
lifecycle { ignore_changes = [value] }`),
		ConfigPlanChecks:  resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(testWebhookSigningSecretAddress, plancheck.ResourceActionUpdate)}},
		ConfigStateChecks: webhookSigningSecretStateChecks(knownvalue.StringExact(testWebhookSigningSecretUpdatedValue)),
	})
	testAccMockedResource(t, recorder, resource.TestCase{Steps: steps, CheckDestroy: func(_ *terraform.State) error {
		response, getErr := server.Handler().GetWebhookSigningSecret(t.Context(), cm.GetWebhookSigningSecretParams{SpaceID: "space"})
		require.NoError(t, getErr)
		require.IsType(t, &cm.ErrorStatusCode{}, response)

		return nil
	}})
	recorder.requireMutations(t, []string{"/spaces/space/webhook_settings/signing_secret", "/spaces/space/webhook_settings/signing_secret"}, []string{
		`{"value":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaAb09+/=_-"}`,
		`{"value":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbZy87-/=_+"}`,
	}, 1)
}

func TestAccWebhookSigningSecretResourceImport(t *testing.T) {
	t.Parallel()

	for _, mode := range []string{"cli", "id", "identity"} {
		for _, ignore := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/ignore=%t", mode, ignore), func(t *testing.T) {
				t.Parallel()

				server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
				require.NoError(t, err)
				server.RegisterSpaceEnvironment("space", "master")
				setTestWebhookSigningSecret(t, server, "space", strings.Repeat("z", 64))
				recorder := &webhookSigningSecretRecorder{handler: server}
				extra := ""

				var expected knownvalue.Check = knownvalue.StringExact(testWebhookSigningSecretValue)

				if ignore {
					extra = `lifecycle { ignore_changes = [value] }`
					expected = knownvalue.Null()
				}

				configuration := webhookSigningSecretConfig("space", testWebhookSigningSecretValue, extra)

				var steps []resource.TestStep

				switch mode {
				case "cli":
					steps = append(steps, resource.TestStep{Config: configuration, PlanOnly: true, ExpectNonEmptyPlan: true}, resource.TestStep{
						Config: configuration, ImportState: true, ImportStateId: "space", ImportStatePersist: true, ResourceName: testWebhookSigningSecretAddress,
						ImportStateCheck: testAccImportAttributes(map[string]string{"id": "space"}),
					})
				case "id":
					configuration += "\nimport {\n id = \"space\"\n to = contentful_webhook_signing_secret.test\n}\n"
				case "identity":
					configuration += "\nimport {\n identity = { space_id = \"space\" }\n to = contentful_webhook_signing_secret.test\n}\n"
				}

				steps = append(steps, resource.TestStep{Config: configuration, ConfigStateChecks: webhookSigningSecretStateChecks(expected)}, resource.TestStep{
					Config:            webhookSigningSecretConfig("space", testWebhookSigningSecretValue, extra+"\ntimeouts = { read = \"1m\" }"),
					ConfigStateChecks: webhookSigningSecretStateChecks(expected),
				})
				testAccMockedResource(t, recorder, resource.TestCase{Steps: steps})

				var paths, bodies []string
				if !ignore {
					paths = []string{"/spaces/space/webhook_settings/signing_secret"}
					bodies = []string{`{"value":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaAb09+/=_-"}`}
				}

				recorder.requireMutations(t, paths, bodies, 1)
			})
		}
	}
}

func TestAccWebhookSigningSecretResourceDisappearsAndReplacesScope(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "master")
	server.RegisterSpaceEnvironment("other", "master")
	recorder := &webhookSigningSecretRecorder{handler: server}
	initial := webhookSigningSecretConfig("space", testWebhookSigningSecretValue, "")
	testAccMockedResource(t, recorder, resource.TestCase{Steps: []resource.TestStep{
		{Config: initial},
		{PreConfig: func() {
			resp, deleteErr := server.Handler().DeleteWebhookSigningSecret(t.Context(), cm.DeleteWebhookSigningSecretParams{SpaceID: "space"})
			require.NoError(t, deleteErr)
			require.IsType(t, &cm.NoContent{}, resp)
		}, Config: initial, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(testWebhookSigningSecretAddress, plancheck.ResourceActionCreate)}}},
		{Config: webhookSigningSecretConfig("other", testWebhookSigningSecretValue, ""), ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(testWebhookSigningSecretAddress, plancheck.ResourceActionDestroyBeforeCreate)}}, ConfigStateChecks: []statecheck.StateCheck{statecheck.ExpectKnownValue(testWebhookSigningSecretAddress, tfjsonpath.New("id"), knownvalue.StringExact("other"))}},
	}})
	recorder.requireMutations(t, []string{"/spaces/space/webhook_settings/signing_secret", "/spaces/space/webhook_settings/signing_secret", "/spaces/other/webhook_settings/signing_secret"}, []string{
		`{"value":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaAb09+/=_-"}`,
		`{"value":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaAb09+/=_-"}`,
		`{"value":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaAb09+/=_-"}`,
	}, 2)
}
