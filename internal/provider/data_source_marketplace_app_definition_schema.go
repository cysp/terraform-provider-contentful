//nolint:dupl
package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func MarketplaceAppDefinitionDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Retrieves a Contentful Marketplace App Definition.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"organization_id": schema.StringAttribute{
				Computed: true,
			},
			"app_definition_id": schema.StringAttribute{
				Description: "The unique identifier for the marketplace app definition.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the marketplace app.",
				Computed:    true,
			},
			"src": schema.StringAttribute{
				Description: "The URL where the marketplace app is hosted.",
				Computed:    true,
			},
			"bundle_id": schema.StringAttribute{
				Description: "The bundle identifier for the marketplace app.",
				Computed:    true,
			},
			"locations": schema.ListNestedAttribute{
				Description: "Locations where the marketplace app can be rendered in the Contentful web app.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"location": schema.StringAttribute{
							Description: "The location where the app can be rendered.",
							Computed:    true,
						},
						"field_types": schema.ListNestedAttribute{
							Description: "Field types that this location supports.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										Description: "The field type.",
										Computed:    true,
									},
									"link_type": schema.StringAttribute{
										Description: "For Link fields, the type of linked resource.",
										Computed:    true,
									},
									"items": schema.ListNestedAttribute{
										Description: "For Array fields, the type of items in the array.",
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"type": schema.StringAttribute{
													Description: "The type of array items.",
													Computed:    true,
												},
												"link_type": schema.StringAttribute{
													Description: "For arrays of Links, the type of linked resource.",
													Computed:    true,
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
							Description: "Navigation item configuration for this location.",
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Description: "Display name for the navigation item.",
									Computed:    true,
								},
								"path": schema.StringAttribute{
									Description: "Path for the navigation item.",
									Computed:    true,
								},
							},
							Computed: true,
						},
					},
				},
				Computed: true,
			},
			"parameters": schema.SingleNestedAttribute{
				Description: "Configuration parameters for the marketplace app.",
				Attributes: map[string]schema.Attribute{
					"installation": schema.ListNestedAttribute{
						Description: "Installation-level parameters for the app.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: MarketplaceAppDefinitionParameterDataSourceSchemaAttributes(ctx),
						},
						Computed: true,
					},
					"instance": schema.ListNestedAttribute{
						Description: "Instance-level parameters for the app.",
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
			Description: "The type of this parameter.",
			Computed:    true,
		},
		"name": schema.StringAttribute{
			Description: "The name of this parameter.",
			Computed:    true,
		},
		"description": schema.StringAttribute{
			Description: "Description of this parameter.",
			Computed:    true,
		},
		"required": schema.BoolAttribute{
			Description: "Whether this parameter is required.",
			Computed:    true,
		},
		"default": schema.StringAttribute{
			Description: "Default value for this parameter in JSON format.",
			CustomType:  jsontypes.NormalizedType{},
			Computed:    true,
		},
		"options": schema.ListAttribute{
			Description: "Available options for this parameter.",
			ElementType: jsontypes.NormalizedType{},
			CustomType:  NewTypedListNull[jsontypes.Normalized]().CustomType(ctx),
			Computed:    true,
		},
		"labels": schema.SingleNestedAttribute{
			Description: "Labels for boolean parameter values.",
			Attributes: map[string]schema.Attribute{
				"empty": schema.StringAttribute{
					Description: "Label for empty value.",
					Computed:    true,
				},
				"true": schema.StringAttribute{
					Description: "Label for true value.",
					Computed:    true,
				},
				"false": schema.StringAttribute{
					Description: "Label for false value.",
					Computed:    true,
				},
			},
			Computed: true,
		},
	}
}
