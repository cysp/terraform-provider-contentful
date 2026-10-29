package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	datasourcetimeouts "github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func publishEnvironmentStatusReadyResponse(
	ctx context.Context,
	state *tfsdk.State,
	response cm.Environment,
	requested cm.GetEnvironmentParams,
	configuredTimeouts datasourcetimeouts.Value,
) (bool, diag.Diagnostics) {
	if response.Sys.Space.Sys.ID != requested.SpaceID || response.Sys.ID != requested.EnvironmentID {
		return false, diag.Diagnostics{diag.NewErrorDiagnostic("Unexpected environment readiness identity", "Contentful did not return the requested space and environment identity. This response cannot establish readiness.")}
	}

	model := NewEnvironmentStatusReadyModelFromResponse(response)
	diags := diag.Diagnostics{}

	model.Timeouts = configuredTimeouts
	diags.Append(state.Set(ctx, &model)...)

	if diags.HasError() {
		return false, diags
	}

	if model.Status.ValueString() == environmentStatusFailedValue {
		diags.AddError(
			"Contentful environment failed to become ready",
			"Contentful reported the environment status as failed.",
		)

		return false, diags
	}

	return model.Status.ValueString() != environmentStatusReadyValue, diags
}
