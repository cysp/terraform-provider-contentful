package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func LocaleDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Retrieves a Contentful Locale.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"space_id": schema.StringAttribute{
				Description: "ID of the space containing the locale.",
				Required:    true,
			},
			"environment_id": schema.StringAttribute{
				Description: "ID of the environment containing the locale.",
				Required:    true,
			},
			"locale_id": schema.StringAttribute{
				Description: "ID of the locale.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the locale.",
				Computed:    true,
			},
			"code": schema.StringAttribute{
				Description: "Locale code, for example en-US.",
				Computed:    true,
			},
			"fallback_code": schema.StringAttribute{
				Description: "Code of the locale to use as a fallback when this locale is empty.",
				Computed:    true,
			},
			"content_delivery_api": schema.BoolAttribute{
				Description: "Whether the locale is available through the Content Delivery API.",
				Computed:    true,
			},
			"content_management_api": schema.BoolAttribute{
				Description: "Whether the locale is available through the Content Management API.",
				Computed:    true,
			},
			"optional": schema.BoolAttribute{
				Description: "Whether the locale can be empty for required localized fields.",
				Computed:    true,
			},
			"default": schema.BoolAttribute{
				Description: "Whether this is the environment's default locale.",
				Computed:    true,
			},
			"internal_code": schema.StringAttribute{
				Description: "Contentful internal locale code.",
				Computed:    true,
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}
