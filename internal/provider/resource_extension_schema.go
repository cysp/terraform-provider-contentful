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
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"environment_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"extension_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
					UseStateForUnknown(),
				},
			},
			"extension": schema.SingleNestedAttribute{
				Attributes: ExtensionResourceExtensionSchemaAttributes(ctx),
				Required:   true,
			},
			"parameters": schema.StringAttribute{
				CustomType: jsontypes.NormalizedType{},
				Optional:   true,
				Computed:   true,
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
			Required: true,
		},
		"src": schema.StringAttribute{
			Optional: true,
			Computed: true,
			Default:  stringdefault.StaticString(""),
			PlanModifiers: []planmodifier.String{
				UseStateForUnknown(),
			},
		},
		"srcdoc": schema.StringAttribute{
			Optional: true,
			Computed: true,
			Default:  stringdefault.StaticString(""),
			PlanModifiers: []planmodifier.String{
				UseStateForUnknown(),
			},
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
					"items": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"type": schema.StringAttribute{
								Required: true,
							},
							"link_type": schema.StringAttribute{
								Optional: true,
							},
						},
						Required: true,
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
