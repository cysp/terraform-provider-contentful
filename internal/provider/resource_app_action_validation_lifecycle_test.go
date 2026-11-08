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
	"github.com/stretchr/testify/require"
)

func TestAccAppActionResourceParameterValidationRecovery(t *testing.T) {
	t.Parallel()
	runtime := newTerraformTestRuntime(t)
	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.SetAppDefinition("org", "app", cm.AppDefinitionData{Name: "App"})
	result, err := server.Handler().CreateAppAction(t.Context(), &cm.AppActionCreateData{Name: "Action", Category: "Custom", Type: "endpoint", URL: cm.NewOptString("https://example.invalid/action"), Parameters: []byte("[]")}, cm.CreateAppActionParams{OrganizationID: "org", AppDefinitionID: "app"})
	require.NoError(t, err)

	action, ok := result.(*cm.AppAction)
	require.True(t, ok)

	var writes, deletes atomic.Int64

	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			deletes.Add(1)
		} else if r.Method != http.MethodGet {
			writes.Add(1)
		}

		if r.Method == http.MethodGet && r.URL.Path == actionCollectionPath+"/"+action.Sys.ID {
			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, r)

			var document map[string]any

			err := json.Unmarshal(recorder.Body.Bytes(), &document)
			if err != nil {
				t.Error(err)
				w.WriteHeader(http.StatusInternalServerError)

				return
			}

			document["parameters"] = []any{nil}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(recorder.Code)

			err = json.NewEncoder(w).Encode(document)
			if err != nil {
				t.Error(err)
			}

			return
		}

		server.ServeHTTP(w, r)
	}))
	t.Cleanup(httpServer.Close)
	runtime.providerURL = httpServer.URL
	configuration := func(name, parameters string) string {
		return fmt.Sprintf(`
 resource "contentful_app_action" "test" {
 organization_id="org"
 app_definition_id="app"
 name=%q
 category="Custom"
 type="endpoint"
 url="https://example.invalid/action"
 parameters=%q
 lifecycle { ignore_changes=[parameters] }
 }
 `, name, parameters)
	}
	runtime.writeConfig(t, configuration("Action", "[]"))
	output, err := runtime.run(t.Context(), "import", "-input=false", "-no-color", actionAddress, "org/app/"+action.Sys.ID)
	require.NoError(t, err, output)
	output, err = runtime.run(t.Context(), "plan", "-input=false", "-no-color", "-detailed-exitcode")
	require.NoError(t, err, output)
	runtime.writeConfig(t, configuration("Changed", "[]"))
	output, err = runtime.run(t.Context(), "apply", "-auto-approve", "-input=false", "-no-color")
	require.Error(t, err, output)
	require.Contains(t, output, "Invalid App Action parameters")
	require.Zero(t, writes.Load())
	runtime.writeConfig(t, configuration("Action", "[null]"))
	output, err = runtime.run(t.Context(), "destroy", "-auto-approve", "-input=false", "-no-color")
	require.Error(t, err, output)
	require.Contains(t, output, "Invalid App Action parameters")
	require.Zero(t, deletes.Load())
	runtime.writeConfig(t, configuration("Action", "[]"))
	output, err = runtime.run(t.Context(), "destroy", "-auto-approve", "-input=false", "-no-color")
	require.NoError(t, err, output)
	require.Zero(t, writes.Load())
	require.EqualValues(t, 1, deletes.Load())
}
