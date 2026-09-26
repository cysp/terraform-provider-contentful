package provider

import "github.com/hashicorp/terraform-plugin-framework/datasource/schema"

func appDefinitionDataSourceLocationAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
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
						Description: "For Array fields, a one-element list when Contentful supplies an item definition; null otherwise. Check for null before accessing `items[0].type` or `items[0].link_type`.",
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
	}
}
