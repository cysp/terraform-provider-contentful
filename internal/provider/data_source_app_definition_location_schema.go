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
						Description: "Type of linked resource for `Link` or `ResourceLink` fields, or null when absent.",
						Computed:    true,
					},
					"items": schema.SingleNestedAttribute{
						Description: "Array item definition, or null when absent. Check for null before accessing `items.type` or `items.link_type`.",
						Attributes: map[string]schema.Attribute{
							"type": schema.StringAttribute{
								Description: "The type of array items.",
								Computed:    true,
							},
							"link_type": schema.StringAttribute{
								Description: "Type of linked resource for `Link` or `ResourceLink` array items, or null when absent.",
								Computed:    true,
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
