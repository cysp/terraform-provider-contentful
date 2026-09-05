package provider_test

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/require"
)

// Real plugin processes are necessary: the reattached acceptance-test provider
// does not reproduce the isolation between provider configurations.
func TestAccEditorInterfaceResourceProviderProcesses(t *testing.T) {
	t.Parallel()

	runtime := newTerraformTestRuntime(t)
	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.RegisterSpaceEnvironment("space", "environment")
	handler := &editorInterfaceRequestRecorder{next: server}
	httpServer := httptest.NewServer(handler)
	t.Cleanup(httpServer.Close)
	runtime.providerURL = httpServer.URL

	configuration := func(provider, description, helpText string) string {
		editorInterface := strings.Replace(editorInterfaceActivationConfig(helpText),
			`resource "contentful_editor_interface" "test" {`,
			"resource \"contentful_editor_interface\" \"test\" {\n provider = "+provider, 1)

		return fmt.Sprintf(`
provider "contentful" {
 alias = "editor"
 access_token = %[1]q
 url = %[2]q
}
`, cmt.ValidAccessToken, httpServer.URL) + editorInterfaceActivationContentTypeConfig(description) + editorInterface
	}

	runtime.writeConfig(t, configuration("contentful", "initial", "initial"))
	output, err := runtime.run(t.Context(), "apply", "-auto-approve", "-input=false", "-no-color")
	require.NoError(t, err, output)

	// The same provider accounts for activation: interface version 2 advances to 3.
	runtime.writeConfig(t, configuration("contentful", "updated", "updated"))
	output, err = runtime.run(t.Context(), "apply", "-auto-approve", "-input=false", "-no-color")
	require.NoError(t, err, output)

	// A different provider cannot account for activation from version 4 to 5.
	runtime.writeConfig(t, configuration("contentful.editor", "aliased", "aliased"))
	output, err = runtime.run(t.Context(), "apply", "-auto-approve", "-input=false", "-no-color")
	require.Error(t, err, output)
	require.Contains(t, output, "Editor Interface version mismatch")
	require.Contains(t, output, "another provider configuration")

	// An explicit new apply refreshes the baseline; the failed write was not retried.
	output, err = runtime.run(t.Context(), "apply", "-auto-approve", "-input=false", "-no-color")
	require.NoError(t, err, output)

	// A saved plan must still conflict with a later edit, preserving its payload.
	runtime.writeConfig(t, configuration("contentful.editor", "aliased", "third"))
	output, err = runtime.run(t.Context(), "plan", "-out=update.tfplan", "-input=false", "-no-color")
	require.NoError(t, err, output)
	response, err := server.Handler().PutEditorInterface(t.Context(), &cm.EditorInterfaceData{
		Controls: cm.NewOptNilEditorInterfaceDataControlsItemArray([]cm.EditorInterfaceDataControlsItem{{
			FieldId: "name", WidgetNamespace: cm.NewOptString("builtin"), WidgetId: cm.NewOptString("singleLine"),
			Settings: []byte(`{"helpText":"external"}`),
		}}),
	}, cm.PutEditorInterfaceParams{
		SpaceID: "space", EnvironmentID: "environment", ContentTypeID: "editor-offset", XContentfulVersion: 6,
	})
	require.NoError(t, err)
	require.IsType(t, &cm.EditorInterfaceStatusCode{}, response)

	output, err = runtime.run(t.Context(), "apply", "-input=false", "-no-color", "update.tfplan")
	require.Error(t, err, output)
	require.Contains(t, output, "Editor Interface version mismatch")
	require.Equal(t, []string{
		"PUT:1", "GET", "PUT:3", "GET", "PUT:4", "GET", "PUT:5", "GET", "PUT:6",
	}, handler.Requests())

	observed, err := server.Handler().GetEditorInterface(t.Context(), cm.GetEditorInterfaceParams{
		SpaceID: "space", EnvironmentID: "environment", ContentTypeID: "editor-offset",
	})
	require.NoError(t, err)

	editorInterface, ok := observed.(*cm.EditorInterface)
	require.True(t, ok)
	require.Equal(t, 7, editorInterface.Sys.Version)
	controls, ok := editorInterface.Controls.Get()
	require.True(t, ok)
	require.Len(t, controls, 1)
	require.JSONEq(t, `{"helpText":"external"}`, string(controls[0].Settings))

	output, err = runtime.run(t.Context(), "destroy", "-auto-approve", "-input=false", "-no-color")
	require.NoError(t, err, output)
}
