package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ExtensionResourceModel struct {
	ID            types.String                    `tfsdk:"id"`
	SpaceID       types.String                    `tfsdk:"space_id"`
	EnvironmentID types.String                    `tfsdk:"environment_id"`
	ExtensionID   types.String                    `tfsdk:"extension_id"`
	Extension     ExtensionResourceModelExtension `tfsdk:"extension"`
	Parameters    jsontypes.Normalized            `tfsdk:"parameters"`
}

type ExtensionResourceModelExtension struct {
	Src        types.String                                        `tfsdk:"src"`
	SrcDoc     types.String                                        `tfsdk:"srcdoc"`
	FieldTypes TypedList[ExtensionResourceModelExtensionFieldType] `tfsdk:"field_types"`
	Sidebar    types.Bool                                          `tfsdk:"sidebar"`
	Parameters jsontypes.Normalized                                `tfsdk:"parameters"`
}

type ExtensionResourceModelExtensionFieldType struct {
	Type     types.String                                  `tfsdk:"type"`
	LinkType types.String                                  `tfsdk:"link_type"`
	Items    ExtensionResourceModelExtensionFieldTypeItems `tfsdk:"items"`
}

type ExtensionResourceModelExtensionFieldTypeItems struct {
	Type     types.String `tfsdk:"type"`
	LinkType types.String `tfsdk:"link_type"`
}

func ExtensionResourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
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
					stringplanmodifier.RequiresReplace(),
				},
			},
			"extension": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"src": schema.StringAttribute{
						Optional: true,
					},
					"srcdoc": schema.StringAttribute{
						Optional: true,
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
					},
					"parameters": schema.StringAttribute{
						CustomType: jsontypes.NormalizedType{},
						Optional:   true,
					},
				},
				Required: true,
			},
			"parameters": schema.StringAttribute{
				CustomType: jsontypes.NormalizedType{},
				Optional:   true,
			},
		},
	}
}
