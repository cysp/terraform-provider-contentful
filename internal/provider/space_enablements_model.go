package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SpaceEnablementsModel struct {
	ID                types.String `tfsdk:"id"`
	SpaceID           types.String `tfsdk:"space_id"`
	CrossSpaceLinks   types.Bool   `tfsdk:"cross_space_links"`
	SpaceTemplates    types.Bool   `tfsdk:"space_templates"`
	StudioExperiences types.Bool   `tfsdk:"studio_experiences"`
	SuggestConcepts   types.Bool   `tfsdk:"suggest_concepts"`
}

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
