package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type WebhookModel struct {
	Active            types.Bool                 `tfsdk:"active"`
	Filters           types.List                 `tfsdk:"filters"`
	Headers           types.Map                  `tfsdk:"headers"`
	HTTPBasicPassword types.String               `tfsdk:"http_basic_password"`
	HTTPBasicUsername types.String               `tfsdk:"http_basic_username"`
	Name              types.String               `tfsdk:"name"`
	SpaceID           types.String               `tfsdk:"space_id"`
	Topics            types.List                 `tfsdk:"topics"`
	Transformation    WebhookTransformationValue `tfsdk:"transformation"`
	URL               types.String               `tfsdk:"url"`
	WebhookID         types.String               `tfsdk:"webhook_id"`
}

func WebhookResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"active": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"filters": WebhookFiltersSchema(ctx, true),
			"headers": WebhookHeadersSchema(ctx, true),
			"http_basic_password": schema.StringAttribute{
				Optional: true,
			},
			"http_basic_username": schema.StringAttribute{
				Optional: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"space_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"topics": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			"transformation": WebhookTransformationSchema(ctx, true),
			"url": schema.StringAttribute{
				Required: true,
			},
			"webhook_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}
