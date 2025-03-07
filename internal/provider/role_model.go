package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

type RoleModel struct {
	Description types.String `tfsdk:"description"`
	Name        types.String `tfsdk:"name"`
	Permissions types.Map    `tfsdk:"permissions"`
	Policies    types.List   `tfsdk:"policies"`
	RoleID      types.String `tfsdk:"role_id"`
	SpaceID     types.String `tfsdk:"space_id"`
}
