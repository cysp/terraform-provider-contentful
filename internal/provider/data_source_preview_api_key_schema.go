package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func PreviewAPIKeyDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"space_id": schema.StringAttribute{
				Required: true,
			},
			"preview_api_key_id": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"description": schema.StringAttribute{
				Computed: true,
			},
			"environments": schema.ListAttribute{
				ElementType: types.StringType,
				CustomType:  NewTypedListNull[types.String]().CustomType(ctx),
				Computed:    true,
			},
			"access_token": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}
