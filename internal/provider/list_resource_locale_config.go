package provider

import (
	"context"

	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type localeListResourceConfig struct {
	SpaceID       types.String            `tfsdk:"space_id"`
	EnvironmentID types.String            `tfsdk:"environment_id"`
	Order         TypedList[types.String] `tfsdk:"order"`
}

func LocaleListResourceConfigSchema(ctx context.Context) listschema.Schema {
	return listschema.Schema{
		Description: "List Contentful Locales.",
		Attributes: map[string]listschema.Attribute{
			"space_id": listschema.StringAttribute{
				Description: "The ID of the space for which to list locales.",
				Required:    true,
			},
			"environment_id": listschema.StringAttribute{
				Description: "The ID of the environment for which to list locales.",
				Required:    true,
			},
			"order": listschema.ListAttribute{
				Description: "Order locales by one or more attributes.",
				ElementType: types.StringNull().Type(ctx),
				CustomType:  NewTypedListNull[types.String]().CustomType(ctx),
				Optional:    true,
			},
		},
	}
}
