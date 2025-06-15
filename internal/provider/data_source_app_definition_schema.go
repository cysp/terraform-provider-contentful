package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func AppDefinitionDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"organization_id": schema.StringAttribute{
				Required: true,
			},
			"app_definition_id": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"src": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"bundle_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"locations": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"location": schema.StringAttribute{
							Computed: true,
						},
						"field_types": schema.ListNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										Computed: true,
									},
									"link_type": schema.StringAttribute{
										Optional: true,
										Computed: true,
									},
									"items": schema.ListNestedAttribute{
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"type": schema.StringAttribute{
													Computed: true,
												},
												"link_type": schema.StringAttribute{
													Optional: true,
													Computed: true,
												},
											},
										},
										Optional: true,
										Computed: true,
									},
								},
							},
							Optional: true,
							Computed: true,
						},
						"navigation_item": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Computed: true,
								},
								"path": schema.StringAttribute{
									Computed: true,
								},
							},
							Optional: true,
							Computed: true,
						},
					},
				},
				Computed: true,
			},
			"parameters": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"installation": schema.ListNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: AppDefinitionParameterDataSourceSchemaAttributes(ctx),
						},
						Optional: true,
						Computed: true,
					},
					"instance": schema.ListNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: AppDefinitionParameterDataSourceSchemaAttributes(ctx),
						},
						Optional: true,
						Computed: true,
					},
				},
				Optional: true,
				Computed: true,
			},
		},
	}
}

func AppDefinitionParameterDataSourceSchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed: true,
		},
		"type": schema.StringAttribute{
			Computed: true,
		},
		"name": schema.StringAttribute{
			Computed: true,
		},
		"description": schema.StringAttribute{
			Optional: true,
			Computed: true,
		},
		"required": schema.BoolAttribute{
			Optional: true,
			Computed: true,
		},
		"default": schema.StringAttribute{
			CustomType: jsontypes.NormalizedType{},
			Optional:   true,
			Computed:   true,
		},
		"options": schema.ListAttribute{
			ElementType: jsontypes.NormalizedType{},
			CustomType:  NewTypedListNull[jsontypes.Normalized](ctx).CustomType(ctx),
			Optional:    true,
			Computed:    true,
		},
		"labels": schema.SingleNestedAttribute{
			Attributes: map[string]schema.Attribute{
				"empty": schema.StringAttribute{
					Optional: true,
					Computed: true,
				},
				"true": schema.StringAttribute{
					Optional: true,
					Computed: true,
				},
				"false": schema.StringAttribute{
					Optional: true,
					Computed: true,
				},
			},
			Optional: true,
			Computed: true,
		},
	}
}
