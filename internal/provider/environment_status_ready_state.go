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
	configuredTimeouts datasourcetimeouts.Value,
) (bool, diag.Diagnostics) {
	model, diags := NewEnvironmentStatusReadyModelFromResponse(ctx, response)

	return publishEnvironmentStatusReadyConversion(ctx, state, configuredTimeouts, model, diags)
}

func publishEnvironmentStatusReadyConversion(
	ctx context.Context,
	state *tfsdk.State,
	configuredTimeouts datasourcetimeouts.Value,
	model EnvironmentStatusReadyModel,
	diags diag.Diagnostics,
) (bool, diag.Diagnostics) {
	if diags.HasError() {
		return false, diags
	}

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
