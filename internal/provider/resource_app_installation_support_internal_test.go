package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

func appInstallationTestPlan(t *testing.T, parameters string) tfsdk.Plan {
	t.Helper()
	schema := AppInstallationResourceSchema(t.Context())
	raw, err := tftypes.ValueFromJSONWithOpts([]byte(testJSON(map[string]any{
		"id": "space/master/app", "space_id": "space", "environment_id": "master", "app_definition_id": "app", "parameters": parameters,
	})), schema.Type().TerraformType(t.Context()), tftypes.ValueFromJSONOpts{})
	require.NoError(t, err)

	return tfsdk.Plan{Schema: schema, Raw: raw}
}
