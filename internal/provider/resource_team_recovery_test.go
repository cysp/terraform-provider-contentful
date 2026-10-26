package provider_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccTeamResourceIdentityContradictionPreservesRecoveryState(t *testing.T) {
	t.Parallel()

	for _, operation := range []string{"create", "update"} {
		t.Run(operation, func(t *testing.T) {
			t.Parallel()

			server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
			require.NoError(t, err)

			adapter := &mutationJSONResponseAdapter{delegate: server}
			httpServer := httptest.NewServer(adapter)
			t.Cleanup(httpServer.Close)
			runtime := newTerraformTestRuntime(t)
			runtime.providerURL = httpServer.URL

			config := func(name string) string {
				return fmt.Sprintf(`resource "contentful_team" "test" {
 organization_id = "org"
 name = %q
 description = ""
}`, name)
			}
			if operation == "update" {
				runtime.writeConfig(t, config("Prior"))
				output, err := runtime.run(t.Context(), "apply", "-auto-approve", "-input=false", "-no-color")
				require.NoError(t, err, output)
			}

			method := http.MethodPost
			if operation == "update" {
				method = http.MethodPut
			}

			var createdID atomic.Value

			adapter.mutateNext(method, func(response map[string]any) {
				sys, ok := response["sys"].(map[string]any)
				if !assert.True(t, ok) {
					return
				}

				teamID, ok := sys["id"].(string)
				if !assert.True(t, ok) {
					return
				}

				createdID.Store(teamID)

				sys["organization"] = map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Organization", "id": "other-organization"}}
			})
			runtime.writeConfig(t, config("Planned"))
			output, err := runtime.run(t.Context(), "apply", "-auto-approve", "-input=false", "-no-color")
			require.Error(t, err, output)
			assert.Contains(t, output, "Contentful returned a different team identity")

			teamID, ok := createdID.Load().(string)
			require.True(t, ok, "mutation response adapter did not run")
			require.NotEmpty(t, teamID)

			output, err = runtime.run(t.Context(), "state", "pull")
			require.NoError(t, err, output)

			var state struct {
				Resources []struct {
					Instances []struct {
						Attributes map[string]string `json:"attributes"`
						Identity   map[string]string `json:"identity"`
					} `json:"instances"`
				} `json:"resources"`
			}
			require.NoError(t, json.Unmarshal([]byte(output), &state))
			require.Len(t, state.Resources, 1)
			require.Len(t, state.Resources[0].Instances, 1)
			instance := state.Resources[0].Instances[0]
			assert.Equal(t, "org/"+teamID, instance.Attributes["id"])
			assert.Equal(t, "org", instance.Attributes["organization_id"])
			assert.Equal(t, teamID, instance.Attributes["team_id"])
			assert.Equal(t, "Planned", instance.Attributes["name"])
			assert.Equal(t, map[string]string{"organization_id": "org", "team_id": teamID}, instance.Identity)

			output, err = runtime.run(t.Context(), "plan", "-input=false", "-no-color")
			require.NoError(t, err, output)

			if operation == "create" {
				assert.Contains(t, output, "1 to add, 0 to change, 1 to destroy")
			} else {
				assert.Contains(t, output, "No changes.")
			}

			output, err = runtime.run(t.Context(), "destroy", "-auto-approve", "-input=false", "-no-color")
			require.NoError(t, err, output)
			remote, err := server.Handler().GetTeam(t.Context(), cm.GetTeamParams{OrganizationID: "org", TeamID: teamID})
			require.NoError(t, err)

			missing, ok := remote.(*cm.ErrorStatusCode)
			require.True(t, ok)
			assert.Equal(t, http.StatusNotFound, missing.StatusCode)
		})
	}
}
