package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

func SpaceEnablementsResourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"space_id": schema.StringAttribute{
				Required: true,
			},
			"cross_space_links": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					UseStateForUnknown(),
				},
			},
			"space_templates": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					UseStateForUnknown(),
				},
			},
			"studio_experiences": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					UseStateForUnknown(),
				},
			},
			"suggest_concepts": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					UseStateForUnknown(),
				},
			},
		},
	}
}
