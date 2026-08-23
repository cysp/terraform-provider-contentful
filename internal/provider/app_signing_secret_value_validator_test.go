package provider_test

import (
	"strings"
	"testing"

	"github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppSigningSecretValueValidatorAcceptsEveryPermittedCharacter(t *testing.T) {
	t.Parallel()

	valueValidator := appSigningSecretValueAttribute(t).Validators[0]

	for _, character := range "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ+/=_-" {
		t.Run(string(character), func(t *testing.T) {
			t.Parallel()

			var response validator.StringResponse
			valueValidator.ValidateString(t.Context(), validator.StringRequest{
				Path:        path.Root("value"),
				ConfigValue: types.StringValue(strings.Repeat(string(character), 64)),
			}, &response)

			require.Empty(t, response.Diagnostics)
		})
	}
}

func TestAppSigningSecretValueValidatorRejectsInvalidValuesWithoutExposingThem(t *testing.T) {
	t.Parallel()

	valueValidator := appSigningSecretValueAttribute(t).Validators[0]
	invalidCharacterPrefix := "DO_NOT_ECHO_INVALID_APP_SIGNING_SECRET!"
	tests := map[string]string{
		"too short":         strings.Repeat("s", 63),
		"too long":          strings.Repeat("s", 65),
		"invalid character": invalidCharacterPrefix + strings.Repeat("X", 64-len(invalidCharacterPrefix)),
	}

	for name, value := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var response validator.StringResponse
			valueValidator.ValidateString(t.Context(), validator.StringRequest{
				Path:        path.Root("value"),
				ConfigValue: types.StringValue(value),
			}, &response)

			require.True(t, response.Diagnostics.HasError())

			for _, diagnostic := range response.Diagnostics {
				assert.NotContains(t, diagnostic.Summary(), value)
				assert.NotContains(t, diagnostic.Detail(), value)
			}
		})
	}
}

func appSigningSecretValueAttribute(t *testing.T) schema.StringAttribute {
	t.Helper()

	resourceSchema := provider.AppSigningSecretResourceSchema(t.Context())
	valueAttribute, ok := resourceSchema.Attributes["value"].(schema.StringAttribute)
	require.True(t, ok)
	require.Len(t, valueAttribute.Validators, 1)

	return valueAttribute
}
