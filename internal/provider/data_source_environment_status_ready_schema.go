package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func EnvironmentStatusReadyDataSourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Description: `Waits until a Contentful environment reaches ready status.

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
				Description: "Status of the environment.",
				Computed:    true,
			},
		},
	}
}
