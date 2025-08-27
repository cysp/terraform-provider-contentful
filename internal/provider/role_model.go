package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RoleModel struct {
	ID          types.String                            `tfsdk:"id"`
	SpaceID     types.String                            `tfsdk:"space_id"`
	RoleID      types.String                            `tfsdk:"role_id"`
	Name        types.String                            `tfsdk:"name"`
	Description types.String                            `tfsdk:"description"`
	Permissions TypedMap[TypedList[types.String]]       `tfsdk:"permissions"`
	Policies    TypedList[TypedObject[RolePolicyValue]] `tfsdk:"policies"`
}

type RolePolicyValue struct {
	Actions    TypedList[types.String] `tfsdk:"actions"`
	Constraint jsontypes.Normalized    `tfsdk:"constraint"`
	Effect     types.String            `tfsdk:"effect"`
}
