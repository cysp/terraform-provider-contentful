package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func SpaceEnablementsResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Manages Contentful Space Enablements.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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
				Description: "Enable cross-space references to link content across multiple spaces. Current Contentful CMA mutations require cross_space_links and space_templates to both be present with equal boolean values; configure both on initial Create. This attribute remains independently Optional+Computed, and the provider forwards its effective planned value without inferring space_templates.",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"space_templates": schema.BoolAttribute{
				Description: "Enable the space templates feature. Current Contentful CMA mutations require space_templates and cross_space_links to both be present with equal boolean values; configure both on initial Create. This attribute remains independently Optional+Computed, and the provider forwards its effective planned value without inferring cross_space_links.",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"studio_experiences": schema.BoolAttribute{
				Description: "Enable Studio Experiences feature.",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"suggest_concepts": schema.BoolAttribute{
				Description: "Enable concept suggestions feature.",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{Create: true, Read: true, Update: true}),
		},
	}
}
