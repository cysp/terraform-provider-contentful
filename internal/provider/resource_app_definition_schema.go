package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func AppDefinitionResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"app_definition_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"src": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
			"bundle_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
			"locations": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"location": schema.StringAttribute{
							Required: true,
						},
						"field_types": schema.ListNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										Required: true,
									},
									"link_type": schema.StringAttribute{
										Optional: true,
									},
									"items": schema.ListNestedAttribute{
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"type": schema.StringAttribute{
													Required: true,
												},
												"link_type": schema.StringAttribute{
													Optional: true,
												},
											},
										},
										Optional: true,
									},
								},
							},
							Optional: true,
						},
						"navigation_item": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Required: true,
								},
								"path": schema.StringAttribute{
									Required: true,
								},
							},
							Optional: true,
						},
					},
				},
				Required: true,
			},
			"parameters": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"installation": schema.ListNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: AppDefinitionParameterSchemaAttributes(ctx),
						},
						Optional: true,
					},
					"instance": schema.ListNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: AppDefinitionParameterSchemaAttributes(ctx),
						},
						Optional: true,
					},
				},
				Optional: true,
			},
		},
	}
}

func AppDefinitionParameterSchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Required: true,
		},
		"type": schema.StringAttribute{
			Required: true,
		},
		"name": schema.StringAttribute{
			Required: true,
		},
		"description": schema.StringAttribute{
			Optional: true,
		},
		"required": schema.BoolAttribute{
			Optional: true,
		},
		"default": schema.StringAttribute{
			CustomType: jsontypes.NormalizedType{},
			Optional:   true,
		},
		"options": schema.ListAttribute{
			ElementType: jsontypes.NormalizedType{},
			CustomType:  NewTypedListNull[jsontypes.Normalized](ctx).CustomType(ctx),
			Optional:    true,
		},
		"labels": schema.SingleNestedAttribute{
			Attributes: map[string]schema.Attribute{
				"empty": schema.StringAttribute{
					Optional: true,
				},
				"true": schema.StringAttribute{
					Optional: true,
				},
				"false": schema.StringAttribute{
					Optional: true,
				},
			},
			Optional: true,
		},
	}
}
