package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RoleDataSourceItemModel struct {
	RoleID      types.String                            `tfsdk:"role_id"`
	Name        types.String                            `tfsdk:"name"`
	Description types.String                            `tfsdk:"description"`
	Permissions TypedMap[TypedList[types.String]]       `tfsdk:"permissions"`
	Policies    TypedList[TypedObject[RolePolicyValue]] `tfsdk:"policies"`
}

type RoleDataSourceModel struct {
	IDIdentityModel
	RoleDataSourceItemModel

	SpaceID  types.String   `tfsdk:"space_id"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

type RolesDataSourceModel struct {
	IDIdentityModel

	SpaceID  types.String              `tfsdk:"space_id"`
	Roles    []RoleDataSourceItemModel `tfsdk:"roles"`
	Timeouts timeouts.Value            `tfsdk:"timeouts"`
}
