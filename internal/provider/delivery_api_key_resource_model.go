package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DeliveryAPIKeyResourceModel struct {
	ID              types.String            `tfsdk:"id"`
	SpaceID         types.String            `tfsdk:"space_id"`
	APIKeyID        types.String            `tfsdk:"api_key_id"`
	Name            types.String            `tfsdk:"name"`
	Description     types.String            `tfsdk:"description"`
	Environments    TypedList[types.String] `tfsdk:"environments"`
	AccessToken     types.String            `tfsdk:"access_token"`
	PreviewAPIKeyID types.String            `tfsdk:"preview_api_key_id"`
}

func DeliveryAPIKeyResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"space_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"api_key_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
			},
			"environments": schema.ListAttribute{
				ElementType: types.StringType,
				CustomType:  NewTypedListNull[types.String](ctx).CustomType(ctx),
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					UseStateForUnknown(),
				},
			},
			"access_token": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
			"preview_api_key_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
		},
	}
}
