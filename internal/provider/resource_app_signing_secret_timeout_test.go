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

func TestAccAppSigningSecretResourceTimeoutUpdatesPreserveRemoteRotation(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.SetAppDefinition("organization", "app", cm.AppDefinitionData{Name: "App"})

	initialValue, rotatedValue := strings.Repeat("a", 64), strings.Repeat("c", 64)
	recorder := &appSigningSecretPutRecorder{handler: server}
	rotateExternally := func() {
		server.SetAppSigningSecret("organization", "app", cm.AppSigningSecretRequestData{Value: strings.Repeat("b", 64)})
	}

	steps := []resource.TestStep{{
		Config: appSigningSecretTimeoutConfig(initialValue, ""),
		ConfigStateChecks: []statecheck.StateCheck{
			statecheck.ExpectKnownValue("contentful_app_signing_secret.test", tfjsonpath.New("value"), knownvalue.StringExact(initialValue)),
		},
	}}

	for _, change := range []struct {
		settings string
		timeouts knownvalue.Check
	}{
		{settings: `timeouts = { read = "1m" }`, timeouts: knownvalue.ObjectExact(map[string]knownvalue.Check{"create": knownvalue.Null(), "read": knownvalue.StringExact("1m"), "update": knownvalue.Null(), "delete": knownvalue.Null()})},
		{settings: `timeouts = { read = "2m" }`, timeouts: knownvalue.ObjectExact(map[string]knownvalue.Check{"create": knownvalue.Null(), "read": knownvalue.StringExact("2m"), "update": knownvalue.Null(), "delete": knownvalue.Null()})},
		{settings: "", timeouts: knownvalue.Null()},
	} {
		steps = append(steps, resource.TestStep{
			PreConfig: rotateExternally,
			Config:    appSigningSecretTimeoutConfig(initialValue, change.settings),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_app_signing_secret.test", plancheck.ResourceActionUpdate),
			}},
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("contentful_app_signing_secret.test", tfjsonpath.New("value"), knownvalue.StringExact(initialValue)),
				statecheck.ExpectKnownValue("contentful_app_signing_secret.test", tfjsonpath.New("timeouts"), change.timeouts),
				statecheck.ExpectIdentityValueMatchesState("contentful_app_signing_secret.test", tfjsonpath.New("organization_id")),
				statecheck.ExpectIdentityValueMatchesState("contentful_app_signing_secret.test", tfjsonpath.New("app_definition_id")),
			},
			Check: func(_ *terraform.State) error {
				recorder.requireValues(t, initialValue)
				requireAppSigningSecretRedactedValue(t, server, "bbbb")

				return nil
			},
		})
	}

	steps = append(steps,
		resource.TestStep{
			Config: appSigningSecretTimeoutConfig(rotatedValue, `timeouts = { read = "3m" }`),
			Check: func(_ *terraform.State) error {
				recorder.requireValues(t, initialValue, rotatedValue)
				requireAppSigningSecretRedactedValue(t, server, "cccc")

				return nil
			},
		},
		resource.TestStep{
			PreConfig: rotateExternally,
			Config: appSigningSecretTimeoutConfig(strings.Repeat("d", 64), `
timeouts = { read = "4m" }
lifecycle { ignore_changes = [value] }
`),
			ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("contentful_app_signing_secret.test", plancheck.ResourceActionUpdate),
			}},
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("contentful_app_signing_secret.test", tfjsonpath.New("value"), knownvalue.StringExact(rotatedValue)),
				statecheck.ExpectKnownValue("contentful_app_signing_secret.test", tfjsonpath.New("timeouts").AtMapKey("read"), knownvalue.StringExact("4m")),
			},
			Check: func(_ *terraform.State) error {
				recorder.requireValues(t, initialValue, rotatedValue)
				requireAppSigningSecretRedactedValue(t, server, "bbbb")

				return nil
			},
		},
	)

	testAccMockedResource(t, recorder, resource.TestCase{Steps: steps})
}

func TestAccAppSigningSecretResourceImportedTimeoutUpdatePreservesNullValue(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.SetAppDefinition("organization", "app", cm.AppDefinitionData{Name: "App"})
	server.SetAppSigningSecret("organization", "app", cm.AppSigningSecretRequestData{Value: strings.Repeat("b", 64)})
	recorder := &appSigningSecretPutRecorder{handler: server}
	steps := make([]resource.TestStep, 0, 2)

	for _, settings := range []string{"", `timeouts = { read = "2m" }`} {
		steps = append(steps, resource.TestStep{
			Config: `
import {
  id = "organization/app"
  to = contentful_app_signing_secret.test
}
` + appSigningSecretTimeoutConfig(strings.Repeat("a", 64), "lifecycle { ignore_changes = [value] }\n"+settings),
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue("contentful_app_signing_secret.test", tfjsonpath.New("value"), knownvalue.Null()),
			},
			Check: func(_ *terraform.State) error {
				recorder.requireValues(t)
				requireAppSigningSecretRedactedValue(t, server, "bbbb")

				return nil
			},
		})
	}

	testAccMockedResource(t, recorder, resource.TestCase{Steps: steps})
}

func appSigningSecretTimeoutConfig(value, settings string) string {
	return fmt.Sprintf(`
resource "contentful_app_signing_secret" "test" {
  organization_id = "organization"
  app_definition_id = "app"
  value = %q
  %s
}
`, value, settings)
}

func requireAppSigningSecretRedactedValue(t *testing.T, server *cmt.Server, expected string) {
	t.Helper()

	response, err := server.Handler().GetAppSigningSecret(t.Context(), cm.GetAppSigningSecretParams{OrganizationID: "organization", AppDefinitionID: "app"})
	require.NoError(t, err)

	secret, ok := response.(*cm.AppSigningSecret)
	require.True(t, ok, "unexpected signing secret response: %T", response)
	require.Equal(t, expected, secret.RedactedValue)
}

type appSigningSecretPutRecorder struct {
	handler http.Handler
	mu      sync.Mutex
	bodies  []string
}

func (r *appSigningSecretPutRecorder) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodPut {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		request.Body = io.NopCloser(bytes.NewReader(body))

		r.mu.Lock()
		r.bodies = append(r.bodies, request.URL.Path+" "+string(body))
		r.mu.Unlock()
	}

	r.handler.ServeHTTP(w, request)
}

func (r *appSigningSecretPutRecorder) requireValues(t *testing.T, values ...string) {
	t.Helper()

	r.mu.Lock()
	defer r.mu.Unlock()

	require.Len(t, r.bodies, len(values))

	for i, value := range values {
		const path = "/organizations/organization/app_definitions/app/signing_secret "
		require.True(t, strings.HasPrefix(r.bodies[i], path))
		assert.JSONEq(t, fmt.Sprintf(`{"value":%q}`, value), strings.TrimPrefix(r.bodies[i], path))
	}
}
