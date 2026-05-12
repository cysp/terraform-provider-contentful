package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func LocaleResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Manages a Contentful Locale.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				Description: "ID of the space containing the locale.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"environment_id": schema.StringAttribute{
				Description: "ID of the environment containing the locale.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"locale_id": schema.StringAttribute{
				Description: "ID of the locale.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the locale.",
				Required:    true,
			},
			"code": schema.StringAttribute{
				Description: "Locale code, for example en-US.",
				Required:    true,
			},
			"fallback_code": schema.StringAttribute{
				Description: "Code of the locale to use as a fallback when this locale is empty.",
				Optional:    true,
			},
			"content_delivery_api": schema.BoolAttribute{
				Description: "Whether the locale is available through the Content Delivery API.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"content_management_api": schema.BoolAttribute{
				Description: "Whether the locale is available through the Content Management API.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"optional": schema.BoolAttribute{
				Description: "Whether the locale can be empty for required localized fields.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"default": schema.BoolAttribute{
				Description: "Whether this is the environment's default locale.",
				Computed:    true,
			},
			"internal_code": schema.StringAttribute{
				Description: "Contentful internal locale code.",
				Computed:    true,
			},
			"timeouts": timeouts.AttributesAll(ctx),
		},
	}
}
