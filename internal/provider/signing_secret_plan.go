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
	resp.Diagnostics.Append(writeOnlySecretContextDiagnostics(ctx)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Without an ordinary value, any planned update can receive a different
	// write-only value at apply, even when planning Config omits it.
	possibleWrite := !planValue.Equal(stateValue) || (planValue.IsNull() && !req.Plan.Raw.Equal(req.State.Raw))
	if !configuredWO.IsNull() {
		// Write-only values may differ at apply. Any other planned change must retain
		// room for a write even when the planning-time value matches.
		possibleWrite = true

		hashes, hashDiags := readSigningSecretHashes(ctx, req.Private)
		for _, diagnostic := range hashDiags {
			switch invalid := diagnostic.(type) {
			case invalidWriteOnlySecretHashesDiagnostic:
				resp.Diagnostics.AddWarning(invalid.Summary(), invalid.reason+" The provider cannot compare the configured secret. An in-place update will fail before writing the secret. Restore valid Terraform state before updating, or explicitly replace or destroy the resource.")
			default:
				resp.Diagnostics.Append(diagnostic)
			}
		}

		if resp.Diagnostics.HasError() {
			return
		}

		if !hashDiags.HasError() && !configuredWO.IsUnknown() && planValue.IsNull() {
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
