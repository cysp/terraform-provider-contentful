//nolint:testpackage // Exercise lifecycle publication and effective-Plan boundaries directly.
package provider

import (
	"bytes"
	"context"
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
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflogtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errAppEventTestTransport = errors.New("transport sentinel")

const (
	appEventTestPath     = "/organizations/organization/app_definitions/app/event_subscription"
	appEventTestSys      = `"sys":{"type":"AppEventSubscription","organization":{"sys":{"type":"Link","linkType":"Organization","id":"organization"}},"appDefinition":{"sys":{"type":"Link","linkType":"AppDefinition","id":"app"}}}`
	appEventHTTPResponse = `{` + appEventTestSys + `,"topics":["Entry.publish","Asset.publish"],"targetUrl":"https://example.invalid/events?secret=sentinel"}`
)

func appEventTestModel() AppEventSubscriptionModel {
	return AppEventSubscriptionModel{IDIdentityModel: IDIdentityModel{ID: types.StringUnknown()}, OrganizationID: types.StringValue("organization"), AppDefinitionID: types.StringValue("app"), Topics: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("Entry.publish"), types.StringValue("Asset.publish")}), TargetURL: types.StringValue("https://example.invalid/events?secret=sentinel"), FilterFunctionID: types.StringNull(), TransformationFunctionID: types.StringNull(), HandlerFunctionID: types.StringNull(), Timeouts: TimeoutsNull()}
}

func appEventTestResource(t *testing.T, handler http.HandlerFunc) *appEventSubscriptionResource {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	client, err := cm.NewClient(server.URL, cm.NewAccessTokenSecuritySource("test-token"), cm.WithClient(newContentfulHTTPClient(server.Client())))
	require.NoError(t, err)

	return &appEventSubscriptionResource{providerData: ContentfulProviderData{client: client}}
}

func appEventTestPlan(t *testing.T, model AppEventSubscriptionModel) tfsdk.Plan {
	t.Helper()
	plan := tfsdk.Plan{Schema: AppEventSubscriptionResourceSchema(t.Context())}
	require.False(t, plan.Set(t.Context(), model).HasError())

	return plan
}

func appEventTestIdentity() *tfsdk.ResourceIdentity {
	s := resourceIdentitySchema(appEventSubscriptionIdentityAttributeNames())

	return &tfsdk.ResourceIdentity{Schema: s, Raw: tftypes.NewValue(s.Type().TerraformType(context.Background()), nil)}
}

func TestAppEventSubscriptionMutationResponses(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		status    int
		body      string
		errorText string
		stored    bool
	}{
		"created": {201, appEventHTTPResponse, "", true}, "updated": {200, appEventHTTPResponse, "", true},
		"duplicate topics equal set": {200, strings.Replace(appEventHTTPResponse, `"Entry.publish","Asset.publish"`, `"Entry.publish","Asset.publish","Entry.publish"`, 1), "cannot fully represent", true},
		"different topics":           {200, strings.Replace(appEventHTTPResponse, "Asset.publish", "Entry.save", 1), "different app event topics", true},
		"missing target":             {200, `{` + appEventTestSys + `,"topics":["Entry.publish","Asset.publish"]}`, "different target_url", true},
		"different parent":           {200, strings.Replace(appEventHTTPResponse, `"id":"app"`, `"id":"other"`, 1), "different app event subscription identity", true},
		"empty functions":            {200, strings.TrimSuffix(appEventHTTPResponse, "}") + `,"functions":{}}`, "", true},
		"null functions":             {200, strings.TrimSuffix(appEventHTTPResponse, "}") + `,"functions":null}`, "Failed to upsert", false},
		"null function role":         {200, strings.TrimSuffix(appEventHTTPResponse, "}") + `,"functions":{"filter":null}}`, "Failed to upsert", false},
		"empty function link":        {200, strings.TrimSuffix(appEventHTTPResponse, "}") + `,"functions":{"filter":{}}}`, "Failed to upsert", false},
		"missing topics":             {200, `{` + appEventTestSys + `}`, "Failed to upsert", false},
		"missing parent":             {200, `{"sys":{"type":"AppEventSubscription"},"topics":["Entry.publish"]}`, "Failed to upsert", false},
		"null topics":                {200, `{` + appEventTestSys + `,"topics":null}`, "Failed to upsert", false},
		"malformed":                  {201, `{"sentinel":`, "Failed to upsert", false},
		"unauthorized":               {401, `{"sys":{"type":"Error","id":"sentinel"},"message":"sentinel"}`, "Failed to upsert", false},
		"server failure":             {500, `{"sys":{"type":"Error","id":"sentinel"},"message":"sentinel"}`, "Failed to upsert", false},
	}
	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var count atomic.Int64

			implementation := appEventTestResource(t, func(w http.ResponseWriter, r *http.Request) {
				count.Add(1)
				assert.Equal(t, http.MethodPut, r.Method)
				assert.Equal(t, appEventTestPath, r.URL.Path)
				assert.Empty(t, r.Header.Get("X-Contentful-Version"))
				assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
				assert.Equal(t, "application/vnd.contentful.management.v1+json", r.Header.Get("Content-Type"))
				raw, err := io.ReadAll(r.Body)
				assert.NoError(t, err)
				assert.JSONEq(t, `{"topics":["Asset.publish","Entry.publish"],"targetUrl":"https://example.invalid/events?secret=sentinel"}`, string(raw))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(test.status)
				_, err = w.Write([]byte(test.body))
				assert.NoError(t, err)
			})
			plan := appEventTestPlan(t, appEventTestModel())
			response := resource.CreateResponse{State: tfsdk.State{Schema: plan.Schema}, Identity: appEventTestIdentity()}
			implementation.Create(t.Context(), resource.CreateRequest{Plan: plan}, &response)
			require.EqualValues(t, 1, count.Load())

			if test.errorText == "" {
				require.False(t, response.Diagnostics.HasError(), response.Diagnostics)
			} else {
				require.True(t, response.Diagnostics.HasError())
				assert.Contains(t, response.Diagnostics.Errors()[0].Summary(), test.errorText)
			}

			for _, d := range response.Diagnostics {
				assert.NotContains(t, d.Detail(), "sentinel")
			}

			if test.stored {
				var state AppEventSubscriptionModel
				require.False(t, response.State.Get(t.Context(), &state).HasError())
				assert.Equal(t, "organization/app", state.ID.ValueString())
				assert.Equal(t, "app", state.AppDefinitionID.ValueString())

				if name == "missing target" {
					assert.True(t, state.TargetURL.IsNull())
				}

				if name == "different topics" {
					assert.Contains(t, state.Topics.Elements(), types.StringValue("Entry.save"))
				}
			}
		})
	}
}

func TestAppEventSubscriptionInvalidValuesBlockMutation(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		change    func(*AppEventSubscriptionModel)
		summaries []string
	}{
		"organization unknown": {
			change:    func(m *AppEventSubscriptionModel) { m.OrganizationID = types.StringUnknown() },
			summaries: []string{"Unexpected unknown string"},
		},
		"organization null": {
			change:    func(m *AppEventSubscriptionModel) { m.OrganizationID = types.StringNull() },
			summaries: []string{"Unexpected null string"},
		},
		"organization empty": {
			change:    func(m *AppEventSubscriptionModel) { m.OrganizationID = types.StringValue("") },
			summaries: []string{"Invalid app event subscription identity"},
		},
		"app definition unknown": {
			change:    func(m *AppEventSubscriptionModel) { m.AppDefinitionID = types.StringUnknown() },
			summaries: []string{"Unexpected unknown string"},
		},
		"app definition null": {
			change:    func(m *AppEventSubscriptionModel) { m.AppDefinitionID = types.StringNull() },
			summaries: []string{"Unexpected null string"},
		},
		"app definition empty": {
			change:    func(m *AppEventSubscriptionModel) { m.AppDefinitionID = types.StringValue("") },
			summaries: []string{"Invalid app event subscription identity"},
		},
		"target_url": {
			change:    func(m *AppEventSubscriptionModel) { m.TargetURL = types.StringUnknown() },
			summaries: []string{"Unexpected unknown string"},
		},
		"filter_function_id": {
			change:    func(m *AppEventSubscriptionModel) { m.FilterFunctionID = types.StringUnknown() },
			summaries: []string{"Unexpected unknown string"},
		},
		"transformation_function_id": {
			change:    func(m *AppEventSubscriptionModel) { m.TransformationFunctionID = types.StringUnknown() },
			summaries: []string{"Unexpected unknown string"},
		},
		"handler_function_id": {
			change:    func(m *AppEventSubscriptionModel) { m.HandlerFunctionID = types.StringUnknown() },
			summaries: []string{"Unexpected unknown string"},
		},
		"topics unknown": {
			change:    func(m *AppEventSubscriptionModel) { m.Topics = types.SetUnknown(types.StringType) },
			summaries: []string{"Unexpected unknown string set"},
		},
		"topics null": {
			change:    func(m *AppEventSubscriptionModel) { m.Topics = types.SetNull(types.StringType) },
			summaries: []string{"Invalid app event topics"},
		},
		"topics empty": {
			change:    func(m *AppEventSubscriptionModel) { m.Topics = types.SetValueMust(types.StringType, []attr.Value{}) },
			summaries: []string{"Invalid app event topics"},
		},
		"topic unknown": {
			change: func(m *AppEventSubscriptionModel) {
				m.Topics = types.SetValueMust(types.StringType, []attr.Value{types.StringUnknown()})
			},
			summaries: []string{"Unexpected unknown string"},
		},
		"topic null": {
			change: func(m *AppEventSubscriptionModel) {
				m.Topics = types.SetValueMust(types.StringType, []attr.Value{types.StringNull()})
			},
			summaries: []string{"Unexpected null string"},
		},
		"topic empty": {
			change: func(m *AppEventSubscriptionModel) {
				m.Topics = types.SetValueMust(types.StringType, []attr.Value{types.StringValue("")})
			},
			summaries: []string{"Invalid app event topic"},
		},
		"target empty": {
			change:    func(m *AppEventSubscriptionModel) { m.TargetURL = types.StringValue("") },
			summaries: []string{"Invalid app event target URL"},
		},
		"function empty": {
			change:    func(m *AppEventSubscriptionModel) { m.FilterFunctionID = types.StringValue("") },
			summaries: []string{"Invalid Function ID"},
		},
		"independent errors": {
			change: func(m *AppEventSubscriptionModel) {
				m.OrganizationID = types.StringUnknown()
				m.AppDefinitionID = types.StringValue("")
				m.Topics = types.SetValueMust(types.StringType, []attr.Value{types.StringNull()})
				m.TargetURL = types.StringValue("")
			},
			summaries: []string{"Unexpected unknown string", "Invalid app event subscription identity", "Unexpected null string", "Invalid app event target URL"},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var count atomic.Int64

			implementation := appEventTestResource(t, func(http.ResponseWriter, *http.Request) { count.Add(1) })
			model := appEventTestModel()
			test.change(&model)
			plan := appEventTestPlan(t, model)
			response := resource.CreateResponse{State: tfsdk.State{Schema: plan.Schema}, Identity: appEventTestIdentity()}
			implementation.Create(t.Context(), resource.CreateRequest{Plan: plan}, &response)
			assert.Zero(t, count.Load())
			prior := appEventTestPlan(t, appEventTestModel())
			update := resource.UpdateResponse{State: tfsdk.State{Schema: plan.Schema, Raw: prior.Raw}, Identity: appEventTestIdentity()}
			implementation.Update(t.Context(), resource.UpdateRequest{Plan: plan, Config: tfsdk.Config(prior), State: tfsdk.State(prior)}, &update)
			assert.Zero(t, count.Load())

			for _, diagnostics := range []diag.Diagnostics{response.Diagnostics, update.Diagnostics} {
				assert.True(t, diagnostics.HasError())

				summaries := make([]string, 0, len(diagnostics))
				for _, diagnostic := range diagnostics {
					summaries = append(summaries, diagnostic.Summary())
				}

				assert.ElementsMatch(t, test.summaries, summaries)
			}
		})
	}
}

func TestAppEventSubscriptionErrorRedaction(t *testing.T) {
	t.Parallel()

	for _, id := range []string{"ValidationFailed", "sentinel"} {
		response := &cm.ErrorStatusCode{StatusCode: 422, Response: cm.NewErrorApplicationJSONError(cm.Error{Sys: cm.ErrorSys{Type: cm.ErrorSysTypeError, ID: id}, Message: cm.NewOptString("sentinel"), Details: []byte(`{"value":"sentinel"}`)})}
		detail := appEventSubscriptionErrorDetail(response, errAppEventTestTransport)
		assert.NotContains(t, detail, "sentinel")
		assert.Contains(t, detail, "422")
	}

	assert.NotContains(t, appEventSubscriptionErrorDetail(nil, errAppEventTestTransport), "sentinel")
}

func TestAppEventSubscriptionReadDeleteFailures(t *testing.T) {
	t.Parallel()

	for _, method := range []string{http.MethodGet, http.MethodDelete} {
		for _, status := range []int{204, 404, 403, 503, 429} {
			if method == http.MethodGet && status == 204 {
				continue
			}

			t.Run(fmt.Sprintf("%s_%d", method, status), func(t *testing.T) {
				t.Parallel()

				var count atomic.Int64

				implementation := appEventTestResource(t, appEventRetryHandler(t, method, status, &count))
				model := appEventTestModel()
				model.ID = types.StringValue("organization/app")
				plan := appEventTestPlan(t, model)
				state := tfsdk.State(plan)

				ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
				defer cancel()

				if method == http.MethodGet {
					response := resource.ReadResponse{State: state, Identity: appEventTestIdentity()}
					implementation.Read(ctx, resource.ReadRequest{State: state}, &response)
					assert.Equal(t, status == 403, response.Diagnostics.HasError(), response.Diagnostics)

					if status == 404 {
						assert.True(t, response.State.Raw.IsNull())
					}
				} else {
					response := resource.DeleteResponse{State: state}
					implementation.Delete(ctx, resource.DeleteRequest{State: state}, &response)
					assert.Equal(t, status == 403 || status == 503 || status == 429, response.Diagnostics.HasError(), response.Diagnostics)
				}

				expected := int64(1)
				if method == http.MethodGet && (status == 429 || status == 503) {
					expected = 2
				}

				assert.Equal(t, expected, count.Load())
			})
		}
	}
}

func TestAppEventSubscriptionFunctionContradictions(t *testing.T) {
	t.Parallel()

	for _, role := range []string{"filter", "transformation", "handler"} {
		t.Run(role, func(t *testing.T) {
			t.Parallel()

			responseBody := strings.TrimSuffix(appEventHTTPResponse, "}") + `,"functions":{"` + role + `":{"sys":{"type":"Link","linkType":"Function","id":"retained"}}}}`
			implementation := appEventTestResource(t, func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(responseBody))
			})
			data, diags, consistency := implementation.put(t.Context(), appEventTestModel())
			assert.Empty(t, diags)
			require.True(t, consistency.HasError())
			assert.Contains(t, consistency.Errors()[0].Summary(), role+"_function_id")

			switch role {
			case "filter":
				assert.Equal(t, "retained", data.FilterFunctionID.ValueString())
			case "transformation":
				assert.Equal(t, "retained", data.TransformationFunctionID.ValueString())
			case "handler":
				assert.Equal(t, "retained", data.HandlerFunctionID.ValueString())
			}
		})
	}
}

func TestAppEventSubscriptionReadIdentityMismatch(t *testing.T) {
	t.Parallel()
	implementation := appEventTestResource(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(strings.Replace(appEventHTTPResponse, `"id":"organization"`, `"id":"other"`, 1)))
	})
	model := appEventTestModel()
	model.ID = types.StringValue("organization/app")
	plan := appEventTestPlan(t, model)
	state := tfsdk.State(plan)
	response := resource.ReadResponse{State: state, Identity: appEventTestIdentity()}
	implementation.Read(t.Context(), resource.ReadRequest{State: state}, &response)
	require.True(t, response.Diagnostics.HasError())

	var actual AppEventSubscriptionModel
	require.False(t, response.State.Get(t.Context(), &actual).HasError())
	assert.Equal(t, "organization", actual.OrganizationID.ValueString())
	assert.Equal(t, "organization/app", actual.ID.ValueString())
}

func appEventRetryHandler(t *testing.T, method string, status int, count *atomic.Int64) http.HandlerFunc {
	t.Helper()

	return func(w http.ResponseWriter, r *http.Request) {
		attempt := count.Add(1)

		assert.Equal(t, method, r.Method)
		assert.Equal(t, appEventTestPath, r.URL.Path)
		assert.Empty(t, r.Header.Get("X-Contentful-Version"))
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Retry-After", "0")

		if status == 429 && attempt > 1 {
			if method == http.MethodGet {
				_, _ = w.Write([]byte(appEventHTTPResponse))
			} else {
				w.WriteHeader(http.StatusNoContent)
			}

			return
		}

		if status == 503 && method == http.MethodGet && attempt > 1 {
			_, _ = w.Write([]byte(appEventHTTPResponse))

			return
		}

		w.Header().Set("X-Contentful-Ratelimit-Reset", "0")
		w.WriteHeader(status)

		if status != 204 {
			_, _ = w.Write([]byte(`{"sys":{"type":"Error","id":"sentinel"},"message":"sentinel"}`))
		}
	}
}

func TestAppEventSubscriptionUpsertRateLimitIsNotRetried(t *testing.T) {
	t.Parallel()

	var count atomic.Int64

	implementation := appEventTestResource(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if count.Add(1) == 1 {
			w.Header().Set("X-Contentful-Ratelimit-Reset", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"sys":{"type":"Error","id":"RateLimitExceeded"},"message":"sentinel"}`))

			return
		}

		_, _ = w.Write([]byte(appEventHTTPResponse))
	})

	var logs bytes.Buffer

	ctx := tflogtest.RootLogger(t.Context(), &logs)
	_, diags, consistency := implementation.put(ctx, appEventTestModel())
	require.True(t, diags.HasError())
	assert.Contains(t, diags.Errors()[0].Detail(), "429")
	assert.NotContains(t, diags.Errors()[0].Detail(), "sentinel")
	assert.Empty(t, consistency)
	assert.EqualValues(t, 1, count.Load())
	assert.NotContains(t, logs.String(), "sentinel")
	assert.NotContains(t, logs.String(), "test-token")
}

func TestAppEventSubscriptionUpsertDeadline(t *testing.T) {
	t.Parallel()

	var count atomic.Int64

	release := make(chan struct{})
	defer close(release)

	implementation := appEventTestResource(t, func(_ http.ResponseWriter, r *http.Request) {
		count.Add(1)

		_, _ = io.Copy(io.Discard, r.Body)
		select {
		case <-r.Context().Done():
		case <-release:
		}
	})

	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()

	_, diags, _ := implementation.put(ctx, appEventTestModel())
	require.True(t, diags.HasError())
	assert.Contains(t, diags.Errors()[0].Detail(), "deadline")
	assert.EqualValues(t, 1, count.Load())
}

func TestAppEventSubscriptionCanceledUpdate(t *testing.T) {
	t.Parallel()

	var count atomic.Int64

	implementation := appEventTestResource(t, func(http.ResponseWriter, *http.Request) { count.Add(1) })
	model := appEventTestModel()
	model.ID = types.StringValue("organization/app")
	prior := appEventTestPlan(t, model)
	model.TargetURL = types.StringValue("https://example.invalid/changed")
	plan := appEventTestPlan(t, model)
	response := resource.UpdateResponse{State: tfsdk.State(prior), Identity: appEventTestIdentity()}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	implementation.Update(ctx, resource.UpdateRequest{State: tfsdk.State(prior), Plan: plan}, &response)
	require.True(t, response.Diagnostics.HasError())
	assert.Contains(t, response.Diagnostics.Errors()[0].Detail(), "The Contentful request was canceled.")
	assert.Zero(t, count.Load())
	assert.True(t, response.State.Raw.Equal(prior.Raw))
}

func TestAppEventSubscriptionUpdateRecoveryState(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		status     int
		body       string
		checkpoint bool
	}{
		"contradictory response":   {200, strings.Replace(appEventHTTPResponse, `"id":"app"`, `"id":"other"`, 1), true},
		"malformed success":        {200, `{"sentinel":`, false},
		"ambiguous server failure": {500, `{"sys":{"type":"Error","id":"sentinel"},"message":"sentinel"}`, false},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var count atomic.Int64

			implementation := appEventTestResource(t, func(w http.ResponseWriter, r *http.Request) {
				count.Add(1)
				assert.Equal(t, http.MethodPut, r.Method)
				assert.Equal(t, appEventTestPath, r.URL.Path)
				assert.Empty(t, r.Header.Get("X-Contentful-Version"))
				raw, err := io.ReadAll(r.Body)
				assert.NoError(t, err)
				assert.JSONEq(t, `{"topics":["Asset.publish","Entry.publish"],"targetUrl":"https://example.invalid/planned"}`, string(raw))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(test.status)
				_, _ = w.Write([]byte(test.body))
			})
			priorModel := appEventTestModel()
			priorModel.ID = types.StringValue("organization/app")
			priorModel.TargetURL = types.StringValue("https://example.invalid/prior")
			prior := appEventTestPlan(t, priorModel)
			model := priorModel
			model.TargetURL = types.StringValue("https://example.invalid/planned")
			plan := appEventTestPlan(t, model)
			response := resource.UpdateResponse{State: tfsdk.State(prior), Identity: appEventTestIdentity()}
			implementation.Update(t.Context(), resource.UpdateRequest{State: tfsdk.State(prior), Plan: plan, Config: tfsdk.Config(prior)}, &response)
			require.True(t, response.Diagnostics.HasError())
			assert.EqualValues(t, 1, count.Load())

			var actual AppEventSubscriptionModel
			require.False(t, response.State.Get(t.Context(), &actual).HasError())
			assert.Equal(t, "organization/app", actual.ID.ValueString())
			assert.Equal(t, "app", actual.AppDefinitionID.ValueString())

			if test.checkpoint {
				assert.Equal(t, "https://example.invalid/events?secret=sentinel", actual.TargetURL.ValueString())
			} else {
				assert.True(t, response.State.Raw.Equal(prior.Raw))
			}
		})
	}
}
