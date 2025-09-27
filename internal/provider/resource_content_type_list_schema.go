package provider

import (
	"context"

	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func ContentTypeListResourceSchema(_ context.Context) listschema.Schema {
	return listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"space_id": schema.StringAttribute{
				Required: true,
			},
			"environment_id": schema.StringAttribute{
				Required: true,
			},
		},
	}
}
