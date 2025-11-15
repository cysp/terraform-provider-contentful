package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func ExtensionResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Manages a Contentful UI Extension.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				Description: "ID of the space containing the extension.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"environment_id": schema.StringAttribute{
				Description: "ID of the environment where the extension is installed.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"extension_id": schema.StringAttribute{
				Description: "ID of the extension.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
					UseStateForUnknown(),
				},
			},
			"extension": schema.SingleNestedAttribute{
				Description: "Extension configuration.",
				Attributes:  ExtensionResourceExtensionSchemaAttributes(ctx),
				Required:    true,
			},
			"parameters": schema.StringAttribute{
				Description: "Definitions of configuration parameters.",
				CustomType:  jsontypes.NormalizedType{},
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
		},
	}
}

func ExtensionResourceExtensionSchemaAttributes(ctx context.Context) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Description: "Extension name.",
			Required:    true,
		},
		"src": schema.StringAttribute{
			Description: "URL where the root HTML document of the extension can be found. Must be HTTPS.",
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(""),
			PlanModifiers: []planmodifier.String{
				UseStateForUnknown(),
			},
		},
		"srcdoc": schema.StringAttribute{
			Description: "String representation of the extension (e.g. inline HTML code).",
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(""),
			PlanModifiers: []planmodifier.String{
				UseStateForUnknown(),
			},
		},
		"field_types": schema.ListNestedAttribute{
			Description: "Field types where an extension can be used.",
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Description: "Field type (e.g., Symbol, Text, Integer).",
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
								Description: "Link type for array items.",
								Optional:    true,
							},
						},
						Optional: true,
					},
				},
			},
			Required: true,
		},
		"sidebar": schema.BoolAttribute{
			Optional: true,
			Computed: true,
			Default:  booldefault.StaticBool(false),
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
	}
}
