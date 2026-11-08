package provider_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	cmt "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go/testing"
	"github.com/stretchr/testify/require"
)

func TestAccContentTypeResourceValidationRecovery(t *testing.T) {
	t.Parallel()

	server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
	require.NoError(t, err)
	server.SetContentType("space", "master", "type", cm.ContentTypeRequestData{
		Name: "Type", Description: cm.NewOptNilString(""), DisplayField: "field",
		Fields: []cm.ContentTypeRequestDataFieldsItem{{ID: "field", Name: "Field", Type: "Symbol", Localized: cm.NewOptBool(false), Required: cm.NewOptBool(false), Disabled: cm.NewOptBool(false), Omitted: cm.NewOptBool(false), DefaultValue: []byte("[]")}},
	})

	var writes, deletes atomic.Int64

	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut, http.MethodPost:
			writes.Add(1)
		case http.MethodDelete:
			deletes.Add(1)
		}

		server.ServeHTTP(w, r)
	}))
	t.Cleanup(httpServer.Close)
	runtime := newTerraformTestRuntime(t)
	runtime.providerURL = httpServer.URL
	config := func(description, value string) string {
		return fmt.Sprintf(`
 resource "contentful_content_type" "test" {
 space_id = "space"
 environment_id = "master"
 content_type_id = "type"
 name = "Type"
 description = %q
 display_field = "field"
 fields = [{ id = "field", name = "Field", type = "Symbol", localized = false, required = false, default_value = %q }]
 lifecycle { ignore_changes = [fields] }
 }
 `, description, value)
	}
	runtime.writeConfig(t, config("", "[]"))
	output, err := runtime.run(t.Context(), "apply", "-auto-approve", "-input=false", "-no-color")
	require.Error(t, err, output)
	require.Contains(t, output, "Invalid Content Type JSON object")
	require.Zero(t, writes.Load())
	runtime.writeConfig(t, config("", "{}"))
	output, err = runtime.run(t.Context(), "import", "-input=false", "-no-color", "contentful_content_type.test", "space/master/type")
	require.NoError(t, err, output)
	output, err = runtime.run(t.Context(), "plan", "-input=false", "-no-color", "-detailed-exitcode")
	require.NoError(t, err, output)
	runtime.writeConfig(t, config("Changed", "{}"))
	output, err = runtime.run(t.Context(), "apply", "-auto-approve", "-input=false", "-no-color")
	require.Error(t, err, output)
	require.Contains(t, output, "Invalid Content Type JSON object")
	require.Zero(t, writes.Load())
	runtime.writeConfig(t, config("", "[]"))
	output, err = runtime.run(t.Context(), "destroy", "-auto-approve", "-input=false", "-no-color")
	require.Error(t, err, output)
	require.Contains(t, output, "Invalid Content Type JSON object")
	require.Zero(t, deletes.Load())
	runtime.writeConfig(t, config("", "{}"))
	output, err = runtime.run(t.Context(), "destroy", "-auto-approve", "-input=false", "-no-color")
	require.NoError(t, err, output)
	require.Zero(t, writes.Load())
	require.Positive(t, deletes.Load())
}
