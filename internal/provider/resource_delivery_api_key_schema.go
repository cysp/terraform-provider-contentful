package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func DeliveryAPIKeyResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Manages a Contentful Delivery API Key.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				Description: "ID of the space for the API key.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"api_key_id": schema.StringAttribute{
				Description: "System ID of the API key.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Human-readable name for the API key.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the API key.",
				Optional:    true,
			},
			"environments": schema.ListAttribute{
				Description: "List of environment IDs that the token can access. Only the environments specified in this property can be accessed using this token.",
				ElementType: types.StringType,
				CustomType:  NewTypedListNull[types.String]().CustomType(ctx),
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					UseStateForUnknown(),
				},
			},
			"access_token": schema.StringAttribute{
				Description: "The delivery API access token.",
				Computed:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
			"preview_api_key_id": schema.StringAttribute{
				Description: "ID of the corresponding preview API key.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
		},
	}
}
