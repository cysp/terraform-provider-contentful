package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func ResourceTypeResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Manages a Contentful App Resource Type.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite Terraform resource identifier in organization_id/app_definition_id/resource_type_id form.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				Description: "ID of the organization. Changing this value replaces the resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"app_definition_id": schema.StringAttribute{
				Description: "ID of the app definition. Changing this value replaces the resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"resource_provider_id": schema.StringAttribute{
				Description: "ID of the parent resource provider.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"resource_type_id": schema.StringAttribute{
				Description: "ID of the resource type. Changing this value replaces the resource.",
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
				Description: "Maps external resource data to the values displayed in the Contentful web app. Use JSON-pointer templates such as `{ /title }`.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"title": schema.StringAttribute{
						Description: "Template for the title.",
						Required:    true,
					},
					"subtitle": schema.StringAttribute{
						Description: "Template for the subtitle.",
						Optional:    true,
					},
					"description": schema.StringAttribute{
						Description: "Template for the description.",
						Optional:    true,
					},
					"external_url": schema.StringAttribute{
						Description: "Template for the external URL.",
						Optional:    true,
					},
					"image": schema.SingleNestedAttribute{
						Description: "Image field mapping.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"url": schema.StringAttribute{
								Description: "Template for the image URL.",
								Required:    true,
							},
							"alt_text": schema.StringAttribute{
								Description: "Template for the image alt text.",
								Optional:    true,
							},
						},
					},
					"badge": schema.SingleNestedAttribute{
						Description: "Badge field mapping.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"label": schema.StringAttribute{
								Description: "Template for the badge label.",
								Required:    true,
							},
							"variant": schema.StringAttribute{
								Description: "Template for the badge variant.",
								Required:    true,
							},
						},
					},
				},
			},
			"timeouts": timeouts.AttributesAll(ctx),
		},
	}
}
