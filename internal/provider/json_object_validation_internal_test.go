package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONObjectValidation(t *testing.T) {
	t.Parallel()

	v := jsonObjectValidator{summary: "Invalid object", detail: "Provide an object."}
	valuePath := path.Root("settings")

	for _, test := range []struct {
		name, raw string
		valid     bool
	}{
		{"empty object", "{}", true},
		{"whitespace", " \n { \"value\": 9007199254740993 } \t", true},
		{"opaque members", `{"a":null,"b":[false,1e100],"a":{"x":-0}}`, true},
		{"null", "null", false},
		{"array", "[]", false},
		{"string", `"{}"`, false},
		{"number", "42", false},
		{"boolean", "false", false},
		{"empty", "", false},
		{"malformed", "{", false},
		{"second document", "{} {}", false},
		{"trailing junk", "{}x", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var response validator.StringResponse
			v.ValidateString(t.Context(), validator.StringRequest{ConfigValue: types.StringValue(test.raw), Path: valuePath}, &response)

			requestDiags := v.validate(test.raw, valuePath)
			if test.valid {
				require.Empty(t, response.Diagnostics)
				require.Empty(t, requestDiags)

				return
			}

			require.Len(t, response.Diagnostics, 1)
			require.Equal(t, response.Diagnostics, requestDiags)
			diagnostic := response.Diagnostics[0]
			assert.Equal(t, "Invalid object", diagnostic.Summary())
			assert.Equal(t, "Provide an object.", diagnostic.Detail())
			assert.Equal(t, []string{"settings"}, mutationDiagnosticPaths(t, requestDiags))
		})
	}

	for _, value := range []types.String{types.StringNull(), types.StringUnknown()} {
		var response validator.StringResponse
		v.ValidateString(t.Context(), validator.StringRequest{ConfigValue: value, Path: valuePath}, &response)
		require.Empty(t, response.Diagnostics)
	}
}
