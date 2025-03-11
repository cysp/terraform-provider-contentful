package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PreviewAPIKeyDataSourceModel struct {
	SpaceID         types.String `tfsdk:"space_id"`
	PreviewAPIKeyID types.String `tfsdk:"preview_api_key_id"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	Environments    types.List   `tfsdk:"environments"`
	AccessToken     types.String `tfsdk:"access_token"`
}

func PreviewAPIKeyDataSourceSchema(_ context.Context) schema.Schema {
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
				Computed:    true,
			},
			"access_token": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}
