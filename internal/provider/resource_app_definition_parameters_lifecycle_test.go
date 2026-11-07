package provider_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func appDefinitionParameterDestroyConfig(invalid, ignored bool) string {
	options, lifecycle := "", ""
	if invalid {
		options = "options = []"
	}

	if ignored {
		lifecycle = "lifecycle { ignore_changes = [parameters] }"
	}

	return fmt.Sprintf(`resource "contentful_app_definition" "test" {
 organization_id = "org"
 name = "App"
 src = "https://example.invalid/app"
 locations = [{location = "app-config"}]
 parameters = { instance = [{id = "setting", name = "Setting", type = "Enum", %s}] }
 %s
}`, options, lifecycle)
}

func TestAccAppDefinitionResourceParameterDestroyRecovery(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name             string
		invalid, ignored bool
	}{
		{name: "valid configuration with irregular imported state"},
		{name: "invalid configuration blocks destroy then correction recovers", invalid: true},
		{name: "ignore_changes does not bypass invalid configuration", invalid: true, ignored: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var deletes, writes atomic.Int64

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/organizations/org/app_definitions/app", r.URL.Path)
				assert.Empty(t, r.Header.Get("X-Contentful-Version"))
				w.Header().Set("Content-Type", "application/json")

				switch r.Method {
				case http.MethodGet:
					fmt.Fprint(w, testJSON(map[string]any{"sys": map[string]any{"type": "AppDefinition", "id": "app", "organization": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Organization", "id": "org"}}}, "name": "App", "src": "https://example.invalid/app", "locations": []any{map[string]any{"location": "app-config"}}, "parameters": map[string]any{"instance": []any{map[string]any{"id": "setting", "name": "Setting", "type": "Enum", "options": []any{}}}}}))
				case http.MethodDelete:
					deletes.Add(1)
					w.WriteHeader(http.StatusNoContent)
				default:
					writes.Add(1)
					t.Errorf("unexpected %s", r.Method)
					w.WriteHeader(http.StatusInternalServerError)
				}
			}))
			defer server.Close()

			runtime := newTerraformTestRuntime(t)
			runtime.providerURL = server.URL
			runtime.writeConfig(t, appDefinitionParameterDestroyConfig(false, test.ignored))
			output, err := runtime.run(t.Context(), "import", "-input=false", "-no-color", "contentful_app_definition.test", "org/app")
			require.NoError(t, err, output)

			if test.ignored {
				output, err = runtime.run(t.Context(), "plan", "-input=false", "-no-color", "-detailed-exitcode")
				require.NoError(t, err, output)

				renamed := strings.Replace(appDefinitionParameterDestroyConfig(false, true), `name = "App"`, `name = "Changed"`, 1)
				runtime.writeConfig(t, renamed)
				output, err = runtime.run(t.Context(), "apply", "-auto-approve", "-input=false", "-no-color")
				require.Error(t, err, output)
				require.Contains(t, output, "options must contain at least one option")
				require.Zero(t, writes.Load())
				runtime.writeConfig(t, appDefinitionParameterDestroyConfig(false, true))
			}

			if test.invalid {
				runtime.writeConfig(t, appDefinitionParameterDestroyConfig(true, test.ignored))
				output, err = runtime.run(t.Context(), "destroy", "-auto-approve", "-input=false", "-json")
				require.Error(t, err, output)

				diagnostics := 0

				for line := range strings.SplitSeq(output, "\n") {
					var event struct {
						Type       string `json:"type"`
						Diagnostic struct {
							Severity string `json:"severity"`
							Summary  string `json:"summary"`
							Detail   string `json:"detail"`
						} `json:"diagnostic"`
					}
					if json.Unmarshal([]byte(line), &event) != nil || event.Type != "diagnostic" || event.Diagnostic.Severity != "error" {
						continue
					}

					diagnostics++

					require.Equal(t, "Invalid App Definition parameter configuration", event.Diagnostic.Summary)
					require.Equal(t, "When configured, options must contain at least one option. Omit options to leave them unspecified.", event.Diagnostic.Detail)
				}

				require.Positive(t, diagnostics, output)
				require.Zero(t, deletes.Load())
				require.Zero(t, writes.Load())
				t.Logf("invalid destroy: error diagnostics=%d, DELETE=%d, POST/PUT=%d", diagnostics, deletes.Load(), writes.Load())
				runtime.writeConfig(t, appDefinitionParameterDestroyConfig(false, test.ignored))
			}

			output, err = runtime.run(t.Context(), "destroy", "-auto-approve", "-input=false", "-no-color")
			require.NoError(t, err, output)
			require.EqualValues(t, 1, deletes.Load())
			require.Zero(t, writes.Load())
			t.Logf("valid/corrected destroy: DELETE=%d, POST/PUT=%d", deletes.Load(), writes.Load())
		})
	}
}
