package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RoleResourceModel struct {
	ID          types.String              `tfsdk:"id"`
	SpaceID     types.String              `tfsdk:"space_id"`
	RoleID      types.String              `tfsdk:"role_id"`
	Name        types.String              `tfsdk:"name"`
	Description types.String              `tfsdk:"description"`
	Permissions map[string][]types.String `tfsdk:"permissions"`
	Policies    []RolePolicyValue         `tfsdk:"policies"`
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
				ElementType: NewTypedListNull[types.String](ctx).Type(ctx),
				CustomType:  NewTypedMapNull[TypedList[types.String]](ctx).CustomType(ctx),
				Required:    true,
			},
			"policies": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: RolePolicyValue{}.SchemaAttributes(ctx),
				},
				Required: true,
			},
		},
	}
}
