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

func TestAccExtensionResourceConfigurationValidationRecovery(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, fields, parameters, ignored, diagnostic string
		remote                                        cm.ExtensionExtension
	}{
		{name: "field types", fields: `[{type="Link",link_type="Entry"}]`, parameters: "null", ignored: "extension.field_types", diagnostic: "Invalid Extension field type configuration", remote: cm.ExtensionExtension{Name: "Extension", Srcdoc: cm.NewOptString(""), Sidebar: cm.NewOptBool(false), FieldTypes: []cm.ExtensionExtensionFieldTypesItem{{Type: "Link"}}}},
		{name: "parameter definitions", fields: "[]", parameters: `{installation=[{id="value",name="Value",type="Number",default=jsonencode(42)}]}`, ignored: "extension.parameters", diagnostic: "Invalid Extension parameter configuration", remote: cm.ExtensionExtension{Name: "Extension", Srcdoc: cm.NewOptString(""), Sidebar: cm.NewOptBool(false), FieldTypes: []cm.ExtensionExtensionFieldTypesItem{}, Parameters: cm.NewOptAppDefinitionParameters(cm.AppDefinitionParameters{Installation: []cm.AppDefinitionParameter{{ID: "value", Name: "Value", Type: "Number", Default: []byte(`"42"`)}}})}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			runtime := newTerraformTestRuntime(t)
			server, err := cmt.NewContentfulManagementServer(cmt.WithRateLimitPerSecond(1000))
			require.NoError(t, err)
			server.RegisterSpaceEnvironment("space", "master")
			result, err := server.Handler().PutExtension(t.Context(), &cm.ExtensionData{Extension: test.remote}, cm.PutExtensionParams{SpaceID: "space", EnvironmentID: "master", ExtensionID: "extension"})
			require.NoError(t, err)

			_, ok := result.(*cm.ExtensionStatusCode)
			require.True(t, ok)

			var writes, deletes atomic.Int64

			httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodDelete {
					deletes.Add(1)
				} else if r.Method != http.MethodGet {
					writes.Add(1)
				}

				server.ServeHTTP(w, r)
			}))
			t.Cleanup(httpServer.Close)
			runtime.providerURL = httpServer.URL
			configuration := func(name string) string {
				return fmt.Sprintf(`
 resource "contentful_extension" "test" {
 space_id="space"
 environment_id="master"
 extension_id="extension"
 extension={name=%q,srcdoc="",field_types=%s,parameters=%s}
 lifecycle {ignore_changes=[%s]}
 }
 `, name, test.fields, test.parameters, test.ignored)
			}
			runtime.writeConfig(t, configuration("Extension"))
			output, err := runtime.run(t.Context(), "import", "-input=false", "-no-color", "contentful_extension.test", "space/master/extension")
			require.NoError(t, err, output)
			output, err = runtime.run(t.Context(), "plan", "-input=false", "-no-color", "-detailed-exitcode")
			require.NoError(t, err, output)
			runtime.writeConfig(t, configuration("Changed"))
			output, err = runtime.run(t.Context(), "apply", "-auto-approve", "-input=false", "-no-color")
			require.Error(t, err, output)
			require.Contains(t, output, test.diagnostic)
			require.Zero(t, writes.Load())
			runtime.writeConfig(t, `resource "contentful_extension" "test" {
 space_id="space"
 environment_id="master"
 extension_id="extension"
 extension={name="Extension",srcdoc="",field_types=[{type="Link"}]}
 }`)
			output, err = runtime.run(t.Context(), "destroy", "-auto-approve", "-input=false", "-no-color")
			require.Error(t, err, output)
			require.Contains(t, output, "Invalid Extension field type configuration")
			require.Zero(t, deletes.Load())
			runtime.writeConfig(t, configuration("Extension"))
			output, err = runtime.run(t.Context(), "destroy", "-auto-approve", "-input=false", "-no-color")
			require.NoError(t, err, output)
			require.Zero(t, writes.Load())
			require.EqualValues(t, 1, deletes.Load())
		})
	}
}
