package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflogtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This fixture contains only opaque redaction metadata.
//
//nolint:gosec
const webhookSigningSecretTestResponse = `{"sys":{"type":"WebhookSigningSecret","space":{"sys":{"type":"Link","linkType":"Space","id":"space"}}},"redactedValue":"opaque"}`

const (
	webhookSigningSecretTestValue        = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaAb09+/=_-"
	webhookSigningSecretTestUpdatedValue = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbZy87-/=_+"
)

func webhookSigningSecretTestState(t *testing.T, value types.String) (tfsdk.State, *tfsdk.ResourceIdentity) {
	t.Helper()
	ctx := t.Context()
	model := WebhookSigningSecretModel{
		IDIdentityModel:                   NewIDIdentityModelFromMultipartID("space"),
		WebhookSigningSecretIdentityModel: WebhookSigningSecretIdentityModel{SpaceID: types.StringValue("space")},
		Value:                             value,
		Timeouts:                          TimeoutsNull(),
	}
	state := tfsdk.State{Schema: WebhookSigningSecretResourceSchema(ctx)}
	require.Empty(t, state.Set(ctx, &model))

	identitySchema := resourceIdentitySchema([]string{"space_id"})
	identity := &tfsdk.ResourceIdentity{Schema: identitySchema, Raw: tftypes.NewValue(identitySchema.Type().TerraformType(ctx), nil)}
	require.Empty(t, identity.Set(ctx, &model.WebhookSigningSecretIdentityModel))

	return state, identity
}

func webhookSigningSecretTestClient(t *testing.T, server *httptest.Server) *webhookSigningSecretResource {
	t.Helper()

	client, err := cm.NewClient(server.URL, cm.NewAccessTokenSecuritySource("test-token"), cm.WithClient(newContentfulHTTPClient(server.Client())))
	require.NoError(t, err)

	return &webhookSigningSecretResource{providerData: ContentfulProviderData{client: client}}
}

func TestWebhookSigningSecretSuccessLogsRedactValues(t *testing.T) {
	t.Parallel()

	for _, operation := range []string{"create", "read", "update", "delete"} {
		t.Run(operation, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.Method == http.MethodDelete {
					w.WriteHeader(http.StatusNoContent)

					return
				}

				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, webhookSigningSecretTestResponse)
			}))
			t.Cleanup(server.Close)
			implementation := webhookSigningSecretTestClient(t, server)
			state, identity := webhookSigningSecretTestState(t, types.StringValue(webhookSigningSecretTestValue))
			plan, _ := webhookSigningSecretTestState(t, types.StringValue(webhookSigningSecretTestUpdatedValue))

			var logs bytes.Buffer

			ctx := tflogtest.RootLogger(t.Context(), &logs)

			var diagnostics diag.Diagnostics

			switch operation {
			case "create":
				response := resource.CreateResponse{State: tfsdk.State{Schema: state.Schema}, Identity: identity}
				implementation.Create(ctx, resource.CreateRequest{Plan: tfsdk.Plan(plan)}, &response)
				diagnostics = response.Diagnostics
			case "read":
				response := resource.ReadResponse{State: state, Identity: identity}
				implementation.Read(ctx, resource.ReadRequest{State: state}, &response)
				diagnostics = response.Diagnostics
			case "update":
				response := resource.UpdateResponse{State: state, Identity: identity}
				implementation.Update(ctx, resource.UpdateRequest{State: state, Plan: tfsdk.Plan(plan)}, &response)
				diagnostics = response.Diagnostics
			case "delete":
				response := resource.DeleteResponse{State: state, Identity: identity}
				implementation.Delete(ctx, resource.DeleteRequest{State: state}, &response)
				diagnostics = response.Diagnostics
			}

			require.Empty(t, diagnostics)
			assert.NotContains(t, logs.String(), webhookSigningSecretTestValue)
			assert.NotContains(t, logs.String(), webhookSigningSecretTestUpdatedValue)
			assert.NotContains(t, logs.String(), "opaque")
			assert.Contains(t, logs.String(), "webhook_signing_secret."+operation)
		})
	}
}

func TestWebhookSigningSecretFailuresRetainStateAndRedactKnownValues(t *testing.T) {
	t.Parallel()

	const sentinel = "UPSTREAM_DETAIL"

	errorsByName := map[string]struct {
		status int
		body   string
	}{
		"bad request":       {400, `{"sys":{"type":"Error","id":"BadRequest"},"message":"Invalid request payload JSON format"}`},
		"validation echo":   {422, `{"sys":{"type":"Error","id":"ValidationFailed"},"message":"Validation error","details":{"errors":[{"details":"Expected string","path":["value"],"value":"` + sentinel + `"}]}}`},
		"auth":              {401, `{"sys":{"type":"Error","id":"AccessDenied"},"message":"` + sentinel + `"}`},
		"forbidden":         {403, `{"sys":{"type":"Error","id":"` + sentinel + `"}}`},
		"other typed 404":   {404, `{"sys":{"type":"Error","id":"` + sentinel + `"}}`},
		"plain 404":         {404, sentinel},
		"rate limit":        {429, `{"sys":{"type":"Error","id":"RateLimitExceeded"},"message":"` + sentinel + `"}`},
		"server failure":    {500, `{"sys":{"type":"Error","id":"ServerError"},"message":"` + sentinel + `"}`},
		"malformed success": {200, sentinel},
		"wrong space":       {200, strings.ReplaceAll(webhookSigningSecretTestResponse, `"id":"space"`, `"id":"`+sentinel+`"`)},
		"empty space":       {200, strings.ReplaceAll(webhookSigningSecretTestResponse, `"id":"space"`, `"id":""`)},
		"missing sys":       {200, `{"redactedValue":"` + sentinel + `"}`},
		"wrong type":        {200, strings.ReplaceAll(webhookSigningSecretTestResponse, "WebhookSigningSecret", sentinel)},
		"wrong link":        {200, strings.ReplaceAll(webhookSigningSecretTestResponse, `"Space"`, `"`+sentinel+`"`)},
		"missing redaction": {200, `{"sys":{"type":"WebhookSigningSecret","space":{"sys":{"type":"Link","linkType":"Space","id":"space"}}}}`},
		"null redaction":    {200, strings.ReplaceAll(webhookSigningSecretTestResponse, `"opaque"`, `null`)},
		"unexpected status": {202, webhookSigningSecretTestResponse},
	}
	for name, test := range errorsByName {
		for _, operation := range []string{"create", "update", "delete", "read"} {
			// Safe reads retry 429/5xx; this matrix exercises terminal read errors
			// and single-attempt mutations without spending a retry deadline.
			if operation == "read" && (test.status == 429 || test.status == 500) {
				continue
			}

			t.Run(name+"/"+operation, func(t *testing.T) {
				t.Parallel()

				echo, redactedEcho := webhookSigningSecretTestValue, "***"

				switch operation {
				case "create":
					echo = webhookSigningSecretTestUpdatedValue
				case "update":
					echo += " / " + webhookSigningSecretTestUpdatedValue
					redactedEcho += " / ***"
				}

				body := strings.ReplaceAll(test.body, sentinel, echo+"; upstream detail")

				var count atomic.Int64

				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					count.Add(1)
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(test.status)
					_, _ = io.WriteString(w, body)
				}))
				t.Cleanup(server.Close)
				implementation := webhookSigningSecretTestClient(t, server)
				state, identity := webhookSigningSecretTestState(t, types.StringValue(webhookSigningSecretTestValue))
				plan, _ := webhookSigningSecretTestState(t, types.StringValue(webhookSigningSecretTestUpdatedValue))
				identityBefore := identity.Raw.Copy()

				var logs bytes.Buffer

				ctx := tflogtest.RootLogger(t.Context(), &logs)

				var diagnostics diag.Diagnostics

				switch operation {
				case "create":
					nullState := tfsdk.State{Schema: state.Schema, Raw: tftypes.NewValue(state.Schema.Type().TerraformType(ctx), nil)}
					resp := resource.CreateResponse{State: nullState, Identity: identity}
					implementation.Create(ctx, resource.CreateRequest{Plan: tfsdk.Plan(plan)}, &resp)
					diagnostics = resp.Diagnostics
					assert.True(t, resp.State.Raw.IsNull())
				case "update":
					resp := resource.UpdateResponse{State: state, Identity: identity}
					implementation.Update(ctx, resource.UpdateRequest{State: state, Plan: tfsdk.Plan(plan)}, &resp)
					diagnostics = resp.Diagnostics
					assert.True(t, state.Raw.Equal(resp.State.Raw))
				case "read":
					resp := resource.ReadResponse{State: state, Identity: identity}
					implementation.Read(ctx, resource.ReadRequest{State: state}, &resp)
					diagnostics = resp.Diagnostics
					assert.True(t, state.Raw.Equal(resp.State.Raw))
				case "delete":
					resp := resource.DeleteResponse{State: state, Identity: identity}
					implementation.Delete(ctx, resource.DeleteRequest{State: state}, &resp)
					diagnostics = resp.Diagnostics
					assert.True(t, state.Raw.Equal(resp.State.Raw))
				}

				require.True(t, diagnostics.HasError(), diagnostics)
				assert.True(t, identityBefore.Equal(identity.Raw))
				assert.EqualValues(t, 1, count.Load(), "no mutation replay or recovery GET")
				assert.NotContains(t, logs.String(), "upstream detail")

				assert.NotContains(t, logs.String(), webhookSigningSecretTestValue)
				assert.NotContains(t, logs.String(), webhookSigningSecretTestUpdatedValue)

				for _, diagnostic := range diagnostics {
					assert.NotContains(t, diagnostic.Detail(), webhookSigningSecretTestValue)
					assert.NotContains(t, diagnostic.Detail(), webhookSigningSecretTestUpdatedValue)
				}

				if name == "auth" {
					want := "Error: AccessDenied: " + redactedEcho + "; upstream detail"
					if operation != "read" {
						want += webhookSigningSecretMutationRecovery
					}

					assert.Equal(t, want, diagnostics[0].Detail())
				}
			})
		}
	}
}

func TestWebhookSigningSecretNotFoundLifecycle(t *testing.T) {
	t.Parallel()

	for _, parent := range []string{"Space", "WebhookSigningSecret"} {
		t.Run(parent, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_, _ = io.WriteString(w, `{"sys":{"type":"Error","id":"NotFound"},"message":"`+parent+` not found"}`)
			}))
			t.Cleanup(server.Close)
			implementation := webhookSigningSecretTestClient(t, server)
			state, identity := webhookSigningSecretTestState(t, types.StringValue(webhookSigningSecretTestValue))
			read := resource.ReadResponse{State: state, Identity: identity}
			implementation.Read(t.Context(), resource.ReadRequest{State: state}, &read)
			require.Empty(t, read.Diagnostics)
			assert.True(t, read.State.Raw.IsNull())
			deleted := resource.DeleteResponse{State: state, Identity: identity}
			implementation.Delete(t.Context(), resource.DeleteRequest{State: state}, &deleted)
			require.Empty(t, deleted.Diagnostics)

			created := resource.CreateResponse{State: tfsdk.State{Schema: state.Schema}, Identity: identity}
			implementation.Create(t.Context(), resource.CreateRequest{Plan: tfsdk.Plan(state)}, &created)
			require.True(t, created.Diagnostics.HasError())
		})
	}
}

func TestWebhookSigningSecretValueValidationAndRequestBoundary(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		value                     types.String
		configError, requestError bool
	}{
		"valid":   {types.StringValue(webhookSigningSecretTestValue), false, false},
		"null":    {types.StringNull(), false, true},
		"unknown": {types.StringUnknown(), false, true},
		"empty":   {types.StringValue(""), true, true},
		"short":   {types.StringValue(strings.Repeat("x", 63)), true, true},
		"long":    {types.StringValue(strings.Repeat("x", 65)), true, true},
		"invalid": {types.StringValue(strings.Repeat("!", 64)), true, true},
		"unicode": {types.StringValue(strings.Repeat("é", 64)), true, true},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var validated validator.StringResponse
			webhookSigningSecretValueValidator{}.ValidateString(t.Context(), validator.StringRequest{ConfigValue: test.value, Path: path.Root("value")}, &validated)
			assert.Equal(t, test.configError, validated.Diagnostics.HasError())
			model := WebhookSigningSecretModel{Value: test.value}
			request, diags := model.ToWebhookSigningSecretRequest(t.Context(), path.Empty())
			assert.Equal(t, test.requestError, diags.HasError())

			if !test.requestError {
				assert.Equal(t, webhookSigningSecretTestValue, request.Value)
			}

			if !test.requestError {
				return
			}
			// Direct lifecycle invocation bypasses configuration validation.
			var requests atomic.Int64

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				requests.Add(1)
				w.WriteHeader(http.StatusInternalServerError)
			}))
			t.Cleanup(server.Close)
			state, identity := webhookSigningSecretTestState(t, test.value)
			resp := resource.CreateResponse{State: tfsdk.State{Schema: state.Schema}, Identity: identity}
			webhookSigningSecretTestClient(t, server).Create(t.Context(), resource.CreateRequest{Plan: tfsdk.Plan(state)}, &resp)
			assert.True(t, resp.Diagnostics.HasError())
			assert.Zero(t, requests.Load())
		})
	}
}

var errWebhookSigningSecretTransport = errors.New("transport failed")

func TestWebhookSigningSecretTransportFailureAndCancellation(t *testing.T) {
	t.Parallel()

	for _, operation := range []string{"put", "delete"} {
		for _, failure := range []error{errWebhookSigningSecretTransport, context.Canceled, context.DeadlineExceeded} {
			t.Run(operation+"/"+failure.Error(), func(t *testing.T) {
				t.Parallel()

				var count atomic.Int64

				client, err := cm.NewClient("https://contentful.invalid", cm.NewAccessTokenSecuritySource("test"), cm.WithClient(newContentfulHTTPClient(&http.Client{Transport: contentfulRetryTestRoundTripper(func(req *http.Request) (*http.Response, error) {
					count.Add(1)

					deadline, ok := req.Context().Deadline()
					assert.True(t, ok)
					assert.LessOrEqual(t, time.Until(deadline), 2*time.Minute)

					return nil, fmt.Errorf("%w: %s", failure, webhookSigningSecretTestValue)
				})})))
				require.NoError(t, err)

				implementation := &webhookSigningSecretResource{providerData: ContentfulProviderData{client: client}}
				state, identity := webhookSigningSecretTestState(t, types.StringValue(webhookSigningSecretTestValue))

				var diags diag.Diagnostics
				if operation == "put" {
					_, diags = implementation.put(t.Context(), mustWebhookSigningSecretModel(t, state), types.StringNull())
				} else {
					resp := resource.DeleteResponse{State: state, Identity: identity}
					implementation.Delete(t.Context(), resource.DeleteRequest{State: state}, &resp)
					diags = resp.Diagnostics
				}

				require.True(t, diags.HasError())
				assert.EqualValues(t, 1, count.Load())

				for _, d := range diags {
					assert.Contains(t, d.Detail(), failure.Error())
					assert.Contains(t, d.Detail(), "***")
					assert.NotContains(t, d.Detail(), webhookSigningSecretTestValue)
				}
			})
		}
	}
}

func mustWebhookSigningSecretModel(t *testing.T, state tfsdk.State) WebhookSigningSecretModel {
	t.Helper()

	var model WebhookSigningSecretModel
	require.Empty(t, state.Get(t.Context(), &model))

	return model
}

func TestWebhookSigningSecretRedactedMetadataIsOpaque(t *testing.T) {
	t.Parallel()

	for _, redacted := range []string{"", "opaque", "+/=_-", webhookSigningSecretTestValue} {
		var response cm.WebhookSigningSecret
		require.NoError(t, json.Unmarshal([]byte(webhookSigningSecretTestResponse), &response))
		response.RedactedValue = redacted
		model, diags := NewWebhookSigningSecretResourceModelFromResponse(response, "space", types.StringValue(webhookSigningSecretTestUpdatedValue))
		require.Empty(t, diags)
		assert.Equal(t, webhookSigningSecretTestUpdatedValue, model.Value.ValueString())
	}
}

func TestWebhookSigningSecretRejectsUnresolvedScopeBeforeHTTP(t *testing.T) {
	t.Parallel()

	for _, spaceID := range []types.String{types.StringNull(), types.StringUnknown(), types.StringValue("")} {
		var count atomic.Int64

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			count.Add(1)
			w.WriteHeader(http.StatusBadRequest)
		}))
		t.Cleanup(server.Close)
		state, identity := webhookSigningSecretTestState(t, types.StringValue(webhookSigningSecretTestValue))
		require.Empty(t, state.SetAttribute(t.Context(), path.Root("space_id"), spaceID))
		implementation := webhookSigningSecretTestClient(t, server)
		resp := resource.CreateResponse{State: tfsdk.State{Schema: state.Schema}, Identity: identity}
		implementation.Create(t.Context(), resource.CreateRequest{Plan: tfsdk.Plan(state)}, &resp)
		require.True(t, resp.Diagnostics.HasError())
		assert.Zero(t, count.Load())

		read := resource.ReadResponse{State: state, Identity: identity}
		implementation.Read(t.Context(), resource.ReadRequest{State: state}, &read)
		require.True(t, read.Diagnostics.HasError())
		assert.True(t, state.Raw.Equal(read.State.Raw))
		assert.Zero(t, count.Load())

		deleted := resource.DeleteResponse{State: state, Identity: identity}
		implementation.Delete(t.Context(), resource.DeleteRequest{State: state}, &deleted)
		require.True(t, deleted.Diagnostics.HasError())
		assert.True(t, state.Raw.Equal(deleted.State.Raw))
		assert.Zero(t, count.Load())
	}
}
