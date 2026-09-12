package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func EnvironmentAliasResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Manages a Contentful Environment Alias, which routes requests from a stable alias ID to a selected environment. The target environment must already exist. Use `contentful_environment_status_ready` to wait for a newly copied environment before directing requests to it.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite Terraform resource identifier in space_id/environment_alias_id form.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				Description: "ID of the space containing the environment alias. Changing this value replaces the resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"environment_alias_id": schema.StringAttribute{
				Description: "ID of the environment alias. Changing this value replaces the resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target_environment_id": schema.StringAttribute{
				Description: "ID of the environment reached through this alias. Changing the target redirects subsequent requests that use the alias.",
				Required:    true,
			},
			"timeouts": timeouts.AttributesAll(ctx),
		},
	}
}
