//nolint:testpackage // The package-private Webhook request converter is intentionally tested directly.
package provider

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestWebhookOptionalNullableString(t *testing.T) {
	t.Parallel()

	for name, test := range map[string]struct {
		value         types.String
		expected      cm.OptNilString
		expectedPaths []string
	}{
		"known":   {value: types.StringValue("value"), expected: cm.NewOptNilString("value"), expectedPaths: []string{}},
		"null":    {value: types.StringNull(), expected: cm.NewOptNilStringNull(), expectedPaths: []string{}},
		"unknown": {value: types.StringUnknown(), expectedPaths: []string{"value"}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual, diags := webhookOptionalNullableString(test.value, path.Root("value"))

			assert.Equal(t, test.expected, actual)
			assert.Equal(t, test.expectedPaths, requestDiagnosticPaths(t, diags))
		})
	}
}
