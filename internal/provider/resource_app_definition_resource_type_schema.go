package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func AppDefinitionResourceTypeResourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Description: "DEPRECATED: Manages a Contentful App Resource Type. Use contentful_resource_type instead.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				Description: "The ID of the organization.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"app_definition_id": schema.StringAttribute{
				Description: "The ID of the app definition.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"resource_provider_id": schema.StringAttribute{
				Description: "The ID of the resource provider.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"resource_type_id": schema.StringAttribute{
				Description: "The unique identifier for this resource type.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the resource type.",
				Required:    true,
			},
			"default_field_mapping": schema.SingleNestedAttribute{
				Description: "Default field mappings for displaying resource instances.",
				Required:    true,
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
		DeprecationMessage: "Use contentful_resource_type instead. Existing resources may be moved from contentful_app_definition_resource_type to contentful_resource_type. contentful_app_definition_resource_type will be removed in a future version.",
	}
}
