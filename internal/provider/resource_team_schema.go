package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func TeamResourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Description: "Manages a Contentful Team.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				Description: "ID of the organization that owns the team.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"team_id": schema.StringAttribute{
				Description: "System ID of the team.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "A human-readable name of the team.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "A description of the team.",
				Optional:    true,
			},
		},
	}
}
