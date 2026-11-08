package provider_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/require"
)

func TestAccEditorInterfaceResourceValidationRecovery(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.SetEditorInterface("space", "master", "type", cm.EditorInterfaceData{
		Controls: cm.NewOptNilEditorInterfaceControlsItemArray([]cm.EditorInterfaceControlsItem{{FieldId: "field", Settings: []byte("[]")}}),
	})
	recorder := &editorInterfaceRequestRecorder{next: server}
	httpServer := httptest.NewServer(recorder)
	t.Cleanup(httpServer.Close)
	runtime := newTerraformTestRuntime(t)
	runtime.providerURL = httpServer.URL
	config := func(settings, other string) string {
		return fmt.Sprintf(`
 resource "contentful_editor_interface" "test" {
 space_id = "space"
 environment_id = "master"
 content_type_id = "type"
 controls = [{ field_id = "field", settings = %q }]
 %s
 lifecycle { ignore_changes = [controls] }
 }
 `, settings, other)
	}
	runtime.writeConfig(t, config("[]", ""))
	output, err := runtime.run(t.Context(), "apply", "-auto-approve", "-input=false", "-no-color")
	require.Error(t, err, output)
	require.Contains(t, output, "Invalid Editor Interface settings")
	require.Empty(t, recorder.Requests())
	runtime.writeConfig(t, config("{}", ""))
	output, err = runtime.run(t.Context(), "import", "-input=false", "-no-color", "contentful_editor_interface.test", "space/master/type")
	require.NoError(t, err, output)
	output, err = runtime.run(t.Context(), "plan", "-input=false", "-no-color", "-detailed-exitcode")
	require.NoError(t, err, output)
	runtime.writeConfig(t, config("{}", "sidebar = []"))
	output, err = runtime.run(t.Context(), "apply", "-auto-approve", "-input=false", "-no-color")
	require.Error(t, err, output)
	require.Contains(t, output, "Invalid Editor Interface settings")

	for _, request := range recorder.Requests() {
		require.False(t, strings.HasPrefix(request, http.MethodPut))
	}

	runtime.writeConfig(t, config("[]", ""))
	output, err = runtime.run(t.Context(), "destroy", "-auto-approve", "-input=false", "-no-color")
	require.Error(t, err, output)
	require.Contains(t, output, "Invalid Editor Interface settings")
	runtime.writeConfig(t, config("{}", ""))
	output, err = runtime.run(t.Context(), "destroy", "-auto-approve", "-input=false", "-no-color")
	require.NoError(t, err, output)

	for _, request := range recorder.Requests() {
		require.Equal(t, http.MethodGet, request)
	}
}
