package provider

import (
	"context"

	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
)

func ContentTypeListResourceSchema(_ context.Context) listschema.Schema {
	return listschema.Schema{
		Description: "List Contentful Content Types.",
		Attributes: map[string]listschema.Attribute{
			"space_id": listschema.StringAttribute{
				Description: "The ID of the space for which to list content types.",
				Required:    true,
			},
			"environment_id": listschema.StringAttribute{
				Description: "The ID of the environment for which to list content types.",
				Required:    true,
			},
		},
	}
}
