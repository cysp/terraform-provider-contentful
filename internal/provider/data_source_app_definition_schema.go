//nolint:dupl
package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func AppDefinitionDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Retrieves a Contentful App Definition.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite Terraform identifier in organization_id/app_definition_id form; not a Contentful system ID.",
				Computed:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "The ID of the organization.",
				Required:    true,
			},
			"app_definition_id": schema.StringAttribute{
				Description: "The unique identifier for the app definition.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the app.",
				Computed:    true,
			},
			"src": schema.StringAttribute{
				Description: "The URL where the app is hosted.",
				Computed:    true,
			},
			"bundle_id": schema.StringAttribute{
				Description: "The bundle identifier for the app.",
				Computed:    true,
			},
			"locations": schema.ListNestedAttribute{
				Description: "Locations where the app can be rendered in the Contentful web app.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"location": schema.StringAttribute{
							Description: "Location identifier in the Contentful web app, such as `entry-field` or `entry-sidebar`.",
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
				Description: "Parameter definitions for configuring the app.",
				Attributes: map[string]schema.Attribute{
					"installation": schema.ListNestedAttribute{
						Description: "Parameter definitions for each app installation.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: AppDefinitionParameterDataSourceSchemaAttributes(ctx),
						},
						Computed: true,
					},
					"instance": schema.ListNestedAttribute{
						Description: "Parameter definitions for each use of the app, such as a field editor.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: AppDefinitionParameterDataSourceSchemaAttributes(ctx),
						},
						Computed: true,
					},
				},
				Computed: true,
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}

//nolint:dupl
func AppDefinitionParameterDataSourceSchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Unique identifier for the parameter.",
			Computed:    true,
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
			Description: "Allowed options for an Enum parameter, each encoded as JSON. An option may be a string or an object with string values.",
			ElementType: jsontypes.NormalizedType{},
			CustomType:  NewTypedListNull[jsontypes.Normalized]().CustomType(ctx),
			Computed:    true,
		},
		"labels": schema.SingleNestedAttribute{
			Description: "Display labels for Boolean values and the empty Enum selection.",
			Attributes: map[string]schema.Attribute{
				"empty": schema.StringAttribute{
					Description: "Label displayed when no Enum option is selected.",
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
