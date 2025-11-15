package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func PersonalAccessTokenResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Manages a Contentful Personal Access Token.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the token.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"expires_in": schema.Int64Attribute{
				Description: "Time-to-live (TTL) of the token expressed in seconds. If not provided, the token will not auto-expire.",
				Optional:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"expires_at": schema.StringAttribute{
				Description: "Timestamp when the token expires.",
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
			},
			"revoked_at": schema.StringAttribute{
				Description: "Timestamp when the token was revoked.",
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
			},
			"scopes": schema.ListAttribute{
				Description: "Scopes used to limit a token's access. Supported scopes are 'content_management_read' (Read-only access) and 'content_management_manage' (Read and write access).",
				ElementType: types.StringType,
				CustomType:  NewTypedListNull[types.String]().CustomType(ctx),
				Required:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"token": schema.StringAttribute{
				Description: "The access token for the Content Management API. This is only available immediately after creation.",
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}
