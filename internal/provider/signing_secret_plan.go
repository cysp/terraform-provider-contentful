package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// value_wo is null in both plan and state. Mark the timestamps unknown when
// the secret may have changed so Terraform can schedule Update.
func modifySigningSecretPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return
	}

	var stateValue, planValue, configuredWO types.String
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("value"), &stateValue)...)
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("value"), &planValue)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("value_wo"), &configuredWO)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Any planned update can receive a different write-only value at apply,
	// including when ignore_changes hides that value during planning.
	possibleWrite := !req.Plan.Raw.Equal(req.State.Raw)
	if !configuredWO.IsNull() {
		// Write-only values may differ at apply. Any other planned change must retain
		// room for a write even when the planning-time value matches.
		possibleWrite = true
		hashes, hashDiags := readSigningSecretHashes(ctx, req.Private)
		resp.Diagnostics.Append(hashDiags...)

		if resp.Diagnostics.HasError() {
			return
		}

		if !configuredWO.IsUnknown() && planValue.IsNull() {
			matches, matchDiags := signingSecretValueMatches(ctx, &hashes, stateValue, configuredWO)
			resp.Diagnostics.Append(matchDiags...)

			if resp.Diagnostics.HasError() {
				return
			}

			possibleWrite = !matches || !req.Plan.Raw.Equal(req.State.Raw)
		}
	}

	for _, name := range []string{"created_at", "updated_at"} {
		value := timetypes.NewRFC3339Unknown()
		if !possibleWrite {
			resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root(name), &value)...)
		}

		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root(name), value)...)
	}
}
