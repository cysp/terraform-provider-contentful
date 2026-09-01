package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func ExtensionResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Manages a Contentful UI Extension.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"extension": schema.SingleNestedAttribute{
				Description: "Extension configuration.",
				Attributes:  ExtensionResourceExtensionSchemaAttributes(ctx),
				Required:    true,
			},
			"parameters": schema.StringAttribute{
				Description: "Definitions of configuration parameters. Use a sensitive Terraform expression when this mixed-use value contains secrets. Sensitivity obscures CLI output; it does not encrypt or omit plan or state data.",
				CustomType:  jsontypes.NormalizedType{},
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"timeouts": timeouts.AttributesAll(ctx),
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
			Description: "URL where the root HTML document of the extension can be found. Must be non-empty and HTTPS, except that Contentful also accepts localhost HTTP URLs. Cannot be configured with srcdoc; both may be omitted to preserve an imported source.",
			Optional:    true,
			Computed:    true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
				stringvalidator.ConflictsWith(
					path.MatchRelative().AtParent().AtName("srcdoc"),
				),
			},
		},
		"srcdoc": schema.StringAttribute{
			Description: "String representation of the extension (e.g. inline HTML code). Cannot be configured with src; both may be omitted to preserve an imported source. Contentful accepts an explicitly empty srcdoc.",
			Optional:    true,
			Computed:    true,
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
			Validators: []validator.List{
				listvalidator.NoNullValues(),
			},
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
					Validators: []validator.List{
						listvalidator.NoNullValues(),
					},
				},
				"instance": schema.ListNestedAttribute{
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
	}
}
