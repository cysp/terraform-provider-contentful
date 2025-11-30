package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type EntryListConfigModel struct {
	SpaceID       types.String `tfsdk:"space_id"`
	EnvironmentID types.String `tfsdk:"environment_id"`
	ContentType   types.String `tfsdk:"content_type"`
	Select        types.String `tfsdk:"select"`
	Limit         types.Int64  `tfsdk:"limit"`
	Skip          types.Int64  `tfsdk:"skip"`
	Order         types.String `tfsdk:"order"`
	Query         types.String `tfsdk:"query"`
}

func EntryListResourceSchema(_ context.Context) schema.Schema {
	return schema.Schema{
		Description: "List entries from a Contentful space and environment",
		Attributes: map[string]schema.Attribute{
			"space_id": schema.StringAttribute{
				Description: "The space ID",
				Required:    true,
			},
			"environment_id": schema.StringAttribute{
				Description: "The environment ID",
				Required:    true,
			},
			"content_type": schema.StringAttribute{
				Description: "Filter by content type ID",
				Optional:    true,
			},
			"select": schema.StringAttribute{
				Description: "Select specific fields (comma-separated)",
				Optional:    true,
			},
			"limit": schema.Int64Attribute{
				Description: "Maximum number of entries to return",
				Optional:    true,
			},
			"skip": schema.Int64Attribute{
				Description: "Number of entries to skip",
				Optional:    true,
			},
			"order": schema.StringAttribute{
				Description: "Order entries by field (e.g., 'sys.createdAt' or '-sys.updatedAt')",
				Optional:    true,
			},
			"query": schema.StringAttribute{
				Description: "Full-text search query",
				Optional:    true,
			},
		},
	}
}
