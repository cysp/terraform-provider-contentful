package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func RoleResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Manages a Contentful Role.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				Description: "ID of the space where the role exists.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role_id": schema.StringAttribute{
				Description: "System ID of the role.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the role.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the role.",
				Optional:    true,
			},
			"permissions": schema.MapAttribute{
				Description: "Basic rules which define whether a user can read or create content types, settings and entries.",
				ElementType: NewTypedListNull[types.String]().Type(ctx),
				CustomType:  NewTypedMapNull[TypedList[types.String]]().CustomType(ctx),
				Required:    true,
			},
			"policies": schema.ListNestedAttribute{
				Description: "Policies allow or deny access to resources in fine-grained detail. For example, limit read access to only entries of a specific content type or write access to only certain parts of an entry (e.g. a specific locale).",
				NestedObject: schema.NestedAttributeObject{
					Attributes: RolePolicyValue{}.SchemaAttributes(ctx),
					CustomType: NewTypedObjectUnknown[RolePolicyValue]().CustomType(ctx),
				},
				CustomType: TypedList[TypedObject[RolePolicyValue]]{}.CustomType(ctx),
				Required:   true,
			},
		},
	}
}

func (v RolePolicyValue) SchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"actions": schema.ListAttribute{
			Description: "Actions that the policy allows or denies (e.g., read, create, update, delete, publish).",
			ElementType: types.StringType,
			CustomType:  TypedList[types.String]{}.CustomType(ctx),
			Required:    true,
		},
		"constraint": schema.StringAttribute{
			Description: "JSON constraint that defines the scope of the policy.",
			CustomType:  jsontypes.NormalizedType{},
			Optional:    true,
		},
		"effect": schema.StringAttribute{
			Description: "Whether the policy allows or denies the specified actions.",
			Required:    true,
		},
	}
}
