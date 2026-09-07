package provider_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflogtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:ireturn
func livePreviewVariablesProtocolServer(t *testing.T, handler http.Handler) tfprotov6.ProviderServer {
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

func livePreviewVariablesModelFromDynamicValue(t *testing.T, value *tfprotov6.DynamicValue) LivePreviewVariablesModel {
	t.Helper()
	require.NotNil(t, value)
	resourceSchema := LivePreviewVariablesResourceSchema(t.Context())
	raw, err := value.Unmarshal(resourceSchema.Type().TerraformType(t.Context()))
	require.NoError(t, err)

	state := tfsdk.State{Schema: resourceSchema, Raw: raw}

	var model LivePreviewVariablesModel
	require.Empty(t, state.Get(t.Context(), &model))

	return model
}

func TestLivePreviewVariablesMutationPublishesResponseAndVersion(t *testing.T) {
	t.Parallel()

	for _, operation := range []string{"create", "update"} {
		for name, test := range map[string]struct {
			variables, spaceID, environmentID string
			plannedID                         string
			paths                             []string
		}{
			"equivalent":                       {variables: `{"a":"VALUE_DO_NOT_ECHO","b":null}`, spaceID: "space", environmentID: "environment"},
			"different variables":              {variables: `{"remote":"VALUE_DO_NOT_ECHO"}`, spaceID: "space", environmentID: "environment", paths: []string{"variables"}},
			"different space":                  {variables: `{"a":"VALUE_DO_NOT_ECHO","b":null}`, spaceID: "other", environmentID: "environment", paths: []string{"space_id"}},
			"different environment":            {variables: `{"a":"VALUE_DO_NOT_ECHO","b":null}`, spaceID: "space", environmentID: "other", paths: []string{"environment_id"}},
			"different legacy ID":              {variables: `{"a":"VALUE_DO_NOT_ECHO","b":null}`, spaceID: "space", environmentID: "environment", plannedID: "other/identity", paths: []string{"id"}},
			"different identity and variables": {variables: `{"remote":"VALUE_DO_NOT_ECHO"}`, spaceID: "other", environmentID: "other", paths: []string{"space_id", "environment_id", "variables"}},
		} {
			t.Run(operation+"/"+name, func(t *testing.T) {
				t.Parallel()

				var requests atomic.Int64

				providerServer := livePreviewVariablesProtocolServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					requests.Add(1)
					assert.Equal(t, http.MethodPut, r.Method)
					assert.Equal(t, "/spaces/space/environments/environment/live_preview/variables", r.URL.Path)

					version := "17"
					if operation == "create" {
						version = "0"
					}

					assert.Equal(t, version, r.Header.Get("X-Contentful-Version"))
					body, err := io.ReadAll(r.Body)
					assert.NoError(t, err)
					assert.JSONEq(t, `{"variables":{"a":"VALUE_DO_NOT_ECHO","b":null}}`, string(body))
					w.Header().Set("Content-Type", "application/json")
					_, _ = fmt.Fprintf(w, `{"sys":{"space":{"sys":{"type":"Link","linkType":"Space","id":%q}},"environment":{"sys":{"type":"Link","linkType":"Environment","id":%q}},"version":43},"variables":%s}`, test.spaceID, test.environmentID, test.variables)
				}))
				model := livePreviewVariablesModel(` { "a": "VALUE_DO_NOT_ECHO", "b": null } `)

				prior := resourceModelDynamicValue(t, LivePreviewVariablesResourceSchema(t.Context()), livePreviewVariablesModel(`{"before":"old"}`))
				if operation == "create" {
					prior = nullResourceDynamicValue(t, LivePreviewVariablesResourceSchema(t.Context()))
					model.ID = types.StringUnknown()
				}

				if test.plannedID != "" {
					model.ID = types.StringValue(test.plannedID)
				}

				planned := resourceModelDynamicValue(t, LivePreviewVariablesResourceSchema(t.Context()), model)
				configured := livePreviewVariablesModel(`{"config":"must-not-be-sent"}`)
				config := resourceModelDynamicValue(t, LivePreviewVariablesResourceSchema(t.Context()), configured)

				var logs bytes.Buffer

				response, err := providerServer.ApplyResourceChange(tflogtest.RootLogger(t.Context(), &logs), &tfprotov6.ApplyResourceChangeRequest{
					TypeName: "contentful_live_preview_variables", PriorState: &prior, PlannedState: &planned, Config: &config, PlannedPrivate: privateVersionBytes(t, 17),
				})
				require.NoError(t, err)
				require.EqualValues(t, 1, requests.Load())

				require.Len(t, response.Diagnostics, len(test.paths))

				for index, attribute := range test.paths {
					assert.Equal(t, tfprotov6.DiagnosticSeverityError, response.Diagnostics[index].Severity)
					assert.Equal(t, tftypes.NewAttributePath().WithAttributeName(attribute), response.Diagnostics[index].Attribute)
				}

				state := livePreviewVariablesModelFromDynamicValue(t, response.NewState)
				require.Equal(t, "space/environment", state.ID.ValueString())
				require.Equal(t, "space", state.SpaceID.ValueString())
				require.Equal(t, "environment", state.EnvironmentID.ValueString())

				if len(test.paths) == 0 {
					require.Equal(t, model.Variables.ValueString(), state.Variables.ValueString())
				} else {
					// Response fixtures are normalized; preserve their exact representation on any mismatch.
					require.Equal(t, test.variables, state.Variables.ValueString())
				}

				var private map[string][]byte
				require.NoError(t, json.Unmarshal(response.Private, &private))
				require.Equal(t, "43", string(private["version"]))
				require.NotNil(t, response.NewIdentity)
				assert.NotContains(t, logs.String(), "VALUE_DO_NOT_ECHO")
			})
		}
	}
}

type livePreviewVariablesErrorTest struct {
	status   int
	body     string
	absent   bool
	conflict bool
}

func TestLivePreviewVariablesErrorsAndAbsence(t *testing.T) {
	t.Parallel()

	for _, operation := range []string{"create", "update", "read", "delete"} {
		for name, test := range map[string]livePreviewVariablesErrorTest{
			"missing document":               {status: 404, body: `{"sys":{"type":"Error","id":"NotFound"},"message":"The resource could not be found."}`, absent: true},
			"missing parent":                 {status: 404, body: `{"sys":{"type":"Error","id":"NotFound"},"message":"Missing environment","details":{"type":"Environment","id":"environment"}}`, absent: true},
			"enterprise feature unavailable": {status: 403, body: `{"statusCode":403,"error":"Forbidden","message":"previewLocalization is not enabled"}`},
			"service not found":              {status: 404, body: `{"statusCode":404,"error":"Not Found","message":"Not Found"}`},
			"wrong CMA ID":                   {status: 404, body: `{"sys":{"type":"Error","id":"AccessDenied"},"message":"Denied"}`},
			"denied":                         {status: 403, body: `{"sys":{"type":"Error","id":"AccessDenied"},"message":"Denied"}`},
			"wrong status for NotFound":      {status: 400, body: `{"sys":{"type":"Error","id":"NotFound"},"message":"Bad response"}`},
			"conflict":                       {status: 409, body: `{"sys":{"type":"Error","id":"VersionMismatch"},"message":"Version mismatch"}`, conflict: true},
			"other conflict":                 {status: 409, body: `{"sys":{"type":"Error","id":"Conflict"},"message":"Other conflict"}`},
			"validation value omitted from diagnostics": {status: 422, body: `{"sys":{"type":"Error","id":"ValidationFailed"},"message":"Validation error","details":{"errors":[{"name":"type","value":"VALUE_DO_NOT_ECHO","details":"Expected Text","path":["variable","en-US"]}]}}`},
			"server failure": {status: 500, body: `{"statusCode":500,"error":"Internal Server Error","message":"Internal error"}`},
			"malformed":      {status: 200, body: `{`},
		} {
			// The shared transport suite owns read retries; this server error exercises mutation guidance.
			if operation == "read" && test.status >= 500 {
				continue
			}

			t.Run(operation+"/"+name, func(t *testing.T) {
				t.Parallel()
				runLivePreviewVariablesErrorTest(t, operation, test)
			})
		}
	}
}

func runLivePreviewVariablesErrorTest(t *testing.T, operation string, test livePreviewVariablesErrorTest) {
	t.Helper()

	var requests atomic.Int64

	providerServer := livePreviewVariablesProtocolServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		assert.Equal(t, "/spaces/space/environments/environment/live_preview/variables", r.URL.Path)

		if operation == "delete" {
			assert.Empty(t, r.Header.Get("X-Contentful-Version"))
			assert.Empty(t, r.Header.Get("If-Match"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(test.status)
		_, _ = io.WriteString(w, test.body)
	}))
	resourceSchema := LivePreviewVariablesResourceSchema(t.Context())
	state := resourceModelDynamicValue(t, resourceSchema, livePreviewVariablesModel(`{"before":"old"}`))

	var (
		diagnostics []*tfprotov6.Diagnostic
		returned    *tfprotov6.DynamicValue
	)

	if operation == "read" {
		response, err := providerServer.ReadResource(t.Context(), &tfprotov6.ReadResourceRequest{TypeName: "contentful_live_preview_variables", CurrentState: &state, CurrentIdentity: &tfprotov6.ResourceIdentityData{IdentityData: &tfprotov6.DynamicValue{JSON: []byte(`{"space_id":"space","environment_id":"environment"}`)}}, Private: privateVersionBytes(t, 17)})
		require.NoError(t, err)

		diagnostics, returned = response.Diagnostics, response.NewState
	} else {
		prior, planned := state, state
		if operation == "create" {
			prior = nullResourceDynamicValue(t, resourceSchema)
		}

		if operation == "delete" {
			planned = nullResourceDynamicValue(t, resourceSchema)
		}

		response, err := providerServer.ApplyResourceChange(t.Context(), &tfprotov6.ApplyResourceChangeRequest{TypeName: "contentful_live_preview_variables", PriorState: &prior, PlannedState: &planned, Config: &planned, PlannedPrivate: privateVersionBytes(t, 17)})
		require.NoError(t, err)

		diagnostics, returned = response.Diagnostics, response.NewState
	}

	require.EqualValues(t, 1, requests.Load())

	if test.absent && (operation == "read" || operation == "delete") {
		for _, diagnostic := range diagnostics {
			require.NotEqual(t, tfprotov6.DiagnosticSeverityError, diagnostic.Severity, diagnostic.Summary+": "+diagnostic.Detail)
		}

		raw, err := returned.Unmarshal(resourceSchema.Type().TerraformType(t.Context()))
		require.NoError(t, err)
		require.True(t, raw.IsNull())

		return
	}

	require.NotEmpty(t, diagnostics)
	require.Equal(t, tfprotov6.DiagnosticSeverityError, diagnostics[0].Severity)
	assert.NotContains(t, diagnostics[0].Detail, "VALUE_DO_NOT_ECHO")
	assert.Equal(t, operation != "read" && (test.status >= 500 || test.body == `{`), strings.Contains(diagnostics[0].Detail, "may have committed"))
	assert.Equal(t, operation == "create" && test.conflict, strings.Contains(diagnostics[0].Detail, "import it"))

	if operation != "create" {
		require.Equal(t, livePreviewVariablesModel(`{"before":"old"}`).Variables, livePreviewVariablesModelFromDynamicValue(t, returned).Variables)
	}
}

func TestLivePreviewVariablesUnknownInputsDoNotMutate(t *testing.T) {
	t.Parallel()

	for _, operation := range []string{"create", "update"} {
		for _, attribute := range []string{"space_id", "environment_id", "variables"} {
			t.Run(operation+"/"+attribute, func(t *testing.T) {
				t.Parallel()

				var requests atomic.Int64

				providerServer := livePreviewVariablesProtocolServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					requests.Add(1)
					w.WriteHeader(http.StatusInternalServerError)
				}))
				model := livePreviewVariablesModel(`{}`)

				switch attribute {
				case "space_id":
					model.SpaceID = types.StringUnknown()
				case "environment_id":
					model.EnvironmentID = types.StringUnknown()
				case "variables":
					model.Variables = jsontypes.NewNormalizedUnknown()
				}

				resourceSchema := LivePreviewVariablesResourceSchema(t.Context())
				planned := resourceModelDynamicValue(t, resourceSchema, model)

				prior := resourceModelDynamicValue(t, resourceSchema, livePreviewVariablesModel(`{}`))
				if operation == "create" {
					prior = nullResourceDynamicValue(t, resourceSchema)
				}

				response, err := providerServer.ApplyResourceChange(t.Context(), &tfprotov6.ApplyResourceChangeRequest{TypeName: "contentful_live_preview_variables", PriorState: &prior, PlannedState: &planned, Config: &planned, PlannedPrivate: privateVersionBytes(t, 17)})
				require.NoError(t, err)
				require.NotEmpty(t, response.Diagnostics)
				require.Zero(t, requests.Load())
			})
		}
	}
}

func TestLivePreviewVariablesUsesDefaultMutationRetryPolicy(t *testing.T) {
	t.Parallel()

	var requests atomic.Int64

	providerServer := livePreviewVariablesProtocolServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "0", r.Header.Get("X-Contentful-Version"))
		w.Header().Set("Content-Type", "application/json")

		if requests.Add(1) == 1 {
			w.Header().Set("X-Contentful-Ratelimit-Reset", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = io.WriteString(w, `{"sys":{"type":"Error","id":"RateLimitExceeded"},"message":"Rate limited"}`)

			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{"statusCode":500,"error":"Internal Server Error","message":"Unknown outcome"}`)
	}))
	resourceSchema := LivePreviewVariablesResourceSchema(t.Context())
	prior := nullResourceDynamicValue(t, resourceSchema)
	planned := resourceModelDynamicValue(t, resourceSchema, livePreviewVariablesModel(`{}`))
	response, err := providerServer.ApplyResourceChange(t.Context(), &tfprotov6.ApplyResourceChangeRequest{TypeName: "contentful_live_preview_variables", PriorState: &prior, PlannedState: &planned, Config: &planned})
	require.NoError(t, err)
	require.NotEmpty(t, response.Diagnostics)
	require.EqualValues(t, 2, requests.Load())
}

func TestLivePreviewVariablesReadProjectsResponse(t *testing.T) {
	t.Parallel()

	for _, knownIdentity := range []bool{false, true} {
		t.Run(fmt.Sprintf("known identity %t", knownIdentity), func(t *testing.T) {
			t.Parallel()

			var requests atomic.Int64

			providerServer := livePreviewVariablesProtocolServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/spaces/space/environments/environment/live_preview/variables", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, `{"sys":{"space":{"sys":{"type":"Link","linkType":"Space","id":"returned-space"}},"environment":{"sys":{"type":"Link","linkType":"Environment","id":"returned-environment"}},"version":29},"variables":{"remote":"new"}}`)
			}))
			state := resourceModelDynamicValue(t, LivePreviewVariablesResourceSchema(t.Context()), livePreviewVariablesModel(`{"before":"old"}`))

			request := &tfprotov6.ReadResourceRequest{TypeName: "contentful_live_preview_variables", CurrentState: &state, Private: privateVersionBytes(t, 17)}
			if knownIdentity {
				request.CurrentIdentity = &tfprotov6.ResourceIdentityData{IdentityData: &tfprotov6.DynamicValue{JSON: []byte(`{"space_id":"space","environment_id":"environment"}`)}}
			}

			response, err := providerServer.ReadResource(t.Context(), request)
			require.NoError(t, err)
			require.EqualValues(t, 1, requests.Load())

			if knownIdentity {
				// The framework owns the check against an existing immutable resource identity.
				require.Len(t, response.Diagnostics, 1)
				require.Equal(t, tfprotov6.DiagnosticSeverityError, response.Diagnostics[0].Severity)
				require.Equal(t, "Unexpected Identity Change", response.Diagnostics[0].Summary)
				require.NotContains(t, response.Diagnostics[0].Detail, "recovery state")
			} else {
				require.Empty(t, response.Diagnostics)
			}

			refreshed := livePreviewVariablesModelFromDynamicValue(t, response.NewState)
			require.Equal(t, "returned-space", refreshed.SpaceID.ValueString())
			require.Equal(t, "returned-environment", refreshed.EnvironmentID.ValueString())
			require.Equal(t, "returned-space/returned-environment", refreshed.ID.ValueString())
			require.Equal(t, jsontypes.NewNormalizedValue(`{"remote":"new"}`), refreshed.Variables)
			require.NotNil(t, response.NewIdentity)

			var private map[string][]byte
			require.NoError(t, json.Unmarshal(response.Private, &private))
			require.Equal(t, "29", string(private["version"]))
		})
	}
}

func TestLivePreviewVariablesMultipleValidationErrorDiagnostics(t *testing.T) {
	t.Parallel()

	var requests atomic.Int64

	providerServer := livePreviewVariablesProtocolServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/spaces/space/environments/environment/live_preview/variables", r.URL.Path)
		assert.Equal(t, "17", r.Header.Get("X-Contentful-Version"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = io.WriteString(w, `{"sys":{"type":"Error","id":"ValidationFailed"},"message":"Validation error","details":{"errors":[
			{"name":"type","type":"Text","value":981723,"details":"The type of \"value\" is incorrect, expected type: Text","path":["global"]},
			{"name":"type","type":"Text","value":["REJECTED_ARRAY_VALUE"],"details":"The type of \"value\" is incorrect, expected type: Text","path":["localized","en-US"]},
			{"name":"unknown","value":"REJECTED_LOCALE_VALUE","details":"The property \"zz-ZZ\" is not allowed here.","path":["localized","zz-ZZ"]}
		]}}`)
	}))
	resourceSchema := LivePreviewVariablesResourceSchema(t.Context())
	prior := resourceModelDynamicValue(t, resourceSchema, livePreviewVariablesModel(`{"before":"old"}`))
	planned := resourceModelDynamicValue(t, resourceSchema, livePreviewVariablesModel(`{"after":"new"}`))
	response, err := providerServer.ApplyResourceChange(t.Context(), &tfprotov6.ApplyResourceChangeRequest{TypeName: "contentful_live_preview_variables", PriorState: &prior, PlannedState: &planned, Config: &planned, PlannedPrivate: privateVersionBytes(t, 17)})
	require.NoError(t, err)
	require.EqualValues(t, 1, requests.Load())
	require.Len(t, response.Diagnostics, 1)
	require.Equal(t, tfprotov6.DiagnosticSeverityError, response.Diagnostics[0].Severity)
	require.Equal(t, "HTTP 422: Error: ValidationFailed: Validation error\n  global: The type of \"value\" is incorrect, expected type: Text\n  localized.en-US: The type of \"value\" is incorrect, expected type: Text\n  localized.zz-ZZ: The property \"zz-ZZ\" is not allowed here.", response.Diagnostics[0].Detail)
	require.Equal(t, livePreviewVariablesModel(`{"before":"old"}`).Variables, livePreviewVariablesModelFromDynamicValue(t, response.NewState).Variables)

	var private map[string][]byte
	require.NoError(t, json.Unmarshal(response.Private, &private))
	require.Equal(t, "17", string(private["version"]))
}

func TestLivePreviewVariablesReadFeedsTheNextUpdateVersion(t *testing.T) {
	t.Parallel()

	var requests atomic.Int64

	providerServer := livePreviewVariablesProtocolServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestNumber := requests.Add(1)

		assert.Equal(t, "/spaces/space/environments/routing-id/live_preview/variables", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")

		if requestNumber == 1 {
			assert.Equal(t, http.MethodGet, r.Method)

			_, _ = io.WriteString(w, `{"sys":{"space":{"sys":{"type":"Link","linkType":"Space","id":"space"}},"environment":{"sys":{"type":"Link","linkType":"Environment","id":"routing-id"}},"version":29},"variables":{"remote":{"future":[true,9007199254740993]}}}`)

			return
		}

		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "29", r.Header.Get("X-Contentful-Version"))

		_, _ = io.WriteString(w, `{"sys":{"space":{"sys":{"type":"Link","linkType":"Space","id":"space"}},"environment":{"sys":{"type":"Link","linkType":"Environment","id":"routing-id"}},"version":30},"variables":{}}`)
	}))
	resourceSchema := LivePreviewVariablesResourceSchema(t.Context())
	model := livePreviewVariablesModel(`{"old":"value"}`)
	model.EnvironmentID = types.StringValue("routing-id")
	model.ID = types.StringValue("space/routing-id")
	state := resourceModelDynamicValue(t, resourceSchema, model)
	read, err := providerServer.ReadResource(t.Context(), &tfprotov6.ReadResourceRequest{TypeName: "contentful_live_preview_variables", CurrentState: &state, Private: privateVersionBytes(t, 1)})
	require.NoError(t, err)
	require.Empty(t, read.Diagnostics)
	refreshed := livePreviewVariablesModelFromDynamicValue(t, read.NewState)
	//nolint:testifylint // JSONEq rounds numbers through float64 and cannot prove preservation here.
	require.Equal(t, `{"remote":{"future":[true,9007199254740993]}}`, refreshed.Variables.ValueString())
	require.Equal(t, model.EnvironmentID, refreshed.EnvironmentID)
	require.Equal(t, model.ID, refreshed.ID)
	refreshed.Variables = jsontypes.NewNormalizedValue(`{}`)
	planned := resourceModelDynamicValue(t, resourceSchema, refreshed)
	updated, err := providerServer.ApplyResourceChange(t.Context(), &tfprotov6.ApplyResourceChangeRequest{TypeName: "contentful_live_preview_variables", PriorState: read.NewState, PlannedState: &planned, Config: &planned, PlannedPrivate: read.Private, PlannedIdentity: read.NewIdentity})
	require.NoError(t, err)
	require.Empty(t, updated.Diagnostics)
	require.EqualValues(t, 2, requests.Load())
}

func TestLivePreviewVariablesUpdateCanRecreateAnAbsentDocument(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")

	var requests atomic.Int64

	providerServer := livePreviewVariablesProtocolServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "17", r.Header.Get("X-Contentful-Version"))
		server.ServeHTTP(w, r)
	}))
	resourceSchema := LivePreviewVariablesResourceSchema(t.Context())
	prior := resourceModelDynamicValue(t, resourceSchema, livePreviewVariablesModel(`{"previous":"lifetime"}`))
	planned := resourceModelDynamicValue(t, resourceSchema, livePreviewVariablesModel(`{}`))
	response, err := providerServer.ApplyResourceChange(t.Context(), &tfprotov6.ApplyResourceChangeRequest{TypeName: "contentful_live_preview_variables", PriorState: &prior, PlannedState: &planned, Config: &planned, PlannedPrivate: privateVersionBytes(t, 17)})
	require.NoError(t, err)
	require.Empty(t, response.Diagnostics)
	require.EqualValues(t, 1, requests.Load())

	var private map[string][]byte
	require.NoError(t, json.Unmarshal(response.Private, &private))
	require.Equal(t, "1", string(private["version"]))
}
