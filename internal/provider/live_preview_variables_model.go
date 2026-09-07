package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type LivePreviewVariablesIdentityModel struct {
	SpaceID       types.String `tfsdk:"space_id"`
	EnvironmentID types.String `tfsdk:"environment_id"`
}

type LivePreviewVariablesModel struct {
	IDIdentityModel
	LivePreviewVariablesIdentityModel

	Variables jsontypes.Normalized `tfsdk:"variables"`
	Timeouts  timeouts.Value       `tfsdk:"timeouts"`
}
