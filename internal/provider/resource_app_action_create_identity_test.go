package provider_test

import (
	"bytes"
	"io"
	"net/http"
	"regexp"
	"sync"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccAppActionSpecifiedIDContradictionRetainsDeleteTarget(t *testing.T) {
	t.Parallel()

	for _, test := range []struct{ name, id string }{
		{"another action", "unrelated"},
		{"unaddressable ID", "unusable/id"},
		{"empty ID", ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			testAppActionSpecifiedIDRecovery(t, test.id)
		})
	}
}

func testAppActionSpecifiedIDRecovery(t *testing.T, responseID string) {
	t.Helper()

	server := newTestCMAServer(t)
	server.SetAppDefinition("org", "app", cm.AppDefinitionData{Name: "App"})
	unrelated, err := server.Handler().CreateAppAction(t.Context(), &cm.AppActionCreateData{
		ID: cm.NewOptString("unrelated"), Name: "Unrelated", Category: "Custom", Type: "function-invocation",
		Function: cm.NewOptFunctionLink(cm.NewFunctionLink("other-function")), Parameters: []byte(`[]`),
	}, cm.CreateAppActionParams{OrganizationID: "org", AppDefinitionID: "app"})
	require.NoError(t, err)
	require.IsType(t, &cm.AppAction{}, unrelated)

	adapter := &mutationJSONResponseAdapter{delegate: server}
	adapter.mutateNext(http.MethodPost, func(response map[string]any) {
		sys, ok := response["sys"].(map[string]any)
		if assert.True(t, ok) {
			sys["id"] = responseID
			response["name"] = "Returned"
		}
	})

	var (
		requestMutex sync.Mutex
		writes       []string
		postBody     string
	)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			body, readErr := io.ReadAll(r.Body)
			assert.NoError(t, readErr)

			r.Body = io.NopCloser(bytes.NewReader(body))

			requestMutex.Lock()

			writes = append(writes, r.Method+" "+r.URL.Path)
			if r.Method == http.MethodPost {
				postBody = string(body)
			}
			requestMutex.Unlock()
		}

		adapter.ServeHTTP(w, r)
	})
	config := appActionFunctionWithID("chosenAction")
	testAccMockedResource(t, handler, resource.TestCase{
		AdditionalCLIOptions: &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
		Steps: []resource.TestStep{
			{Config: config, ExpectError: regexp.MustCompile(`Contentful returned a different`)},
			{
				Config: config, PlanOnly: true, ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{PostApplyPreRefresh: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(actionAddress, plancheck.ResourceActionDestroyBeforeCreate),
					expectRecoveryTaint{address: actionAddress, tainted: true},
					testAccPriorState{check: statecheck.ExpectKnownValue(actionAddress, tfjsonpath.New("id"), knownvalue.StringExact("org/app/chosenAction"))},
					testAccPriorState{check: statecheck.ExpectKnownValue(actionAddress, tfjsonpath.New("app_action_id"), knownvalue.StringExact("chosenAction"))},
					testAccPriorState{check: statecheck.ExpectKnownValue(actionAddress, tfjsonpath.New("name"), knownvalue.StringExact("Returned"))},
					testAccPriorState{check: statecheck.ExpectIdentity(actionAddress, map[string]knownvalue.Check{
						"organization_id": knownvalue.StringExact("org"), "app_definition_id": knownvalue.StringExact("app"), "app_action_id": knownvalue.StringExact("chosenAction"),
					})},
				}},
			},
			{Config: config, Destroy: true},
		},
	})
	requestMutex.Lock()
	assert.Equal(t, []string{"POST /organizations/org/app_definitions/app/actions", "DELETE /organizations/org/app_definitions/app/actions/chosenAction"}, writes)
	assert.JSONEq(t, testJSON(map[string]any{
		"id": "chosenAction", "name": "Action", "category": "Custom", "type": "function-invocation", "parameters": []any{},
		"function": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Function", "id": "function"}},
	}), postBody)
	requestMutex.Unlock()

	remaining, err := server.Handler().GetAppAction(t.Context(), cm.GetAppActionParams{OrganizationID: "org", AppDefinitionID: "app", AppActionID: "unrelated"})
	require.NoError(t, err)

	action, ok := remaining.(*cm.AppAction)
	require.True(t, ok, "the unrelated action must remain untouched")
	assert.Equal(t, "Unrelated", action.Name)
	assert.Equal(t, "other-function", action.Function.Value.Sys.ID)
	removed, err := server.Handler().GetAppAction(t.Context(), cm.GetAppActionParams{OrganizationID: "org", AppDefinitionID: "app", AppActionID: "chosenAction"})
	require.NoError(t, err)

	missing, ok := removed.(*cm.ErrorStatusCode)
	require.True(t, ok)
	assert.Equal(t, http.StatusNotFound, missing.StatusCode)
}
