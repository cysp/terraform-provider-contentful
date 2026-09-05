package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func EnvironmentStatusReadyDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: `Waits until a Contentful environment reaches ready status.

The data source polls while Contentful reports queued or inProgress. It returns an error immediately if Contentful reports failed. Unrecognized status values remain pollable so that a newly introduced status does not fail prematurely.

The readiness wait is controlled by this data source's timeouts.read value.

This may be referenced in depends_on chains when creating resources that require an environment to be fully ready.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
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
