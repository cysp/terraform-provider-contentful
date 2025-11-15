package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func EnvironmentAliasResourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Description: "Manages a Contentful Environment Alias.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				Description: "ID of the space containing the environment alias.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"environment_alias_id": schema.StringAttribute{
				Description: "ID of the environment alias.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target_environment_id": schema.StringAttribute{
				Description: "ID of the environment which the environment alias references. Allows you to access and modify the data of this target environment through a different static identifier.",
				Required:    true,
			},
		},
	}
}
