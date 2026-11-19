package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AppInstallationIdentityModel struct {
	SpaceID         types.String `tfsdk:"space_id"`
	EnvironmentID   types.String `tfsdk:"environment_id"`
	AppDefinitionID types.String `tfsdk:"app_definition_id"`
}

type AppInstallationModel struct {
	IDIdentityModel
	AppInstallationIdentityModel

	Marketplace types.Set `tfsdk:"marketplace"`
	Parameters  types.Map `tfsdk:"parameters"`

	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

func emptyAppInstallationMarketplace() types.Set {
	return types.SetValueMust(types.StringType, nil)
}

func (model *AppInstallationModel) defaultMarketplace() {
	// Imports and state written before the schema default can have a null set.
	if model.Marketplace.IsNull() {
		model.Marketplace = emptyAppInstallationMarketplace()
	}
}

func (model *AppInstallationModel) readState(ctx context.Context, state tfsdk.State) diag.Diagnostics {
	diags := state.Get(ctx, model)
	model.defaultMarketplace()

	return diags
}
