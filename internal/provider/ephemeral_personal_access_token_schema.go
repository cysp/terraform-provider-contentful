package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func PersonalAccessTokenEphemeralResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"expires_in": schema.Int64Attribute{
				Optional: true,
			},
			"expires_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"revoked_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"scopes": schema.ListAttribute{
				ElementType: types.StringType,
				CustomType:  NewTypedListNull[types.String](ctx).CustomType(ctx),
				Required:    true,
			},
			"token": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}
