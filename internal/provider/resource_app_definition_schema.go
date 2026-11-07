package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func AppDefinitionResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Manages a Contentful App Definition: the app's name, hosting source, locations in the web app, and parameter definitions. Use `contentful_app_installation` to install it in an environment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite Terraform resource identifier in `organization_id/app_definition_id` form.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				Description: "ID of the organization that owns the app definition. Changing this value replaces the resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"app_definition_id": schema.StringAttribute{
				Description: "System ID of the app definition.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "A human-readable name of the app.",
				Required:    true,
			},
			"src": schema.StringAttribute{
				Description: "Nonempty frontend URL using HTTPS, or HTTP on localhost. Omission retains the existing URL during in-place updates. Configure explicitly to retain it during replacement.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"bundle_id": schema.StringAttribute{
				Description: "ID of the App Bundle.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"locations": schema.ListNestedAttribute{
				Description: "Locations where the app can be rendered in the Contentful web app. Use `[]` for no locations. Rendering locations require a frontend hosting source. When the provider sends a `src` URL, including one retained from state, Contentful requires at least one location other than `dialog`.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"location": schema.StringAttribute{
							Description: "Location identifier (e.g., entry-field, entry-sidebar, app-config).",
							Required:    true,
						},
						"field_types": schema.ListNestedAttribute{
							Description: "Field definitions for `entry-field` locations. At least one is required; omit for other locations.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										Description: "Field type (e.g., Symbol, Text, Integer, Link, Array).",
										Required:    true,
									},
									"link_type": schema.StringAttribute{
										Description: "Type of linked resource. Required for `Link` and `ResourceLink` fields; omit for other field types. For `Link`, use `Entry` or `Asset`. For linked array items, configure `items.link_type`. See [Field types](#field-types) for the ResourceLink support limitation.",
										Optional:    true,
									},
									"items": schema.SingleNestedAttribute{
										Description: "Required item type definition for `Array` fields. Omit for other field types.",
										Attributes: map[string]schema.Attribute{
											"type": schema.StringAttribute{
												Description: "Array item type, such as `Symbol` or `Link`.",
												Required:    true,
											},
											"link_type": schema.StringAttribute{
												Description: "Type of linked resource. Required for `Link` and `ResourceLink` items; omit for other item types. For `Link`, use `Entry` or `Asset`. See [Field types](#field-types) for the ResourceLink support limitation.",
												Optional:    true,
											},
										},
										Optional: true,
									},
								},
							},
							Optional: true,
							Validators: []validator.List{
								listvalidator.NoNullValues(),
							},
						},
						"navigation_item": schema.SingleNestedAttribute{
							Description: "Navigation item for `page` locations; omit for other locations. Omission removes existing navigation configuration.",
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Description: "Nonempty display name for the navigation item.",
									Required:    true,
								},
								"path": schema.StringAttribute{
									Description: "Nonempty navigation path beginning with `/`, for example `/dashboard`.",
									Required:    true,
								},
							},
							Optional: true,
						},
					},
				},
				Required: true,
				Validators: []validator.List{
					listvalidator.NoNullValues(),
					appDefinitionLocationsValidator{},
				},
			},
			"parameters": schema.SingleNestedAttribute{
				Description: "App configuration parameter definitions.",
				Attributes: map[string]schema.Attribute{
					"installation": schema.ListNestedAttribute{
						Description: "Parameter definitions for each app installation.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: AppDefinitionParameterSchemaAttributes(ctx),
						},
						Optional: true,
						Validators: []validator.List{
							listvalidator.NoNullValues(),
						},
					},
					"instance": schema.ListNestedAttribute{
						Description: "Parameter definitions for each use of the app, such as a field editor. Maximum eight definitions.",
						NestedObject: schema.NestedAttributeObject{
							Attributes: AppDefinitionParameterSchemaAttributes(ctx),
						},
						Optional: true,
						Validators: []validator.List{
							listvalidator.NoNullValues(),
						},
					},
				},
				Optional: true,
			},
			"timeouts": timeouts.AttributesAll(ctx),
		},
	}
}

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
			Description: "Whether a parameter value must be provided.",
			Optional:    true,
		},
		"default": schema.StringAttribute{
			Description: "Default value encoded as JSON.",
			CustomType:  jsontypes.NormalizedType{},
			Optional:    true,
		},
		"options": schema.ListAttribute{
			Description: "Options, each encoded as JSON.",
			ElementType: jsontypes.NormalizedType{},
			CustomType:  NewTypedListNull[jsontypes.Normalized]().CustomType(ctx),
			Optional:    true,
			Validators: []validator.List{
				listvalidator.NoNullValues(),
			},
		},
		"labels": schema.SingleNestedAttribute{
			Description: "Display labels for Boolean values and the empty Enum selection.",
			Attributes: map[string]schema.Attribute{
				"empty": schema.StringAttribute{
					Description: "Label displayed when no Enum option is selected.",
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
