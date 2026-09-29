package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func roleDataSourceItemAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"role_id":     schema.StringAttribute{Description: "System ID of the role.", Computed: true},
		"name":        schema.StringAttribute{Description: "Name of the role.", Computed: true},
		"description": schema.StringAttribute{Description: "Description of the role, or null when absent.", Computed: true},
		"permissions": schema.MapAttribute{
			Description: "Permission names mapped to value lists. Contentful's scalar `\"all\"` appears as `[\"all\"]`; empty lists and duplicate values are preserved.",
			ElementType: NewTypedListNull[types.String]().Type(ctx),
			CustomType:  NewTypedMapNull[TypedList[types.String]]().CustomType(ctx),
			Computed:    true,
		},
		"policies": schema.ListNestedAttribute{
			Description: "Policies that allow or deny actions on selected resources, in the order returned by Contentful.",
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"actions": schema.ListAttribute{
						Description: "Actions allowed or denied by this policy. Scalar `\"all\"` is represented as `[\"all\"]`; action order and duplicates are preserved.",
						ElementType: types.StringType,
						CustomType:  TypedList[types.String]{}.CustomType(ctx),
						Computed:    true,
					},
					"constraint": schema.StringAttribute{Description: "Normalized JSON constraint, or null when absent.", CustomType: jsontypes.NormalizedType{}, Computed: true},
					"effect":     schema.StringAttribute{Description: "Policy effect: `allow` or `deny` for the specified actions.", Computed: true},
				},
				CustomType: NewTypedObjectUnknown[RolePolicyValue]().CustomType(ctx),
			},
			CustomType: TypedList[TypedObject[RolePolicyValue]]{}.CustomType(ctx),
			Computed:   true,
		},
	}
}

func RoleDataSourceSchema(ctx context.Context) schema.Schema {
	attributes := roleDataSourceItemAttributes(ctx)
	attributes["space_id"] = schema.StringAttribute{Description: "ID of the space containing the role.", Required: true, Validators: []validator.String{discoveryIDValidator{}}}
	attributes["role_id"] = schema.StringAttribute{Description: "System ID of the role to look up.", Required: true, Validators: []validator.String{discoveryIDValidator{}}}
	attributes["id"] = schema.StringAttribute{Description: "Composite Terraform identifier in `space_id/role_id` form.", Computed: true}
	attributes["timeouts"] = timeouts.Attributes(ctx)

	return schema.Schema{Description: "Retrieves an existing Contentful Role in a space.", Attributes: attributes}
}

func RolesDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Retrieves all Contentful Roles in a space. The read timeout covers all pages, and a failed page causes the lookup to fail.",
		Attributes: map[string]schema.Attribute{
			"space_id": schema.StringAttribute{Description: "ID of the space containing the roles.", Required: true, Validators: []validator.String{discoveryIDValidator{}}},
			"id":       schema.StringAttribute{Description: "Terraform identifier equal to `space_id`.", Computed: true},
			"roles":    schema.ListNestedAttribute{Description: "Roles in the order returned by Contentful. An empty collection returns an empty list.", Computed: true, NestedObject: schema.NestedAttributeObject{Attributes: roleDataSourceItemAttributes(ctx)}},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}
