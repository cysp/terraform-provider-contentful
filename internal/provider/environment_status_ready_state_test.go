//nolint:testpackage
package provider

import (
	"testing"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	datasourcetimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func environmentStatusReadyTimeoutAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"read": types.StringType,
	}
}

func TestPublishEnvironmentStatusReadyConversionErrorDoesNotPublish(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	state := tfsdk.State{Schema: EnvironmentStatusReadyDataSourceSchema(ctx)}
	current := environmentStatusReadyTestModel(
		types.StringValue("queued"),
		datasourcetimeouts.Value{
			Object: types.ObjectNull(environmentStatusReadyTimeoutAttributeTypes()),
		},
	)
	require.False(t, state.Set(ctx, &current).HasError())
	before := state.Raw
	conversionDiags := diag.Diagnostics{
		diag.NewWarningDiagnostic("Incomplete Contentful response", "conversion warning"),
		diag.NewErrorDiagnostic("Malformed Contentful response", "conversion failed"),
	}

	continuePolling, diags := publishEnvironmentStatusReadyConversion(
		ctx,
		&state,
		datasourcetimeouts.Value{Object: types.ObjectUnknown(environmentStatusReadyTimeoutAttributeTypes())},
		EnvironmentStatusReadyModel{Status: types.StringUnknown()},
		conversionDiags,
	)

	require.True(t, diags.HasError())
	assert.False(t, continuePolling)
	assert.Equal(t, conversionDiags, diags)
	assert.True(t, before.Equal(state.Raw))
}

func TestPublishEnvironmentStatusReadyConversionWarningPublishes(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	state := tfsdk.State{Schema: EnvironmentStatusReadyDataSourceSchema(ctx)}
	configuredTimeouts := environmentStatusReadyTimeoutValue(types.StringUnknown())
	conversionDiags := diag.Diagnostics{
		diag.NewWarningDiagnostic("Incomplete Contentful response", "conversion warning"),
	}

	continuePolling, diags := publishEnvironmentStatusReadyConversion(
		ctx,
		&state,
		configuredTimeouts,
		environmentStatusReadyTestModel(types.StringValue("queued"), datasourcetimeouts.Value{}),
		conversionDiags,
	)

	require.False(t, diags.HasError())
	assert.True(t, continuePolling)
	assert.Equal(t, conversionDiags, diags)

	var published EnvironmentStatusReadyModel
	require.False(t, state.Get(ctx, &published).HasError())
	assert.Equal(t, types.StringValue("space/environment"), published.ID)
	assert.Equal(t, types.StringValue("space"), published.SpaceID)
	assert.Equal(t, types.StringValue("environment"), published.EnvironmentID)
	assert.Equal(t, types.StringValue("queued"), published.Status)
	assert.True(t, configuredTimeouts.Equal(published.Timeouts))
}

func TestPublishEnvironmentStatusReadyResponsePreservesTimeoutsAndControlsPolling(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		configuredTimeouts datasourcetimeouts.Value
		status             string
		continuePolling    bool
		expectedError      string
	}{
		"null timeout and queued": {
			configuredTimeouts: datasourcetimeouts.Value{
				Object: types.ObjectNull(environmentStatusReadyTimeoutAttributeTypes()),
			},
			status:          "queued",
			continuePolling: true,
		},
		"null timeout and in progress": {
			configuredTimeouts: datasourcetimeouts.Value{
				Object: types.ObjectNull(environmentStatusReadyTimeoutAttributeTypes()),
			},
			status:          "inProgress",
			continuePolling: true,
		},
		"unknown timeout and ready": {
			configuredTimeouts: datasourcetimeouts.Value{
				Object: types.ObjectUnknown(environmentStatusReadyTimeoutAttributeTypes()),
			},
			status:          environmentStatusReadyValue,
			continuePolling: false,
		},
		"known timeout and ready": {
			configuredTimeouts: environmentStatusReadyTimeoutValue(types.StringValue("30m")),
			status:             environmentStatusReadyValue,
			continuePolling:    false,
		},
		"known timeout and failed": {
			configuredTimeouts: environmentStatusReadyTimeoutValue(types.StringValue("30m")),
			status:             "failed",
			continuePolling:    false,
			expectedError:      "Contentful environment failed to become ready",
		},
		"unknown future status remains pollable": {
			configuredTimeouts: environmentStatusReadyTimeoutValue(types.StringValue("30m")),
			status:             "futureStatus",
			continuePolling:    true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			state := tfsdk.State{Schema: EnvironmentStatusReadyDataSourceSchema(ctx)}
			response := cm.Environment{
				Sys:  cm.NewEnvironmentSys("space", "environment", test.status),
				Name: "environment",
			}

			continuePolling, diags := publishEnvironmentStatusReadyResponse(
				ctx,
				&state,
				response,
				test.configuredTimeouts,
			)

			if test.expectedError == "" {
				require.False(t, diags.HasError(), diags.Errors())
			} else {
				require.Len(t, diags.Errors(), 1)
				assert.Equal(t, test.expectedError, diags.Errors()[0].Summary())
				assert.Equal(
					t,
					"Contentful reported the environment status as failed.",
					diags.Errors()[0].Detail(),
				)
			}

			assert.Equal(t, test.continuePolling, continuePolling)

			var published EnvironmentStatusReadyModel
			require.False(t, state.Get(ctx, &published).HasError())
			assert.Equal(t, types.StringValue(test.status), published.Status)
			assert.True(t, test.configuredTimeouts.Equal(published.Timeouts))
		})
	}
}

func TestPublishEnvironmentStatusReadyStatePublicationErrorStopsPollingAndLeavesStateUntouched(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	state := tfsdk.State{Schema: EnvironmentStatusReadyDataSourceSchema(ctx)}
	current := environmentStatusReadyTestModel(
		types.StringValue("queued"),
		datasourcetimeouts.Value{Object: types.ObjectNull(environmentStatusReadyTimeoutAttributeTypes())},
	)
	require.False(t, state.Set(ctx, &current).HasError())
	before := state.Raw
	state.Schema = datasourceschema.Schema{}
	response := cm.Environment{
		Sys:  cm.NewEnvironmentSys("space", "environment", environmentStatusReadyValue),
		Name: "environment",
	}

	continuePolling, diags := publishEnvironmentStatusReadyResponse(
		ctx,
		&state,
		response,
		environmentStatusReadyTimeoutValue(types.StringNull()),
	)

	require.True(t, diags.HasError())
	assert.False(t, continuePolling)
	assert.True(t, before.Equal(state.Raw))
}

func environmentStatusReadyTimeoutValue(read types.String) datasourcetimeouts.Value {
	return datasourcetimeouts.Value{
		Object: types.ObjectValueMust(
			environmentStatusReadyTimeoutAttributeTypes(),
			map[string]attr.Value{"read": read},
		),
	}
}

func environmentStatusReadyTestModel(
	status types.String,
	timeouts datasourcetimeouts.Value,
) EnvironmentStatusReadyModel {
	return EnvironmentStatusReadyModel{
		IDIdentityModel: NewIDIdentityModelFromMultipartID("space", "environment"),
		EnvironmentIdentityModel: EnvironmentIdentityModel{
			SpaceID:       types.StringValue("space"),
			EnvironmentID: types.StringValue("environment"),
		},
		Status:   status,
		Timeouts: timeouts,
	}
}
