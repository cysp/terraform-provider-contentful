package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func LocaleDataSourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Description: "Retrieves a Contentful Locale.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID of this data source result.",
				Computed:    true,
			},
			"space_id": schema.StringAttribute{
				Description: "ID of the space containing the locale.",
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
				Description: "Locale code (for example, en-US).",
				Computed:    true,
			},
			"fallback_code": schema.StringAttribute{
				Description: "Fallback locale code.",
				Computed:    true,
			},
			"optional": schema.BoolAttribute{
				Description: "Whether this locale is optional for entries.",
				Computed:    true,
			},
			"default": schema.BoolAttribute{
				Description: "Whether this locale is the default locale.",
				Computed:    true,
			},
			"content_delivery_api": schema.BoolAttribute{
				Description: "Whether this locale is enabled for the Content Delivery API.",
				Computed:    true,
			},
			"content_management_api": schema.BoolAttribute{
				Description: "Whether this locale is enabled for the Content Management API.",
				Computed:    true,
			},
		},
	}
}
