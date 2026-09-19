package provider

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
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

func TestAppActionMutationNoReplay(t *testing.T) {
	t.Parallel()

	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete} {
		for _, status := range []int{http.StatusTooManyRequests, http.StatusTemporaryRedirect, http.StatusInternalServerError} {
			t.Run(method+strconv.Itoa(status), func(t *testing.T) {
				t.Parallel()

				var count atomic.Int64

				implementation := appActionTestResource(t, func(w http.ResponseWriter, r *http.Request) {
					count.Add(1)
					assert.Equal(t, method, r.Method)
					assert.Empty(t, r.Header.Get("X-Contentful-Version"))
					w.Header().Set("Content-Type", "application/json")
					w.Header().Set("Location", "/redirect")
					w.Header().Set("X-Contentful-Ratelimit-Reset", "0")
					w.WriteHeader(status)
					_, _ = w.Write([]byte(`{"sys":{"type":"Error","id":"Failure"}}`))
				})
				plan := appActionTestPlan(t, appActionTestModel())
				config := tfsdk.Config(plan)
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

				assert.EqualValues(t, 1, count.Load())
			})
		}
	}
}

func TestAppActionMutationPublishesRecoveryState(t *testing.T) {
	t.Parallel()

	for _, create := range []bool{false, true} {
		t.Run(strconv.FormatBool(create), func(t *testing.T) {
			t.Parallel()
			implementation := appActionTestResource(t, func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")

				if create {
					w.WriteHeader(http.StatusCreated)
				}

				_, _ = w.Write([]byte(strings.Replace(actionDiscoveryFixture, `"name":"Action"`, `"name":"Changed"`, 1)))
			})

			model := appActionTestModel()
			if create {
				model.ID = types.StringUnknown()
				model.AppActionID = types.StringUnknown()
			}

			plan := appActionTestPlan(t, model)
			config := tfsdk.Config(plan)

			identitySchema := resourceIdentitySchema(appActionIdentityAttributeNames())
			identity := &tfsdk.ResourceIdentity{Schema: identitySchema, Raw: tftypes.NewValue(identitySchema.Type().TerraformType(t.Context()), nil)}

			var stored tfsdk.State

			if create {
				response := resource.CreateResponse{State: tfsdk.State{Schema: plan.Schema}, Identity: identity}
				implementation.Create(t.Context(), resource.CreateRequest{Plan: plan, Config: config}, &response)
				require.True(t, response.Diagnostics.HasError())
				stored = response.State
			} else {
				response := resource.UpdateResponse{State: tfsdk.State(plan), Identity: identity}
				previous := appActionTestModel()
				previous.Name = types.StringValue("Previous")
				implementation.Update(t.Context(), resource.UpdateRequest{Plan: plan, Config: config, State: tfsdk.State(appActionTestPlan(t, previous))}, &response)
				require.True(t, response.Diagnostics.HasError())
				stored = response.State
			}

			var actual AppActionModel
			require.False(t, stored.Get(t.Context(), &actual).HasError())
			assert.Equal(t, "Changed", actual.Name.ValueString())
			assert.Equal(t, "org/app/action", actual.ID.ValueString())
		})
	}
}

func TestAppActionResourceNotFound(t *testing.T) {
	t.Parallel()
	implementation := appActionTestResource(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"sys":{"type":"Error","id":"NotFound"}}`))
	})
	plan := appActionTestPlan(t, appActionTestModel())
	state := tfsdk.State(plan)
	read := resource.ReadResponse{State: state}
	implementation.Read(t.Context(), resource.ReadRequest{State: state}, &read)
	require.False(t, read.Diagnostics.HasError())
	assert.True(t, read.State.Raw.IsNull())
	deleted := resource.DeleteResponse{State: state}
	implementation.Delete(t.Context(), resource.DeleteRequest{State: state}, &deleted)
	require.False(t, deleted.Diagnostics.HasError())
}

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
				_, _ = w.Write([]byte(strings.Replace(actionDiscoveryFixture, `"id":"action"`, `"id":"`+actionID+`"`, 1)))
			})
			model := appActionTestModel()
			model.ID = types.StringUnknown()
			model.AppActionID = types.StringUnknown()
			plan := appActionTestPlan(t, model)
			response := resource.CreateResponse{State: tfsdk.State{Schema: plan.Schema}}
			implementation.Create(t.Context(), resource.CreateRequest{Plan: plan, Config: tfsdk.Config(plan)}, &response)
			require.Len(t, response.Diagnostics, 1)
			require.True(t, response.Diagnostics.HasError())
			assert.Equal(t, "Invalid App Action response identity", response.Diagnostics[0].Summary())
			assert.EqualValues(t, 1, calls.Load())
			assert.True(t, response.State.Raw.IsNull())
		})
	}
}
