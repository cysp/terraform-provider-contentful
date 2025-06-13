// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

// copied from https://github.com/hashicorp/terraform-plugin-framework/tree/main/resource/schema/planmodifier and adjusted to
// preserve null state values after initial resource creation.

//nolint:revive
func UseStateForUnknown() useStateForUnknownModifier {
	return useStateForUnknownModifier{}
}

// useStateForUnknownModifier implements the plan modifier.
type useStateForUnknownModifier struct{}

var (
	_ planmodifier.Bool   = useStateForUnknownModifier{}
	_ planmodifier.List   = useStateForUnknownModifier{}
	_ planmodifier.Map    = useStateForUnknownModifier{}
	_ planmodifier.String = useStateForUnknownModifier{}
)

func (m useStateForUnknownModifier) Description(_ context.Context) string {
	return "Once set, the value of this attribute in state will not change."
}

func (m useStateForUnknownModifier) MarkdownDescription(_ context.Context) string {
	return "Once set, the value of this attribute in state will not change."
}

func (m useStateForUnknownModifier) PlanModifyBool(_ context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	if req.State.Raw.IsNull() || !req.PlanValue.IsUnknown() || req.ConfigValue.IsUnknown() {
		return
	}

	resp.PlanValue = req.StateValue
}

func (m useStateForUnknownModifier) PlanModifyList(_ context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	if req.State.Raw.IsNull() || !req.PlanValue.IsUnknown() || req.ConfigValue.IsUnknown() {
		return
	}

	resp.PlanValue = req.StateValue
}

func (m useStateForUnknownModifier) PlanModifyMap(_ context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
	if req.State.Raw.IsNull() || !req.PlanValue.IsUnknown() || req.ConfigValue.IsUnknown() {
		return
	}

	resp.PlanValue = req.StateValue
}

func (m useStateForUnknownModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.State.Raw.IsNull() || !req.PlanValue.IsUnknown() || req.ConfigValue.IsUnknown() {
		return
	}

	resp.PlanValue = req.StateValue
}
