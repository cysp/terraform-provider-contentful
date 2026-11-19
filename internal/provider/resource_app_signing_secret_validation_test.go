package provider_test

import (
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppSigningSecretRejectsInvalidResolvedValueBeforeMutation(t *testing.T) {
	t.Parallel()

	const typeName = "contentful_app_signing_secret"

	for _, test := range []struct {
		name  string
		value string
		prior bool
	}{
		{name: "create", value: strings.Repeat("s", 63)},
		{name: "update", value: strings.Repeat("s", 63) + "!", prior: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var requests atomic.Int64

			server := configuredProtocolServer(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				requests.Add(1)
				w.WriteHeader(http.StatusBadRequest)
			}))

			schema := provider.AppSigningSecretResourceSchema(t.Context())
			model := provider.AppSigningSecretModel{
				OrganizationID:  types.StringValue("org"),
				AppDefinitionID: types.StringValue("app"),
				Value:           types.StringUnknown(),
				Timeouts:        provider.TimeoutsNull(),
			}
			config := resourceModelDynamicValue(t, schema, model)
			validation, err := server.ValidateResourceConfig(t.Context(), &tfprotov6.ValidateResourceConfigRequest{
				TypeName: typeName, Config: &config,
			})
			require.NoError(t, err)
			require.Empty(t, validation.Diagnostics)

			model.Value = types.StringValue(test.value)
			config = resourceModelDynamicValue(t, schema, model)
			model.ID = types.StringUnknown()
			prior := nullResourceDynamicValue(t, schema)

			if test.prior {
				model.ID = types.StringValue("org/app")
				old := model
				old.Value = types.StringValue(strings.Repeat("a", 64))
				prior = resourceModelDynamicValue(t, schema, old)
			}

			planned := resourceModelDynamicValue(t, schema, model)
			result, err := server.ApplyResourceChange(t.Context(), &tfprotov6.ApplyResourceChangeRequest{
				TypeName: typeName, PriorState: &prior, PlannedState: &planned, Config: &config,
			})
			require.NoError(t, err)
			assert.Zero(t, requests.Load())
			require.Len(t, result.Diagnostics, 1)
			diagnostic := result.Diagnostics[0]
			assert.Equal(t, tfprotov6.DiagnosticSeverityError, diagnostic.Severity)
			assert.Equal(t, "Invalid app signing secret value", diagnostic.Summary)
			assert.Equal(t, "app signing secret must be exactly 64 characters and contain only letters, digits, +, /, =, _, or -.", diagnostic.Detail)
			assert.Equal(t, tftypes.NewAttributePath().WithAttributeName("value"), diagnostic.Attribute)
		})
	}
}
