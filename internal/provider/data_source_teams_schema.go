package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func TeamsDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Retrieves all Contentful Teams in an organization.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"organization_id": schema.StringAttribute{
				Description: "The ID of the organization.",
				Required:    true,
			},
			"teams": schema.ListNestedAttribute{
				Description: "The teams in the organization, ordered lexicographically by team_id.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"team_id": schema.StringAttribute{
							Description: "System ID of the team.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "A human-readable name of the team.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "A description of the team.",
							Computed:    true,
						},
					},
				},
				Computed: true,
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}
