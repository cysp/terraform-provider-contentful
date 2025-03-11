package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RoleResourceModel struct {
	ID          types.String `tfsdk:"id"`
	SpaceID     types.String `tfsdk:"space_id"`
	RoleID      types.String `tfsdk:"role_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Permissions types.Map    `tfsdk:"permissions"`
	Policies    types.List   `tfsdk:"policies"`
}

func RoleResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"space_id": schema.StringAttribute{
				Required: true,
			},
			"role_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
			},
			"permissions": schema.MapAttribute{
				ElementType: types.ListType{
					ElemType: types.StringType,
				},
				Required: true,
			},
			"policies": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: RolePolicyValue{}.SchemaAttributes(ctx),
					CustomType: RolePolicyValue{}.CustomType(ctx),
				},
				Required: true,
			},
		},
	}
}
