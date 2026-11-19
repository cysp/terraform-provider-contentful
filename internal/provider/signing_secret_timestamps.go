package provider

import (
	"context"
	"time"

	cm "github.com/cysp/terraform-provider-contentful/internal/contentful-management-go"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func signingSecretTimestamp(value cm.OptDateTime) types.String {
	if timestamp, ok := value.Get(); ok {
		return types.StringValue(timestamp.UTC().Format(time.RFC3339Nano))
	}

	return types.StringNull()
}

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
		value := types.StringUnknown()
		if planValue.Equal(stateValue) {
			resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root(name), &value)...)
		}

		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root(name), value)...)
	}
}
