//nolint:dupl
package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func MarketplaceAppDefinitionDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"organization_id": schema.StringAttribute{
				Computed: true,
			},
			"app_definition_id": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"src": schema.StringAttribute{
				Computed: true,
			},
			"bundle_id": schema.StringAttribute{
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
										Computed: true,
									},
									"items": schema.ListNestedAttribute{
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"type": schema.StringAttribute{
													Computed: true,
												},
												"link_type": schema.StringAttribute{
													Computed: true,
												},
											},
										},
										Computed: true,
									},
								},
							},
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
							Attributes: MarketplaceAppDefinitionParameterDataSourceSchemaAttributes(ctx),
						},
						Computed: true,
					},
					"instance": schema.ListNestedAttribute{
						NestedObject: schema.NestedAttributeObject{
							Attributes: MarketplaceAppDefinitionParameterDataSourceSchemaAttributes(ctx),
						},
						Computed: true,
					},
				},
				Computed: true,
			},
		},
	}
}

//nolint:dupl
func MarketplaceAppDefinitionParameterDataSourceSchemaAttributes(ctx context.Context) map[string]schema.Attribute {
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
			Computed: true,
		},
		"required": schema.BoolAttribute{
			Computed: true,
		},
		"default": schema.StringAttribute{
			CustomType: jsontypes.NormalizedType{},
			Computed:   true,
		},
		"options": schema.ListAttribute{
			ElementType: jsontypes.NormalizedType{},
			CustomType:  NewTypedListNull[jsontypes.Normalized]().CustomType(ctx),
			Computed:    true,
		},
		"labels": schema.SingleNestedAttribute{
			Attributes: map[string]schema.Attribute{
				"empty": schema.StringAttribute{
					Computed: true,
				},
				"true": schema.StringAttribute{
					Computed: true,
				},
				"false": schema.StringAttribute{
					Computed: true,
				},
			},
			Computed: true,
		},
	}
}
