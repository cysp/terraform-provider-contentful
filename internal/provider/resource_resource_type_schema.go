package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func ResourceTypeResourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Description: "Manages a Contentful App Resource Type.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				Description: "ID of the organization.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"app_definition_id": schema.StringAttribute{
				Description: "ID of the app definition.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"resource_provider_id": schema.StringAttribute{
				Description: "ID of the parent resource provider.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
			"resource_type_id": schema.StringAttribute{
				Description: "ID of the resource type.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the resource type.",
				Required:    true,
			},
			"default_field_mapping": schema.SingleNestedAttribute{
				Description: "Default field mapping configuration for the resource type.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"title": schema.StringAttribute{
						Description: "Field path for the title.",
						Required:    true,
					},
					"subtitle": schema.StringAttribute{
						Description: "Field path for the subtitle.",
						Optional:    true,
					},
					"description": schema.StringAttribute{
						Description: "Field path for the description.",
						Optional:    true,
					},
					"external_url": schema.StringAttribute{
						Description: "Field path for the external URL.",
						Optional:    true,
					},
					"image": schema.SingleNestedAttribute{
						Description: "Image field mapping.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"url": schema.StringAttribute{
								Description: "Field path for the image URL.",
								Required:    true,
							},
							"alt_text": schema.StringAttribute{
								Description: "Field path for the image alt text.",
								Optional:    true,
							},
						},
					},
					"badge": schema.SingleNestedAttribute{
						Description: "Badge field mapping.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"label": schema.StringAttribute{
								Description: "Field path for the badge label.",
								Required:    true,
							},
							"variant": schema.StringAttribute{
								Description: "Field path for the badge variant.",
								Required:    true,
							},
						},
					},
				},
			},
		},
	}
}
