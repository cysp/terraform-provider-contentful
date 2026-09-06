package provider_test

import (
	"strings"
	"testing"

	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestPreviewEnvironmentIDSchemaMatchesEndpointEnvelope(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		value       types.String
		expectError bool
	}{
		"omitted":                   {value: types.StringNull()},
		"unknown":                   {value: types.StringUnknown()},
		"one character":             {value: types.StringValue("a")},
		"64 characters":             {value: types.StringValue(strings.Repeat("A", 64))},
		"uppercase and punctuation": {value: types.StringValue("-A_z9-")},
		"empty":                     {value: types.StringValue(""), expectError: true},
		"65 characters":             {value: types.StringValue(strings.Repeat("a", 65)), expectError: true},
		"dot":                       {value: types.StringValue("preview.id"), expectError: true},
		"space":                     {value: types.StringValue("preview id"), expectError: true},
		"at sign":                   {value: types.StringValue("preview@id"), expectError: true},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			attribute, ok := PreviewEnvironmentResourceSchema(t.Context()).Attributes["preview_environment_id"].(schema.StringAttribute)
			require.True(t, ok)

			request := validator.StringRequest{
				Path:        path.Root("preview_environment_id"),
				ConfigValue: test.value,
			}

			response := validator.StringResponse{}
			for _, valueValidator := range attribute.Validators {
				valueValidator.ValidateString(t.Context(), request, &response)
			}

			require.Equal(t, test.expectError, response.Diagnostics.HasError(), response.Diagnostics)
		})
	}
}
