//nolint:testpackage // Create endpoint selection depends on package-local resource implementation details.
package provider

import (
	"net/http"
	"net/http/httptest"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreviewEnvironmentCreateEndpointFollowsConfiguredIDOwnership(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		configID       types.String
		planID         types.String
		expectedMethod string
		expectedPath   string
		expectedNoIO   bool
	}{
		"omitted ID with unknown plan uses generated-ID endpoint": {
			configID:       types.StringNull(),
			planID:         types.StringUnknown(),
			expectedMethod: http.MethodPost,
			expectedPath:   "/spaces/space/preview_environments",
		},
		"omitted ID with carried known plan uses generated-ID endpoint": {
			configID:       types.StringNull(),
			planID:         types.StringValue("prior-response-id"),
			expectedMethod: http.MethodPost,
			expectedPath:   "/spaces/space/preview_environments",
		},
		"configured known ID uses selected-ID endpoint": {
			configID:       types.StringValue("selected-id"),
			planID:         types.StringValue("selected-id"),
			expectedMethod: http.MethodPut,
			expectedPath:   "/spaces/space/preview_environments/selected-id",
		},
		"configured known ID with unknown plan fails closed": {
			configID:     types.StringValue("selected-id"),
			planID:       types.StringUnknown(),
			expectedNoIO: true,
		},
		"unknown configured ID fails closed": {
			configID:     types.StringUnknown(),
			planID:       types.StringUnknown(),
			expectedNoIO: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var requests []struct {
				method string
				path   string
			}

			testServer := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				requests = append(requests, struct {
					method string
					path   string
				}{method: request.Method, path: request.URL.Path})

				response.WriteHeader(http.StatusInternalServerError)
			}))
			t.Cleanup(testServer.Close)

			client, err := cm.NewClient(
				testServer.URL,
				cm.NewAccessTokenSecuritySource("access-token"),
				cm.WithClient(testServer.Client()),
			)
			require.NoError(t, err)

			ctx := t.Context()
			resourceSchema := PreviewEnvironmentResourceSchema(ctx)
			plan := previewEnvironmentCreateTestPlan(t, resourceSchema, test.planID)
			configPlan := previewEnvironmentCreateTestPlan(t, resourceSchema, test.configID)
			config := tfsdk.Config{Raw: configPlan.Raw, Schema: resourceSchema}

			implementation := previewEnvironmentResource{providerData: ContentfulProviderData{client: client}}
			response := resource.CreateResponse{State: tfsdk.State{Schema: resourceSchema}}
			implementation.Create(ctx, resource.CreateRequest{Config: config, Plan: plan}, &response)

			if test.expectedNoIO {
				require.True(t, response.Diagnostics.HasError())
				assert.Contains(t, mutationDiagnosticPaths(t, response.Diagnostics), "preview_environment_id")
				assert.Empty(t, requests)

				return
			}

			require.Len(t, requests, 1)
			assert.Equal(t, test.expectedMethod, requests[0].method)
			assert.Equal(t, test.expectedPath, requests[0].path)
		})
	}
}

func previewEnvironmentCreateTestPlan(
	t *testing.T,
	resourceSchema schema.Schema,
	previewEnvironmentID types.String,
) tfsdk.Plan {
	t.Helper()

	model := PreviewEnvironmentModel{
		IDIdentityModel: IDIdentityModel{ID: types.StringUnknown()},
		PreviewEnvironmentIdentityModel: PreviewEnvironmentIdentityModel{
			SpaceID:              types.StringValue("space"),
			PreviewEnvironmentID: previewEnvironmentID,
		},
		Name:        types.StringValue("Preview"),
		Description: types.StringValue(""),
		ContentTypeConfigurations: NewTypedMap(map[string]TypedObject[PreviewEnvironmentContentTypeConfigurationValue]{
			"page": NewTypedObject(PreviewEnvironmentContentTypeConfigurationValue{
				URL: types.StringValue("https://preview.invalid/page"),
			}),
		}),
		Timeouts: TimeoutsNull(),
	}

	plan := tfsdk.Plan{Schema: resourceSchema}
	diagnostics := plan.Set(t.Context(), &model)
	require.False(t, diagnostics.HasError(), diagnostics.Errors())

	return plan
}
