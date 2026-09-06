package provider_test

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreviewEnvironmentRequestConversionErrorsStopBeforeAPIRequest(t *testing.T) {
	t.Parallel()

	for _, operation := range []string{"create", "update"} {
		t.Run(operation, func(t *testing.T) {
			t.Parallel()

			var requestCount atomic.Int64

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				requestCount.Add(1)
				w.WriteHeader(http.StatusInternalServerError)
			}))
			t.Cleanup(server.Close)

			client, err := cm.NewClient(server.URL, cm.NewAccessTokenSecuritySource("test-token"), cm.WithClient(server.Client()))
			require.NoError(t, err)

			ctx := t.Context()
			resourceSchema := PreviewEnvironmentResourceSchema(ctx)
			prior := previewEnvironmentCheckpointModel(types.StringValue("space/preview"), types.StringValue("preview"), "Prior")
			planned := prior
			planned.Name = types.StringUnknown()

			if operation == "create" {
				planned.ID = types.StringUnknown()
				planned.PreviewEnvironmentID = types.StringUnknown()
			}

			plan := tfsdk.Plan{Schema: resourceSchema}
			require.Empty(t, plan.Set(ctx, planned))

			implementation := NewPreviewEnvironmentResourceWithClient(client)

			var diagnostics diag.Diagnostics

			switch operation {
			case "create":
				configured := planned
				configured.ID = types.StringNull()
				configured.PreviewEnvironmentID = types.StringNull()
				configured.Name = types.StringValue("Configured")
				configPlan := tfsdk.Plan{Schema: resourceSchema}
				require.Empty(t, configPlan.Set(ctx, configured))

				response := resource.CreateResponse{State: tfsdk.State{Schema: resourceSchema}}
				implementation.Create(ctx, resource.CreateRequest{
					Config: tfsdk.Config{Raw: configPlan.Raw, Schema: resourceSchema},
					Plan:   plan,
				}, &response)
				diagnostics = response.Diagnostics

			case "update":
				state := tfsdk.State{Schema: resourceSchema}
				require.Empty(t, state.Set(ctx, prior))

				response := resource.UpdateResponse{State: tfsdk.State{Schema: resourceSchema}}
				implementation.Update(ctx, resource.UpdateRequest{State: state, Plan: plan}, &response)
				diagnostics = response.Diagnostics
			}

			require.True(t, diagnostics.HasError())
			assert.Contains(t, attributeDiagnosticPaths(t, diagnostics), "name")
			assert.Zero(t, requestCount.Load())
		})
	}
}
