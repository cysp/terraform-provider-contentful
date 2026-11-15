package provider_test

import (
	"fmt"
	"net/http"
	"regexp"
	"sync"
	"sync/atomic"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccTeamResourceIdentityContradictionPreservesRecoveryState(t *testing.T) {
	t.Parallel()

	for _, operation := range []string{"create", "update"} {
		t.Run(operation, func(t *testing.T) {
			t.Parallel()

			testTeamIdentityRecovery(t, operation)
		})
	}
}

func testTeamIdentityRecovery(t *testing.T, operation string) {
	t.Helper()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)

	adapter := &mutationJSONResponseAdapter{delegate: server}

	var (
		requestsMu sync.Mutex
		requests   []string
	)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestsMu.Lock()

		requests = append(requests, r.Method+" "+r.URL.Path+" version="+r.Header.Get("X-Contentful-Version"))
		requestsMu.Unlock()
		adapter.ServeHTTP(w, r)
	})

	config := func(name string) string {
		return fmt.Sprintf(`resource "contentful_team" "test" {
 organization_id = "org"
 name = %q
 description = ""
}`, name)
	}

	var steps []resource.TestStep

	if operation == "update" {
		steps = append(steps, resource.TestStep{Config: config("Prior")})
	}

	method := http.MethodPost
	if operation == "update" {
		method = http.MethodPut
	}

	var createdID atomic.Value

	steps = append(steps, resource.TestStep{
		PreConfig: func() {
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
				server.SetTeam("other-organization", teamID, cm.TeamData{Name: "Untouched", Description: cm.NewNilString("")})

				sys["version"] = 73

				sys["organization"] = map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Organization", "id": "other-organization"}}
			})
		},
		Config:      config("Planned"),
		ExpectError: regexp.MustCompile(`Contentful returned a different team identity`),
	})

	checkID := func(prefix string) knownvalue.Check {
		return knownvalue.StringFunc(func(value string) error {
			teamID, ok := createdID.Load().(string)
			if !ok || teamID == "" || value != prefix+teamID {
				return fmt.Errorf("%w: recovery ID = %q, want %q followed by generated team ID %v", errUnexpectedTerraformPlan, value, prefix, createdID.Load())
			}

			return nil
		})
	}

	recoveryAction := plancheck.ResourceActionNoop
	if operation == "create" {
		recoveryAction = plancheck.ResourceActionDestroyBeforeCreate
	}

	steps = append(steps, resource.TestStep{
		Config: config("Planned"), PlanOnly: true, ExpectNonEmptyPlan: operation == "create",
		ConfigPlanChecks: resource.ConfigPlanChecks{PostApplyPreRefresh: []plancheck.PlanCheck{
			plancheck.ExpectResourceAction("contentful_team.test", recoveryAction),
			expectRecoveryTaint{address: "contentful_team.test", tainted: operation == "create"},
			testAccPriorState{check: statecheck.ExpectKnownValue("contentful_team.test", tfjsonpath.New("id"), checkID("org/"))},
			testAccPriorState{check: statecheck.ExpectKnownValue("contentful_team.test", tfjsonpath.New("organization_id"), knownvalue.StringExact("org"))},
			testAccPriorState{check: statecheck.ExpectKnownValue("contentful_team.test", tfjsonpath.New("team_id"), checkID(""))},
			testAccPriorState{check: statecheck.ExpectKnownValue("contentful_team.test", tfjsonpath.New("name"), knownvalue.StringExact("Planned"))},
			testAccPriorState{check: statecheck.ExpectIdentity("contentful_team.test", map[string]knownvalue.Check{
				"organization_id": knownvalue.StringExact("org"), "team_id": checkID(""),
			})},
		}},
	})

	var beforeRetry int

	if operation == "update" {
		// Protocol tests inspect private-version bytes directly. Here Core
		// must persist their revocation and block a retry without refresh.
		steps = append(steps, resource.TestStep{
			PreConfig: func() {
				requestsMu.Lock()
				beforeRetry = len(requests)
				requestsMu.Unlock()
			},
			Config: config("Next"), ExpectError: regexp.MustCompile(`Private version is unavailable`),
		})
	}

	cliOptions := &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}}

	steps = append(steps, resource.TestStep{
		PreConfig: func() {
			if operation == "update" {
				requestsMu.Lock()
				assert.Len(t, requests, beforeRetry, "missing version must stop before a network mutation")
				requestsMu.Unlock()
			}

			cliOptions.Plan.NoRefresh = false
		},
		Config: config("Planned"), PlanOnly: true, ExpectNonEmptyPlan: operation == "create",
		ConfigPlanChecks: resource.ConfigPlanChecks{PostApplyPostRefresh: []plancheck.PlanCheck{
			plancheck.ExpectResourceAction("contentful_team.test", recoveryAction),
		}},
	})
	testAccMockedResource(t, handler, resource.TestCase{AdditionalCLIOptions: cliOptions, Steps: steps})

	teamID, ok := createdID.Load().(string)
	require.True(t, ok, "mutation response adapter did not run")
	require.NotEmpty(t, teamID)
	remote, err := server.Handler().GetTeam(t.Context(), cm.GetTeamParams{OrganizationID: "org", TeamID: teamID})
	require.NoError(t, err)

	missing, ok := remote.(*cm.ErrorStatusCode)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, missing.StatusCode)
	unrelated, err := server.Handler().GetTeam(t.Context(), cm.GetTeamParams{OrganizationID: "other-organization", TeamID: teamID})
	require.NoError(t, err)

	other, ok := unrelated.(*cm.Team)
	require.True(t, ok)
	assert.Equal(t, "Untouched", other.Name)
	requestsMu.Lock()
	assert.Contains(t, requests, "DELETE /organizations/org/teams/"+teamID+" version=")

	for _, request := range requests {
		assert.NotContains(t, request, "/organizations/other-organization/")
	}
	requestsMu.Unlock()
}
