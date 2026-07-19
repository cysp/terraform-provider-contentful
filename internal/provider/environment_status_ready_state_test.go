//nolint:testpackage
package provider

import (
	"context"
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	datasourcetimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvironmentStatusReadyConversionErrorLeavesStateUntouched(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	dataSourceSchema := EnvironmentStatusReadyDataSourceSchema(ctx)
	state := tfsdk.State{Schema: dataSourceSchema}
	current := EnvironmentStatusReadyModel{
		IDIdentityModel: NewIDIdentityModelFromMultipartID("space", "environment"),
		EnvironmentIdentityModel: EnvironmentIdentityModel{
			SpaceID:       types.StringValue("space"),
			EnvironmentID: types.StringValue("environment"),
		},
		Status: types.StringValue("queued"),
		Timeouts: datasourcetimeouts.Value{Object: types.ObjectNull(map[string]attr.Type{
			"read": types.StringType,
		})},
	}
	require.False(t, state.Set(ctx, &current).HasError())
	before := state.Raw

	result, diags := setEnvironmentStatusReadyState(
		ctx,
		&state,
		current,
		cm.Environment{},
		func(context.Context, cm.Environment) (EnvironmentStatusReadyModel, diag.Diagnostics) {
			return EnvironmentStatusReadyModel{Status: types.StringUnknown()}, diag.Diagnostics{
				diag.NewErrorDiagnostic("Malformed Contentful response", "conversion failed"),
			}
		},
	)

	require.True(t, diags.HasError())
	assert.Equal(t, EnvironmentStatusReadyModel{}, result)
	assert.True(t, before.Equal(state.Raw))
}
