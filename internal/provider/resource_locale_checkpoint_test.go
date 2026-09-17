package provider_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocaleCreateCheckpointsResponseAndVersion(t *testing.T) {
	t.Parallel()

	for name, missingVersion := range map[string]bool{
		"returned version": false,
		"missing version":  true,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			versionJSON := `,"version":9`
			if missingVersion {
				versionJSON = ""
			}

			responseJSON := `{"name":"Returned","code":"en-AU","fallbackCode":null,"contentDeliveryApi":true,"contentManagementApi":true,"optional":false,"default":false,"sys":{"type":"Locale","id":"locale","space":{"sys":{"type":"Link","linkType":"Space","id":"space"}},"environment":{"sys":{"type":"Link","linkType":"Environment","id":"environment"}}` + versionJSON + `}}`

			var method, target, version, body string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				method, target, version = r.Method, r.URL.Path, r.Header.Get("X-Contentful-Version")

				data, err := io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)

					return
				}

				body = string(data)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_, _ = io.WriteString(w, responseJSON)
			}))
			t.Cleanup(server.Close)
			factories := makeTestAccProtoV6ProviderFactories(provider.WithContentfulURL(server.URL), provider.WithAccessToken("test"))
			protocol, err := factories["contentful"]()
			require.NoError(t, err)
			config, err := providerConfigDynamicValue(map[string]any{"url": server.URL, "access_token": "test"})
			require.NoError(t, err)
			configured, err := protocol.ConfigureProvider(t.Context(), &tfprotov6.ConfigureProviderRequest{Config: &config})
			require.NoError(t, err)
			require.Empty(t, configured.Diagnostics)

			model := validLocaleRequestModel()
			model.Name = types.StringValue("Planned")
			model.ID, model.LocaleID, model.Default = types.StringUnknown(), types.StringUnknown(), types.BoolUnknown()
			model.Timeouts = provider.TimeoutsNull()
			schema := provider.LocaleResourceSchema(t.Context())
			prior := nullResourceDynamicValue(t, schema)
			plan := resourceModelDynamicValue(t, schema, model)
			applied, err := protocol.ApplyResourceChange(t.Context(), &tfprotov6.ApplyResourceChangeRequest{TypeName: "contentful_locale", PriorState: &prior, PlannedState: &plan, Config: &plan})
			require.NoError(t, err)
			assert.Equal(t, http.MethodPost, method)
			assert.Equal(t, "/spaces/space/environments/environment/locales", target)
			assert.Empty(t, version)
			assert.JSONEq(t, `{"name":"Planned","code":"en-AU","fallbackCode":null,"contentDeliveryApi":true,"contentManagementApi":true,"optional":false}`, body)

			expectedDiagnostics := 1
			if missingVersion {
				expectedDiagnostics++
			}

			require.Len(t, applied.Diagnostics, expectedDiagnostics)
			assert.Equal(t, "Contentful returned a different locale name", applied.Diagnostics[expectedDiagnostics-1].Summary)
			assert.Equal(t, "Contentful accepted the request but returned a value that differs from the value Terraform applied. Terraform retained the returned value in state rather than substituting the planned value. Review the locale and configuration before applying again.", applied.Diagnostics[expectedDiagnostics-1].Detail)
			require.NotNil(t, applied.NewState)
			raw, err := applied.NewState.Unmarshal(schema.Type().TerraformType(t.Context()))
			require.NoError(t, err)

			state := tfsdk.State{Raw: raw, Schema: schema}

			var result provider.LocaleModel
			require.Empty(t, state.Get(t.Context(), &result))
			assert.Equal(t, types.StringValue("Returned"), result.Name)
			assert.Equal(t, types.StringValue("space/environment/locale"), result.ID)
			require.NotNil(t, applied.NewIdentity)

			var values map[string][]byte
			if len(applied.Private) > 0 {
				require.NoError(t, json.Unmarshal(applied.Private, &values))
			}

			if missingVersion {
				assert.Empty(t, values["version"])
				assert.Equal(t, "Missing locale version", applied.Diagnostics[0].Summary)
			} else {
				assert.Equal(t, []byte("9"), values["version"])
			}
		})
	}
}
