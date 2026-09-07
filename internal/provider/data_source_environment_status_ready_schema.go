package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func EnvironmentStatusReadyDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: `Waits until a Contentful environment reaches ready status.

Use this data source in a depends_on relationship before creating resources that need a ready environment. It polls queued, inProgress, and unrecognized statuses, and fails immediately if Contentful reports failed. Set timeouts.read to control how long it waits.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Composite Terraform identifier in space_id/environment_id form; not a Contentful system ID.",
				Computed:    true,
			},
			"space_id": schema.StringAttribute{
				Description: "ID of the space containing the environment.",
				Required:    true,
			},
			"environment_id": schema.StringAttribute{
				Description: "ID of the environment to wait for.",
				Required:    true,
			},
			"status": schema.StringAttribute{
				Description: "Latest status reported for the environment.",
				Computed:    true,
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}
