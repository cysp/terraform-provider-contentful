package provider_test

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	. "github.com/cysp/terraform-provider-contentful/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppInstallationResponsePreservesReferenceValues(t *testing.T) {
	t.Parallel()

	reference := AppInstallationModel{
		Marketplace: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("terms"), types.StringValue("privacy")}),
		Parameters:  jsontypes.NewNormalizedValue(testJSON(map[string]any{"visible": "prior", "secret": "retained secret"})),
		Timeouts: timeouts.Value{Object: types.ObjectValueMust(
			map[string]attr.Type{"create": types.StringType, "update": types.StringType},
			map[string]attr.Value{"create": types.StringValue("2m"), "update": types.StringValue("3m")},
		)},
	}
	response := cm.AppInstallation{
		Sys:        cm.NewAppInstallationSys("space", "environment", "app"),
		Parameters: []byte(testJSON(map[string]any{"visible": "remote", "secret": "***"})),
	}

	actual, diags := NewAppInstallationResourceModelFromResponse(response, reference)

	require.Empty(t, diags)
	assert.Equal(t, AppInstallationModel{
		ID:              types.StringValue("space/environment/app"),
		SpaceID:         types.StringValue("space"),
		EnvironmentID:   types.StringValue("environment"),
		AppDefinitionID: types.StringValue("app"),
		Marketplace:     reference.Marketplace,
		Timeouts:        reference.Timeouts,
		Parameters:      jsontypes.NewNormalizedValue(testJSON(map[string]any{"secret": "retained secret", "visible": "remote"})),
	}, actual)
}
