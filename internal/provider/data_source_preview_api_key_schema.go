package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func PreviewAPIKeyDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Reads an existing Contentful Preview API Key and its access token for previewing unpublished content. The corresponding `contentful_delivery_api_key` resource exposes the `preview_api_key_id` needed for this lookup.",
		Attributes: map[string]schema.Attribute{
			"space_id": schema.StringAttribute{
				Description: "The ID of the space.",
				Required:    true,
			},
			"preview_api_key_id": schema.StringAttribute{
				Description: "The unique identifier for the preview API key.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the preview API key.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the preview API key.",
				Computed:    true,
			},
			"environments": schema.ListAttribute{
				Description: "List of environment IDs this preview API key has access to.",
				ElementType: types.StringType,
				CustomType:  NewTypedListNull[types.String]().CustomType(ctx),
				Computed:    true,
			},
			"access_token": schema.StringAttribute{
				Description: "Preview API access token. This sensitive value is stored in Terraform state.",
				Computed:    true,
				Sensitive:   true,
			},
			"timeouts": timeouts.Attributes(ctx),
		},
	}
}
