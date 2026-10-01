package provider_test

import (
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppSigningSecretMutationRequiresConfirmedIdentity(t *testing.T) {
	t.Parallel()

	for name, response := range map[string]struct {
		organizationID, appDefinitionID string
		diagnosticAttributes            []string
	}{
		"organization":   {"other-org", "app", []string{"organization_id"}},
		"app definition": {"org", "other-app", []string{"app_definition_id"}},
		"both":           {"other-org", "other-app", []string{"organization_id", "app_definition_id"}},
	} {
		for _, update := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/update=%t", name, update), func(t *testing.T) {
				t.Parallel()

				var calls atomic.Int64

				protocol := configuredProtocolServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					calls.Add(1)
					assert.Equal(t, http.MethodPut, r.Method)
					assert.Equal(t, "/organizations/org/app_definitions/app/signing_secret", r.URL.Path)
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(testJSON(map[string]any{
						"sys": map[string]any{
							"type":          "AppSigningSecret",
							"organization":  map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Organization", "id": response.organizationID}},
							"appDefinition": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "AppDefinition", "id": response.appDefinitionID}},
						},
						"redactedValue": "****aaaa",
					})))
				}))
				resourceSchema := AppSigningSecretResourceSchema(t.Context())
				values := map[string]any{
					"id": "org/app", "organization_id": "org", "app_definition_id": "app", "value": strings.Repeat("a", 64),
				}
				planned := mutationEndpointDynamicValue(t, resourceSchema, values)
				prior := nullResourceDynamicValue(t, resourceSchema)

				if update {
					values["value"] = strings.Repeat("b", 64)
					prior = mutationEndpointDynamicValue(t, resourceSchema, values)
				}

				result, err := protocol.ApplyResourceChange(t.Context(), &tfprotov6.ApplyResourceChangeRequest{
					TypeName: "contentful_app_signing_secret", PriorState: &prior, PlannedState: &planned, Config: &planned,
				})
				require.NoError(t, err)
				require.EqualValues(t, 1, calls.Load(), result.Diagnostics)
				require.Len(t, result.Diagnostics, len(response.diagnosticAttributes)+1)

				for index, attribute := range response.diagnosticAttributes {
					diagnostic := result.Diagnostics[index]
					assert.Equal(t, tfprotov6.DiagnosticSeverityError, diagnostic.Severity)
					assert.Equal(t, "Unexpected app signing secret identity", diagnostic.Summary)
					assert.Equal(t, tftypes.NewAttributePath().WithAttributeName(attribute), diagnostic.Attribute)
				}

				recoveryDiagnostic := result.Diagnostics[len(response.diagnosticAttributes)]
				assert.Equal(t, tfprotov6.DiagnosticSeverityError, recoveryDiagnostic.Severity)
				assert.Equal(t, "Unconfirmed app signing secret write", recoveryDiagnostic.Summary)
				require.NotNil(t, result.NewState)
				priorRaw, err := prior.Unmarshal(resourceSchema.Type().TerraformType(t.Context()))
				require.NoError(t, err)
				actualRaw, err := result.NewState.Unmarshal(resourceSchema.Type().TerraformType(t.Context()))
				require.NoError(t, err)
				assert.True(t, priorRaw.Equal(actualRaw), "unconfirmed response must not acknowledge a new secret")
			})
		}
	}
}
