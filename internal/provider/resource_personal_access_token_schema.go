package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func PersonalAccessTokenResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Manages a Contentful Personal Access Token. Changing name, scopes, or expires_in requires a new token; Terraform revokes the old token during replacement. Changing only timeouts preserves the existing token. Destroy revokes the token.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "System ID of the personal access token. This is distinct from the secret `token` value.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the token. Changing this value replaces the resource.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"expires_in": schema.Int64Attribute{
				Description: "Time-to-live (TTL) of the token expressed in seconds. If not provided, the token will not auto-expire. Changing this value replaces the resource.",
				Optional:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"expires_at": schema.StringAttribute{
				Description: "RFC 3339 timestamp when the token expires.",
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
			},
			"revoked_at": schema.StringAttribute{
				Description: "RFC 3339 timestamp when the token was revoked.",
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
			},
			"scopes": schema.ListAttribute{
				Description: "Access granted to the token: `content_management_read` for read access or `content_management_manage` for read and write access. Changing this value replaces the resource.",
				ElementType: types.StringType,
				CustomType:  NewTypedListNull[types.String]().CustomType(ctx),
				Required:    true,
				Validators: []validator.List{
					listvalidator.NoNullValues(),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"token": schema.StringAttribute{
				Description: "The access token for the Content Management API. Contentful returns it only on creation; Terraform retains the known value in state during later refreshes. Import cannot recover the token and leaves this attribute null.",
				Computed:    true,
				Sensitive:   true,
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{Create: true, Read: true, Delete: true}),
		},
	}
}
