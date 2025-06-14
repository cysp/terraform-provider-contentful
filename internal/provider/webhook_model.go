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
	ID                types.String                  `tfsdk:"id"`
	SpaceID           types.String                  `tfsdk:"space_id"`
	WebhookID         types.String                  `tfsdk:"webhook_id"`
	Name              types.String                  `tfsdk:"name"`
	URL               types.String                  `tfsdk:"url"`
	Topics            TypedList[types.String]       `tfsdk:"topics"`
	Filters           TypedList[WebhookFilterValue] `tfsdk:"filters"`
	HTTPBasicPassword types.String                  `tfsdk:"http_basic_password"`
	HTTPBasicUsername types.String                  `tfsdk:"http_basic_username"`
	Headers           TypedMap[WebhookHeaderValue]  `tfsdk:"headers"`
	Transformation    WebhookTransformationValue    `tfsdk:"transformation"`
	Active            types.Bool                    `tfsdk:"active"`
}

func WebhookResourceSchema(ctx context.Context) schema.Schema {
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
			"webhook_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					UseStateForUnknown(),
				},
			},
			"active": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"url": schema.StringAttribute{
				Required: true,
			},
			"topics": schema.ListAttribute{
				ElementType: types.StringType,
				CustomType:  TypedList[types.String]{}.CustomType(ctx),
				Optional:    true,
			},
			"filters": WebhookFiltersSchema(ctx, true),
			"http_basic_password": schema.StringAttribute{
				Optional: true,
			},
			"http_basic_username": schema.StringAttribute{
				Optional: true,
			},
			"headers":        WebhookHeadersSchema(ctx, true),
			"transformation": WebhookTransformationSchema(ctx, true),
		},
	}
}
