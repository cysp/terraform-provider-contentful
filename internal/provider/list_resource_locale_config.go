package provider

import (
	"context"

	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type localeListResourceConfig struct {
	SpaceID       types.String `tfsdk:"space_id"`
	EnvironmentID types.String `tfsdk:"environment_id"`
}

func LocaleListResourceConfigSchema(_ context.Context) listschema.Schema {
	return listschema.Schema{
		Description: "Lists Contentful Locales in an existing space and environment.",
		Attributes: map[string]listschema.Attribute{
			"space_id": listschema.StringAttribute{
				Description: "ID of the space from which to list Locales.",
				Required:    true,
			},
			"environment_id": listschema.StringAttribute{
				Description: "ID of the environment from which to list Locales.",
				Required:    true,
			},
		},
	}
}
