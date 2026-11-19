package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Schema 0 stored the entire parameters document as one string.
func (r *appInstallationResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	prior := schema.Schema{Attributes: map[string]schema.Attribute{
		"id":                schema.StringAttribute{Computed: true},
		"space_id":          schema.StringAttribute{Required: true},
		"environment_id":    schema.StringAttribute{Required: true},
		"app_definition_id": schema.StringAttribute{Required: true},
		"parameters":        schema.StringAttribute{Optional: true},
		"marketplace":       schema.SetAttribute{Optional: true, ElementType: types.StringType},
		"timeouts": schema.SingleNestedAttribute{Optional: true, Attributes: map[string]schema.Attribute{
			"create": schema.StringAttribute{Optional: true},
			"read":   schema.StringAttribute{Optional: true},
			"update": schema.StringAttribute{Optional: true},
			"delete": schema.StringAttribute{Optional: true},
		}},
	}}

	return map[int64]resource.StateUpgrader{0: {PriorSchema: &prior, StateUpgrader: upgradeAppInstallationStateV0}}
}

func upgradeAppInstallationStateV0(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	var prior struct {
		ID              types.String `tfsdk:"id"`
		SpaceID         types.String `tfsdk:"space_id"`
		EnvironmentID   types.String `tfsdk:"environment_id"`
		AppDefinitionID types.String `tfsdk:"app_definition_id"`
		Marketplace     types.Set    `tfsdk:"marketplace"`
		Parameters      types.String `tfsdk:"parameters"`
		Timeouts        types.Object `tfsdk:"timeouts"`
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var raw []byte
	if !prior.Parameters.IsNull() {
		raw = []byte(prior.Parameters.ValueString())
	}

	parametersPath := path.Root("parameters")

	parameters, diags := decodeInstallationParameters(parametersPath, raw)
	if errors := diags.Errors(); len(errors) > 0 {
		resp.Diagnostics.AddAttributeError(parametersPath, "Cannot upgrade App Installation parameters", errors[0].Detail()+" If the previous provider can read this state and Contentful holds representable parameters, apply a refresh-only plan with that provider and configuration. See the App Installation resource's upgrade guide for recovery options.")

		return
	}

	state := AppInstallationModel{
		ID:              prior.ID,
		SpaceID:         prior.SpaceID,
		EnvironmentID:   prior.EnvironmentID,
		AppDefinitionID: prior.AppDefinitionID,
		Parameters:      parameters,
		Marketplace:     prior.Marketplace,
		Timeouts:        timeouts.Value{Object: prior.Timeouts},
	}
	state.defaultMarketplace()
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
