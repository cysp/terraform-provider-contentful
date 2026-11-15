package provider_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
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

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
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

	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	}))
	t.Cleanup(httpServer.Close)
	runtime := newTerraformTestRuntime(t)
	runtime.providerURL = httpServer.URL
	runtime.writeConfig(t, appActionFunctionWithID("chosenAction"))

	output, err := runtime.run(t.Context(), "apply", "-auto-approve", "-input=false", "-no-color")
	require.Error(t, err, output)
	assert.Contains(t, output, "Contentful returned a different")

	output, err = runtime.run(t.Context(), "state", "pull")
	require.NoError(t, err, output)

	var state struct {
		Resources []struct {
			Instances []struct {
				Status     string                     `json:"status"`
				Attributes map[string]json.RawMessage `json:"attributes"`
				Identity   map[string]string          `json:"identity"`
			} `json:"instances"`
		} `json:"resources"`
	}
	require.NoError(t, json.Unmarshal([]byte(output), &state))
	require.Len(t, state.Resources, 1)
	require.Len(t, state.Resources[0].Instances, 1)
	instance := state.Resources[0].Instances[0]
	assert.Equal(t, "tainted", instance.Status)
	assert.JSONEq(t, `"org/app/chosenAction"`, string(instance.Attributes["id"]))
	assert.JSONEq(t, `"chosenAction"`, string(instance.Attributes["app_action_id"]))
	assert.JSONEq(t, `"Returned"`, string(instance.Attributes["name"]))
	assert.Equal(t, map[string]string{"organization_id": "org", "app_definition_id": "app", "app_action_id": "chosenAction"}, instance.Identity)

	output, err = runtime.run(t.Context(), "destroy", "-refresh=false", "-auto-approve", "-input=false", "-no-color")
	require.NoError(t, err, output)
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
