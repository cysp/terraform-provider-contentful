//nolint:testpackage // The package-local request configuration helper is tested directly.
package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRejectUnknownConfigurationOwnedRequestValue(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		configured   types.String
		planned      types.String
		expectedPath string
	}{
		"config null plan unknown is response-owned": {
			configured: types.StringNull(),
			planned:    types.StringUnknown(),
		},
		"config null plan known uses plan": {
			configured: types.StringNull(),
			planned:    types.StringValue("planned default or prior state"),
		},
		"config known plan known uses plan": {
			configured: types.StringValue("configured expression"),
			planned:    types.StringValue("resolved planned value"),
		},
		"config known plan unknown is rejected": {
			configured:   types.StringValue("configuration-owned"),
			planned:      types.StringUnknown(),
			expectedPath: "value",
		},
		"known empty remains known": {
			configured: types.StringValue(""),
			planned:    types.StringValue(""),
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := rejectUnknownConfigurationOwnedRequestValue(test.planned, test.configured, path.Root("value"))

			if test.expectedPath == "" {
				assert.False(t, diags.HasError(), diags.Errors())

				return
			}

			require.True(t, diags.HasError())
			assert.Equal(t, []string{test.expectedPath}, mutationDiagnosticPaths(t, diags))
		})
	}
}
