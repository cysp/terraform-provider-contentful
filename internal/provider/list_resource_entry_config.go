package provider

import (
	"context"

	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type entryListResourceConfig struct {
	SpaceID       types.String            `tfsdk:"space_id"`
	EnvironmentID types.String            `tfsdk:"environment_id"`
	ContentType   types.String            `tfsdk:"content_type"`
	Order         TypedList[types.String] `tfsdk:"order"`
	Query         TypedMap[types.String]  `tfsdk:"query"`
}

func EntryListResourceConfigSchema(ctx context.Context) listschema.Schema {
	return listschema.Schema{
		Description: "List entries from a Contentful space and environment",
		Attributes: map[string]listschema.Attribute{
			"space_id": listschema.StringAttribute{
				Description: "The ID of the space for which to list entries.",
				Required:    true,
			},
			"environment_id": listschema.StringAttribute{
				Description: "The ID of the environment for which to list entries.",
				Required:    true,
			},
			"content_type": listschema.StringAttribute{
				Description: "Query entries for a specific content type.",
				Optional:    true,
			},
			"order": listschema.ListAttribute{
				Description: "Order entries by one or more attributes.",
				ElementType: types.StringNull().Type(ctx),
				CustomType:  NewTypedListNull[types.String]().CustomType(ctx),
				Optional:    true,
			},
			"query": listschema.MapAttribute{
				Description: "Query parameters to filter the entries listed.",
				ElementType: types.StringNull().Type(ctx),
				CustomType:  NewTypedMapNull[types.String]().CustomType(ctx),
				Optional:    true,
			},
		},
	}
}
