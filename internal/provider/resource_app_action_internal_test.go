package provider

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func appActionTestResource(t *testing.T, handler http.HandlerFunc) *appActionResource {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	client, err := cm.NewClient(server.URL, cm.NewAccessTokenSecuritySource("test-token"), cm.WithClient(newContentfulHTTPClient(server.Client())))
	require.NoError(t, err)

	return &appActionResource{providerData: ContentfulProviderData{client: client}}
}

func appActionTestPlan(t *testing.T, model AppActionModel) tfsdk.Plan {
	t.Helper()
	plan := tfsdk.Plan{Schema: AppActionResourceSchema(t.Context())}
	require.False(t, plan.Set(t.Context(), model).HasError())

	return plan
}

func TestAppActionMutationsStopAfterUncertainFailureOrRedirect(t *testing.T) {
	t.Parallel()

	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete} {
		for _, status := range []int{http.StatusTooManyRequests, http.StatusTemporaryRedirect, http.StatusInternalServerError} {
			t.Run(method+strconv.Itoa(status), func(t *testing.T) {
				t.Parallel()

				var count atomic.Int64

				implementation := appActionTestResource(t, func(w http.ResponseWriter, r *http.Request) {
					attempt := count.Add(1)

					assert.Equal(t, method, r.Method)
					assert.Empty(t, r.Header.Get("X-Contentful-Version"))
					w.Header().Set("Content-Type", "application/json")
					w.Header().Set("Location", "/redirect")
					w.Header().Set("X-Contentful-Ratelimit-Reset", "0")

					responseStatus := status
					if status == http.StatusTooManyRequests && attempt > 1 {
						responseStatus = http.StatusInternalServerError
					}

					w.WriteHeader(responseStatus)
					_, _ = w.Write([]byte(testJSON(map[string]any{"sys": map[string]any{"type": "Error", "id": "Failure"}})))
				})
				plan := appActionTestPlan(t, appActionTestModel())
				config := tfsdk.Config(plan)

				if method == http.MethodPost {
					unconfigured := appActionTestModel()
					unconfigured.AppActionID = types.StringNull()
					config = tfsdk.Config(appActionTestPlan(t, unconfigured))
				}

				previous := appActionTestModel()
				previous.Name = types.StringValue("Previous")
				state := tfsdk.State(appActionTestPlan(t, previous))

				switch method {
				case http.MethodPost:
					response := resource.CreateResponse{State: tfsdk.State{Schema: plan.Schema}}
					implementation.Create(t.Context(), resource.CreateRequest{Plan: plan, Config: config}, &response)
					require.True(t, response.Diagnostics.HasError())
				case http.MethodPut:
					response := resource.UpdateResponse{State: state}
					implementation.Update(t.Context(), resource.UpdateRequest{Plan: plan, Config: config, State: state}, &response)
					require.True(t, response.Diagnostics.HasError())
				case http.MethodDelete:
					response := resource.DeleteResponse{State: state}
					implementation.Delete(t.Context(), resource.DeleteRequest{State: state}, &response)
					require.True(t, response.Diagnostics.HasError())
				}

				expected := int64(1)
				if status == http.StatusTooManyRequests {
					expected = 2
				}

				assert.Equal(t, expected, count.Load())
			})
		}
	}
}

func TestAppActionDeleteNotFound(t *testing.T) {
	t.Parallel()
	implementation := appActionTestResource(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(testJSON(map[string]any{"sys": map[string]any{"type": "Error", "id": "NotFound"}})))
	})
	plan := appActionTestPlan(t, appActionTestModel())
	state := tfsdk.State(plan)
	deleted := resource.DeleteResponse{State: state}
	implementation.Delete(t.Context(), resource.DeleteRequest{State: state}, &deleted)
	require.False(t, deleted.Diagnostics.HasError())
}

//nolint:forcetypeassert // Fixture mutations target independently defined object shapes.
func TestAppActionCreateRejectsUnaddressableID(t *testing.T) {
	t.Parallel()

	for _, actionID := range []string{"", ".", "..", "a/b"} {
		t.Run(actionID, func(t *testing.T) {
			t.Parallel()

			var calls atomic.Int64

			implementation := appActionTestResource(t, func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				assert.Equal(t, http.MethodPost, r.Method)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(mutateTestJSON(actionDiscoveryFixture, func(document map[string]any) { document["sys"].(map[string]any)["id"] = actionID })))
			})
			model := appActionTestModel()
			model.ID = types.StringUnknown()
			model.AppActionID = types.StringUnknown()
			plan := appActionTestPlan(t, model)
			response := resource.CreateResponse{State: tfsdk.State{Schema: plan.Schema}}
			configModel := model
			configModel.AppActionID = types.StringNull()
			implementation.Create(t.Context(), resource.CreateRequest{Plan: plan, Config: tfsdk.Config(appActionTestPlan(t, configModel))}, &response)
			require.Len(t, response.Diagnostics, 1)
			require.True(t, response.Diagnostics.HasError())
			assert.Equal(t, "Invalid App Action response identity", response.Diagnostics[0].Summary())
			assert.EqualValues(t, 1, calls.Load())
			assert.True(t, response.State.Raw.IsNull())
		})
	}
}

func TestAppActionValidateConfigUnknownValues(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name  string
		alter func(*AppActionModel)
	}{
		{"category", func(model *AppActionModel) { model.Category = types.StringUnknown() }},
		{"parameters", func(model *AppActionModel) { model.Parameters = jsontypes.NewNormalizedUnknown() }},
		{"schema", func(model *AppActionModel) {
			model.Parameters = jsontypes.NewNormalizedNull()
			model.ParametersSchema = jsontypes.NewNormalizedUnknown()
		}},
		{"executor", func(model *AppActionModel) {
			model.Type = types.StringUnknown()
			model.URL = types.StringNull()
			model.FunctionID = types.StringUnknown()
		}},
		{"builtin parameters", func(model *AppActionModel) {
			model.Category = types.StringValue("Entries.v1.0")
			model.Parameters = jsontypes.NewNormalizedUnknown()
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			model := appActionTestModel()
			test.alter(&model)
			model.AppActionID = types.StringNull()
			request := resource.ValidateConfigRequest{Config: tfsdk.Config(appActionTestPlan(t, model))}
			response := resource.ValidateConfigResponse{}
			implementation := &appActionResource{}
			implementation.ValidateConfig(t.Context(), request, &response)
			assert.False(t, response.Diagnostics.HasError(), response.Diagnostics)
		})
	}
}

func TestAppActionUpdateRejectsBuiltinParameterChange(t *testing.T) {
	t.Parallel()

	implementation := appActionTestResource(t, func(http.ResponseWriter, *http.Request) {
		t.Error("invalid built-in parameters must not reach Contentful")
	})
	previous := appActionTestModel()
	previous.Category = types.StringValue("Entries.v1.0")
	previous.Parameters = jsontypes.NewNormalizedNull()
	model := previous
	model.Parameters = jsontypes.NewNormalizedValue(testJSON([]any{}))
	config := model
	config.Category = types.StringValue("Custom")
	state := tfsdk.State(appActionTestPlan(t, previous))
	request := resource.UpdateRequest{Plan: appActionTestPlan(t, model), Config: tfsdk.Config(appActionTestPlan(t, config)), State: state}
	response := resource.UpdateResponse{State: state}
	implementation.Update(t.Context(), request, &response)

	require.Len(t, response.Diagnostics, 1)
	assert.Equal(t, "Built-in App Action parameters are read-only", response.Diagnostics[0].Summary())
	assert.True(t, response.State.Raw.Equal(state.Raw))
}

func TestAppActionMutationsRetryRateLimit(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		method string
		path   string
		status int
	}{
		{method: http.MethodPost, path: "/organizations/org/app_definitions/app/actions", status: http.StatusCreated},
		{method: http.MethodPut, path: "/organizations/org/app_definitions/app/actions/action", status: http.StatusOK},
		{method: http.MethodDelete, path: "/organizations/org/app_definitions/app/actions/action", status: http.StatusNoContent},
	} {
		t.Run(test.method, func(t *testing.T) {
			t.Parallel()

			expectedBody := testJSON(map[string]any{
				"name":       "Action",
				"category":   "Custom",
				"type":       "endpoint",
				"url":        "https://example.invalid/action",
				"parameters": []any{},
			})

			var bodies []string

			httpClient, retryClient := contentfulRetryTestClient(t, &http.Client{
				Transport: contentfulRetryTestRoundTripper(func(request *http.Request) (*http.Response, error) {
					assert.Equal(t, test.method, request.Method)
					assert.Equal(t, test.path, request.URL.Path)
					assert.Empty(t, request.Header.Values("X-Contentful-Version"))

					var data []byte

					if request.Body != nil {
						var err error

						data, err = io.ReadAll(request.Body)
						require.NoError(t, err)
					}

					bodies = append(bodies, string(data))

					status := test.status
					if len(bodies) == 1 {
						status = http.StatusTooManyRequests
					}

					response := contentfulRetryTestResponse(request, status)
					response.Header.Set("Content-Type", "application/json")

					body := actionDiscoveryFixture
					if status == http.StatusTooManyRequests {
						body = testJSON(map[string]any{"sys": map[string]any{"type": "Error", "id": "RateLimitExceeded"}, "message": "rate limited"})
					}

					response.Body = io.NopCloser(strings.NewReader(body))

					return response, nil
				}),
			})
			removeContentfulRetryTestDelay(retryClient)

			client, err := cm.NewClient("https://contentful.invalid", cm.NewAccessTokenSecuritySource("test"), cm.WithClient(httpClient))
			require.NoError(t, err)

			implementation := &appActionResource{
				providerData: ContentfulProviderData{
					client: client,
				},
			}
			plan := appActionTestPlan(t, appActionTestModel())
			state := tfsdk.State(plan)
			identitySchema := resourceIdentitySchema(appActionIdentityAttributeNames())
			identity := &tfsdk.ResourceIdentity{
				Schema: identitySchema,
				Raw:    tftypes.NewValue(identitySchema.Type().TerraformType(t.Context()), nil),
			}

			switch test.method {
			case http.MethodPost:
				configured := appActionTestModel()
				configured.AppActionID = types.StringNull()
				response := resource.CreateResponse{
					State: tfsdk.State{
						Schema: plan.Schema,
					},
					Identity: identity,
				}
				implementation.Create(t.Context(), resource.CreateRequest{
					Plan:   plan,
					Config: tfsdk.Config(appActionTestPlan(t, configured)),
				}, &response)
				require.False(t, response.Diagnostics.HasError(), response.Diagnostics)

				var created AppActionModel
				require.Empty(t, response.State.Get(t.Context(), &created))
				assert.Equal(t, "org/app/action", created.ID.ValueString())
				assert.Equal(t, "Action", created.Name.ValueString())
			case http.MethodPut:
				previous := appActionTestModel()
				previous.Name = types.StringValue("Previous action")
				state = tfsdk.State(appActionTestPlan(t, previous))
				response := resource.UpdateResponse{
					State:    state,
					Identity: identity,
				}
				implementation.Update(t.Context(), resource.UpdateRequest{
					State:  state,
					Plan:   plan,
					Config: tfsdk.Config(plan),
				}, &response)
				require.False(t, response.Diagnostics.HasError(), response.Diagnostics)

				var updated AppActionModel
				require.Empty(t, response.State.Get(t.Context(), &updated))
				assert.Equal(t, "Action", updated.Name.ValueString())
			case http.MethodDelete:
				response := resource.DeleteResponse{
					State: state,
				}
				implementation.Delete(t.Context(), resource.DeleteRequest{
					State: state,
				}, &response)
				require.False(t, response.Diagnostics.HasError(), response.Diagnostics)
			}

			require.Len(t, bodies, 2)

			if test.method == http.MethodDelete {
				assert.Empty(t, bodies[0])
			} else {
				assert.JSONEq(t, expectedBody, bodies[0])
			}

			assert.Equal(t, bodies[0], bodies[1])
		})
	}
}
