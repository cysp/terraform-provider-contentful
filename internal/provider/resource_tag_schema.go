package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func TagResourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Description: "Manages a Contentful Tag.",
		Attributes: map[string]schema.Attribute{
			"space_id": schema.StringAttribute{
				Description: "ID of the space containing the tag.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"environment_id": schema.StringAttribute{
				Description: "ID of the environment containing the tag.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tag_id": schema.StringAttribute{
				Description: "ID of the tag.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the tag.",
				Required:    true,
			},
			"visibility": schema.StringAttribute{
				Description: "Visibility of the tag. Can be 'public' or 'private'. Defaults to 'public'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("public"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}
