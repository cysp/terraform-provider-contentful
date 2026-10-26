package provider_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccTeamResourceReadIdentityPreservesPriorState(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name   string
		legacy bool
		args   []string
	}{
		{name: "established refresh", args: []string{"apply", "-refresh-only", "-auto-approve", "-input=false", "-no-color"}},
		{name: "legacy refresh", legacy: true, args: []string{"apply", "-refresh-only", "-auto-approve", "-input=false", "-no-color"}},
		{name: "legacy plan", legacy: true, args: []string{"plan", "-input=false", "-no-color"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			server, adapter, runtime := teamReadIdentityRuntime(t)
			server.SetTeam("org", "team", cm.TeamData{Name: "Initial", Description: cm.NewNilString("")})

			output, err := runtime.run(t.Context(), "import", "-input=false", "-no-color", "contentful_team.test", "org/team")
			require.NoError(t, err, output)

			if test.legacy {
				removeTeamTestResourceIdentity(t, runtime)
			}

			before, err := runtime.run(t.Context(), "state", "pull")
			require.NoError(t, err, before)
			adapter.mutateNext(http.MethodGet, func(body map[string]any) { corruptTeamReadIdentity(t, body) })

			output, err = runtime.run(t.Context(), test.args...)
			require.Error(t, err, output)
			assert.Contains(t, output, "Unexpected Contentful read identity")
			after, err := runtime.run(t.Context(), "state", "pull")
			require.NoError(t, err, after)
			assert.JSONEq(t, before, after)
			output, err = runtime.run(t.Context(), "plan", "-detailed-exitcode", "-input=false", "-no-color")
			require.NoError(t, err, output)
			assert.Contains(t, output, "No changes.")
			output, err = runtime.run(t.Context(), "destroy", "-auto-approve", "-input=false", "-no-color")
			require.NoError(t, err, output)
			remote, err := server.Handler().GetTeam(t.Context(), cm.GetTeamParams{OrganizationID: "org", TeamID: "team"})
			require.NoError(t, err)

			missing, ok := remote.(*cm.ErrorStatusCode)
			require.True(t, ok)
			assert.Equal(t, http.StatusNotFound, missing.StatusCode)
		})
	}
}

func TestAccTeamResourceImportRejectsContradictoryReadIdentity(t *testing.T) {
	t.Parallel()
	server, adapter, runtime := teamReadIdentityRuntime(t)
	server.SetTeam("org", "team", cm.TeamData{Name: "Initial", Description: cm.NewNilString("")})
	adapter.mutateNext(http.MethodGet, func(body map[string]any) { corruptTeamReadIdentity(t, body) })

	output, err := runtime.run(t.Context(), "import", "-input=false", "-no-color", "contentful_team.test", "org/team")
	require.Error(t, err, output)
	assert.Contains(t, output, "Unexpected Contentful read identity")

	_, stateErr := os.Stat(filepath.Join(runtime.workingDirectory, "terraform.tfstate"))
	if !os.IsNotExist(stateErr) {
		require.NoError(t, stateErr)
		output, err = runtime.run(t.Context(), "state", "list")
		require.NoError(t, err, output)
		assert.NotContains(t, output, "contentful_team.test")
	}

	output, err = runtime.run(t.Context(), "import", "-input=false", "-no-color", "contentful_team.test", "org/team")
	require.NoError(t, err, output)
	output, err = runtime.run(t.Context(), "plan", "-detailed-exitcode", "-input=false", "-no-color")
	require.NoError(t, err, output)
	assert.Contains(t, output, "No changes.")
}

func teamReadIdentityRuntime(t *testing.T) (*cmt.Server, *mutationJSONResponseAdapter, *terraformTestRuntime) {
	t.Helper()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)

	adapter := &mutationJSONResponseAdapter{delegate: server}
	httpServer := httptest.NewServer(adapter)
	t.Cleanup(httpServer.Close)
	runtime := newTerraformTestRuntime(t)
	runtime.providerURL = httpServer.URL
	runtime.writeConfig(t, `resource "contentful_team" "test" {
 organization_id = "org"
 name = "Initial"
 description = ""
}`)

	return server, adapter, runtime
}

func corruptTeamReadIdentity(t *testing.T, body map[string]any) {
	t.Helper()

	sys, ok := body["sys"].(map[string]any)
	if !assert.True(t, ok) {
		return
	}

	sys["organization"] = map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Organization", "id": "wrong-org"}}
	body["name"] = "Unrelated"
}

func removeTeamTestResourceIdentity(t *testing.T, runtime *terraformTestRuntime) {
	t.Helper()
	output, err := runtime.run(t.Context(), "state", "pull")
	require.NoError(t, err, output)

	var state map[string]json.RawMessage
	require.NoError(t, json.Unmarshal([]byte(output), &state))

	var resources []map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(state["resources"], &resources))
	require.Len(t, resources, 1)

	var instances []map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(resources[0]["instances"], &instances))
	require.Len(t, instances, 1)
	delete(instances[0], "identity")
	delete(instances[0], "identity_schema_version")
	resources[0]["instances"], err = json.Marshal(instances)
	require.NoError(t, err)
	state["resources"], err = json.Marshal(resources)
	require.NoError(t, err)
	encoded, err := json.Marshal(state)
	require.NoError(t, err)

	filename := filepath.Join(runtime.workingDirectory, "legacy-state.json")
	require.NoError(t, os.WriteFile(filename, encoded, 0o600))
	output, err = runtime.run(t.Context(), "state", "push", "-force", filename)
	require.NoError(t, err, output)
}
