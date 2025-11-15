package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func SpaceEnablementsResourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Description: "Manages Contentful Space Enablements.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				Description: "ID of the space for which enablements are configured.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cross_space_links": schema.BoolAttribute{
				Description: "Enable cross-space references to link content across multiple spaces. Must be set together with space_templates.",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					UseStateForUnknown(),
				},
			},
			"space_templates": schema.BoolAttribute{
				Description: "Enable space templates feature. Must be set together with cross_space_links.",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					UseStateForUnknown(),
				},
			},
			"studio_experiences": schema.BoolAttribute{
				Description: "Enable Studio Experiences feature.",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					UseStateForUnknown(),
				},
			},
			"suggest_concepts": schema.BoolAttribute{
				Description: "Enable concept suggestions feature.",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					UseStateForUnknown(),
				},
			},
		},
	}
}
