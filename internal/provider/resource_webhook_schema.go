package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func WebhookResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Description: "Manages a Contentful Webhook.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
			"space_id": schema.StringAttribute{
				Description: "ID of the space containing the webhook.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"webhook_id": schema.StringAttribute{
				Description: "System ID of the webhook.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
			"active": schema.BoolAttribute{
				Description: "Whether the webhook is active.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"name": schema.StringAttribute{
				Description: "Name of the webhook.",
				Required:    true,
			},
			"url": schema.StringAttribute{
				Description: "Preconfigured HTTP endpoint that is called when content has changed.",
				Required:    true,
			},
			"topics": schema.ListAttribute{
				Description: "List of event topics to which the webhook subscribes.",
				ElementType: types.StringType,
				CustomType:  TypedList[types.String]{}.CustomType(ctx),
				Optional:    true,
			},
			"filters": WebhookFiltersSchema(ctx, true),
			"http_basic_password": schema.StringAttribute{
				Description: "HTTP Basic authentication password.",
				Optional:    true,
			},
			"http_basic_username": schema.StringAttribute{
				Description: "HTTP Basic authentication username.",
				Optional:    true,
			},
			"headers":        WebhookHeadersSchema(ctx, true),
			"transformation": WebhookTransformationSchema(ctx, true),
		},
	}
}
