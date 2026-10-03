package provider_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
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
	values["created_at"] = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
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

func testSigningSecretWriteOnlyAcknowledgement(t *testing.T, kind string) {
	t.Helper()

	initial, changed := strings.Repeat("a", 64), strings.Repeat("b", 64)
	fixture := &signingSecretFixture{kind: kind, timestamps: true}

	type failureCase struct {
		name, summary, attribute               string
		echo, invalidTimestamp, invalidPrivate bool
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
			response, err = signingSecretResponseWithOtherIdentity(kind, fault.attribute, response)
			assert.NoError(t, err)
		}

		if fault.invalidTimestamp {
			response = strings.ReplaceAll(response, "2026-09-02T00:00:00Z", "invalid")
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
	require.Contains(t, string(private["write_only_secret_hashes"]), "$argon2id$")
	require.NotContains(t, string(private["write_only_secret_hashes"]), initial)
	// Same acknowledged value retains the exact salt and encoded verifier.
	response, err = providerServer.ApplyResourceChange(t.Context(), &tfprotov6.ApplyResourceChangeRequest{TypeName: "contentful_" + kind + "_signing_secret", PriorState: acknowledged.NewState, PlannedState: &plan, Config: &config, PlannedPrivate: acknowledged.Private, PlannedIdentity: acknowledged.NewIdentity})
	require.NoError(t, err)
	require.Empty(t, response.Diagnostics)
	require.Equal(t, 1, requests)
	require.Equal(t, acknowledged.Private, response.Private)
	_, config = signingSecretDynamicValue(t, kind, changed)
	expected = changed

	failures := []failureCase{
		{name: "timestamp", summary: "Failed to write " + kind + " signing secret", invalidTimestamp: true},
		{name: "echo", summary: "Failed to write " + kind + " signing secret", echo: true},
		{name: "corrupt private data", summary: "Invalid write-only secret private state", invalidPrivate: true},
	}
	if kind == "app" {
		failures = append(failures,
			failureCase{name: "organization identity", summary: "Unexpected app signing secret identity", attribute: "organization_id"},
			failureCase{name: "app definition identity", summary: "Unexpected app signing secret identity", attribute: "app_definition_id"},
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
		}

		before := requests

		var logs bytes.Buffer

		response, err = providerServer.ApplyResourceChange(tflogtest.RootLogger(t.Context(), &logs), &tfprotov6.ApplyResourceChangeRequest{TypeName: "contentful_" + kind + "_signing_secret", PriorState: acknowledged.NewState, PlannedState: &plan, Config: &config, PlannedPrivate: plannedPrivate, PlannedIdentity: acknowledged.NewIdentity})
		require.NoError(t, err)
		require.NotEmpty(t, response.Diagnostics, failure.name)
		assert.Equal(t, tfprotov6.DiagnosticSeverityError, response.Diagnostics[0].Severity, failure.name)
		assert.Equal(t, failure.summary, response.Diagnostics[0].Summary, failure.name)

		if failure.attribute != "" {
			assert.Equal(t, tftypes.NewAttributePath().WithAttributeName(failure.attribute), response.Diagnostics[0].Attribute, failure.name)
		}

		require.Equal(t, acknowledged.NewIdentity, response.NewIdentity, failure.name)
		require.Equal(t, plannedPrivate, response.Private)
		require.Equal(t, acknowledged.NewState, response.NewState)
		require.NotContains(t, logs.String(), changed)

		for _, diagnostic := range response.Diagnostics {
			require.NotContains(t, diagnostic.Detail, changed)
		}

		if fault.invalidPrivate {
			require.Equal(t, before, requests)
		} else {
			require.Equal(t, before+1, requests)
		}
	}
	// Even after a request may have reached Contentful, reverting to
	// the acknowledged value does not claim remote rollback.
	before := requests
	_, config = signingSecretDynamicValue(t, kind, initial)
	response, err = providerServer.ApplyResourceChange(t.Context(), &tfprotov6.ApplyResourceChangeRequest{TypeName: "contentful_" + kind + "_signing_secret", PriorState: acknowledged.NewState, PlannedState: &plan, Config: &config, PlannedPrivate: acknowledged.Private, PlannedIdentity: acknowledged.NewIdentity})
	require.NoError(t, err)
	require.Empty(t, response.Diagnostics)
	require.Equal(t, before, requests)
	require.Equal(t, acknowledged.Private, response.Private)
}

func signingSecretResponseWithOtherIdentity(kind, attribute, response string) (string, error) {
	var secret json.Marshaler

	switch kind {
	case "app":
		var app cm.AppSigningSecret

		err := json.Unmarshal([]byte(response), &app)
		if err != nil {
			return "", fmt.Errorf("decode app signing secret fixture: %w", err)
		}

		if attribute == "organization_id" {
			app.Sys.Organization = cm.NewOrganizationLink("other")
		} else {
			app.Sys.AppDefinition = cm.NewAppDefinitionLink("other")
		}

		secret = &app
	default:
		var webhook cm.WebhookSigningSecret

		err := json.Unmarshal([]byte(response), &webhook)
		if err != nil {
			return "", fmt.Errorf("decode webhook signing secret fixture: %w", err)
		}

		webhook.Sys.Space = cm.NewSpaceLink("other")
		secret = &webhook
	}

	encoded, err := json.Marshal(secret)
	if err != nil {
		return "", fmt.Errorf("encode signing secret fixture: %w", err)
	}

	return string(encoded), nil
}

func TestSigningSecretReadAbsentIdentity(t *testing.T) {
	t.Parallel()

	for _, kind := range []string{"app", "webhook"} {
		for _, knownIdentity := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/prior_identity=%t", kind, knownIdentity), func(t *testing.T) {
				t.Parallel()

				expectedPath := "/organizations/organization/app_definitions/app/signing_secret"
				identity := map[string]string{"organization_id": "organization", "app_definition_id": "app"}
				state := map[string]any{
					"id": "organization/app", "organization_id": "organization", "app_definition_id": "app",
					"value": nil, "value_wo": nil, "created_at": nil, "updated_at": nil, "timeouts": nil,
				}
				identityType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{"organization_id": tftypes.String, "app_definition_id": tftypes.String}}

				if kind == "webhook" {
					expectedPath = "/spaces/space/webhook_settings/signing_secret"
					identity = map[string]string{"space_id": "space"}
					state = map[string]any{
						"id": "space", "space_id": "space",
						"value": nil, "value_wo": nil, "created_at": nil, "updated_at": nil, "timeouts": nil,
					}
					identityType = tftypes.Object{AttributeTypes: map[string]tftypes.Type{"space_id": tftypes.String}}
				}

				stateJSON, err := json.Marshal(state)
				require.NoError(t, err)
				identityJSON, err := json.Marshal(identity)
				require.NoError(t, err)

				requests := 0
				providerServer := signingSecretProtocolServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					requests++

					assert.Equal(t, http.MethodGet, r.Method)
					assert.Equal(t, expectedPath, r.URL.Path)
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusNotFound)

					response := cmt.NewContentfulManagementError("NotFound", new("Absent"), nil)
					assert.NoError(t, json.NewEncoder(w).Encode(&response))
				}))

				request := &tfprotov6.ReadResourceRequest{
					TypeName: "contentful_" + kind + "_signing_secret", CurrentState: &tfprotov6.DynamicValue{JSON: stateJSON},
				}
				if knownIdentity {
					request.CurrentIdentity = &tfprotov6.ResourceIdentityData{IdentityData: &tfprotov6.DynamicValue{JSON: identityJSON}}
				}

				response, err := providerServer.ReadResource(t.Context(), request)
				require.NoError(t, err)
				require.Equal(t, 1, requests)

				for _, diagnostic := range response.Diagnostics {
					require.NotEqual(t, tfprotov6.DiagnosticSeverityError, diagnostic.Severity, diagnostic.Detail)
				}

				resourceSchema, _ := signingSecretDynamicValue(t, kind, nil)
				refreshed, err := response.NewState.Unmarshal(resourceSchema.Type().TerraformType(t.Context()))
				require.NoError(t, err)
				require.True(t, refreshed.IsNull())
				require.NotNil(t, response.NewIdentity)
				actual, err := response.NewIdentity.IdentityData.Unmarshal(identityType)
				require.NoError(t, err)
				expected, err := (&tfprotov6.DynamicValue{JSON: identityJSON}).Unmarshal(identityType)
				require.NoError(t, err)
				require.True(t, expected.Equal(actual))
			})
		}
	}
}
