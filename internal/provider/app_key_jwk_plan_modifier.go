package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const appKeyJWKConfiguredPrivateStateKey = "app_key_jwk_configured"

func appKeyJWKPlanModifierFor(ctx context.Context) appKeyJWKPlanModifier {
	return appKeyJWKPlanModifier{
		attributeTypes: NewTypedObjectNull[AppKeyJWKModel]().CustomType(ctx).AttributeTypes(),
	}
}

type appKeyJWKPlanModifier struct {
	attributeTypes map[string]attr.Type
}

func (m appKeyJWKPlanModifier) Description(_ context.Context) string {
	return "Preserves generated JWK state and replaces the resource when a configured JWK is changed or removed."
}

func (m appKeyJWKPlanModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m appKeyJWKPlanModifier) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	// Do nothing for initial create, destroy, or unknown configuration.
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	if req.ConfigValue.IsNull() {
		if appKeyJWKWasConfigured(ctx, req.Private) {
			resp.PlanValue = types.ObjectNull(m.attributeTypes)
			resp.RequiresReplace = true
			resp.Diagnostics.Append(resp.Private.SetKey(ctx, appKeyJWKConfiguredPrivateStateKey, []byte("false"))...)

			return
		}

		if req.PlanValue.IsUnknown() {
			resp.PlanValue = req.StateValue
		}

		return
	}

	resp.Diagnostics.Append(resp.Private.SetKey(ctx, appKeyJWKConfiguredPrivateStateKey, []byte("true"))...)

	if !req.PlanValue.Equal(req.StateValue) {
		resp.RequiresReplace = true
	}
}

func appKeyJWKWasConfigured(ctx context.Context, providerData PrivateProviderData) bool {
	var diags diag.Diagnostics

	value, getDiags := providerData.GetKey(ctx, appKeyJWKConfiguredPrivateStateKey)
	diags.Append(getDiags...)

	if diags.HasError() {
		return false
	}

	return string(value) == "true"
}
