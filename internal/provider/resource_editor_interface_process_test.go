package provider_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/require"
)

// The reattached in-process acceptance provider does not isolate configurations
// like separate Terraform plugin processes. Exercise that distinction through CLI.
func TestAccEditorInterfaceResourceProviderProcesses(t *testing.T) {
	t.Parallel()

	for _, separateProvider := range []bool{false, true} {
		t.Run(fmt.Sprintf("separate_provider=%t", separateProvider), func(t *testing.T) {
			t.Parallel()

			runtime := newTerraformTestRuntime(t)
			server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
			require.NoError(t, err)
			server.RegisterSpaceEnvironment("space", "environment")

			var (
				requestsMu sync.Mutex
				requests   []string
			)

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodPut && (strings.HasSuffix(r.URL.Path, "/editor_interface") || strings.HasSuffix(r.URL.Path, "/published")) {
					requestsMu.Lock()

					requests = append(requests, r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:]+":"+r.Header.Get("X-Contentful-Version"))
					requestsMu.Unlock()
				}

				server.ServeHTTP(w, r)
			})
			httpServer := httptest.NewServer(handler)
			t.Cleanup(httpServer.Close)
			runtime.providerURL = httpServer.URL

			configuration := func(description, helpText string) string {
				editorInterface := editorInterfaceActivationConfig(helpText)
				if separateProvider {
					editorInterface = strings.Replace(editorInterface, `resource "contentful_editor_interface" "test" {`, `resource "contentful_editor_interface" "test" {
 provider = contentful.editor`, 1)
				}

				return fmt.Sprintf(`
provider "contentful" {
 alias = "editor"
 access_token = %[1]q
 url = %[2]q
}
`, cmt.ValidAccessToken, httpServer.URL) + editorInterfaceActivationContentTypeConfig(description) + editorInterface
			}
			snapshot := func() []string {
				requestsMu.Lock()
				defer requestsMu.Unlock()

				return append([]string(nil), requests...)
			}

			runtime.writeConfig(t, configuration("initial", "initial"))
			output, err := runtime.run(t.Context(), "apply", "-auto-approve", "-input=false", "-no-color")
			require.NoError(t, err, output)
			require.Equal(t, []string{"published:1", "editor_interface:1"}, snapshot())

			runtime.writeConfig(t, configuration("updated", "updated"))

			output, err = runtime.run(t.Context(), "apply", "-auto-approve", "-input=false", "-no-color")
			if separateProvider {
				require.Error(t, err, output)
				require.Contains(t, output, "Editor Interface version mismatch")
				require.Contains(t, output, "another provider configuration")
				require.Equal(t, []string{"published:1", "editor_interface:1", "published:3", "editor_interface:2"}, snapshot())

				// A new refresh repairs the baseline; the provider did not silently retry.
				output, err = runtime.run(t.Context(), "apply", "-auto-approve", "-input=false", "-no-color")
				require.NoError(t, err, output)
				require.Equal(t, []string{"published:1", "editor_interface:1", "published:3", "editor_interface:2", "editor_interface:3"}, snapshot())
			} else {
				require.NoError(t, err, output)
				require.Equal(t, []string{"published:1", "editor_interface:1", "published:3", "editor_interface:3"}, snapshot())
			}

			// Preserve optimistic concurrency between planning and applying a saved plan.
			// This mutation uses the mock's public CMA handler, bypassing only the recorder.
			runtime.writeConfig(t, configuration("updated", "third"))
			output, err = runtime.run(t.Context(), "plan", "-out=update.tfplan", "-input=false", "-no-color")
			require.NoError(t, err, output)
			response, err := server.Handler().PutEditorInterface(t.Context(), &cm.EditorInterfaceData{
				Controls: cm.NewOptNilEditorInterfaceDataControlsItemArray([]cm.EditorInterfaceDataControlsItem{{
					FieldId: "name", WidgetNamespace: cm.NewOptString("builtin"), WidgetId: cm.NewOptString("singleLine"),
					Settings: []byte(`{"helpText":"external"}`),
				}}),
			}, cm.PutEditorInterfaceParams{
				SpaceID: "space", EnvironmentID: "environment", ContentTypeID: "editor-offset", XContentfulVersion: 4,
			})
			require.NoError(t, err)
			require.IsType(t, &cm.EditorInterfaceStatusCode{}, response)

			before := snapshot()
			output, err = runtime.run(t.Context(), "apply", "-input=false", "-no-color", "update.tfplan")
			require.Error(t, err, output)
			require.Contains(t, output, "Editor Interface version mismatch")
			require.Equal(t, append(before, "editor_interface:4"), snapshot())

			observed, err := server.Handler().GetEditorInterface(t.Context(), cm.GetEditorInterfaceParams{
				SpaceID: "space", EnvironmentID: "environment", ContentTypeID: "editor-offset",
			})
			require.NoError(t, err)
			require.IsType(t, &cm.EditorInterface{}, observed)
			editorInterface, ok := observed.(*cm.EditorInterface)
			require.True(t, ok)
			require.Equal(t, 5, editorInterface.Sys.Version)
			controls, ok := editorInterface.Controls.Get()
			require.True(t, ok)
			require.Len(t, controls, 1)
			require.JSONEq(t, `{"helpText":"external"}`, string(controls[0].Settings))

			output, err = runtime.run(t.Context(), "destroy", "-auto-approve", "-input=false", "-no-color")
			require.NoError(t, err, output)
		})
	}
}
