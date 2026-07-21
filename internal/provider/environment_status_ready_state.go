package provider

import (
	"context"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type environmentStatusReadyResponseConverter func(context.Context, cm.Environment) (EnvironmentStatusReadyModel, diag.Diagnostics)

func setEnvironmentStatusReadyState(
	ctx context.Context,
	state *tfsdk.State,
	current EnvironmentStatusReadyModel,
	response cm.Environment,
	convert environmentStatusReadyResponseConverter,
) (EnvironmentStatusReadyModel, diag.Diagnostics) {
	model, diags := convert(ctx, response)
	if diags.HasError() {
		return EnvironmentStatusReadyModel{}, diags
	}

	model.Timeouts = current.Timeouts
	diags.Append(state.Set(ctx, &model)...)

	if diags.HasError() {
		return EnvironmentStatusReadyModel{}, diags
	}

	return model, diags
}
