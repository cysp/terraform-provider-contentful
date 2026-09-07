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
		Description: "Manages Contentful Space Enablements. Destroying this resource removes it from Terraform state without disabling or resetting the remote enablements. Import the existing Space Enablements to resume management. This retention applies while the parent space exists; this resource does not manage the space lifecycle.",
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
				Description: "Enable cross-space references to link content across multiple spaces. Contentful may reject unsupported combinations with other space enablements.",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"space_templates": schema.BoolAttribute{
				Description: "Enable the space templates feature. Contentful may reject unsupported combinations with other space enablements.",
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
