package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *extensionResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	priorSchema := ExtensionResourceSchema(ctx)
	priorSchema.Version = 0

	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema:   &priorSchema,
			StateUpgrader: upgradeExtensionStateV0,
		},
	}
}

func upgradeExtensionStateV0(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	var state ExtensionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if state.Extension == nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("extension"),
			"Missing extension configuration",
			"The prior state does not contain the extension configuration required for a state upgrade.",
		)

		return
	}

	if state.Extension.Src.ValueString() == "" {
		state.Extension.Src = types.StringNull()
	} else if state.Extension.SrcDoc.ValueString() == "" {
		state.Extension.SrcDoc = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
