package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type TagIdentityModel struct {
	SpaceID       types.String `tfsdk:"space_id"`
	EnvironmentID types.String `tfsdk:"environment_id"`
	TagID         types.String `tfsdk:"tag_id"`
}

type TagModel struct {
	IDIdentityModel
	TagIdentityModel

	Name       types.String `tfsdk:"name"`
	Visibility types.String `tfsdk:"visibility"`
}
