package provider_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflogtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:ireturn
func signingSecretProtocolServer(t *testing.T, handler http.Handler) tfprotov6.ProviderServer {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	providerServer, err := makeTestAccProtoV6ProviderFactories(WithContentfulURL(server.URL), WithAccessToken(cmt.ValidAccessToken))["contentful"]()
	require.NoError(t, err)
	config, err := providerConfigDynamicValue(map[string]any{"url": server.URL, "access_token": cmt.ValidAccessToken})
	require.NoError(t, err)
	response, err := providerServer.ConfigureProvider(t.Context(), &tfprotov6.ConfigureProviderRequest{Config: &config})
	require.NoError(t, err)
	require.Empty(t, response.Diagnostics)

	return providerServer
}

func signingSecretDynamicValue(t *testing.T, kind string, valueWO any) (schema.Schema, tfprotov6.DynamicValue) {
	t.Helper()
	resourceSchema := AppSigningSecretResourceSchema(t.Context())
	scope := map[string]string{"organization_id": "organization", "app_definition_id": "app", "id": "organization/app"}

	if kind == "webhook" {
		resourceSchema = WebhookSigningSecretResourceSchema(t.Context())
		scope = map[string]string{"space_id": "space", "id": "space"}
	}

	typ := resourceSchema.Type().TerraformType(t.Context())
	object, ok := typ.(tftypes.Object)
	require.True(t, ok)

	values := map[string]tftypes.Value{}
	for name, attributeType := range object.AttributeTypes {
		values[name] = tftypes.NewValue(attributeType, nil)
	}

	for name, value := range scope {
		values[name] = tftypes.NewValue(tftypes.String, value)
	}

	values["value_wo"] = tftypes.NewValue(tftypes.String, valueWO)
	values["updated_at"] = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
	result, err := tfprotov6.NewDynamicValue(typ, tftypes.NewValue(typ, values))
	require.NoError(t, err)

	return resourceSchema, result
}

func TestSigningSecretWriteOnlyAcknowledgement(t *testing.T) {
	t.Parallel()

	for _, kind := range []string{"app", "webhook"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()

			testSigningSecretWriteOnlyAcknowledgement(t, kind)
		})
	}
}

//nolint:forcetypeassert // Fixture mutations target independently defined object shapes.
func testSigningSecretWriteOnlyAcknowledgement(t *testing.T, kind string) {
	t.Helper()

	initial, changed := strings.Repeat("a", 64), strings.Repeat("b", 64)
	fixture := &signingSecretFixture{kind: kind, withUpdatedAt: true}

	type failureCase struct {
		name, summary, attribute                            string
		echo, invalidTimestamp, invalidPrivate, wrongSuffix bool
	}

	fault := failureCase{}
	requests := 0
	expected := initial
	providerServer := signingSecretProtocolServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++

		assert.Equal(t, http.MethodPut, r.Method)
		assert.Empty(t, r.Header.Get("X-Contentful-Version"))
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		expectedBody, err := json.Marshal(map[string]string{"value": expected})
		assert.NoError(t, err)
		assert.JSONEq(t, string(expectedBody), string(body))
		r.Body = io.NopCloser(bytes.NewReader(body))

		if fault.echo {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)

			response := cmt.NewContentfulManagementError("BadRequest", &changed, nil)
			assert.NoError(t, json.NewEncoder(w).Encode(&response))

			return
		}

		recorder := httptest.NewRecorder()
		fixture.ServeHTTP(recorder, r)

		response := recorder.Body.String()

		if fault.attribute != "" {
			response = mutateTestJSON(response, func(document map[string]any) {
				link := "space"

				switch fault.attribute {
				case "organization_id":
					link = "organization"
				case "app_definition_id":
					link = "appDefinition"
				}

				document["sys"].(map[string]any)[link].(map[string]any)["sys"].(map[string]any)["id"] = "other"
			})
		}

		if fault.wrongSuffix {
			response = mutateTestJSON(response, func(document map[string]any) { document["redactedValue"] = "zzzz" })
		}

		if fault.invalidTimestamp {
			response = mutateTestJSON(response, func(document map[string]any) {
				document["sys"].(map[string]any)["updatedAt"] = "invalid"
			})
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, response)
	}))
	resourceSchema, plan := signingSecretDynamicValue(t, kind, nil)
	_, config := signingSecretDynamicValue(t, kind, initial)
	prior := nullResourceDynamicValue(t, resourceSchema)
	response, err := providerServer.ApplyResourceChange(t.Context(), &tfprotov6.ApplyResourceChangeRequest{TypeName: "contentful_" + kind + "_signing_secret", PriorState: &prior, PlannedState: &plan, Config: &config})
	require.NoError(t, err)
	require.Empty(t, response.Diagnostics)
	require.Equal(t, 1, requests)

	require.NotNil(t, response.NewIdentity)
	acknowledged := response
	raw, err := response.NewState.Unmarshal(resourceSchema.Type().TerraformType(t.Context()))
	require.NoError(t, err)

	attrs := map[string]tftypes.Value{}
	require.NoError(t, raw.As(&attrs))
	require.True(t, attrs["value"].IsNull())
	require.True(t, attrs["value_wo"].IsNull())

	private := map[string][]byte{}
	require.NoError(t, json.Unmarshal(response.Private, &private))
	require.Contains(t, string(private["write_only_secret_hashes"]), "$argon2id$v=19$m=65536,t=3,p=4$")
	require.NotContains(t, string(private["write_only_secret_hashes"]), initial)
	// Same acknowledged value retains the exact salt and encoded verifier.
	response, err = providerServer.ApplyResourceChange(t.Context(), &tfprotov6.ApplyResourceChangeRequest{TypeName: "contentful_" + kind + "_signing_secret", PriorState: acknowledged.NewState, PlannedState: &plan, Config: &config, PlannedPrivate: acknowledged.Private, PlannedIdentity: acknowledged.NewIdentity})
	require.NoError(t, err)
	require.Empty(t, response.Diagnostics)
	require.Equal(t, 1, requests)
	require.Equal(t, acknowledged.Private, response.Private)
	_, config = signingSecretDynamicValue(t, kind, changed)
	expected = changed

	writeFailure := "Failed to write webhook signing secret"
	if kind == "app" {
		writeFailure = "Failed to update app signing secret"
	}

	failures := []failureCase{
		{name: "timestamp", summary: writeFailure, invalidTimestamp: true},
		{name: "echo", summary: writeFailure, echo: true},
		{name: "corrupt private data", summary: "Invalid write-only secret private state", invalidPrivate: true},
	}
	if kind == "app" {
		failures = append(failures,
			failureCase{name: "contradictory redacted value", summary: "Unconfirmed app signing secret mutation", wrongSuffix: true},
			failureCase{name: "organization identity", summary: "Unconfirmed app signing secret mutation", attribute: "organization_id"},
			failureCase{name: "app definition identity", summary: "Unconfirmed app signing secret mutation", attribute: "app_definition_id"},
		)
	} else {
		failures = append(failures, failureCase{name: "space identity", summary: "Unexpected webhook signing secret identity", attribute: "space_id"})
	}

	for _, failure := range failures {
		fault = failure

		plannedPrivate := acknowledged.Private

		if fault.invalidPrivate {
			records, marshalErr := json.Marshal([]map[string]string{{"path": "value_wo", "hash": "invalid"}})
			require.NoError(t, marshalErr)

			plannedPrivate, err = json.Marshal(map[string][]byte{"write_only_secret_hashes": records})
			require.NoError(t, err)

			planned, planErr := providerServer.PlanResourceChange(t.Context(), &tfprotov6.PlanResourceChangeRequest{
				TypeName: "contentful_" + kind + "_signing_secret", PriorState: acknowledged.NewState,
				ProposedNewState: &plan, Config: &config, PriorPrivate: plannedPrivate, PriorIdentity: acknowledged.NewIdentity,
			})
			require.NoError(t, planErr)
			require.Len(t, planned.Diagnostics, 1)
			assert.Equal(t, tfprotov6.DiagnosticSeverityWarning, planned.Diagnostics[0].Severity)
			assert.Equal(t, "Invalid write-only secret private state", planned.Diagnostics[0].Summary)
			assert.Equal(t, "A verifier uses an unsupported algorithm, version, or parameter encoding. The provider cannot compare the configured secret. An in-place update will fail before writing the secret. Restore valid Terraform state before updating, or explicitly replace or destroy the resource.", planned.Diagnostics[0].Detail)
			assert.Equal(t, plannedPrivate, planned.PlannedPrivate)
			plannedRaw, decodeErr := planned.PlannedState.Unmarshal(resourceSchema.Type().TerraformType(t.Context()))
			require.NoError(t, decodeErr)

			var plannedAttrs map[string]tftypes.Value
			require.NoError(t, plannedRaw.As(&plannedAttrs))
			assert.False(t, plannedAttrs["updated_at"].IsKnown())
		}

		before := requests

		var logs bytes.Buffer

		response, err = providerServer.ApplyResourceChange(tflogtest.RootLogger(t.Context(), &logs), &tfprotov6.ApplyResourceChangeRequest{TypeName: "contentful_" + kind + "_signing_secret", PriorState: acknowledged.NewState, PlannedState: &plan, Config: &config, PlannedPrivate: plannedPrivate, PlannedIdentity: acknowledged.NewIdentity})
		require.NoError(t, err)
		require.NotEmpty(t, response.Diagnostics, failure.name)
		assert.Equal(t, tfprotov6.DiagnosticSeverityError, response.Diagnostics[0].Severity, failure.name)
		assert.Equal(t, failure.summary, response.Diagnostics[0].Summary, failure.name)

		if kind == "app" && failure.attribute != "" {
			assert.Contains(t, response.Diagnostics[0].Detail, "requested "+failure.attribute+".", failure.name)
		} else if failure.attribute != "" {
			assert.Equal(t, tftypes.NewAttributePath().WithAttributeName(failure.attribute), response.Diagnostics[0].Attribute, failure.name)
		}

		require.Equal(t, acknowledged.NewIdentity, response.NewIdentity, failure.name)
		require.Equal(t, plannedPrivate, response.Private)
		require.Equal(t, acknowledged.NewState, response.NewState)
		require.NotContains(t, logs.String(), changed)

		for _, diagnostic := range response.Diagnostics {
			require.NotContains(t, diagnostic.Detail, changed)
		}

		if fault.wrongSuffix {
			require.NotContains(t, response.Diagnostics[0].Detail, "zzzz")
			require.NotContains(t, response.Diagnostics[0].Detail, "bbbb")
		}

		if fault.invalidPrivate {
			require.Equal(t, before, requests)
		} else {
			require.Equal(t, before+1, requests)
		}
	}
}
