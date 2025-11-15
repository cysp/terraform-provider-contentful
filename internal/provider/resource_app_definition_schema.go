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
		Description: "Manages a Contentful App Definition.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				Description: "ID of the organization that owns the app definition.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"app_definition_id": schema.StringAttribute{
				Description: "System ID of the app definition.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "A human-readable name of the app.",
				Required:    true,
			},
			"src": schema.StringAttribute{
				Description: "Publicly available source URL of the app. Requires HTTPS with exception of localhost (for development).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
			"bundle_id": schema.StringAttribute{
				Description: "Link to an AppBundle if hosted on Contentful.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
			"locations": schema.ListNestedAttribute{
				Description: "List of places in the web app where the app can be rendered.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"location": schema.StringAttribute{
							Description: "Location identifier (e.g., entry-field, entry-sidebar, app-config).",
							Required:    true,
						},
						"field_types": schema.ListNestedAttribute{
							Description: "Field types where an extension can be used.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										Description: "Field type (e.g., Symbol, Text, Integer, Link, Array).",
										Required:    true,
									},
									"link_type": schema.StringAttribute{
										Description: "Type of linked resource (Entry or Asset).",
										Optional:    true,
									},
									"items": schema.SingleNestedAttribute{
										Description: "Item type definition for Array fields.",
										Attributes: map[string]schema.Attribute{
											"type": schema.StringAttribute{
												Description: "Type of array items.",
												Required:    true,
											},
											"link_type": schema.StringAttribute{
												Description: "Link type for array items (Entry or Asset).",
												Optional:    true,
											},
										},
										Optional: true,
									},
								},
							},
							Optional: true,
						},
						"navigation_item": schema.SingleNestedAttribute{
							Description: "Navigation item configuration for page locations.",
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Description: "Display name for the navigation item.",
									Required:    true,
								},
								"path": schema.StringAttribute{
									Description: "URL path for the navigation item.",
									Required:    true,
								},
							},
							Optional: true,
						},
					},
				},
				Required: true,
			},
			"parameters": schema.SingleNestedAttribute{
				Description: "Definitions of configuration parameters.",
				Attributes: map[string]schema.Attribute{
					"installation": schema.ListNestedAttribute{
						Description: "Installation-level parameter definitions.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: AppDefinitionParameterSchemaAttributes(ctx),
						},
						Optional: true,
					},
					"instance": schema.ListNestedAttribute{
						Description: "Instance-level parameter definitions.",
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

//nolint:dupl
func AppDefinitionParameterSchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Unique identifier for the parameter.",
			Required:    true,
		},
		"type": schema.StringAttribute{
			Description: "Parameter type (e.g., Symbol, Enum, Boolean).",
			Required:    true,
		},
		"name": schema.StringAttribute{
			Description: "Display name for the parameter.",
			Required:    true,
		},
		"description": schema.StringAttribute{
			Description: "Help text describing the parameter.",
			Optional:    true,
		},
		"required": schema.BoolAttribute{
			Description: "Whether the parameter is required.",
			Optional:    true,
		},
		"default": schema.StringAttribute{
			Description: "Default value for the parameter.",
			CustomType:  jsontypes.NormalizedType{},
			Optional:    true,
		},
		"options": schema.ListAttribute{
			Description: "List of allowed values for Enum parameters.",
			ElementType: jsontypes.NormalizedType{},
			CustomType:  NewTypedListNull[jsontypes.Normalized]().CustomType(ctx),
			Optional:    true,
		},
		"labels": schema.SingleNestedAttribute{
			Description: "Custom labels for Boolean parameter values.",
			Attributes: map[string]schema.Attribute{
				"empty": schema.StringAttribute{
					Description: "Label when no value is set.",
					Optional:    true,
				},
				"true": schema.StringAttribute{
					Description: "Label for true value.",
					Optional:    true,
				},
				"false": schema.StringAttribute{
					Description: "Label for false value.",
					Optional:    true,
				},
			},
			Optional: true,
		},
	}
}
