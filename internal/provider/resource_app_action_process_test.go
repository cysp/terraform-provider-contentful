package provider_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccAppActionResourceForcedReplacement(t *testing.T) {
	t.Parallel()
	runtime := newTerraformTestRuntime(t)
	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.SetAppDefinition("org", "app", cm.AppDefinitionData{Name: "App"})

	var (
		requestMutex    sync.Mutex
		methods, bodies []string
		deleted         []string
	)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			requestMutex.Lock()
			defer requestMutex.Unlock()

			methods = append(methods, r.Method)
			if r.Method == http.MethodDelete {
				deleted = append(deleted, strings.TrimPrefix(r.URL.Path, actionCollectionPath+"/"))
			}

			if r.Method == http.MethodPost || r.Method == http.MethodPut {
				body, err := io.ReadAll(r.Body)
				assert.NoError(t, err)

				bodies = append(bodies, string(body))
				r.Body = io.NopCloser(bytes.NewReader(body))
			}
		}

		server.ServeHTTP(w, r)
	})
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)
	runtime.providerURL = ts.URL

	const lifecycle = "\n lifecycle {\n ignore_changes = [parameters]\n create_before_destroy = true\n }\n}"
	runtime.writeConfig(t, strings.Replace(actionLegacyConfig, "\n}", lifecycle, 1))
	out, err := runtime.run(t.Context(), "apply", "-auto-approve", "-input=false", "-no-color")
	require.NoError(t, err, out)
	firstRes, err := server.Handler().GetAppActions(t.Context(), cm.GetAppActionsParams{OrganizationID: "org", AppDefinitionID: "app"})
	require.NoError(t, err)

	first, ok := firstRes.(*cm.AppActionCollection)
	require.True(t, ok)
	require.Len(t, first.Items, 1)
	oldID := first.Items[0].Sys.ID

	runtime.writeConfig(t, strings.Replace(actionBuiltinConfig, "\n}", lifecycle, 1))
	out, err = runtime.run(t.Context(), "plan", "-replace="+actionAddress, "-out=replace.plan", "-input=false", "-no-color")
	require.NoError(t, err, out)
	out, err = runtime.run(t.Context(), "show", "-json", "replace.plan")
	require.NoError(t, err, out)

	var plan tfjson.Plan
	require.NoError(t, json.Unmarshal([]byte(out), &plan))

	found := false

	for _, change := range plan.ResourceChanges {
		if change.Address == actionAddress {
			found = true

			assert.Equal(t, tfjson.Actions{tfjson.ActionCreate, tfjson.ActionDelete}, change.Change.Actions)
			unknown, ok := change.Change.AfterUnknown.(map[string]any)
			require.True(t, ok)
			assert.Equal(t, true, unknown["id"])
			assert.Equal(t, true, unknown["app_action_id"])
		}
	}

	require.True(t, found)
	out, err = runtime.run(t.Context(), "apply", "-input=false", "-no-color", "replace.plan")
	require.NoError(t, err, out)
	secondRes, err := server.Handler().GetAppActions(t.Context(), cm.GetAppActionsParams{OrganizationID: "org", AppDefinitionID: "app"})
	require.NoError(t, err)

	second, ok := secondRes.(*cm.AppActionCollection)
	require.True(t, ok)
	require.Len(t, second.Items, 1)
	newID := second.Items[0].Sys.ID
	assert.NotEqual(t, oldID, newID)
	out, err = runtime.run(t.Context(), "plan", "-detailed-exitcode", "-input=false", "-no-color")
	require.NoError(t, err, out)
	out, err = runtime.run(t.Context(), "destroy", "-auto-approve", "-input=false", "-no-color")
	require.NoError(t, err, out)
	requestMutex.Lock()
	defer requestMutex.Unlock()

	expected := []string{"POST", "POST", "DELETE", "DELETE"}
	assert.Equal(t, expected, methods)
	assert.Equal(t, []string{oldID, newID}, deleted)
	require.Len(t, bodies, 2)
	assert.JSONEq(t, actionLegacyBody, bodies[0])
	assert.JSONEq(t, actionBuiltinBody, bodies[1])
}
