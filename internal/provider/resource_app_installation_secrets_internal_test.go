package provider

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Directly exercise Update so Terraform's independent refresh calls cannot be
// mistaken for provider recovery reads after an uncertain mutation result.
func TestAppInstallationSecretsUnsuccessfulUpdate(t *testing.T) {
	t.Parallel()

	type observation struct {
		method, path, body string
		err                error
	}

	var (
		requestMutex sync.Mutex
		requests     []observation
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)

		requestMutex.Lock()

		requests = append(requests, observation{method: r.Method, path: r.URL.Path, body: string(body), err: err})
		requestMutex.Unlock()
		// The complete request was received, but the response fails.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, testJSON(map[string]any{"sys": map[string]any{"type": "Error", "id": "ServerError"}, "message": "Synthetic failure after receiving request"}))
	}))
	t.Cleanup(server.Close)
	client, err := cm.NewClient(server.URL, cm.NewAccessTokenSecuritySource("test-token"))
	require.NoError(t, err)

	implementation := appInstallationResource{providerData: ContentfulProviderData{client: client}}
	state := tfsdk.State(appInstallationTestPlan(t, map[string]string{"secret": `"previous"`, "visible": `"one"`}))
	plan := appInstallationTestPlan(t, map[string]string{"secret": `"next"`, "visible": `"two"`})
	identitySchema := resourceIdentitySchema(appInstallationIdentityAttributeNames())
	identityRaw, err := tftypes.ValueFromJSONWithOpts([]byte(testJSON(map[string]any{"space_id": "space", "environment_id": "master", "app_definition_id": "app"})), identitySchema.Type().TerraformType(t.Context()), tftypes.ValueFromJSONOpts{})
	require.NoError(t, err)

	identity := &tfsdk.ResourceIdentity{Schema: identitySchema, Raw: identityRaw}
	identityBefore := identity.Raw.Copy()
	response := resource.UpdateResponse{State: state, Identity: identity}
	implementation.Update(t.Context(), resource.UpdateRequest{Plan: plan, State: state}, &response)
	require.True(t, response.Diagnostics.HasError())
	require.Len(t, response.Diagnostics, 1)
	assert.Equal(t, "Failed to update app installation", response.Diagnostics[0].Summary())
	assert.True(t, state.Raw.Equal(response.State.Raw))
	assert.True(t, identityBefore.Equal(response.Identity.Raw))
	requestMutex.Lock()
	defer requestMutex.Unlock()

	require.Len(t, requests, 1)
	observed := requests[0]
	require.NoError(t, observed.err)
	assert.Equal(t, http.MethodPut, observed.method)
	assert.Equal(t, "/spaces/space/environments/master/app_installations/app", observed.path)

	var stored map[string]any
	require.NoError(t, json.Unmarshal([]byte(observed.body), &stored))
	assert.Equal(t, map[string]any{"parameters": map[string]any{"secret": "next", "visible": "two"}}, stored)
}

func TestAppInstallationMutationRejectsContradictoryParameters(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name     string
		returned map[string]any
	}{
		{name: "readable secret", returned: map[string]any{"secret": "different", "visible": "one"}},
		{name: "readable ordinary", returned: map[string]any{"secret": "*****", "visible": "different"}},
		{name: "missing ordinary key", returned: map[string]any{"secret": "*****"}},
		{name: "missing secret key", returned: map[string]any{"visible": "one"}},
		{name: "null parameters"},
		{name: "added key", returned: map[string]any{"secret": "*****", "visible": "one", "added": true}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			plan := appInstallationTestPlan(t, map[string]string{"secret": `"alpha"`, "visible": `"one"`})

			var model AppInstallationModel
			require.Empty(t, plan.Get(t.Context(), &model))

			schema := resourceIdentitySchema(appInstallationIdentityAttributeNames())
			identity := tfsdk.ResourceIdentity{Schema: schema, Raw: tftypes.NewValue(schema.Type().TerraformType(t.Context()), nil)}
			state := tfsdk.State{Schema: plan.Schema}
			endpoint := cm.PutAppInstallationParams{SpaceID: "space", EnvironmentID: "master", AppDefinitionID: "app"}
			response := cm.AppInstallation{Sys: cm.NewAppInstallationSys("space", "master", "app"), Parameters: []byte(testJSON(test.returned))}
			diags := setAppInstallationMutationState(t.Context(), &identity, &state, endpoint, model, response)
			require.Len(t, diags.Errors(), 1)
			assert.Equal(t, "Contentful returned different App Installation parameters", diags.Errors()[0].Summary())
		})
	}
}
