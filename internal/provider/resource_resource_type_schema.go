package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func ResourceTypeResourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"app_definition_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"resource_provider_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
			"resource_type_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"default_field_mapping": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"title": schema.StringAttribute{
						Required: true,
					},
					"subtitle": schema.StringAttribute{
						Optional: true,
					},
					"description": schema.StringAttribute{
						Optional: true,
					},
					"external_url": schema.StringAttribute{
						Optional: true,
					},
					"image": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"url": schema.StringAttribute{
								Required: true,
							},
							"alt_text": schema.StringAttribute{
								Optional: true,
							},
						},
					},
					"badge": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"label": schema.StringAttribute{
								Required: true,
							},
							"variant": schema.StringAttribute{
								Required: true,
							},
						},
					},
				},
			},
		},
	}
}
