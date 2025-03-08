package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ContentTypeResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"space_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"environment_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"content_type_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Required: true,
			},
			"display_field": schema.StringAttribute{
				Required: true,
			},
			"fields": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"default_value": schema.StringAttribute{
							CustomType: jsontypes.NormalizedType{},
							Optional:   true,
							Computed:   true,
							Default:    stringdefault.StaticString(""),
						},
						"disabled": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  booldefault.StaticBool(false),
						},
						"id": schema.StringAttribute{
							Required: true,
						},
						"items": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"link_type": schema.StringAttribute{
									Optional: true,
								},
								"type": schema.StringAttribute{
									Required: true,
								},
								"validations": schema.ListAttribute{
									ElementType: jsontypes.NormalizedType{},
									Optional:    true,
									Computed:    true,
									Default:     listdefault.StaticValue(NewEmptyListMust(jsontypes.NormalizedType{})),
								},
							},
							CustomType: ItemsType{
								ObjectType: types.ObjectType{
									AttrTypes: ItemsValue{}.AttributeTypes(ctx),
								},
							},
							Optional: true,
						},
						"link_type": schema.StringAttribute{
							Optional: true,
						},
						"localized": schema.BoolAttribute{
							Required: true,
						},
						"name": schema.StringAttribute{
							Required: true,
						},
						"omitted": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  booldefault.StaticBool(false),
						},
						"required": schema.BoolAttribute{
							Required: true,
						},
						"type": schema.StringAttribute{
							Required: true,
						},
						"validations": schema.ListAttribute{
							ElementType: jsontypes.NormalizedType{},
							Optional:    true,
							Computed:    true,
							Default:     listdefault.StaticValue(NewEmptyListMust(jsontypes.NormalizedType{})),
						},
					},
					CustomType: FieldsType{
						ObjectType: types.ObjectType{
							AttrTypes: FieldsValue{}.AttributeTypes(ctx),
						},
					},
				},
				Required: true,
			},
		},
	}
}
