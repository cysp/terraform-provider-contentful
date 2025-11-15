package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func AppDefinitionResourceProviderResourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Description: "DEPRECATED: Manages a Contentful App Resource Provider. Use contentful_resource_provider instead.",
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
			"function_id": schema.StringAttribute{
				Description: "The ID of the function handling resource operations.",
				Required:    true,
			},
		},
		DeprecationMessage: "Use contentful_resource_provider instead. Existing resources may be moved from contentful_app_definition_resource_provider to contentful_resource_provider. contentful_app_definition_resource_provider will be removed in a future version.",
	}
}
