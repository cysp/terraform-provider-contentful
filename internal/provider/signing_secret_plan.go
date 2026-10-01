package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Both signing-secret resources only mutate their secret. Their timestamps must
// remain unknown for a possible replacement and retain state for local changes.
func modifySigningSecretPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return
	}

	var stateValue, planValue types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("value"), &stateValue)...)
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("value"), &planValue)...)

	if resp.Diagnostics.HasError() {
		return
	}

	for _, name := range []string{"created_at", "updated_at"} {
		value := timetypes.NewRFC3339Unknown()
		if planValue.Equal(stateValue) {
			resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root(name), &value)...)
		}

		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root(name), value)...)
	}
}
