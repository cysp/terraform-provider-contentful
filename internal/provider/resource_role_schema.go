package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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
					UseStateForUnknown(),
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
					CustomType: RolePolicyValue{}.CustomType(ctx),
				},
				CustomType: TypedList[RolePolicyValue]{}.CustomType(ctx),
				Required:   true,
			},
		},
	}
}
